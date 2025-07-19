package decorator

import (
	"context"
	"errors"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cb"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
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
