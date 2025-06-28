package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type SubscriptionRepo interface {
	Create(subscription domain.Subscription) error
	Activate(token uuid.UUID) error
	DeleteByToken(token uuid.UUID) error
}
type confirmationMailer interface {
	SendConfirmation(subscription domain.Subscription) error
}
type SubscriptionInput struct {
	Email     string
	Frequency string
	City      string
}

type SubscriptionService struct {
	repo   SubscriptionRepo
	mailer confirmationMailer
}

func NewSubscriptionService(repo SubscriptionRepo, mailer confirmationMailer) *SubscriptionService {
	return &SubscriptionService{repo: repo, mailer: mailer}
}

func (s *SubscriptionService) Subscribe(subInput SubscriptionInput) error {
	subscription := domain.Subscription{
		ID:        uuid.New(),
		Email:     subInput.Email,
		Frequency: subInput.Frequency,
		City:      subInput.City,
		Activated: false,
		Token:     uuid.New(),
	}
	if err := s.repo.Create(subscription); err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	if err := s.mailer.SendConfirmation(subscription); err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	return nil
}

func (s *SubscriptionService) Activate(token uuid.UUID) error {
	err := s.repo.Activate(token)
	if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	return nil
}

func (s *SubscriptionService) Unsubscribe(token uuid.UUID) error {
	err := s.repo.DeleteByToken(token)
	if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	return nil
}
