package repos

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

	repo := NewWeatherAPIRepo("dummy-api-key", client)
	weather, err := repo.GetCurrentWeather(context.Background(), "Kyiv")

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

	repo := NewWeatherAPIRepo("dummy-api-key", client)
	_, err := repo.GetCurrentWeather(context.Background(), "InvalidCity")

	assert.ErrorIs(t, err, ErrCityNotFound)
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

	repo := NewWeatherAPIRepo("invalid-api-key", client)
	_, err := repo.GetCurrentWeather(context.Background(), "Kyiv")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api key is invalid")
}

func TestGetCurrentWeather_HTTPError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	repo := NewWeatherAPIRepo("dummy-api-key", client)
	_, err := repo.GetCurrentWeather(context.Background(), "Kyiv")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get weather")
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

	repo := NewWeatherAPIRepo("dummy-api-key", client)
	_, err := repo.GetCurrentWeather(context.Background(), "Kyiv")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
}
