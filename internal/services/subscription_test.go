package services_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/models"
	"github.com/velosypedno/genesis-weather-api/internal/services"
)

type mockSubscriptionRepo struct {
	createErr error
}

func (m *mockSubscriptionRepo) CreateSubscription(sub models.Subscription) error {
	return m.createErr
}

func (m *mockSubscriptionRepo) ActivateSubscription(token uuid.UUID) error {
	return nil
}

func (m *mockSubscriptionRepo) DeleteSubscriptionByToken(token uuid.UUID) error {
	return nil
}

type mockMailer struct {
	sendErr error
}

func (m *mockMailer) SendConfirmationEmail(sub models.Subscription) error {
	return m.sendErr
}

func TestSubscriptionService_Subscribe(t *testing.T) {
	tests := []struct {
		name      string
		repoErr   error
		mailerErr error
		wantErr   bool
	}{
		{"success", nil, nil, false},
		{"repo error", errors.New("repo error"), nil, true},
		{"mailer error", nil, errors.New("mailer error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockSubscriptionRepo{createErr: tt.repoErr}
			mailer := &mockMailer{sendErr: tt.mailerErr}
			service := services.NewSubscriptionService(repo, mailer)

			err := service.Subscribe(services.SubscriptionInput{
				Email:     "test@example.com",
				Frequency: "daily",
				City:      "Kyiv",
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
