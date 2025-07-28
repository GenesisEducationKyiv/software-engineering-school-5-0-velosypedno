package mailers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"
)

var ErrInternal = errors.New("internal error")

type Subscription struct {
	Email string
	Token string
}

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

func (m *SubscriptionEmailNotifier) SendConfirmation(subscription Subscription) error {
	to := subscription.Email
	subject := "Subscription Confirmation"
	confirmSubURL := fmt.Sprintf("http://localhost:8080/api/confirm/%s", subscription.Token)
	tmpl, err := template.ParseFiles(m.confirmTmplPath)
	if err != nil {
		log.Printf("sub mailer: %v\n", err)
		return fmt.Errorf("sub mailer: %w", ErrInternal)
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Link": confirmSubURL}); err != nil {
		log.Printf("sub mailer: %v\n", err)
		return fmt.Errorf("sub mailer: %w", ErrInternal)
	}
	err = m.sender.Send(to, subject, body.String())
	if err != nil {
		log.Printf("sub mailer: %v\n", err)
		return fmt.Errorf("sub mailer: %w", ErrInternal)
	}
	return nil
}
