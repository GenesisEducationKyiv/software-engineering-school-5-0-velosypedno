package services

import (
	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type SubscriptionRepo interface {
	CreateSubscription(subscription models.Subscription) error
	ActivateSubscription(token uuid.UUID) error
	DeleteSubscriptionByToken(token uuid.UUID) error
}
type confirmationEmailService interface {
	SendConfirmationEmail(subscription models.Subscription) error
}
type SubscriptionInput struct {
	Email     string
	Frequency string
	City      string
}

type SubscriptionService struct {
	repo   SubscriptionRepo
	mailer confirmationEmailService
}

func NewSubscriptionService(repo SubscriptionRepo, mailer confirmationEmailService) *SubscriptionService {
	return &SubscriptionService{repo: repo, mailer: mailer}
}

func (s *SubscriptionService) Subscribe(subInput SubscriptionInput) error {
	subscription := models.Subscription{
		ID:        uuid.New(),
		Email:     subInput.Email,
		Frequency: subInput.Frequency,
		City:      subInput.City,
		Activated: false,
		Token:     uuid.New(),
	}
	if err := s.repo.CreateSubscription(subscription); err != nil {
		return err
	}
	if err := s.mailer.SendConfirmationEmail(subscription); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) ActivateSubscription(token uuid.UUID) error {
	return s.repo.ActivateSubscription(token)
}

func (s *SubscriptionService) Unsubscribe(token uuid.UUID) error {
	return s.repo.DeleteSubscriptionByToken(token)
}
