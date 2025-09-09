package redis

import (
	"context"
	"fmt"
	"real-time-chat/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	fmt.Println("Redis connection successful.")
	return rdb, nil
}
