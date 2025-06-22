//go:build unit
// +build unit

package repos_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestGetCurrentWeather_Success(t *testing.T) {
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

	repo := repos.NewWeatherAPIRepo("dummy-api-key", "http://dummy-url", client)
	weather, err := repo.GetCurrent(context.Background(), "Kyiv")

	assert.NoError(t, err)
	assert.Equal(t, 10000.0, weather.Temperature)
	assert.Equal(t, 100.0, weather.Humidity)
	assert.Equal(t, "H_E_L_L", weather.Description)
}

func TestGetCurrentWeather_CityNotFound(t *testing.T) {
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

	repo := repos.NewWeatherAPIRepo("dummy-api-key", "http://dummy-url", client)
	_, err := repo.GetCurrent(context.Background(), "InvalidCity")

	assert.ErrorIs(t, err, domain.ErrCityNotFound)
}

func TestGetCurrentWeather_APIKeyInvalid(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	repo := repos.NewWeatherAPIRepo("invalid-api-key", "http://dummy-url", client)
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInternal)
}

func TestGetCurrentWeather_HTTPError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	repo := repos.NewWeatherAPIRepo("dummy-api-key", "http://dummy-url", client)
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInternal)
}

func TestGetCurrentWeather_BadJSON(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{invalid json}`)),
			}, nil
		},
	}

	repo := repos.NewWeatherAPIRepo("dummy-api-key", "http://dummy-url", client)
	_, err := repo.GetCurrent(context.Background(), "Kyiv")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInternal)
}
