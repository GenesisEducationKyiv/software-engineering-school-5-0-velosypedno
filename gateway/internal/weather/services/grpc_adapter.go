package services

import (
	"context"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/weather/domain"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCAdapter struct {
	client pb.WeatherServiceClient
	logger *zap.Logger
}

func NewGRPCAdapter(logger *zap.Logger, client pb.WeatherServiceClient) *GRPCAdapter {
	return &GRPCAdapter{
		client: client,
		logger: logger,
	}
}

func (s *GRPCAdapter) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	req := pb.GetCurrentRequest{
		City: city,
	}

	resp, err := s.client.GetCurrent(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			s.logger.Error("unexpected gRPC error",
				zap.String("city", city),
				zap.Error(err),
			)
			return domain.Weather{}, fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		s.logger.Warn("handled gRPC error",
			zap.String("city", city),
			zap.String("grpc_message", st.Message()),
			zap.String("grpc_code", st.Code().String()),
		)

		return domain.Weather{}, fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}

	return domain.Weather{
		Humidity:    float64(resp.Humidity),
		Temperature: float64(resp.Temperature),
		Description: resp.Description,
	}, nil
}

func gRPCToDomainError(code codes.Code) error {
	switch code {
	case codes.NotFound:
		return domain.ErrCityNotFound
	case codes.Internal:
		return domain.ErrInternal
	case codes.Unavailable:
		return domain.ErrWeatherUnavailable
	default:
		return domain.ErrInternal
	}
}
