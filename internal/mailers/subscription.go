package mailers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"path/filepath"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
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

func (m *SubscriptionMailer) SendConfirmation(subscription domain.Subscription) error {
	to := subscription.Email
	subject := "Subscription Confirmation"
	confirmSubURL := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmplPath := filepath.Join(m.templateDir, m.confirmTmpl)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("sub mailer: %v\n", err)
		return fmt.Errorf("sub mailer: %w", domain.ErrInternal)
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Link": confirmSubURL}); err != nil {
		log.Printf("sub mailer: %v\n", err)
		return fmt.Errorf("sub mailer: %w", domain.ErrInternal)
	}
	err = m.sender.Send(to, subject, body.String())
	if err != nil {
		return fmt.Errorf("sub mailer: %w", err)
	}
	return nil
}
