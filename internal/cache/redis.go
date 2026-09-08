package cache

// This file represents the Redis connection/client

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) *Redis {
	client := redis.NewClient(&redis.Options{Addr: addr})

	return &Redis{client: client}
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.client.Close()
}
