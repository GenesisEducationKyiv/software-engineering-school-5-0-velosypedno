package decorator

import (
	"context"
	"log"
	"time"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type cacheClient interface {
	Get(ctx context.Context, key string, value *domain.Weather) error
	Set(ctx context.Context, key string, value domain.Weather) error
}

type weathMetrics interface {
	CacheHit()
	CacheMiss()
	CacheAccessLatency(duration float64)
}
type CacheDecorator struct {
	inner        weatherRepo
	cacheClient  cacheClient
	weathMetrics weathMetrics
}

func NewCacheDecorator(inner weatherRepo, cacheBack cacheClient, weathMetrics weathMetrics) *CacheDecorator {
	return &CacheDecorator{inner: inner, cacheClient: cacheBack, weathMetrics: weathMetrics}
}

func (d *CacheDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var weather domain.Weather

	now := time.Now()
	if err := d.cacheClient.Get(ctx, city, &weather); err == nil {
		d.weathMetrics.CacheHit()
		d.weathMetrics.CacheAccessLatency(time.Since(now).Seconds())
		log.Println("cache hit")

		return weather, nil
	}

	d.weathMetrics.CacheMiss()
	d.weathMetrics.CacheAccessLatency(time.Since(now).Seconds())
	log.Println("cache miss")

	weather, err := d.inner.GetCurrent(ctx, city)
	if err != nil {
		return weather, err
	}
	err = d.cacheClient.Set(ctx, city, weather)
	if err != nil {
		log.Println("cache error:", err)
	} else {
		log.Println("cache set")
	}
	return weather, nil
}
