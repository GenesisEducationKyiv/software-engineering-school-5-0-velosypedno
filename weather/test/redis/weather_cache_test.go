//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cache"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/decorator"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type weathMetrics struct {
	CacheHitCalled           bool
	CacheMissCalled          bool
	CacheAccessLatencyCalled bool
}

func (m *weathMetrics) CacheHit()                           { m.CacheHitCalled = true }
func (m *weathMetrics) CacheMiss()                          { m.CacheMissCalled = true }
func (m *weathMetrics) CacheAccessLatency(duration float64) { m.CacheAccessLatencyCalled = true }

type weatherRepo struct {
	called  bool
	weather domain.Weather
}

func NewWeatherRepo(weather domain.Weather) *weatherRepo {
	return &weatherRepo{weather: weather, called: false}
}

func (r *weatherRepo) Clear() {
	r.called = false
}

func (r *weatherRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	r.called = true
	return r.weather, nil
}

type mocks struct {
	repo      *weatherRepo
	weather   domain.Weather
	cacheBack *cache.RedisCacheClient[domain.Weather]
	metrics   *weathMetrics
}

func TestCacheWeatherDecorator(main *testing.T) {
	cfg, err := config.Load()
	require.NoError(main, err)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Pass,
	})
	cacheBackend := cache.NewRedisCacheClient[domain.Weather](redisClient, time.Duration(0))
	temp := 20.0
	humidity := 50.0
	mockWeather := domain.Weather{Temperature: temp, Humidity: humidity, Description: "Sunny"}
	repo := NewWeatherRepo(mockWeather)

	setup := func() *mocks {
		err = redisClient.FlushDB(context.Background()).Err()
		require.NoError(main, err)
		repo.Clear()
		return &mocks{repo: repo, weather: mockWeather, cacheBack: cacheBackend, metrics: &weathMetrics{}}
	}

	main.Run("CacheMiss", func(t *testing.T) {
		// Arrange
		mocks := setup()
		require.False(t, mocks.repo.called)
		decoratedRepo := decorator.NewCacheDecorator(zap.NewNop(), mocks.repo, mocks.cacheBack, mocks.metrics)
		city := "Kyiv"

		// Acr
		weather, err := decoratedRepo.GetCurrent(context.Background(), city)

		// Assert
		assert.True(t, mocks.metrics.CacheMissCalled, "Cache miss should be called")
		assert.True(t, mocks.metrics.CacheAccessLatencyCalled, "Cache access latency should be called")
		require.NoError(t, err, "Failed to get weather: %v", err)
		require.True(t, repo.called, "Repo GetCurrent method should be called")
		assert.Equal(t, mocks.weather, weather, "Expected weather %v, got %v", mocks.weather, weather)
	})

	main.Run("CacheHit", func(t *testing.T) {
		// Arrange
		mocks := setup()
		require.False(t, mocks.repo.called)
		city := "Kyiv"
		mocks.cacheBack.Set(context.Background(), city, mocks.weather)
		decoratedRepo := decorator.NewCacheDecorator(zap.NewNop(), mocks.repo, mocks.cacheBack, mocks.metrics)

		// Act
		weather, err := decoratedRepo.GetCurrent(context.Background(), "Kyiv")

		// Assert
		assert.False(t, mocks.metrics.CacheMissCalled, "Cache miss should not be called")
		assert.True(t, mocks.metrics.CacheHitCalled, "Cache hit should be called")
		assert.True(t, mocks.metrics.CacheAccessLatencyCalled, "Cache access latency should be called")
		require.NoError(t, err, "Failed to get weather: %v", err)
		require.False(t, repo.called, "Repo GetCurrent method should not be called")
		assert.Equal(t, mocks.weather, weather, "Expected weather %v, got %v", mocks.weather, weather)
	})

	main.Run("CacheExpired", func(t *testing.T) {
		// Arrange
		mocks := setup()
		ttl := 1 * time.Millisecond
		cacheBackend := cache.NewRedisCacheClient[domain.Weather](redisClient, ttl)
		require.False(t, mocks.repo.called)
		city := "Kyiv"
		cacheBackend.Set(context.Background(), city, mocks.weather)
		decoratedRepo := decorator.NewCacheDecorator(zap.NewNop(), mocks.repo, cacheBackend, mocks.metrics)

		// Act
		<-time.After(ttl * 2)
		weather, err := decoratedRepo.GetCurrent(context.Background(), "Kyiv")

		// Assert
		assert.True(t, mocks.metrics.CacheMissCalled, "Cache miss should be called")
		assert.True(t, mocks.metrics.CacheAccessLatencyCalled, "Cache access latency should be called")
		require.NoError(t, err, "Failed to get weather: %v", err)
		require.True(t, repo.called, "Repo GetCurrent method should be called")
		require.Equal(t, mockWeather, weather, "Expected weather %v, got %v", mockWeather, weather)
	})
}
