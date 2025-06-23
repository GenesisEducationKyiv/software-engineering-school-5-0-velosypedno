package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	services "github.com/velosypedno/genesis-weather-api/internal/services/weather"
)

type mockWeatherRepo struct {
	mock.Mock
}

func (m *mockWeatherRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(domain.Weather), args.Error(1)
}

func TestWeatherService_GetCurrent_Success(t *testing.T) {
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

	ctx := context.Background()
	actual, err := service.GetCurrent(ctx, "Kyiv")

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	mockRepo.AssertExpectations(t)
}

func TestWeatherService_GetCurrent_Error(t *testing.T) {
	mockRepo := new(mockWeatherRepo)
	service := services.NewWeatherService(mockRepo)

	mockRepo.
		On("GetCurrent", mock.Anything, "ZUUUBR").
		Return(domain.Weather{}, domain.ErrCityNotFound)

	ctx := context.Background()
	_, err := service.GetCurrent(ctx, "ZUUUBR")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCityNotFound)

	mockRepo.AssertExpectations(t)
}
