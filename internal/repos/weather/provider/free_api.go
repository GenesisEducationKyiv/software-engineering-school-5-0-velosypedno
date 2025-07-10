package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
)

const noMatchingLocationFoundCode = 1006

type APICfg struct {
	APIKey string
	APIURL string
}
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type FreeWeatherAPI struct {
	cfg    APICfg
	client HTTPClient
}

func NewFreeWeatherAPI(cfg APICfg, client HTTPClient) *FreeWeatherAPI {
	return &FreeWeatherAPI{
		cfg:    cfg,
		client: client,
	}
}

type freeWeatherAPIResponse struct {
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  float64 `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

type freeWeatherAPIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r *FreeWeatherAPI) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	// step 1: format request
	q := url.QueryEscape(city)
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", r.cfg.APIURL, r.cfg.APIKey, q)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("free weather repo: failed to format request for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
	}

	// step 2: send request
	resp, err := r.client.Do(req)
	if err != nil {
		log.Printf("free weather repo: failed to get weather for %s, err:%v\n", city, err)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp body: %v\n", err)
		}
	}()

	// step 3: handle response
	if resp.StatusCode == http.StatusForbidden {
		log.Println("free weather repo: api key is invalid")
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrWeatherUnavailable)
	}
	if resp.StatusCode != http.StatusOK {
		var errResp freeWeatherAPIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Error.Code == noMatchingLocationFoundCode {
				log.Printf("free weather repo: city %s not found\n", city)
				return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrCityNotFound)
			}
			log.Printf("free weather repo: api error: %s\n", errResp.Error.Message)
			return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
		}
		log.Printf("free weather repo: unexpected error %d\n", resp.StatusCode)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
	}

	// step 4: parse response body
	var responseData freeWeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("free weather repo: failed to decode weather data: %v\n", err)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
	}

	return domain.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Condition.Text,
	}, nil
}
