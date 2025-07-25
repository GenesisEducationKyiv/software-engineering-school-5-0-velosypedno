package handlers

import (
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/mailers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
)

type subscribeMailer interface {
	SendConfirmation(subscription mailers.Subscription) error
}
type SubscribeEventHandler struct {
	Mailer subscribeMailer
}

func NewSubscribeEventHandler(mailer subscribeMailer) *SubscribeEventHandler {
	return &SubscribeEventHandler{
		Mailer: mailer,
	}
}

func (h *SubscribeEventHandler) Handle(event messaging.SubscribeEvent) error {
	sub := mailers.Subscription{
		Email: event.Email,
		Token: event.Token,
	}
	err := h.Mailer.SendConfirmation(sub)
	if err != nil {
		return fmt.Errorf("subscribe event handler: %w", err)
	}
	return nil
}
