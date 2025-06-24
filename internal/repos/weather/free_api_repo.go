package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

const noMatchingLocationFoundCode = 1006

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type WeatherAPIRepo struct {
	apiKey string
	client HTTPClient
}

func NewWeatherAPIRepo(apiKey string, client HTTPClient) *WeatherAPIRepo {
	return &WeatherAPIRepo{
		apiKey: apiKey,
		client: client,
	}
}

type weatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  float64 `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

type weatherAPIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r *WeatherAPIRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	// step 1: format request
	q := url.QueryEscape(city)
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", r.apiKey, q)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("weather repo: failed to format request for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("weather repo: %w", domain.ErrInternal)
	}

	// step 2: send request
	resp, err := r.client.Do(req)
	if err != nil {
		log.Printf("weather repo: failed to get weather for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("weather repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp body: %v\n", err)
		}
	}()

	// step 3: handle response
	if resp.StatusCode == http.StatusForbidden {
		log.Println("weather repo: api key is invalid")
		return domain.Weather{}, fmt.Errorf("weather repo: %w", domain.ErrWeatherUnavailable)
	}
	if resp.StatusCode != http.StatusOK {
		var errResp weatherAPIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Error.Code == noMatchingLocationFoundCode {
				return domain.Weather{}, domain.ErrCityNotFound
			}
			log.Printf("weather repo: api error: %s\n", errResp.Error.Message)
			return domain.Weather{}, fmt.Errorf("weather repo: %w", domain.ErrInternal)
		}
		log.Printf("weather repo: unexpected error %d\n", resp.StatusCode)
		return domain.Weather{}, fmt.Errorf("weather repo: %w", domain.ErrInternal)
	}

	// step 4: parse response body
	var responseData weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("weather repo: failed to decode weather data: %v\n", err)
		return domain.Weather{}, fmt.Errorf("weather repo: %w", domain.ErrInternal)
	}

	return domain.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Condition.Text,
	}, nil
}
