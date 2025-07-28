package app

import (
	"path/filepath"
	"sync"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/consumers"
	eventhandlers "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/handlers/event"
	httphandlers "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/handlers/http"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/mailers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/email"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/gin-gonic/gin"
)

const confirmSubTmplName = "confirm_sub.html"
const subscribeConsumerName = "subscribe event"
const weathNotifyConsumerName = "weather notify command"

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

func (a *App) setupSubscribeEventConsumer() (*consumers.GenericConsumer[messaging.SubscribeEvent], error) {
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
	_ = stdoutBackend
	confirmTmplPath := filepath.Join(a.cfg.TemplatesDir, confirmSubTmplName)
	subscribeMailer := mailers.NewSubscriptionEmailNotifier(smtpBackend, confirmTmplPath)
	subscribeEventHandler := eventhandlers.NewSubscribeEventHandler(subscribeMailer)
	subscribeEventConsumer := consumers.NewGenericConsumer(subscribeEventHandler, msgs, subscribeConsumerName)
	return subscribeEventConsumer, nil
}

func (a *App) setupWeatherCommandConsumer() (*consumers.GenericConsumer[messaging.WeatherNotifyCommand], error) {
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
	_ = stdoutBackend
	weatherMailer := mailers.NewWeatherEmailNotifier(smtpBackend)
	weatherCommandHandler := eventhandlers.NewWeatherNotifyCommandHandler(weatherMailer)
	weatherCommandConsumer := consumers.NewGenericConsumer(weatherCommandHandler, msgs, weathNotifyConsumerName)
	return weatherCommandConsumer, nil
}

func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/healthcheck", httphandlers.NewHealthcheckGETHandler(a.rmqCh,
		[]string{messaging.SubscribeQueueName, messaging.WeatherQueueName}))
	return router
}
