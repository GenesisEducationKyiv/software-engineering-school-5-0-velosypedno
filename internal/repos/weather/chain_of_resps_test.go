package repos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathr "github.com/velosypedno/genesis-weather-api/internal/repos/weather"
)

type mockRepo struct {
	resp   domain.Weather
	err    error
	called bool
}

func (m *mockRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	m.called = true
	return m.resp, m.err
}

func TestWeatherRepoChain_FirstSuccess(t *testing.T) {
	// Arrange
	first := &mockRepo{
		resp: domain.Weather{Temperature: 20, Humidity: 50, Description: "Sunny"},
		err:  nil,
	}
	second := &mockRepo{
		err: errors.New("should not be called"),
	}
	chain := weathr.NewChain(first, second)

	// Act
	weather, err := chain.GetCurrent(context.Background(), "Kyiv")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 20.0, weather.Temperature)
	assert.True(t, first.called)
	assert.False(t, second.called)
}

func TestWeatherRepoChain_SecondSuccess(t *testing.T) {
	// Arrange
	first := &mockRepo{
		err: errors.New("first failed"),
	}
	second := &mockRepo{
		resp: domain.Weather{Temperature: 10, Humidity: 80, Description: "Rain"},
		err:  nil,
	}
	chain := weathr.NewChain(first, second)

	// Act
	weather, err := chain.GetCurrent(context.Background(), "Lviv")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 10.0, weather.Temperature)
	assert.True(t, first.called)
	assert.True(t, second.called)
}

func TestWeatherRepoChain_AllFail(t *testing.T) {
	// Arrange
	first := &mockRepo{err: errors.New("first fail")}
	second := &mockRepo{err: errors.New("second fail")}
	chain := weathr.NewChain(first, second)

	// Act
	weather, err := chain.GetCurrent(context.Background(), "Tsrcuny")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chain:")
	assert.Equal(t, domain.Weather{}, weather)
	assert.True(t, first.called)
	assert.True(t, second.called)
}
