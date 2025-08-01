package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
)

const tomorrowCityNotFoundCode = 400001

type TomorrowAPI struct {
	cfg    APICfg
	client HTTPClient
}

func NewTomorrowAPI(cfg APICfg, client HTTPClient) *TomorrowAPI {
	return &TomorrowAPI{
		cfg:    cfg,
		client: client,
	}
}

type tomorrowAPIResponse struct {
	Data struct {
		Values struct {
			Temperature float64 `json:"temperature"`
			Humidity    float64 `json:"humidity"`
			Visibility  float64 `json:"visibility"`
			CloudCover  float64 `json:"cloudCover"`
		} `json:"values"`
	} `json:"data"`
}

type tomorrowAPIErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (r *TomorrowAPI) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	// step 1: format request
	q := url.QueryEscape(city)
	url := fmt.Sprintf("%s/weather/realtime?location=%s&apikey=%s", r.cfg.APIURL, q, r.cfg.APIKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("tomorrow weather repo: failed to format request for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	// step 2: send request
	resp, err := r.client.Do(req)
	if err != nil {
		log.Printf("tomorrow weather repo: failed to get weather for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("tomorrow weather repo: failed to close resp body: %v\n", err)
		}
	}()

	// step 3: handle response
	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("tomorrow weather repo: api key is invalid")
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrWeatherUnavailable)
	}
	if resp.StatusCode != http.StatusOK {
		var errResp tomorrowAPIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Code == tomorrowCityNotFoundCode {
				log.Printf("tomorrow weather repo: city %s not found\n", city)
				return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrCityNotFound)
			}
			log.Printf("tomorrow weather repo: api error: %s\n", errResp.Message)
			return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
		}
		log.Printf("tomorrow weather repo: unexpected error %d\n", resp.StatusCode)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	// step 4: parse response body
	var responseData tomorrowAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("tomorrow weather repo: failed to decode weather data: %v\n", err)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	description := fmt.Sprintf("Cloud cover: %.2f%%", responseData.Data.Values.CloudCover)
	if responseData.Data.Values.Visibility > 0 {
		description += fmt.Sprintf("\nVisibility: %.2f km", responseData.Data.Values.Visibility)
	}
	return domain.Weather{
		Temperature: responseData.Data.Values.Temperature,
		Humidity:    responseData.Data.Values.Humidity,
		Description: description,
	}, nil
}
