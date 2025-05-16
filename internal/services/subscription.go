package services

import (
	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type SubscriptionRepo interface {
	CreateSubscription(subscription models.Subscription) error
}
type EmailService interface {
	SendConfirmationEmail(subscription models.Subscription) error
}
type SubscriptionInput struct {
	Email     string
	Frequency string
	City      string
}

type SubscriptionService struct {
	repo   SubscriptionRepo
	mailer EmailService
}

func NewSubscriptionService(repo SubscriptionRepo, mailer EmailService) *SubscriptionService {
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
	var err error
	err = s.repo.CreateSubscription(subscription)
	if err != nil {
		return err
	}
	err = s.mailer.SendConfirmationEmail(subscription)
	if err != nil {
		return err
	}
	return nil
}
