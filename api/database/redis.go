package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClients struct {
	Cache     *redis.Client
	RateLimit *redis.Client
}

func NewRedisClients(addr, password string) (*RedisClients, error) {
	cache := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	rateLimit := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       1,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cache.Ping(ctx).Err(); err != nil {
		_ = cache.Close()
		_ = rateLimit.Close()
		return nil, fmt.Errorf("unable to connect to redis (cache): %w", err)
	}
	return &RedisClients{
		Cache:     cache,
		RateLimit: rateLimit,
	}, nil
}
func (r *RedisClients) Close() {
	if r.Cache != nil {
		_ = r.Cache.Close()
	}
	if r.RateLimit != nil {
		_ = r.RateLimit.Close()
	}
}
