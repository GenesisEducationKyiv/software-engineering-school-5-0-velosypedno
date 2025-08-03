package services

import (
	"context"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"go.uber.org/zap"
)

type activeSubsRepo interface {
	GetActivatedByFreq(freq domain.Frequency) ([]domain.Subscription, error)
}

type weatherMailer interface {
	SendCurrent(subscription domain.Subscription, weather domain.Weather) error
}

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type WeatherNotificationService struct {
	logger        *zap.Logger
	subRepo       activeSubsRepo
	weatherMailer weatherMailer
	weatherRepo   weatherRepo
}

func NewWeatherNotificationService(
	logger *zap.Logger,
	subRepo activeSubsRepo,
	weatherMailer weatherMailer,
	weatherRepo weatherRepo,
) *WeatherNotificationService {
	return &WeatherNotificationService{
		logger:        logger.With(zap.String("service", "WeatherNotificationService")),
		subRepo:       subRepo,
		weatherMailer: weatherMailer,
		weatherRepo:   weatherRepo,
	}
}

func (s *WeatherNotificationService) SendByFreq(freq domain.Frequency) {
	subscriptions, err := s.subRepo.GetActivatedByFreq(freq)
	if err != nil {
		s.logger.Error("failed to get subscriptions", zap.Error(err))
		return
	}
	for _, sub := range subscriptions {
		weather, err := s.weatherRepo.GetCurrent(context.Background(), sub.City)
		if err != nil {
			err = fmt.Errorf("weather notification service: failed to get weather for %s, err:%v ",
				sub.City, err)
			s.logger.Error("failed to get weather", zap.Error(err), zap.String("city", sub.City))
			continue
		}
		if err := s.weatherMailer.SendCurrent(sub, weather); err != nil {
			err = fmt.Errorf(
				"weather notification service: failed to send email to %s, err:%v ",
				sub.Email, err,
			)
			s.logger.Error(
				"failed to send email",
				zap.Error(err),
				zap.String("email_hash", logging.HashEmail(sub.Email)),
			)
			continue
		}
	}
}
