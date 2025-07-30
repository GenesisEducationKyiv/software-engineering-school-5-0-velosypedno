package handlers

import (
	"context"
	"errors"

	"go.uber.org/zap"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *WeathGRPCServer) GetCurrent(ctx context.Context, req *pb.GetCurrentRequest) (*pb.GetCurrentResponse, error) {
	city := req.City
	if city == "" {
		s.logger.Warn("empty city in request", zap.String("method", "GetCurrent"))
		return nil, status.Errorf(codes.InvalidArgument, "city is empty")
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.requestTimeout)
	defer cancel()

	weather, err := s.weathSvc.GetCurrent(ctxWithTimeout, city)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCityNotFound):
			s.logger.Info("city not found", zap.String("city", city), zap.Error(err))
			return nil, status.Errorf(codes.NotFound, "city not found")

		case errors.Is(err, domain.ErrInternal):
			s.logger.Error("internal error getting weather", zap.String("city", city), zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to get weather")

		case errors.Is(err, domain.ErrWeatherUnavailable):
			s.logger.Error("weather unavailable", zap.String("city", city), zap.Error(err))
			return nil, status.Errorf(codes.Unavailable, "weather unavailable")

		case errors.Is(err, domain.ErrProviderUnreliable):
			s.logger.Error("unreliable provider", zap.String("city", city), zap.Error(err))
			return nil, status.Errorf(codes.Unavailable, "weather provider is unreliable")

		default:
			s.logger.Error("unexpected error getting weather", zap.String("city", city), zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to get weather")
		}
	}

	return &pb.GetCurrentResponse{
		Temperature: float32(weather.Temperature),
		Humidity:    float32(weather.Humidity),
		Description: weather.Description,
	}, nil
}
