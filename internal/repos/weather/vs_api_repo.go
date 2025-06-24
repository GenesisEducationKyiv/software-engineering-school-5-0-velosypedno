package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

const baseURL = "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/"

type VisualCrossingAPIRepo struct {
	apiKey string
	client HTTPClient
}

type visualCrossingAPIResponse struct {
	Current struct {
		TempC       float64 `json:"temp"`
		Humidity    float64 `json:"humidity"`
		Description string  `json:"conditions"`
	} `json:"currentConditions"`
}

func NewVisualCrossingAPIRepo(apiKey string, client HTTPClient) *VisualCrossingAPIRepo {
	return &VisualCrossingAPIRepo{
		apiKey: apiKey,
		client: client,
	}
}

func (r *VisualCrossingAPIRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	// step 1: format request
	q := url.QueryEscape(city)
	url := fmt.Sprintf("%s%s/today?key=%s&include=current&unitGroup=metric", baseURL, q, r.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("visual crossing repo: failed to format request for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	// step 2: send request
	resp, err := r.client.Do(req)
	if err != nil {
		log.Printf("visual crossing repo: failed to get weather for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("visual crossing repo: failed to close resp body: %v\n", err)
		}
	}()

	// step 3: handle response
	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("visual crossing repo: api key is invalid")
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrWeatherUnavailable)
	}
	if resp.StatusCode == http.StatusInternalServerError {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("visual crossing repo: failed to read response body: %v\n", err)
			return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
		}

		log.Printf("visual crossing repo: api error: %s\n", string(bodyBytes))
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
	}
	if resp.StatusCode == http.StatusBadRequest {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("visual crossing repo: failed to read response body: %v\n", err)
			return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
		}

		log.Printf("visual crossing repo: api error: %s\n", string(bodyBytes))
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrCityNotFound)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("visual crossing repo: unexpected error %d\n", resp.StatusCode)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
	}

	// step 4: parse response body
	var responseData visualCrossingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("visual crossing repo: failed to decode weather data: %v\n", err)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
	}
	return domain.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Description,
	}, nil
}
