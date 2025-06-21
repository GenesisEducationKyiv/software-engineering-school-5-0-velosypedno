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

func (m *SubscriptionMailer) SendConfirmation(subscription domain.Subscription) error {
	to := subscription.Email
	subject := "Subscription Confirmation"
	confirmSubURL := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmpl, err := template.ParseFiles("internal/templates/confirm_sub.html")
	if err != nil {
		log.Println(err)
		return ErrInternal
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Link": confirmSubURL}); err != nil {
		log.Println(err)
		return ErrInternal
	}

	err = m.sender.Send(to, subject, body.String())
	if errors.Is(err, email.ErrSendEmail) {
		err = fmt.Errorf("smtp email service: failed to send confirmation email to %s", to)
		log.Println(err)
		return ErrSendEmail
	} else if err != nil {
		log.Println(err)
		return ErrInternal
	}
	return nil
}
