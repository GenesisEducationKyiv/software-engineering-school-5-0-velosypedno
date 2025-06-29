package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBackend[T any] struct {
	client *redis.Client
}

func NewRedisBackend[T any](client *redis.Client) *RedisBackend[T] {
	return &RedisBackend[T]{
		client: client,
	}
}

func (r *RedisBackend[T]) SetStruct(ctx context.Context, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisBackend[T]) GetStruct(ctx context.Context, key string, value *T) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
