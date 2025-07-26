package handlers

import (
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/mailers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
)

type weatherNotifyMailer interface {
	SendCurrent(subscription mailers.Subscription, weather mailers.Weather) error
}
type WeatherNotifyCommandHandler struct {
	Mailer weatherNotifyMailer
}

func NewWeatherNotifyCommandHandler(mailer weatherNotifyMailer) *WeatherNotifyCommandHandler {
	return &WeatherNotifyCommandHandler{
		Mailer: mailer,
	}
}

func (h *WeatherNotifyCommandHandler) Handle(command messaging.WeatherNotifyCommand) error {
	sub := mailers.Subscription{
		Email: command.Email,
		Token: command.Token,
	}
	weather := mailers.Weather{
		Temperature: command.Weather.Temperature,
		Humidity:    command.Weather.Humidity,
		Description: command.Weather.Description,
	}
	err := h.Mailer.SendCurrent(sub, weather)
	if err != nil {
		return fmt.Errorf("weather command handler: %w", err)
	}
	return nil
}
