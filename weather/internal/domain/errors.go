package domain

import "errors"

var (
	ErrInternal           = errors.New("internal error")
	ErrCityNotFound       = errors.New("city not found")
	ErrWeatherUnavailable = errors.New("weather api is unavailable")
	ErrProviderUnreliable = errors.New("weather provider is unreliable")
)
