package mailers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/email"
)

var ErrSendEmail = email.ErrSendEmail
var ErrInternal = errors.New("mailer: internal error")

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
		log.Println(err)
		return ErrInternal
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, map[string]any{
		"Temperature": weather.Temperature,
		"Humidity":    weather.Humidity,
		"Condition":   weather.Description,
		"Link":        unsubscribeURL,
	})
	if err != nil {
		log.Println(err)
		return ErrInternal
	}

	err = m.sender.Send(to, subject, body.String())
	if errors.Is(err, email.ErrSendEmail) {
		err = fmt.Errorf("mailer: failed to send weather email to %s", to)
		log.Println(err)
		return ErrSendEmail
	} else if err != nil {
		log.Println(err)
		return ErrInternal
	}
	return nil
}
