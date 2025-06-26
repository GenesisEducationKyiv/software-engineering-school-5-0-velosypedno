//go:build integration

package integration_test

import (
	"net/http"
	"testing"
)

func TestGetWeatherSuccess(t *testing.T) {
	resp, err := http.Get(apiURL + "/api/weather?city=Kyiv")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from get weather: %d", resp.StatusCode)
	}

}

func TestGetWeatherInvalidCity(t *testing.T) {
	resp, err := http.Get(apiURL + "/api/weather?city=" + invalidCity)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Unexpected status code from get weather: %d", resp.StatusCode)
	}
}
