//go:build unit

package provider_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestFreeApiGetCurrentWeather_Success(t *testing.T) {
	// Arrange
	mockRespBody := `{
		"current": {
			"temp_c": 10000.0,
			"humidity": 100.0,
			"condition": {
				"text": "H_E_L_L"
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
	cfg := provider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := provider.NewFreeWeatherAPI(cfg, client)

	// Act
	weather, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10000.0, weather.Temperature)
	assert.Equal(t, 100.0, weather.Humidity)
	assert.Equal(t, "H_E_L_L", weather.Description)
}

func TestFreeApiGetCurrentWeather_CityNotFound(t *testing.T) {
	// Arrange
	mockRespBody := `{
		"error": {
			"code": 1006,
			"message": "No matching location found."
		}
	}`
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(mockRespBody)),
			}, nil
		},
	}
	cfg := provider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := provider.NewFreeWeatherAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "InvalidCity")

	// Assert
	assert.ErrorIs(t, err, domain.ErrCityNotFound)
}

func TestFreeApiGetCurrentWeather_APIKeyInvalid(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}
	cfg := provider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := provider.NewFreeWeatherAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWeatherUnavailable)
}

func TestFreeApiGetCurrentWeather_HTTPError(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}
	cfg := provider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := provider.NewFreeWeatherAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrWeatherUnavailable)
}

func TestFreeApiGetCurrentWeather_BadJSON(t *testing.T) {
	// Arrange
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{invalid json}`)),
			}, nil
		},
	}
	cfg := provider.APICfg{APIKey: "dummy-api-key", APIURL: "http://dummy-url.com"}
	repo := provider.NewFreeWeatherAPI(cfg, client)

	// Act
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInternal)
}
