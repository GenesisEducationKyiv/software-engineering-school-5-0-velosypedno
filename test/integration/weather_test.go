//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWeatherSuccess(t *testing.T) {
	resp, err := http.Get(apiURL + "/api/weather?city=Kyiv")
	require.NoError(t, err, "Failed to send GET request: %v", err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK, got %d", resp.StatusCode)

}

func TestGetWeatherInvalidCity(t *testing.T) {
	resp, err := http.Get(apiURL + "/api/weather?city=" + invalidCity)
	require.NoError(t, err, "Failed to send GET request: %v", err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected status 404 Not Found, got %d", resp.StatusCode)
}
