package app

import (
	"path/filepath"
	"sync"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/consumers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/handlers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/mailers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/email"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
)

const confirmSubTmplName = "confirm_sub.html"

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
	smtpBackend := email.NewSMTPBackend(
		a.cfg.SMTP.Host,
		a.cfg.SMTP.Port,
		a.cfg.SMTP.User,
		a.cfg.SMTP.Pass,
		a.cfg.SMTP.EmailFrom,
	)
	_ = smtpBackend
	stdoutBackend := email.NewStdoutBackend()
	confirmTmplPath := filepath.Join(a.cfg.TemplatesDir, confirmSubTmplName)
	subscribeMailer := mailers.NewSubscriptionEmailNotifier(stdoutBackend, confirmTmplPath)
	subscribeEventHandler := handlers.NewSubscribeEventHandler(subscribeMailer)
	subscribeEventConsumer := consumers.NewSubscribeEventConsumer(
		subscribeEventHandler,
		msgs,
	)

	return subscribeEventConsumer, nil
}

func (a *App) setupWeatherCommandConsumer() (*consumers.WeatherCommandConsumer, error) {
	err := a.setupExchange()
	if err != nil {
		return nil, err
	}
	q, err := a.rmqCh.QueueDeclare(
		messaging.WeatherQueueName, // name
		true,                       // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		return nil, err
	}

	err = a.rmqCh.QueueBind(
		q.Name,                      // queue name
		messaging.WeatherRoutingKey, // routing key
		messaging.ExchangeName,      // exchange
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
	smtpBackend := email.NewSMTPBackend(
		a.cfg.SMTP.Host,
		a.cfg.SMTP.Port,
		a.cfg.SMTP.User,
		a.cfg.SMTP.Pass,
		a.cfg.SMTP.EmailFrom,
	)
	_ = smtpBackend
	stdoutBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherEmailNotifier(stdoutBackend)
	weatherCommandHandler := handlers.NewWeatherNotifyCommandHandler(weatherMailer)
	weatherCommandConsumer := consumers.NewWeatherCommandConsumer(
		weatherCommandHandler,
		msgs,
	)

	return weatherCommandConsumer, nil
}
