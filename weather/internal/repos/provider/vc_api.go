package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
)

const visualCrossingName = "visualcrossing.com"

type VisualCrossingAPI struct {
	cfg     APICfg
	client  httpClient
	logger  *zap.Logger
	metrics metrics
}

type visualCrossingAPIResponse struct {
	Current struct {
		TempC       float64 `json:"temp"`
		Humidity    float64 `json:"humidity"`
		Description string  `json:"conditions"`
	} `json:"currentConditions"`
}

func NewVisualCrossingAPI(logger *zap.Logger, cfg APICfg, client httpClient, metrics metrics) *VisualCrossingAPI {
	return &VisualCrossingAPI{
		cfg:     cfg,
		client:  client,
		logger:  logger.With(zap.String("provider", visualCrossingName)),
		metrics: metrics,
	}
}

func (r *VisualCrossingAPI) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	q := url.QueryEscape(city)
	url := fmt.Sprintf("%s/%s/today?key=%s&include=current&unitGroup=metric", r.cfg.APIURL, q, r.cfg.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		r.logger.Error("failed to create HTTP request",
			zap.String("city", city),
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
	}

	r.metrics.Request(visualCrossingName)
	start := time.Now()

	resp, err := r.client.Do(req)
	if err != nil {
		r.metrics.Error(visualCrossingName)
		r.metrics.RequestDuration(visualCrossingName, time.Since(start).Seconds())

		r.logger.Error("failed to perform HTTP request",
			zap.String("url", url),
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrWeatherUnavailable)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			r.logger.Warn("failed to close HTTP response body",
				zap.Error(err),
			)
		}
	}()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		r.metrics.Error(visualCrossingName)
		r.metrics.RequestDuration(visualCrossingName, time.Since(start).Seconds())

		r.logger.Error("API key is invalid",
			zap.Int("status_code", resp.StatusCode),
		)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrWeatherUnavailable)

	case http.StatusInternalServerError, http.StatusBadRequest:
		r.metrics.Error(visualCrossingName)
		r.metrics.RequestDuration(visualCrossingName, time.Since(start).Seconds())

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			r.logger.Error("failed to read response body",
				zap.Error(err),
			)
			return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
		}

		r.logger.Error("API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(bodyBytes)),
		)

		if resp.StatusCode == http.StatusBadRequest {
			return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrCityNotFound)
		}
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)

	case http.StatusOK:
	default:
		r.metrics.Error(visualCrossingName)
		r.metrics.RequestDuration(visualCrossingName, time.Since(start).Seconds())

		r.logger.Error("unexpected HTTP response status",
			zap.Int("status_code", resp.StatusCode),
		)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
	}

	var responseData visualCrossingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		r.metrics.Error(visualCrossingName)
		r.metrics.RequestDuration(visualCrossingName, time.Since(start).Seconds())

		r.logger.Error("failed to decode weather API response",
			zap.Error(err),
		)
		return domain.Weather{}, fmt.Errorf("visual crossing repo: %w", domain.ErrInternal)
	}

	r.metrics.RequestDuration(visualCrossingName, time.Since(start).Seconds())

	return domain.Weather{
		Temperature: responseData.Current.TempC,
		Humidity:    responseData.Current.Humidity,
		Description: responseData.Current.Description,
	}, nil
}
