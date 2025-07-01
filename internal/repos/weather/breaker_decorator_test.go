//go:build unit

package repos_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathr "github.com/velosypedno/genesis-weather-api/internal/repos/weather"
	"github.com/velosypedno/genesis-weather-api/pkg/cb"
)

func newTestBreaker() *cb.CircuitBreaker {
	breaker := cb.NewCircuitBreaker(time.Minute, 1)
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
	repo := weathr.NewBreakerDecorator(mock, breaker)

	// Act
	result, err := repo.GetCurrent(context.Background(), "Lviv")

	// Assert
	require.NoError(t, err)
	require.True(t, mock.Called)
	assert.Equal(t, 20.0, result.Temperature)
	assert.True(t, breaker.IsClosed())
}

func TestBreakerDecorator_UnavailableErrorTriggersBreaker(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrWeatherUnavailable,
	}
	breaker := newTestBreaker()
	repo := weathr.NewBreakerDecorator(mock, breaker)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Odessa")

	// Assert
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)
	assert.False(t, breaker.IsClosed())
}

func TestBreakerDecorator_OtherErrorDoesNotTriggerBreaker(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrCityNotFound,
	}
	breaker := newTestBreaker()
	repo := weathr.NewBreakerDecorator(mock, breaker)

	// Act
	_, err := repo.GetCurrent(context.Background(), "!!!")

	// Assert
	require.ErrorIs(t, err, domain.ErrCityNotFound)
	assert.True(t, breaker.IsClosed())
}

func TestBreakerDecorator_BreakerOpenSkipsCall(t *testing.T) {
	// Arrange
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      nil,
	}
	breaker := newTestBreaker()
	breaker.Fail()
	require.False(t, breaker.IsClosed())

	repo := weathr.NewBreakerDecorator(mock, breaker)

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
	breaker := cb.NewCircuitBreaker(time.Minute, 1)
	breaker.Now = func() time.Time {
		return currentTime
	}
	currentTime = currentTime.Add(time.Second)

	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrWeatherUnavailable,
	}
	repo := weathr.NewBreakerDecorator(mock, breaker)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kharkiv")
	require.ErrorIs(t, err, domain.ErrWeatherUnavailable)
	assert.False(t, breaker.IsClosed())
	currentTime = currentTime.Add(time.Minute + time.Second)

	// Assert
	assert.True(t, breaker.IsClosed())
}
