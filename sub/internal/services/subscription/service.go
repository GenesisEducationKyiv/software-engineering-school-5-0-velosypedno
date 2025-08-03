package services

import (
	"errors"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"github.com/google/uuid"
)

type metrics interface {
	IncSubscribe()
	IncSubscribeError()
	IncActivate()
	IncActivateError()
	IncUnsubscribe()
	IncUnsubscribeError()
}

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
	repo    SubscriptionRepo
	mailer  confirmationMailer
	metrics metrics
}

func NewSubscriptionService(repo SubscriptionRepo, mailer confirmationMailer, metrics metrics) *SubscriptionService {
	return &SubscriptionService{repo: repo, mailer: mailer, metrics: metrics}
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

	err := s.repo.Create(subscription)
	if errors.Is(err, domain.ErrInternal) {
		s.metrics.IncSubscribeError()
		return fmt.Errorf("subscription service: %w", err)
	} else if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	err = s.mailer.SendConfirmation(subscription)
	if errors.Is(err, domain.ErrSendEmail) {
		s.metrics.IncSubscribeError()
		return fmt.Errorf("subscription service: %w", err)
	} else if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	s.metrics.IncSubscribe()
	return nil
}

func (s *SubscriptionService) Activate(token uuid.UUID) error {
	err := s.repo.Activate(token)
	if errors.Is(err, domain.ErrInternal) {
		s.metrics.IncActivateError()
		return fmt.Errorf("subscription service: %w", err)
	} else if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	s.metrics.IncActivate()
	return nil
}

func (s *SubscriptionService) Unsubscribe(token uuid.UUID) error {
	err := s.repo.DeleteByToken(token)
	if errors.Is(err, domain.ErrInternal) {
		s.metrics.IncUnsubscribeError()
		return fmt.Errorf("subscription service: %w", err)
	} else if err != nil {
		return fmt.Errorf("subscription service: %w", err)
	}
	s.metrics.IncUnsubscribe()
	return nil
}
