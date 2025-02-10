package redisHandler

import (
	"context"
	"fmt"

	"github.com/abexamir/url-shortener-operator/internal/constants"
	"github.com/go-redis/redis/v8"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(addr string) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisService{client: client}, nil
}

func (s *RedisService) GetURL(ctx context.Context, shortPath string) (string, error) {
	return s.client.Get(ctx, shortPath).Result()
}

func (s *RedisService) SetURL(ctx context.Context, shortPath, targetURL string) error {
	return s.client.Set(ctx, shortPath, targetURL, -1).Err()
}

func (s *RedisService) DeleteURL(ctx context.Context, shortPath string) error {
	return s.client.Del(ctx, shortPath).Err()
}

func (s *RedisService) GetClickCount(ctx context.Context, shortPath string) (int64, error) {
	clickKey := fmt.Sprintf("%s%s", constants.ClickCountKeyPrefix, shortPath)
	return s.client.Get(ctx, clickKey).Int64()
}

func (s *RedisService) IncrementClickCount(ctx context.Context, shortPath string) error {
	clickKey := fmt.Sprintf("%s%s", constants.ClickCountKeyPrefix, shortPath)
	return s.client.Incr(ctx, clickKey).Err()
}
