package app

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
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

const confirmSubTmplName = "confirm_sub.html"

func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()
	weatherRepo := repos.NewWeatherAPIRepo(a.cfg.WeatherAPIKey, a.cfg.WeatherAPIBaseURL, &http.Client{})
	weatherService := weathsvc.NewWeatherService(weatherRepo)
	subRepo := repos.NewSubscriptionDBRepo(a.db)
	smtpEmailBackend := email.NewSMTPBackend(a.cfg.SMTPHost, a.cfg.SMTPPort, a.cfg.SMTPUser, a.cfg.SMTPPass,
		a.cfg.EmailFrom)
	confirmTmplPath := filepath.Join(a.cfg.TemplatesDir, confirmSubTmplName)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend, confirmTmplPath)
	subService := subsvc.NewSubscriptionService(subRepo, subMailer)

	api := router.Group("/api")
	{
		api.GET("/weather", weathh.NewWeatherGETHandler(weatherService))
		api.POST("/subscribe", subh.NewSubscribePOSTHandler(subService))
		api.GET("/confirm/:token", subh.NewConfirmGETHandler(subService))
		api.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(subService))
	}
	return router
}

func (a *App) setupCron() error {
	a.cron = cron.New()
	subRepo := repos.NewSubscriptionDBRepo(a.db)
	weatherRepo := repos.NewWeatherAPIRepo(a.cfg.WeatherAPIKey, a.cfg.WeatherAPIBaseURL, &http.Client{})
	stdoutEmailBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherMailer(stdoutEmailBackend)
	weatherMailerSrv := weathnotsvc.NewWeatherNotificationService(subRepo, weatherMailer, weatherRepo)

	_, err := a.cron.AddFunc("0 * * * *", func() {
		weatherMailerSrv.SendByFreq(domain.FreqHourly)
	})
	if err != nil {
		return err
	}
	_, err = a.cron.AddFunc("0 7 * * *", func() {
		weatherMailerSrv.SendByFreq(domain.FreqDaily)
	})
	if err != nil {
		return err
	}
	return nil
}
