//go:build unit

package decorator_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cb"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/decorator"
)

func newTestBreaker() *cb.CircuitBreaker {
	breaker := cb.NewCircuitBreaker(time.Minute, 1, 1)
	breaker.Now = func() time.Time {
		return time.Now()
	}
	return breaker
}

func TestBreakerDecorator_Success(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{Temperature: 20.0},
		Err:      nil,
	}
	breaker := newTestBreaker()
	repo := decorator.NewBreakerDecorator(mock, breaker)

	// Act
	result, err := repo.GetCurrent(context.Background(), "Lviv")

	// Assert
	require.NoError(t, err)
	require.True(t, mock.Called)
	assert.Equal(t, 20.0, result.Temperature)
	assert.True(t, breaker.Allowed())
}

func TestBreakerDecorator_UnavailableErrorTriggersBreaker(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrWeatherUnavailable,
	}
	breaker := newTestBreaker()
	repo := decorator.NewBreakerDecorator(mock, breaker)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Odessa")

	// Assert
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)
	assert.False(t, breaker.Allowed())
}

func TestBreakerDecorator_OtherErrorDoesNotTriggerBreaker(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrCityNotFound,
	}
	breaker := newTestBreaker()
	repo := decorator.NewBreakerDecorator(mock, breaker)

	// Act
	_, err := repo.GetCurrent(context.Background(), "!!!")

	// Assert
	require.ErrorIs(t, err, domain.ErrCityNotFound)
	assert.True(t, breaker.Allowed())
}

func TestBreakerDecorator_BreakerOpenSkipsCall(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      nil,
	}
	breaker := newTestBreaker()
	breaker.Fail()
	require.False(t, breaker.Allowed())

	repo := decorator.NewBreakerDecorator(mock, breaker)

	// Act
	result, err := repo.GetCurrent(context.Background(), "Dnipro")

	// Assert
	require.ErrorIs(t, err, domain.ErrProviderUnreliable)
	assert.False(t, mock.Called)
	assert.Equal(t, domain.Weather{}, result)
}

func TestBreakerDecorator_ClosedAfterTimeout(t *testing.T) {
	// Arrange
	currentTime := time.Now()
	breaker := cb.NewCircuitBreaker(time.Minute, 1, 1)
	breaker.Now = func() time.Time {
		return currentTime
	}
	currentTime = currentTime.Add(time.Second)

	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrWeatherUnavailable,
	}
	repo := decorator.NewBreakerDecorator(mock, breaker)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kharkiv")
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)
	assert.False(t, breaker.Allowed())
	currentTime = currentTime.Add(time.Minute + time.Second)

	// Assert
	assert.True(t, breaker.Allowed())
}

func TestBreakerDecorator_OpenAgain(t *testing.T) {
	// Arrange
	currentTime := time.Now()
	breaker := cb.NewCircuitBreaker(time.Minute, 2, 2)
	breaker.Now = func() time.Time {
		return currentTime
	}
	currentTime = currentTime.Add(time.Second)

	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrWeatherUnavailable,
	}
	repo := decorator.NewBreakerDecorator(mock, breaker)

	// Act
	city := "Kharkiv"
	_, err := repo.GetCurrent(context.Background(), city)
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)
	assert.True(t, breaker.Allowed())
	_, err = repo.GetCurrent(context.Background(), city)
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)
	assert.False(t, breaker.Allowed())
	require.Equal(t, cb.Open, breaker.State())
	currentTime = currentTime.Add(time.Minute + time.Second)
	require.Equal(t, cb.HalfOpen, breaker.State())
	_, err = repo.GetCurrent(context.Background(), city)
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)

	// Assert
	require.Equal(t, cb.Open, breaker.State())
	assert.False(t, breaker.Allowed())
}
