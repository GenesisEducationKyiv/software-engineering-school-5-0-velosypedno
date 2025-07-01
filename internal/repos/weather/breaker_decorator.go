package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/pkg/cb"
)

type BreakerDecorator struct {
	Inner   weatherRepo
	Breaker *cb.CircuitBreaker
}

func NewBreakerDecorator(inner weatherRepo, breaker *cb.CircuitBreaker) *BreakerDecorator {
	return &BreakerDecorator{Inner: inner, Breaker: breaker}
}

func (d *BreakerDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	if !d.Breaker.Allowed() {
		return domain.Weather{}, fmt.Errorf("circuit breaker: %w", domain.ErrProviderUnreliable)
	}

	weather, err := d.Inner.GetCurrent(ctx, city)
	if errors.Is(err, domain.ErrWeatherUnavailable) {
		d.Breaker.Fail()
	}
	if err != nil {
		return domain.Weather{}, err
	}

	d.Breaker.Success()
	return weather, nil
}
