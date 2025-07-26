package notifiers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
)

type emailBackend interface {
	Send(to, subject, body string) error
}

type SubscriptionEmailNotifier struct {
	sender          emailBackend
	confirmTmplPath string
}

func NewSubscriptionEmailNotifier(sender emailBackend, tmplPath string) *SubscriptionEmailNotifier {
	return &SubscriptionEmailNotifier{
		sender:          sender,
		confirmTmplPath: tmplPath,
	}
}

func (m *SubscriptionEmailNotifier) SendConfirmation(subscription domain.Subscription) error {
	to := subscription.Email
	subject := "Subscription Confirmation"
	confirmSubURL := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmpl, err := template.ParseFiles(m.confirmTmplPath)
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
