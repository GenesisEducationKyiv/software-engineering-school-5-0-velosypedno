package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
)

func SetupScheduler(c *ioc.TaskContainer) *cron.Cron {
	cron := cron.New()
	var err error
	_, err = cron.AddFunc("@every 1m", c.HourlyWeatherNotificationTask)
	if err != nil {
		log.Fatal(err)
	}
	_, err = cron.AddFunc("0 7 * * *", c.DailyWeatherNotificationTask)
	if err != nil {
		log.Fatal(err)
	}
	return cron
}
