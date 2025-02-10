// controllers/shorturl_controller.go
package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"os"

	"github.com/go-redis/redis/v8"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	urlshortenerv1 "github.com/abexamir/url-shortener-operator/api/v1"

	httpserver "github.com/abexamir/url-shortener-operator/pkg/httpserver"
)

// ShortURLReconciler reconciles a ShortURL object
type ShortURLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Redis  *redis.Client
}

//+kubebuilder:rbac:groups=urlshortener.tapsi.ir,resources=shorturls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=urlshortener.tapsi.ir,resources=shorturls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=urlshortener.tapsi.ir,resources=shorturls/finalizers,verbs=update

func (r *ShortURLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Starting Reconcile for ShortURL", "NamespacedName", req.NamespacedName)

	shortURL := &urlshortenerv1.ShortURL{}
	if err := r.Get(ctx, req.NamespacedName, shortURL); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Handle deletion
			log.Info("ShortURL resource not found, cleaning up Redis entries", "NamespacedName", req.NamespacedName)
			if shortURL.Status.ShortPath != "" {
				if err := r.Redis.Del(ctx, shortURL.Status.ShortPath).Err(); err != nil {
					log.Error(err, "Failed to delete Redis entry")
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch ShortURL resource")
		return ctrl.Result{}, err
	}

	if !r.isValidURL(shortURL.Spec.TargetURL) {
		log.Error(nil, "Invalid target URL", "TargetURL", shortURL.Spec.TargetURL)
		return ctrl.Result{}, fmt.Errorf("invalid target URL")
	}

	// Handle new resources or updates
	needsNewShortPath := false

	// For new resources
	if shortURL.Status.ShortPath == "" {
		needsNewShortPath = true
	} else {
		// For existing resources, check if target URL changed
		existingURL, err := r.Redis.Get(ctx, shortURL.Status.ShortPath).Result()
		if err != nil && err != redis.Nil {
			log.Error(err, "Failed to get existing URL from Redis")
			return ctrl.Result{}, err
		}
		// If Redis entry doesn't exist or URL has changed, we need a new short path
		if err == redis.Nil || existingURL != shortURL.Spec.TargetURL {
			needsNewShortPath = true
			// Clean up old path if it exists
			if shortURL.Status.ShortPath != "" {
				if err := r.Redis.Del(ctx, shortURL.Status.ShortPath).Err(); err != nil {
					log.Error(err, "Failed to delete old Redis entry")
					return ctrl.Result{}, err
				}
			}
		}
	}

	if needsNewShortPath {
		shortPath, err := r.generateShortPath(shortURL.Spec.TargetURL)
		if err != nil {
			log.Error(err, "Failed to generate short path")
			return ctrl.Result{}, err
		}

		if err := r.Redis.Set(ctx, shortPath, shortURL.Spec.TargetURL, 0).Err(); err != nil {
			log.Error(err, "Failed to set Redis entry")
			return ctrl.Result{}, err
		}

		shortURL.Status.ShortPath = shortPath
		if err := r.Status().Update(ctx, shortURL); err != nil {
			log.Error(err, "Failed to update ShortURL status")
			return ctrl.Result{}, err
		}
	}

	// Update click count
	clickKey := fmt.Sprintf("clicks:%s", shortURL.Status.ShortPath)
	clickCount, err := r.Redis.Get(ctx, clickKey).Int64()
	if err != nil && err != redis.Nil {
		log.Error(err, "Failed to get click count")
		return ctrl.Result{}, err
	}

	if clickCount != shortURL.Status.ClickCount {
		shortURL.Status.ClickCount = clickCount
		if err := r.Status().Update(ctx, shortURL); err != nil {
			log.Error(err, "Failed to update click count")
			return ctrl.Result{}, err
		}
	}

	// Requeue periodically to update click count
	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *ShortURLReconciler) generateShortPath(url string) (string, error) {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return "/" + encoded[:4], nil
}

func (r *ShortURLReconciler) isValidURL(s string) bool {
	parsed, err := url.ParseRequestURI(s)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return false
	}
	return true
}

func (r *ShortURLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize Redis client
	log := log.Log.WithName("setup")
	log.Info("Initializing Redis client")
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis-service:6379",
	})

	// Wait for Redis to be ready
	for {
		_, err := redisClient.Ping(context.Background()).Result()
		if err == nil {
			break
		}
		log.Info("Waiting for Redis to be ready...")
		time.Sleep(1 * time.Second)
	}

	// Start HTTP server
	redirectServer := httpserver.NewRedirectServer(redisClient)
	go func() {
		if err := redirectServer.Start(); err != nil {
			log.Error(err, "unable to start HTTP server")
			os.Exit(1)
		}
	}()

	r.Redis = redisClient

	return ctrl.NewControllerManagedBy(mgr).
		For(&urlshortenerv1.ShortURL{}).
		Complete(r)
}
