package scheduler

import (
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
)

func SetupScheduler(c *ioc.TaskContainer) *cron.Cron {
	cron := cron.New()
	cron.AddFunc("0 0 * * * *", c.HourlyWeatherNotificationTask)
	cron.AddFunc("0 0 7 * * *", c.DailyWeatherNotificationTask)
	return cron
}
