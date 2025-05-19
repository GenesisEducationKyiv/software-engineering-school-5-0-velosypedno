package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type SmtpEmailService struct {
	Host      string
	Port      string
	User      string
	Pass      string
	EmailFrom string
	Auth      smtp.Auth
}

func NewSmtpEmailService(host, port, user, pass, emailFrom string) *SmtpEmailService {
	auth := smtp.PlainAuth("", user, pass, host)
	return &SmtpEmailService{
		Host:      host,
		Port:      port,
		User:      user,
		Pass:      pass,
		EmailFrom: emailFrom,
		Auth:      auth,
	}
}

func (s *SmtpEmailService) sendEmail(recipient, subject, body string) error {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.EmailFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := s.Host + ":" + s.Port
	if err := smtp.SendMail(addr, s.Auth, s.EmailFrom, []string{recipient}, []byte(msg.String())); err != nil {
		return err
	}
	return nil
}

func (s *SmtpEmailService) SendConfirmationEmail(subscription models.Subscription) error {
	recipient := subscription.Email
	subject := "Subscription Confirmation"
	body := fmt.Sprintf("Hello!\n\nPlease click the following link to confirm your subscription:\nhttp://localhost:8080/api/confirm/%s\n\nThank you!", subscription.Token)

	if err := s.sendEmail(recipient, subject, body); err != nil {
		return fmt.Errorf("smtp email service: failed to send confirmation email to %s: %w", recipient, err)
	}
	return nil
}

func (s *SmtpEmailService) SendWeatherEmail(subscription models.Subscription, weather models.Weather) error {
	recipient := subscription.Email
	subject := "Weather Update"

	unsubscribeURL := fmt.Sprintf("http://localhost:8080/api/unsubscribe/%s", subscription.Token)

	body := fmt.Sprintf(
		`Hello!
		Current weather update:
		Temperature: %.1f°C
		Humidity: %.1f%%
		Condition: %s

		To unsubscribe from weather updates, click here: %s

		Best regards!`,
		weather.Temperature,
		weather.Humidity,
		weather.Description,
		unsubscribeURL,
	)

	if err := s.sendEmail(recipient, subject, body); err != nil {
		return fmt.Errorf("smtp email service: failed to send weather email to %s: %w", recipient, err)
	}
	return nil
}
