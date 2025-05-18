package scheduler

import (
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
)

func SetupScheduler(c *ioc.TaskContainer) *cron.Cron {
	cron := cron.New()
	cron.AddFunc("@every 1m", c.HourlyWeatherNotificationTask)
	cron.AddFunc("@every 2m", c.DailyWeatherNotificationTask)
	return cron
}
