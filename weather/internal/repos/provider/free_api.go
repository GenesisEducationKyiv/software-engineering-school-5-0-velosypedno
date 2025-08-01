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

const freeWeatherName = "weatherapi.com"
const noMatchingLocationFoundCode = 1006

type APICfg struct {
	APIKey string
	APIURL string
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type FreeWeatherAPI struct {
	cfg    APICfg
	client httpClient
	logger *zap.Logger
}

func NewFreeWeatherAPI(logger *zap.Logger, cfg APICfg, client httpClient) *FreeWeatherAPI {
	return &FreeWeatherAPI{
		cfg:    cfg,
		client: client,
		logger: logger.With(zap.String("provider", freeWeatherName)),
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
		r.logger.Error("failed to create HTTP request",
			zap.String("city", city),
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
	}

	// step 2: perform request
	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.Error("failed to perform HTTP request",
			zap.String("url", url),
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			r.logger.Warn("failed to close HTTP response body",
				zap.Error(err),
			)
		}
	}()

	// step 3: handle response
	if resp.StatusCode == http.StatusForbidden {
		r.logger.Error("API key is invalid",
			zap.Int("status_code", resp.StatusCode),
		)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrWeatherUnavailable)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp freeWeatherAPIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Error.Code == noMatchingLocationFoundCode {
				r.logger.Info("city not found",
					zap.String("city", city),
				)
				return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrCityNotFound)
			}
			r.logger.Error("API returned error",
				zap.String("message", errResp.Error.Message),
				zap.Int("error_code", errResp.Error.Code),
			)
			return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
		}
		r.logger.Error("unexpected HTTP response status",
			zap.Int("status_code", resp.StatusCode),
		)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
	}

	var responseData freeWeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		r.logger.Error("failed to decode weather API response",
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("free weather repo: %w", domain.ErrInternal)
	}

	// step 4: return weather
	return domain.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Condition.Text,
	}, nil
}
