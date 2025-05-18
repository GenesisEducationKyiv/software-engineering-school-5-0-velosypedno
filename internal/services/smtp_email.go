package services

import (
	"fmt"
	"net/smtp"

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
	body := fmt.Sprintf("Click this link to confirm: http://localhost:8080/api/confirm/%s", subscription.Token)
	msg := []byte(subject + "\n" + body)
	addr := s.Host + ":" + s.Port
	err := smtp.SendMail(addr, auth, s.EmailFrom, to, msg)
	if err != nil {
		return err
	}
	return nil
}
