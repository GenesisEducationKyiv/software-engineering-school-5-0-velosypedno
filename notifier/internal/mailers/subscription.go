package mailers

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"go.uber.org/zap"
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
	logger          *zap.Logger
	sender          emailBackend
	confirmTmplPath string
}

func NewSubscriptionEmailNotifier(logger *zap.Logger, sender emailBackend, tmplPath string) *SubscriptionEmailNotifier {
	return &SubscriptionEmailNotifier{
		logger:          logger,
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
		m.logger.Error("failed to parse confirmation email template",
			zap.String("templatePath", m.confirmTmplPath),
			zap.Error(err),
		)
		return fmt.Errorf("sub mailer: %w", ErrInternal)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]string{"Link": confirmSubURL}); err != nil {
		m.logger.Error("failed to execute confirmation email template",
			zap.Error(err),
		)
		return fmt.Errorf("sub mailer: %w", ErrInternal)
	}

	if err := m.sender.Send(to, subject, body.String()); err != nil {
		m.logger.Error("failed to send confirmation email",
			zap.Error(err),
		)
		return fmt.Errorf("sub mailer: %w", ErrInternal)
	}

	m.logger.Info("confirmation email sent",
		zap.String("email_hash", logging.HashEmail(to)),
	)
	return nil
}
