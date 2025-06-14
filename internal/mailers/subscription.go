package mailers

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type emailSender interface {
	Send(to, subject, body string) error
}

type SubscriptionMailer struct {
	sender emailSender
}

func NewSubscriptionMailer(sender emailSender) *SubscriptionMailer {
	return &SubscriptionMailer{
		sender: sender,
	}
}

func (m *SubscriptionMailer) SendConfirmation(subscription models.Subscription) error {
	to := subscription.Email
	subject := "Subscription Confirmation"
	confirmSubURL := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmpl, err := template.ParseFiles("internal/templates/confirm_sub.html")
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Link": confirmSubURL}); err != nil {
		return err
	}

	if err := m.sender.Send(to, subject, body.String()); err != nil {
		return fmt.Errorf("smtp email service: failed to send confirmation email to %s: %w", to, err)
	}
	return nil
}
