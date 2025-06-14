package ioc

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/email"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	"github.com/velosypedno/genesis-weather-api/internal/models"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
	"github.com/velosypedno/genesis-weather-api/internal/services"
)

type task func()

type Tasks struct {
	HourlyWeatherNotificationTask task
	DailyWeatherNotificationTask  task
}

func NewTasks(c *config.Config) *Tasks {
	db, err := sql.Open(c.DbDriver, c.DbDSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WeatherAPIKey, &http.Client{})
	subRepo := repos.NewSubscriptionDBRepo(db)
	stdoutEmailBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherMailer(stdoutEmailBackend)
	weatherMailerSrv := services.NewWeatherNotificationService(subRepo, weatherMailer, weatherRepo)

	return &Tasks{
		HourlyWeatherNotificationTask: func() {
			weatherMailerSrv.SendByFreq(models.FreqHourly)
		},
		DailyWeatherNotificationTask: func() {
			weatherMailerSrv.SendByFreq(models.FreqDaily)
		},
	}
}
