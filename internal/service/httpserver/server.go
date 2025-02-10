package httpserver

import (
	"context"
	"net/http"

	redisHandler "github.com/abexamir/url-shortener-operator/internal/service/redis"
	"github.com/go-redis/redis/v8"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type RedirectServer struct {
	redisService *redisHandler.RedisService
}

func NewRedirectServer(redisService *redisHandler.RedisService) *RedirectServer {
	return &RedirectServer{
		redisService: redisService,
	}
}

func (s *RedirectServer) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log := ctrllog.FromContext(ctx, "component", "redirect-server")

	shortPath := r.URL.Path
	log.Info("Handling redirect request",
		"path", shortPath,
		"method", r.Method,
		"remoteAddr", r.RemoteAddr)

	targetURL, err := s.redisService.GetURL(ctx, shortPath)
	if err == redis.Nil {
		log.Info("Short path not found", "path", shortPath)
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Error(err, "Failed to retrieve target URL")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := s.redisService.IncrementClickCount(ctx, shortPath); err != nil {
		log.Error(err, "Failed to increment click count",
			"shortPath", shortPath)
		// Continue with redirect despite click count error
	}

	log.Info("Redirecting",
		"shortPath", shortPath,
		"targetURL", targetURL)
	http.Redirect(w, r, targetURL, http.StatusFound)
}

func (s *RedirectServer) Start() error {
	log := ctrllog.Log.WithName("redirect-server")
	log.Info("Starting HTTP server on :8082")

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HandleRedirect)

	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	return server.ListenAndServe()
}
