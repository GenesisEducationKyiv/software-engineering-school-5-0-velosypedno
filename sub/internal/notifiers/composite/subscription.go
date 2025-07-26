package composite

import "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"

type SubscriptionNotifier interface {
	SendConfirmation(subscription domain.Subscription) error
}

type SubscriptionCompositeNotifier struct {
	Notifiers []SubscriptionNotifier
}

func NewSubscriptionCompositeNotifier(notifiers ...SubscriptionNotifier) *SubscriptionCompositeNotifier {
	return &SubscriptionCompositeNotifier{
		Notifiers: notifiers,
	}
}

func (m *SubscriptionCompositeNotifier) SendConfirmation(subscription domain.Subscription) error {
	for _, notifier := range m.Notifiers {
		err := notifier.SendConfirmation(subscription)
		if err != nil {
			return err
		}
	}
	return nil
}
