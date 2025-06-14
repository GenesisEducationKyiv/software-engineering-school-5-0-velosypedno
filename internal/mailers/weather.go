package mailers

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type WeatherMailer struct {
	sender emailSender
}

func NewWeatherMailer(sender emailSender) *WeatherMailer {
	return &WeatherMailer{
		sender: sender,
	}
}

func (m *WeatherMailer) SendCurrent(subscription models.Subscription, weather models.Weather) error {
	to := subscription.Email
	subject := "Weather Update"

	unsubscribeURL := fmt.Sprintf("http://localhost:8080/api/unsubscribe/%s", subscription.Token)
	tmpl, err := template.ParseFiles("internal/templates/weather.html")
	if err != nil {
		return err
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]any{
		"Temperature": weather.Temperature,
		"Humidity":    weather.Humidity,
		"Condition":   weather.Description,
		"Link":        unsubscribeURL,
	})
	if err != nil {
		return err
	}

	if err := m.sender.Send(to, subject, body.String()); err != nil {
		return fmt.Errorf("smtp email service: failed to send weather email to %s: %w", to, err)
	}
	return nil
}
