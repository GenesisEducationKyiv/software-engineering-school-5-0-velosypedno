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

type weathMetrics interface {
	CacheHit()
	CacheMiss()
	CacheAccessLatency(duration float64)
}
type CacheDecorator struct {
	Inner        weatherRepo
	TimeToLive   time.Duration
	CacheBack    cacheBackend
	weathMetrics weathMetrics
}

func NewCacheDecorator(inner weatherRepo, ttl time.Duration,
	cacheBack cacheBackend, weathMetrics weathMetrics) *CacheDecorator {
	return &CacheDecorator{Inner: inner, TimeToLive: ttl, CacheBack: cacheBack, weathMetrics: weathMetrics}
}

func (d *CacheDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var weather domain.Weather

	now := time.Now()
	if err := d.CacheBack.GetStruct(ctx, city, &weather); err == nil {
		d.weathMetrics.CacheHit()
		d.weathMetrics.CacheAccessLatency(time.Since(now).Seconds())
		log.Println("cache hit")

		return weather, nil
	}

	d.weathMetrics.CacheMiss()
	d.weathMetrics.CacheAccessLatency(time.Since(now).Seconds())
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
