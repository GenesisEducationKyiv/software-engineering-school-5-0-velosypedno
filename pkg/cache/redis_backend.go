package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheClient[T any] struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCacheClient[T any](client *redis.Client, ttl time.Duration) *RedisCacheClient[T] {
	return &RedisCacheClient[T]{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisCacheClient[T]) Set(ctx context.Context, key string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *RedisCacheClient[T]) Get(ctx context.Context, key string, value *T) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
