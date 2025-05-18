package containers

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
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
	db, err := sql.Open(c.DB_DRIVER, c.DB_DSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WEATHER_API_KEY)
	subRepo := repos.NewSubscriptionDBRepo(db)
	emailService := services.NewDebugEmailService()

	weatherMailerSrv := services.NewWeatherMailerService(subRepo, emailService, weatherRepo)
	return &TaskContainer{
		HourlyWeatherNotificationTask: func() { weatherMailerSrv.SendWeatherEmailsByFrequency(models.FreqHourly) },
		DailyWeatherNotificationTask:  func() { weatherMailerSrv.SendWeatherEmailsByFrequency(models.FreqDaily) },
	}
}
