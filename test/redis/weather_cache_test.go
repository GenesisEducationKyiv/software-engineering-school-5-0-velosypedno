//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/velosypedno/genesis-weather-api/internal/cache"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/repos/weather/decorator"
)

type weathMetrics struct{}

func (m *weathMetrics) CacheHit()                           {}
func (m *weathMetrics) CacheMiss()                          {}
func (m *weathMetrics) CacheAccessLatency(duration float64) {}

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
	cacheBack *cache.RedisBackend[domain.Weather]
}

func TestCacheWeatherDecorator(main *testing.T) {
	cfg, err := config.Load()
	require.NoError(main, err)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Pass,
	})
	cacheBackend := cache.NewRedisBackend[domain.Weather](redisClient)
	temp := 20.0
	humidity := 50.0
	mockWeather := domain.Weather{Temperature: temp, Humidity: humidity, Description: "Sunny"}
	repo := NewWeatherRepo(mockWeather)

	setup := func() *mocks {
		err = redisClient.FlushDB(context.Background()).Err()
		require.NoError(main, err)
		repo.Clear()
		return &mocks{repo: repo, weather: mockWeather, cacheBack: cacheBackend}
	}

	main.Run("CacheMiss", func(main *testing.T) {
		// Arrange
		mocks := setup()
		ttl := time.Duration(0)
		require.False(main, mocks.repo.called)
		decoratedRepo := decorator.NewCacheDecorator(mocks.repo, ttl, mocks.cacheBack, &weathMetrics{})
		city := "Kyiv"

		// Acr
		weather, err := decoratedRepo.GetCurrent(context.Background(), city)

		// Assert
		require.NoError(main, err)
		require.True(main, repo.called)
		require.Equal(main, mocks.weather, weather)
	})

	main.Run("CacheHit", func(main *testing.T) {
		// Arrange
		mocks := setup()
		ttl := time.Duration(0)
		require.False(main, mocks.repo.called)
		city := "Kyiv"
		mocks.cacheBack.SetStruct(context.Background(), city, mocks.weather, ttl)
		decoratedRepo := decorator.NewCacheDecorator(mocks.repo, ttl, mocks.cacheBack, &weathMetrics{})

		// Act
		weather, err := decoratedRepo.GetCurrent(context.Background(), "Kyiv")

		// Assert
		require.NoError(main, err)
		require.False(main, repo.called)
		require.Equal(main, mockWeather, weather)
	})

	main.Run("CacheExpired", func(main *testing.T) {
		// Arrange
		mocks := setup()
		ttl := 1 * time.Millisecond
		require.False(main, mocks.repo.called)
		city := "Kyiv"
		mocks.cacheBack.SetStruct(context.Background(), city, mocks.weather, ttl)
		decoratedRepo := decorator.NewCacheDecorator(mocks.repo, ttl, mocks.cacheBack, &weathMetrics{})

		// Act
		<-time.After(ttl * 2)
		weather, err := decoratedRepo.GetCurrent(context.Background(), "Kyiv")

		// Assert
		require.NoError(main, err)
		require.True(main, repo.called)
		require.Equal(main, mockWeather, weather)
	})
}
