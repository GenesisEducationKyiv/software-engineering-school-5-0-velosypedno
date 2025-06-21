package ioc

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/email"
	subh "github.com/velosypedno/genesis-weather-api/internal/handlers/subscription"
	weathh "github.com/velosypedno/genesis-weather-api/internal/handlers/weather"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
	subsvc "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
	weathsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather"
)

type Handlers struct {
	WeatherGETHandler     gin.HandlerFunc
	SubscribePOSTHandler  gin.HandlerFunc
	ConfirmGETHandler     gin.HandlerFunc
	UnsubscribeGETHandler gin.HandlerFunc
}

func NewHandlers(c *config.Config) *Handlers {
	db, err := sql.Open(c.DbDriver, c.DbDSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WeatherAPIKey, &http.Client{})
	weatherService := weathsvc.NewWeatherService(weatherRepo)

	subRepo := repos.NewSubscriptionDBRepo(db)
	smtpEmailBackend := email.NewSMTPBackend(c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPass, c.EmailFrom)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend)
	subService := subsvc.NewSubscriptionService(subRepo, subMailer)

	return &Handlers{
		WeatherGETHandler:     weathh.NewWeatherGETHandler(weatherService),
		SubscribePOSTHandler:  subh.NewSubscribePOSTHandler(subService),
		ConfirmGETHandler:     subh.NewConfirmGETHandler(subService),
		UnsubscribeGETHandler: subh.NewUnsubscribeGETHandler(subService),
	}
}
