//go:build unit

package services_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subsvc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
)

type mockSubscriptionRepo struct {
	createErr error
}

func (m *mockSubscriptionRepo) Create(sub domain.Subscription) error {
	return m.createErr
}

func (m *mockSubscriptionRepo) Activate(token uuid.UUID) error {
	return nil
}

func (m *mockSubscriptionRepo) DeleteByToken(token uuid.UUID) error {
	return nil
}

type mockMailer struct {
	sendErr error
}

func (m *mockMailer) SendConfirmation(sub domain.Subscription) error {
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
			// Arrange
			repo := &mockSubscriptionRepo{createErr: tt.repoErr}
			mailer := &mockMailer{sendErr: tt.mailerErr}
			service := subsvc.NewSubscriptionService(repo, mailer)

			// Act
			err := service.Subscribe(subsvc.SubscriptionInput{
				Email:     "test@example.com",
				Frequency: "daily",
				City:      "Kyiv",
			})

			// Assert
			assert.Equal(t, tt.wantErr, err != nil, "Subscribe() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
