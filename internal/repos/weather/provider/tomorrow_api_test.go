//go:build unit

package provider_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathprovider "github.com/velosypedno/genesis-weather-api/internal/repos/weather/provider"
)

func TestTomorrowGetCurrentWeather_Success(t *testing.T) {
	// Arrange
	mockRespBody := `{
		"data": {
			"values": {
				"temperature": 10000.0,
				"humidity": 100.0,
				"visibility": 12.7,
				"cloudCover": 0.1
			}
		}
	}`
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
			}, nil
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewTomorrowAPI(cfg, client)

	// Act
	weather, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10000.0, weather.Temperature)
	assert.Equal(t, 100.0, weather.Humidity)
}

func TestTomorrowGetCurrentWeather_CityNotFound(t *testing.T) {
	// Arrange
	mockRespBody := `{
		"code": 400001,
		"message": "No matching location found.",
		"type": "error"
	}`
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
			}, nil
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewTomorrowAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "InvalidCity")

	// Assert
	assert.ErrorIs(t, err, domain.ErrCityNotFound)
}

func TestTomorrowGetCurrentWeather_APIKeyInvalid(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewTomorrowAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWeatherUnavailable)
}

func TestTomorrowGetCurrentWeather_HTTPError(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewTomorrowAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWeatherUnavailable)
}

func TestTomorrowGetCurrentWeather_BadJSON(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{invalid json}`)),
			}, nil
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewTomorrowAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInternal)
}
