package mailers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type WeatherMailer struct {
	sender emailSender
}

func NewWeatherMailer(sender emailSender) *WeatherMailer {
	return &WeatherMailer{
		sender: sender,
	}
}

func (m *WeatherMailer) SendCurrent(subscription domain.Subscription, weather domain.Weather) error {
	to := subscription.Email
	subject := "Weather Update"

	unsubscribeURL := fmt.Sprintf("http://localhost:8080/api/unsubscribe/%s", subscription.Token)
	tmpl, err := template.ParseFiles("internal/templates/weather.html")
	if err != nil {
		log.Printf("weather mailer: %v\n", err)
		return fmt.Errorf("weather mailer: %w", domain.ErrInternal)
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]any{
		"Temperature": weather.Temperature,
		"Humidity":    weather.Humidity,
		"Condition":   weather.Description,
		"Link":        unsubscribeURL,
	})
	if err != nil {
		log.Printf("weather mailer: %v\n", err)
		return fmt.Errorf("weather mailer: %w", domain.ErrInternal)
	}

	err = m.sender.Send(to, subject, body.String())
	if err != nil {
		return fmt.Errorf("weather mailer: %w", err)
	}
	return nil
}
