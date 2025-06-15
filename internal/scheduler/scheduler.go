package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
)

func SetupScheduler(t *ioc.Tasks) *cron.Cron {
	cron := cron.New()
	var err error
	_, err = cron.AddFunc("@every 1m", t.HourlyWeatherNotificationTask)
	if err != nil {
		log.Fatal(err)
	}
	_, err = cron.AddFunc("0 7 * * *", t.DailyWeatherNotificationTask)
	if err != nil {
		log.Fatal(err)
	}
	return cron
}
