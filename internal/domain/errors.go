package domain

import "errors"

var (
	ErrInternal           = errors.New("internal error")
	ErrCityNotFound       = errors.New("city not found")
	ErrSubNotFound        = errors.New("subscription not found")
	ErrSubAlreadyExists   = errors.New("subscription already exists")
	ErrSendEmail          = errors.New("failed to send email")
	ErrWeatherUnavailable = errors.New("weather api is unavailable")
)
