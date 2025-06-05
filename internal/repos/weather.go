package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

var ErrCityNotFound = fmt.Errorf("weather repo: city not found")

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

const noMatchingLocationFoundCode = 1006

type weatherAPIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r *WeatherAPIRepo) GetCurrentWeather(ctx context.Context, city string) (models.Weather, error) {
	q := url.QueryEscape(city)
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", r.apiKey, q)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = fmt.Errorf("weather repo: failed to format request for %s, err:%v ", city, err)
		return models.Weather{}, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		err = fmt.Errorf("weather repo: failed to get weather for %s, err:%v ", city, err)
		return models.Weather{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		err = fmt.Errorf("weather repo: api key is invalid")
		return models.Weather{}, err
	}
	if resp.StatusCode != http.StatusOK {
		var errResp weatherAPIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Error.Code == noMatchingLocationFoundCode {
				return models.Weather{}, ErrCityNotFound
			}
			err = fmt.Errorf("weather repo: api error: %s", errResp.Error.Message)
			return models.Weather{}, err
		}
		err = fmt.Errorf("weather repo: unexpected error %d", resp.StatusCode)
		return models.Weather{}, err
	}

	var responseData weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		err = fmt.Errorf("weather repo: failed to decode weather data: %w", err)
		return models.Weather{}, err
	}

	return models.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Condition.Text,
	}, nil
}
