//go:build unit

package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockWeatherRepo struct {
	mock.Mock
}

func (m *mockWeatherRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	args := m.Called(ctx, city)
	weather, ok := args.Get(0).(domain.Weather)
	if !ok {
		return domain.Weather{}, fmt.Errorf("mock: expected models.Weather, got %T", weather)
	}
	return weather, args.Error(1)
}

func TestWeatherService_GetCurrent_Success(t *testing.T) {
	// Arrange
	mockRepo := new(mockWeatherRepo)
	service := services.NewWeatherService(mockRepo)
	expected := domain.Weather{
		Temperature: 20.0,
		Humidity:    80.0,
		Description: "Sunny",
	}
	mockRepo.
		On("GetCurrent", mock.Anything, "Kyiv").
		Return(expected, nil)

	// Act
	ctx := context.Background()
	actual, err := service.GetCurrent(ctx, "Kyiv")

	// Assert
	mockRepo.AssertExpectations(t)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

}

func TestWeatherService_GetCurrent_Error(t *testing.T) {
	// Arrange
	mockRepo := new(mockWeatherRepo)
	service := services.NewWeatherService(mockRepo)
	mockRepo.
		On("GetCurrent", mock.Anything, "ZUUUBR").
		Return(domain.Weather{}, domain.ErrCityNotFound)

	// Act
	ctx := context.Background()
	_, err := service.GetCurrent(ctx, "ZUUUBR")

	// Assert
	mockRepo.AssertExpectations(t)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCityNotFound)

}
