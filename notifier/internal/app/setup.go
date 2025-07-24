package app

import (
	"sync"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/consumers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/handlers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
)

var declareExchangeOnce sync.Once
var declareExchangeErr error

func (a *App) setupExchange() error {
	declareExchangeOnce.Do(func() {
		declareExchangeErr = a.rmqCh.ExchangeDeclare(
			messaging.ExchangeName, // name
			"direct",               // type
			true,                   // durable
			false,                  // auto-deleted
			false,                  // internal
			false,                  // no-wait
			nil,                    // arguments
		)
	})
	return declareExchangeErr
}

func (a *App) setupSubscribeEventConsumer() (*consumers.SubscribeEventConsumer, error) {
	err := a.setupExchange()
	if err != nil {
		return nil, err
	}

	q, err := a.rmqCh.QueueDeclare(
		messaging.SubscribeQueueName, // name
		true,                         // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		return nil, err
	}

	err = a.rmqCh.QueueBind(
		q.Name,                        // queue name
		messaging.SubscribeRoutingKey, // routing key
		messaging.ExchangeName,        // exchange
		false,
		nil)
	if err != nil {
		return nil, err
	}

	msgs, err := a.rmqCh.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}

	subscribeEventConsumer := consumers.NewSubscribeEventConsumer(
		handlers.NewSubscribeEventHandler(),
		msgs,
	)

	return subscribeEventConsumer, nil
}
