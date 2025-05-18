package services

import (
	"context"
	"fmt"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type activatedSubscriptionsRepo interface {
	GetActivatedSubscriptionsByFreq(freq models.Frequency) ([]models.Subscription, error)
}

type currentWeatherEmailService interface {
	SendWeatherEmail(subscription models.Subscription, weather models.Weather) error
}

type weatherRepo interface {
	GetCurrentWeather(ctx context.Context, city string) (models.Weather, error)
}

type WeatherMailerService struct {
	subRepo     activatedSubscriptionsRepo
	emailSrv    currentWeatherEmailService
	weatherRepo weatherRepo
}

func NewWeatherMailerService(
	subRepo activatedSubscriptionsRepo,
	emailSrv currentWeatherEmailService,
	weatherRepo weatherRepo,
) *WeatherMailerService {
	return &WeatherMailerService{
		subRepo:     subRepo,
		emailSrv:    emailSrv,
		weatherRepo: weatherRepo,
	}
}

func (s *WeatherMailerService) SendWeatherEmailsByFrequency(freq models.Frequency) {
	subscriptions, err := s.subRepo.GetActivatedSubscriptionsByFreq(freq)
	if err != nil {
		log.Println(fmt.Errorf("weather mailer service: failed to get subscriptions, err:%v ", err))
		return
	}
	for _, sub := range subscriptions {
		weather, err := s.weatherRepo.GetCurrentWeather(context.Background(), sub.City)
		if err != nil {
			log.Println(fmt.Errorf("weather mailer service: failed to get weather for %s, err:%v ", sub.City, err))
			continue
		}
		if err := s.emailSrv.SendWeatherEmail(sub, weather); err != nil {
			log.Println(fmt.Errorf("weather mailer service: failed to send email to %s, err:%v ", sub.Email, err))
			continue
		}
	}
}
