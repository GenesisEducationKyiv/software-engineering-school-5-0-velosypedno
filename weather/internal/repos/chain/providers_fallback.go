package chain

import (
	"context"
	"errors"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
)

type weatherProvider interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type ProvidersFallbackChain struct {
	Repos []weatherProvider
}

func NewProvidersFallbackChain(repos ...weatherProvider) *ProvidersFallbackChain {
	return &ProvidersFallbackChain{Repos: repos}
}

func (c *ProvidersFallbackChain) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var errs []error

	for _, repo := range c.Repos {
		weather, err := repo.GetCurrent(ctx, city)
		if err != nil {
			err = fmt.Errorf("chain: %w", err)
			errs = append(errs, err)
			continue
		}
		return weather, nil
	}

	if len(errs) == 0 {
		return domain.Weather{}, fmt.Errorf("chain: %w", domain.ErrWeatherUnavailable)
	}

	return domain.Weather{}, errors.Join(errs...)
}
