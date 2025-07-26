package mailers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
)

type Weather struct {
	Temperature float64
	Humidity    float64
	Description string
}

type WeatherEmailNotifier struct {
	sender emailBackend
}

func NewWeatherEmailNotifier(sender emailBackend) *WeatherEmailNotifier {
	return &WeatherEmailNotifier{
		sender: sender,
	}
}

func (m *WeatherEmailNotifier) SendCurrent(subscription Subscription, weather Weather) error {
	to := subscription.Email
	subject := "Weather Update"

	unsubscribeURL := fmt.Sprintf("http://localhost:8080/api/unsubscribe/%s", subscription.Token)
	tmpl, err := template.ParseFiles("internal/templates/weather.html")
	if err != nil {
		log.Printf("weather mailer: %v\n", err)
		return fmt.Errorf("weather mailer: %w", ErrInternal)
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
		return fmt.Errorf("weather mailer: %w", ErrInternal)
	}

	err = m.sender.Send(to, subject, body.String())
	if err != nil {
		log.Printf("weather mailer: %v\n", err)
		return fmt.Errorf("weather mailer: %w", ErrInternal)
	}
	return nil
}
