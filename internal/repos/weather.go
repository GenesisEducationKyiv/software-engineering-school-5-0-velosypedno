package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type WeatherAPIRepo struct {
	apiKey string
	client *http.Client
}

func NewWeatherAPIRepo(apiKey string) *WeatherAPIRepo {
	return &WeatherAPIRepo{
		apiKey: apiKey,
		client: &http.Client{},
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

func (r *WeatherAPIRepo) GetCurrentWeather(ctx context.Context, city string) (models.Weather, error) {
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", r.apiKey, city)
	log.Println(url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Println(err, 1)
		return models.Weather{}, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		log.Println(err, 2)
		return models.Weather{}, err
	}
	defer resp.Body.Close()

	var responseData weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Println(err, 3)
		return models.Weather{}, err
	}

	return models.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Condition.Text,
	}, nil
}
