//go:build unit

package chain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathchain "github.com/velosypedno/genesis-weather-api/internal/repos/weather/chain"
)

type mockProvider struct {
	resp   domain.Weather
	err    error
	called bool
}

func (m *mockProvider) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	m.called = true
	return m.resp, m.err
}

func TestWeatherRepoChain_FirstSuccess(t *testing.T) {
	// Arrange
	first := &mockProvider{
		resp: domain.Weather{Temperature: 20, Humidity: 50, Description: "Sunny"},
		err:  nil,
	}
	second := &mockProvider{
		err: errors.New("should not be called"),
	}
	chain := weathchain.NewProvidersFallbackChain(first, second)

	// Act
	weather, err := chain.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 20.0, weather.Temperature)
	assert.True(t, first.called)
	assert.False(t, second.called)
}

func TestWeatherRepoChain_SecondSuccess(t *testing.T) {
	// Arrange
	first := &mockProvider{
		err: errors.New("first failed"),
	}
	second := &mockProvider{
		resp: domain.Weather{Temperature: 10, Humidity: 80, Description: "Rain"},
		err:  nil,
	}
	chain := weathchain.NewProvidersFallbackChain(first, second)

	// Act
	weather, err := chain.GetCurrent(context.Background(), "Lviv")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10.0, weather.Temperature)
	assert.True(t, first.called)
	assert.True(t, second.called)
}

func TestWeatherRepoChain_AllFail(t *testing.T) {
	// Arrange
	first := &mockProvider{err: errors.New("first fail")}
	second := &mockProvider{err: errors.New("second fail")}
	chain := weathchain.NewProvidersFallbackChain(first, second)

	// Act
	weather, err := chain.GetCurrent(context.Background(), "Tsrcuny")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain:")
	assert.Equal(t, domain.Weather{}, weather)
	assert.True(t, first.called)
	assert.True(t, second.called)
}
