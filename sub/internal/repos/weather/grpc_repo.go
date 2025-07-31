package services

import (
	"context"
	"fmt"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCRepo struct {
	logger *zap.Logger
	client pb.WeatherServiceClient
}

func NewGRPCRepo(logger *zap.Logger, client pb.WeatherServiceClient) *GRPCRepo {
	return &GRPCRepo{
		logger: logger.With(zap.String("repo", "GRPCRepo")),
		client: client,
	}
}

func (s *GRPCRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	logger := s.logger.With(
		zap.String("method", "GetCurrent"),
	)

	req := pb.GetCurrentRequest{
		City: city,
	}
	resp, err := s.client.GetCurrent(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			logger.Error("unexpected gRPC error", zap.Error(err), zap.String("city", city))
			return domain.Weather{}, fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}
		logger.Warn(
			"handled gRPC error",
			zap.Error(err),
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
