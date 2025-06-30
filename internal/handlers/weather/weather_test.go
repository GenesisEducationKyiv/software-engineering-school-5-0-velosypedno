//go:build unit

package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathh "github.com/velosypedno/genesis-weather-api/internal/handlers/weather"
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

func TestWeatherHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWeather := domain.Weather{
		Temperature: 1000.0,
		Humidity:    100.0,
		Description: "H_E_L_L",
	}

	tests := []struct {
		name           string
		city           string
		mockReturn     domain.Weather
		mockError      error
		expectedStatus int
	}{
		{
			name:           "MissParam",
			city:           "",
			mockReturn:     domain.Weather{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "NotFound",
			city:           "Bagatkino",
			mockReturn:     domain.Weather{},
			mockError:      domain.ErrCityNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrInternal",
			city:           "Kyiv",
			mockReturn:     domain.Weather{},
			mockError:      domain.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Success",
			city:           "Kyiv",
			mockReturn:     mockWeather,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	var requestTimeout = 5 * time.Second
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(mockWeatherRepo)
			if tt.city != "" {
				mockRepo.
					On("GetCurrent", mock.Anything, tt.city).
					Return(tt.mockReturn, tt.mockError)
			}
			router := gin.New()
			router.GET("/weather", weathh.NewWeatherGETHandler(mockRepo, requestTimeout))

			// Act
			req := httptest.NewRequest(http.MethodGet, "/weather?city="+tt.city, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

type timeoutErrRepo struct {
}

func (t *timeoutErrRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	select {
	case <-time.After(time.Second):
		return domain.Weather{}, nil
	case <-ctx.Done():
		return domain.Weather{}, ctx.Err()
	}
}

func TestWeatherHandler_Timeout(t *testing.T) {
	// Arrange
	router := gin.New()
	router.GET("/weather", weathh.NewWeatherGETHandler(&timeoutErrRepo{}, time.Millisecond))

	// Act
	req := httptest.NewRequest(http.MethodGet, "/weather?city=Kyiv", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}
