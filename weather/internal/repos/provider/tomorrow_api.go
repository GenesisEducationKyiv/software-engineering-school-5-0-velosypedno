package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
)

const tomorrowIOName = "tomorrow.io"
const tomorrowCityNotFoundCode = 400001

type TomorrowAPI struct {
	cfg    APICfg
	client httpClient
	logger *zap.Logger
}

func NewTomorrowAPI(logger *zap.Logger, cfg APICfg, client httpClient) *TomorrowAPI {
	return &TomorrowAPI{
		cfg:    cfg,
		client: client,
		logger: logger.With(zap.String("provider", tomorrowIOName)),
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
		r.logger.Error("failed to create HTTP request",
			zap.String("city", city),
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	// step 2: perform request
	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.Error("failed to perform HTTP request",
			zap.String("url", url),
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			r.logger.Warn("failed to close HTTP response body",
				zap.Error(err),
			)
		}
	}()

	// step 3: handle response
	if resp.StatusCode == http.StatusUnauthorized {
		r.logger.Error("API key is invalid",
			zap.Int("status_code", resp.StatusCode),
		)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrWeatherUnavailable)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp tomorrowAPIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Code == tomorrowCityNotFoundCode {
				r.logger.Info("city not found",
					zap.String("city", city),
				)
				return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrCityNotFound)
			}
			r.logger.Error("API returned error",
				zap.String("message", errResp.Message),
				zap.Int("error_code", errResp.Code),
			)
			return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
		}
		r.logger.Error("unexpected HTTP response status",
			zap.Int("status_code", resp.StatusCode),
		)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	var responseData tomorrowAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		r.logger.Error("failed to decode weather API response",
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("tomorrow weather repo: %w", domain.ErrInternal)
	}

	// step 4: return weather
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
