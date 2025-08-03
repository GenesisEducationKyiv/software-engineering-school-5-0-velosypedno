package mailers

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"go.uber.org/zap"
)

type Weather struct {
	Temperature float64
	Humidity    float64
	Description string
}

type WeatherEmailNotifier struct {
	logger   *zap.Logger
	sender   emailBackend
	tmplPath string
}

func NewWeatherEmailNotifier(logger *zap.Logger, sender emailBackend, tmplPath string) *WeatherEmailNotifier {
	return &WeatherEmailNotifier{
		logger:   logger,
		sender:   sender,
		tmplPath: tmplPath,
	}
}

func (m *WeatherEmailNotifier) SendCurrent(subscription Subscription, weather Weather) error {
	to := subscription.Email
	subject := "Weather Update"

	unsubscribeURL := fmt.Sprintf("http://localhost:8080/api/unsubscribe/%s", subscription.Token)
	tmpl, err := template.ParseFiles(m.tmplPath)
	if err != nil {
		m.logger.Error("failed to parse weather email template",
			zap.String("templatePath", m.tmplPath),
			zap.Error(err),
		)
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
		m.logger.Error("failed to execute weather email template",
			zap.Error(err),
		)
		return fmt.Errorf("weather mailer: %w", ErrInternal)
	}

	err = m.sender.Send(to, subject, body.String())
	if err != nil {
		m.logger.Error("failed to send weather email",
			zap.Error(err),
		)
		return fmt.Errorf("weather mailer: %w", ErrInternal)
	}

	m.logger.Info("weather email sent",
		zap.String("email_hash", logging.HashEmail(to)),
	)
	return nil
}
