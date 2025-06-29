package decorator

import (
	"context"
	"log"
	"time"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type cacheBackend interface {
	GetStruct(ctx context.Context, key string, value *domain.Weather) error
	SetStruct(ctx context.Context, key string, value domain.Weather, ttl time.Duration) error
}

type CacheDecorator struct {
	Inner      weatherRepo
	TimeToLive time.Duration
	CacheBack  cacheBackend
}

func NewCacheDecorator(inner weatherRepo, ttl time.Duration, cacheBack cacheBackend) *CacheDecorator {
	return &CacheDecorator{Inner: inner, TimeToLive: ttl, CacheBack: cacheBack}
}

func (d *CacheDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var weather domain.Weather
	if err := d.CacheBack.GetStruct(ctx, city, &weather); err == nil {
		log.Println("cache hit")
		return weather, nil
	}
	log.Println("cache miss")
	weather, err := d.Inner.GetCurrent(ctx, city)
	if err != nil {
		return weather, err
	}
	err = d.CacheBack.SetStruct(ctx, city, weather, d.TimeToLive)
	if err != nil {
		log.Println("cache error:", err)
	} else {
		log.Println("cache set")
	}
	return weather, nil
}
