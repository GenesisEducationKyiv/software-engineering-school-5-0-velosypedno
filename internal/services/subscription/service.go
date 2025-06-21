package services

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

var (
	ErrSubNotFound      = errors.New("subscription not found")
	ErrSubAlreadyExists = errors.New("subscription with this email already exists")
	ErrInternal         = errors.New("internal error")
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
		return handleSubRepoError(err)
	}

	err := s.mailer.SendConfirmation(subscription)
	if errors.Is(err, mailers.ErrSendEmail) {
		return ErrInternal
	} else if err != nil {
		log.Println(err)
		return ErrInternal
	}
	return nil
}

func (s *SubscriptionService) Activate(token uuid.UUID) error {
	err := s.repo.Activate(token)
	return handleSubRepoError(err)
}

func (s *SubscriptionService) Unsubscribe(token uuid.UUID) error {
	err := s.repo.DeleteByToken(token)
	return handleSubRepoError(err)
}

func handleSubRepoError(err error) error {
	switch {
	case errors.Is(err, repos.ErrTokenNotFound):
		return ErrSubNotFound
	case errors.Is(err, repos.ErrInternal):
		return ErrInternal
	case errors.Is(err, repos.ErrEmailAlreadyExists):
		return ErrSubAlreadyExists
	case err != nil:
		log.Println(err)
		return ErrInternal
	}
	return nil
}
