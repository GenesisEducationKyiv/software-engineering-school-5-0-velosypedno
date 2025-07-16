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
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	weathprovider "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/provider"
)

func TestVisualCrossingGetCurrentWeather_Success(t *testing.T) {
	// Arrange
	mockRespBody := `{
		"currentConditions": {
			"temp": 10000.0,
			"humidity": 100.0,
			"conditions": "H_E_L_L"
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
	repo := weathprovider.NewVisualCrossingAPI(cfg, client)

	// Act
	weather, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10000.0, weather.Temperature)
	assert.Equal(t, 100.0, weather.Humidity)
	assert.Equal(t, "H_E_L_L", weather.Description)
}

func TestVisualCrossingGetCurrentWeather_CityNotFound(t *testing.T) {
	// Arrange
	mockRespBody := `NOt found`
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
			}, nil
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewVisualCrossingAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "InvalidCity")

	// Assert
	assert.ErrorIs(t, err, domain.ErrCityNotFound)
}

func TestVisualCrossingGetCurrentWeather_APIKeyInvalid(t *testing.T) {
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
	repo := weathprovider.NewVisualCrossingAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWeatherUnavailable)
}

func TestVisualCrossingGetCurrentWeather_HTTPError(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}
	cfg := weathprovider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := weathprovider.NewVisualCrossingAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWeatherUnavailable)
}

func TestVisualCrossingGetCurrentWeather_BadJSON(t *testing.T) {
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
	repo := weathprovider.NewVisualCrossingAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInternal)
}
