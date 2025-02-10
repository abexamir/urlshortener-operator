// pkg/http/server.go
package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

type RedirectServer struct {
	redis *redis.Client
}

func NewRedirectServer(redis *redis.Client) *RedirectServer {
	return &RedirectServer{redis: redis}
}

func (s *RedirectServer) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	shortPath := r.URL.Path
	log.Printf("Looking up short path: %s", shortPath)

	targetURL, err := s.redis.Get(context.Background(), shortPath).Result()
	if err == redis.Nil {
		log.Printf("Short path not found: %s", shortPath)
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("Error retrieving target URL from Redis: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Found target URL: %s -> %s", shortPath, targetURL)

	// Increment click count
	clickKey := fmt.Sprintf("clicks:%s", shortPath)
	log.Printf("Incrementing click count for key: %s", clickKey)
	err = s.redis.Incr(context.Background(), clickKey).Err()
	if err != nil {
		log.Printf("Error incrementing click count: %v", err)
	}

	log.Printf("Redirecting to target URL: %s", targetURL)
	http.Redirect(w, r, targetURL, http.StatusFound)
}

func (s *RedirectServer) Start() error {
	log.Println("Starting RedirectServer on :8082")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling request for host: %s", r.Host)
		// if !strings.HasPrefix(r.Host, "ourtinyurl.local") {
		// 	log.Printf("Invalid host: %s", r.Host)
		// 	http.Error(w, "Invalid host", http.StatusBadRequest)
		// 	return
		// }
		s.HandleRedirect(w, r)
	})

	return http.ListenAndServe(":8082", nil)
}
