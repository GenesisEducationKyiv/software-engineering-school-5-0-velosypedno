package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type SMTPEmailService struct {
	Host      string
	Port      string
	User      string
	Pass      string
	EmailFrom string
	Auth      smtp.Auth
}

func NewSMTPEmailService(host, port, user, pass, emailFrom string) *SMTPEmailService {
	auth := smtp.PlainAuth("", user, pass, host)
	return &SMTPEmailService{
		Host:      host,
		Port:      port,
		User:      user,
		Pass:      pass,
		EmailFrom: emailFrom,
		Auth:      auth,
	}
}

func (s *SMTPEmailService) sendEmail(recipient, subject, body string) error {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.EmailFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := s.Host + ":" + s.Port
	err := smtp.SendMail(addr, s.Auth, s.EmailFrom, []string{recipient}, []byte(msg.String()))
	if err != nil {
		return err
	}
	return nil
}

func (s *SMTPEmailService) SendConfirmationEmail(subscription models.Subscription) error {
	recipient := subscription.Email
	subject := "Subscription Confirmation"
	body := fmt.Sprintf(
		`Hello!

Please click the following link to confirm your subscription:
http://localhost:8080/api/confirm/%s			
	
Thank you!`,
		subscription.Token,
	)

	if err := s.sendEmail(recipient, subject, body); err != nil {
		return fmt.Errorf("smtp email service: failed to send confirmation email to %s: %w", recipient, err)
	}
	return nil
}

func (s *SMTPEmailService) SendWeatherEmail(subscription models.Subscription, weather models.Weather) error {
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
