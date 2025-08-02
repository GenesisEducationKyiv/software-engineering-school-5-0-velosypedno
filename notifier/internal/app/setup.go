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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	confirmSubTmplName      = "confirm_sub.html"
	weatherTmplName         = "weather.html"
	subscribeConsumerName   = "subscribe event"
	weathNotifyConsumerName = "weather notify command"
)

var declareExchangeOnce sync.Once
var declareExchangeErr error

func (a *App) setupExchange() error {
	a.logger.Debug("Declaring exchange...", zap.String("exchange", messaging.ExchangeName))
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
	if declareExchangeErr != nil {
		a.logger.Error("Failed to declare exchange", zap.Error(declareExchangeErr))
	} else {
		a.logger.Debug("Exchange declared")
	}
	return declareExchangeErr
}

func (a *App) setupSubscribeEventConsumer() (*consumers.GenericConsumer[messaging.SubscribeEvent], error) {
	a.logger.Debug("Setting up SubscribeEvent consumer...")

	if err := a.setupExchange(); err != nil {
		return nil, err
	}

	a.logger.Debug("Declaring Subscribe queue", zap.String("queue", messaging.SubscribeQueueName))
	q, err := a.rmqCh.QueueDeclare(
		messaging.SubscribeQueueName,
		true, false, false, false, nil,
	)
	if err != nil {
		a.logger.Error("Failed to declare Subscribe queue", zap.Error(err))
		return nil, err
	}

	a.logger.Debug("Binding Subscribe queue",
		zap.String("queue", q.Name),
		zap.String("routingKey", messaging.SubscribeRoutingKey),
		zap.String("exchange", messaging.ExchangeName),
	)
	if err := a.rmqCh.QueueBind(q.Name, messaging.SubscribeRoutingKey, messaging.ExchangeName, false, nil); err != nil {
		a.logger.Error("Failed to bind Subscribe queue", zap.Error(err))
		return nil, err
	}

	a.logger.Debug("Starting consumption from Subscribe queue")
	msgs, err := a.rmqCh.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		a.logger.Error("Failed to start consuming Subscribe queue", zap.Error(err))
		return nil, err
	}

	confirmTmplPath := filepath.Join(a.cfg.TemplatesDir, confirmSubTmplName)
	smtpBackend := email.NewSMTPBackend(
		a.cfg.SMTP.Host,
		a.cfg.SMTP.Port,
		a.cfg.SMTP.User,
		a.cfg.SMTP.Pass,
		a.cfg.SMTP.EmailFrom,
	)
	mailerLogger := a.logFactory.ForPackage("mailers")
	subscribeMailer := mailers.NewSubscriptionEmailNotifier(mailerLogger, smtpBackend, confirmTmplPath)
	subscribeEventHandler := eventhandlers.NewSubscribeEventHandler(subscribeMailer)
	consumerLogger := a.logFactory.ForPackage("consumers")
	subscribeEventConsumer := consumers.NewGenericConsumer(
		consumerLogger, subscribeEventHandler, msgs,
		subscribeConsumerName, a.metrics.consumers,
	)

	a.logger.Debug("SubscribeEvent consumer successfully created")
	return subscribeEventConsumer, nil
}

func (a *App) setupWeatherCommandConsumer() (*consumers.GenericConsumer[messaging.WeatherNotifyCommand], error) {
	a.logger.Debug("Setting up WeatherNotifyCommand consumer...")

	if err := a.setupExchange(); err != nil {
		return nil, err
	}

	a.logger.Debug("Declaring Weather queue", zap.String("queue", messaging.WeatherQueueName))
	q, err := a.rmqCh.QueueDeclare(
		messaging.WeatherQueueName,
		true, false, false, false, nil,
	)
	if err != nil {
		a.logger.Error("Failed to declare Weather queue", zap.Error(err))
		return nil, err
	}

	a.logger.Debug("Binding Weather queue",
		zap.String("queue", q.Name),
		zap.String("routingKey", messaging.WeatherRoutingKey),
		zap.String("exchange", messaging.ExchangeName),
	)
	if err := a.rmqCh.QueueBind(q.Name, messaging.WeatherRoutingKey, messaging.ExchangeName, false, nil); err != nil {
		a.logger.Error("Failed to bind Weather queue", zap.Error(err))
		return nil, err
	}

	a.logger.Debug("Starting consumption from Weather queue")
	msgs, err := a.rmqCh.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		a.logger.Error("Failed to start consuming Weather queue", zap.Error(err))
		return nil, err
	}

	smtpBackend := email.NewSMTPBackend(
		a.cfg.SMTP.Host,
		a.cfg.SMTP.Port,
		a.cfg.SMTP.User,
		a.cfg.SMTP.Pass,
		a.cfg.SMTP.EmailFrom,
	)
	weatherTmplPath := filepath.Join(a.cfg.TemplatesDir, weatherTmplName)
	mailerLogger := a.logFactory.ForPackage("mailers")
	weatherMailer := mailers.NewWeatherEmailNotifier(mailerLogger, smtpBackend, weatherTmplPath)
	weatherCommandHandler := eventhandlers.NewWeatherNotifyCommandHandler(weatherMailer)
	consumerLogger := a.logFactory.ForPackage("consumers")
	weatherCommandConsumer := consumers.NewGenericConsumer(
		consumerLogger, weatherCommandHandler, msgs,
		weathNotifyConsumerName, a.metrics.consumers,
	)

	a.logger.Debug("WeatherNotifyCommand consumer successfully created")
	return weatherCommandConsumer, nil
}

func (a *App) setupRouter() *gin.Engine {
	a.logger.Debug("Setting up HTTP router")
	router := gin.Default()

	queues := []string{
		messaging.SubscribeQueueName,
		messaging.WeatherQueueName,
	}

	handlerLogger := a.logFactory.ForPackage("handlers/http")
	router.GET("/healthcheck",
		httphandlers.NewHealthcheckGETHandler(
			handlerLogger,
			a.rmqCh,
			queues,
		),
	)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	a.logger.Debug("HTTP router created")
	return router
}
