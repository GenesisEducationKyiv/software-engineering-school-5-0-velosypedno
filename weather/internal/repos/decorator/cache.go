package decorator

import (
	"context"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"go.uber.org/zap"
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
	logger       *zap.Logger
	inner        weatherRepo
	cacheClient  cacheClient
	weathMetrics weathMetrics
}

func NewCacheDecorator(logger *zap.Logger, inner weatherRepo, cacheBack cacheClient, weathMetrics weathMetrics) *CacheDecorator {
	return &CacheDecorator{
		logger:       logger.With(zap.String("decorator", "CacheDecorator")),
		inner:        inner,
		cacheClient:  cacheBack,
		weathMetrics: weathMetrics,
	}
}

func (d *CacheDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var weather domain.Weather

	now := time.Now()
	if err := d.cacheClient.Get(ctx, city, &weather); err == nil {
		d.weathMetrics.CacheHit()
		d.weathMetrics.CacheAccessLatency(time.Since(now).Seconds())
		d.logger.Debug("Cache hit", zap.String("city", city))
		return weather, nil
	}

	d.weathMetrics.CacheMiss()
	d.weathMetrics.CacheAccessLatency(time.Since(now).Seconds())
	d.logger.Debug("Cache miss", zap.String("city", city))

	weather, err := d.inner.GetCurrent(ctx, city)
	if err != nil {
		return weather, err
	}
	err = d.cacheClient.Set(ctx, city, weather)
	if err != nil {
		d.logger.Error("Cache error", zap.String("city", city), zap.Error(err))
	} else {
		d.logger.Debug("Cache set", zap.String("city", city))
	}
	return weather, nil
}
