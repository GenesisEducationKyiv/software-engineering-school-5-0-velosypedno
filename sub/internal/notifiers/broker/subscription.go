package notifiers

import (
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
)

type subscribeEventProducer interface {
	Produce(sub domain.Subscription) error
}

type SubscriptionEventNotifier struct {
	producer subscribeEventProducer
}

func NewSubscriptionEmailNotifier(producer subscribeEventProducer) *SubscriptionEventNotifier {
	return &SubscriptionEventNotifier{
		producer: producer,
	}
}

func (m *SubscriptionEventNotifier) SendConfirmation(subscription domain.Subscription) error {
	err := m.producer.Produce(subscription)
	if err != nil {
		return fmt.Errorf("subscription notifier: %w", err)
	}
	return nil
}
