package mailers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"path/filepath"

	"github.com/velosypedno/genesis-weather-api/internal/email"
	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type emailSender interface {
	Send(to, subject, body string) error
}

type SubscriptionMailer struct {
	sender      emailSender
	templateDir string
	confirmTmpl string
}

func NewSubscriptionMailer(sender emailSender, templateDir, confirmTmpl string) *SubscriptionMailer {
	return &SubscriptionMailer{
		sender:      sender,
		templateDir: templateDir,
		confirmTmpl: confirmTmpl,
	}
}

func (m *SubscriptionMailer) SendConfirmation(subscription models.Subscription) error {
	to := subscription.Email
	subject := "Subscription Confirmation"
	confirmSubURL := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmplPath := filepath.Join(m.templateDir, m.confirmTmpl)
	tmpl, err := template.ParseFiles(tmplPath)
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
