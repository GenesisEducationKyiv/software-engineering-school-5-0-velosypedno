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

type TaskContainer struct {
	HourlyWeatherNotificationTask task
	DailyWeatherNotificationTask  task
}

func BuildTaskContainer(c *config.Config) *TaskContainer {
	db, err := sql.Open(c.DbDriver, c.DbDSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WeatherAPIKey, &http.Client{})
	subRepo := repos.NewSubscriptionDBRepo(db)
	stdoutEmailBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherMailer(stdoutEmailBackend)
	weatherMailerSrv := services.NewWeatherMailerService(subRepo, weatherMailer, weatherRepo)

	return &TaskContainer{
		HourlyWeatherNotificationTask: func() {
			weatherMailerSrv.SendWeatherEmailsByFreq(models.FreqHourly)
		},
		DailyWeatherNotificationTask: func() {
			weatherMailerSrv.SendWeatherEmailsByFreq(models.FreqDaily)
		},
	}
}
