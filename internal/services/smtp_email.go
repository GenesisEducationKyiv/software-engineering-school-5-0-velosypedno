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
}

func NewSmtpEmailService(host, port, user, pass, emailFrom string) *SmtpEmailService {
	return &SmtpEmailService{
		Host:      host,
		Port:      port,
		User:      user,
		Pass:      pass,
		EmailFrom: emailFrom,
	}
}

func (s *SmtpEmailService) SendConfirmationEmail(subscription models.Subscription) error {
	recipient := subscription.Email

	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)
	to := []string{recipient}

	subject := "Subscription Confirmation"
	body := fmt.Sprintf("Hello!\n\nPlease click the following link to confirm your subscription:\nhttp://localhost:8080/api/confirm/%s\n\nThank you!", subscription.Token)

	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.EmailFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := s.Host + ":" + s.Port

	if err := smtp.SendMail(addr, auth, s.EmailFrom, to, []byte(msg.String())); err != nil {
		return fmt.Errorf("smtp email service: failed to send confirmation email to %s: %w", recipient, err)
	}
	return nil
}

func (s *SmtpEmailService) SendWeatherEmail(subscription models.Subscription, weather models.Weather) error {
	recipient := subscription.Email

	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)
	to := []string{recipient}

	subject := "Weather Update"
	body := fmt.Sprintf(
		"Hello!\n\nCurrent weather update:\nTemperature: %.1f°C\nHumidity: %.1f%%\nCondition: %s\n\nBest regards!",
		weather.Temperature,
		weather.Humidity,
		weather.Description,
	)

	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.EmailFrom))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n")
	msg.WriteString(body)

	addr := s.Host + ":" + s.Port

	if err := smtp.SendMail(addr, auth, s.EmailFrom, to, []byte(msg.String())); err != nil {
		return fmt.Errorf("smtp email service: failed to send weather email to %s: %w", recipient, err)
	}
	return nil
}
