package chain

import (
	"context"
	"fmt"
	"log"

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
	var lastError error
	for _, repo := range c.Repos {
		weather, err := repo.GetCurrent(ctx, city)
		if err != nil {
			err = fmt.Errorf("chain: %w", err)
			log.Println(err)
			lastError = err
			continue
		}
		return weather, nil
	}
	return domain.Weather{}, lastError
}
