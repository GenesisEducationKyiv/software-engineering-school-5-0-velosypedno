package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBackend struct {
	client *redis.Client
}

func NewRedisBackend(client *redis.Client) *RedisBackend {
	return &RedisBackend{
		client: client,
	}
}

func (r *RedisBackend) SetStruct(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisBackend) GetStruct(ctx context.Context, key string, value any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
