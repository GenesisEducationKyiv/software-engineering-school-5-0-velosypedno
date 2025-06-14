package services

import (
	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type SubscriptionRepo interface {
	Create(subscription models.Subscription) error
	Activate(token uuid.UUID) error
	DeleteByToken(token uuid.UUID) error
}
type confirmationMailer interface {
	SendConfirmation(subscription models.Subscription) error
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
	subscription := models.Subscription{
		ID:        uuid.New(),
		Email:     subInput.Email,
		Frequency: subInput.Frequency,
		City:      subInput.City,
		Activated: false,
		Token:     uuid.New(),
	}
	if err := s.repo.Create(subscription); err != nil {
		return err
	}
	if err := s.mailer.SendConfirmation(subscription); err != nil {
		return err
	}
	return nil
}

func (s *SubscriptionService) Activate(token uuid.UUID) error {
	return s.repo.Activate(token)
}

func (s *SubscriptionService) Unsubscribe(token uuid.UUID) error {
	return s.repo.DeleteByToken(token)
}
