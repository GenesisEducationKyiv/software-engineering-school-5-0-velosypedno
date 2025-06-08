package services

import (
	"bytes"
	"fmt"
	"html/template"
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
	msg.WriteString("MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n")
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
	confirmSubUrl := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmpl, err := template.ParseFiles("internal/templates/confirm_sub.html")
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Link": confirmSubUrl}); err != nil {
		return err
	}

	if err := s.sendEmail(recipient, subject, body.String()); err != nil {
		return fmt.Errorf("smtp email service: failed to send confirmation email to %s: %w", recipient, err)
	}
	return nil
}

func (s *SMTPEmailService) SendWeatherEmail(subscription models.Subscription, weather models.Weather) error {
	recipient := subscription.Email
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

	if err := s.sendEmail(recipient, subject, body.String()); err != nil {
		return fmt.Errorf("smtp email service: failed to send weather email to %s: %w", recipient, err)
	}
	return nil
}
