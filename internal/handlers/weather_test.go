package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/velosypedno/genesis-weather-api/internal/models"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

type mockWeatherRepo struct {
	mock.Mock
}

func (m *mockWeatherRepo) GetCurrentWeather(ctx context.Context, city string) (models.Weather, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(models.Weather), args.Error(1)
}

func TestNewWeatherGETHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWeather := models.Weather{
		Temperature: 1000.0,
		Humidity:    100.0,
		Description: "H_E_L_L",
	}

	tests := []struct {
		name           string
		city           string
		mockReturn     models.Weather
		mockError      error
		expectedStatus int
	}{
		{
			name:           "missing city parameter",
			city:           "",
			mockReturn:     models.Weather{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "city not found",
			city:           "Bagatkino",
			mockReturn:     models.Weather{},
			mockError:      repos.ErrCityNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal error",
			city:           "Kyiv",
			mockReturn:     models.Weather{},
			mockError:      errors.New("api failure"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "successful weather fetch",
			city:           "Kyiv",
			mockReturn:     mockWeather,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockWeatherRepo)

			if tt.city != "" {
				mockRepo.
					On("GetCurrentWeather", mock.Anything, tt.city).
					Return(tt.mockReturn, tt.mockError)
			}

			router := gin.New()
			router.GET("/weather", NewWeatherGETHandler(mockRepo))
			req := httptest.NewRequest(http.MethodGet, "/weather?city="+tt.city, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedStatus, resp.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
