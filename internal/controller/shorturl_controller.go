// controllers/shorturl_controller.go
package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/go-redis/redis/v8"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	urlshortenerv1 "github.com/abexamir/url-shortener-operator/api/v1"
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

	shortURL := &urlshortenerv1.ShortURL{}
	if err := r.Get(ctx, req.NamespacedName, shortURL); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate short path if not exists
	if shortURL.Status.ShortPath == "" {
		shortPath, err := r.generateShortPath(shortURL.Spec.TargetURL)
		if err != nil {
			log.Error(err, "Failed to generate short path")
			return ctrl.Result{}, err
		}

		// Store in Redis
		err = r.Redis.Set(ctx, shortPath, shortURL.Spec.TargetURL, 0).Err()
		if err != nil {
			log.Error(err, "Failed to store in Redis")
			return ctrl.Result{}, err
		}

		// Update status
		shortURL.Status.ShortPath = shortPath
		shortURL.Status.ClickCount = 0
		if err := r.Status().Update(ctx, shortURL); err != nil {
			log.Error(err, "Failed to update ShortURL status")
			return ctrl.Result{}, err
		}
	}

	// Update click count
	clickCount, err := r.Redis.Get(ctx, fmt.Sprintf("clicks:%s", shortURL.Status.ShortPath)).Int64()
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

	return ctrl.Result{}, nil
}

func (r *ShortURLReconciler) generateShortPath(url string) (string, error) {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return "/" + encoded[:3], nil
}

func (r *ShortURLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize Redis client
	r.Redis = redis.NewClient(&redis.Options{
		Addr: "redis-service:6379", // Update this to match your Redis service name
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&urlshortenerv1.ShortURL{}).
		Complete(r)
}
