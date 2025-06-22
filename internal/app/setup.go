package app

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/email"
	subh "github.com/velosypedno/genesis-weather-api/internal/handlers/subscription"
	weathh "github.com/velosypedno/genesis-weather-api/internal/handlers/weather"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
	subsvc "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
	weathsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather"
	weathnotsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather_notification"
)

func setupRouter(db *sql.DB, router *gin.Engine, cfg *config.Config) {
	weatherRepo := repos.NewWeatherAPIRepo(cfg.WeatherAPIKey, &http.Client{})
	weatherService := weathsvc.NewWeatherService(weatherRepo)
	subRepo := repos.NewSubscriptionDBRepo(db)
	smtpEmailBackend := email.NewSMTPBackend(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass,
		cfg.EmailFrom)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend)
	subService := subsvc.NewSubscriptionService(subRepo, subMailer)

	api := router.Group("/api")
	{
		api.GET("/weather", weathh.NewWeatherGETHandler(weatherService))
		api.POST("/subscribe", subh.NewSubscribePOSTHandler(subService))
		api.GET("/confirm/:token", subh.NewConfirmGETHandler(subService))
		api.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(subService))
	}
}

func setupCron(db *sql.DB, cron *cron.Cron, cfg *config.Config) error {
	subRepo := repos.NewSubscriptionDBRepo(db)
	weatherRepo := repos.NewWeatherAPIRepo(cfg.WeatherAPIKey, &http.Client{})
	stdoutEmailBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherMailer(stdoutEmailBackend)
	weatherMailerSrv := weathnotsvc.NewWeatherNotificationService(subRepo, weatherMailer, weatherRepo)

	_, err := cron.AddFunc("0 * * * *", func() {
		weatherMailerSrv.SendByFreq(domain.FreqHourly)
	})
	if err != nil {
		return err
	}
	_, err = cron.AddFunc("0 7 * * *", func() {
		weatherMailerSrv.SendByFreq(domain.FreqDaily)
	})
	if err != nil {
		return err
	}
	return nil
}
