package services

import (
	"context"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/weather/domain"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCAdapter struct {
	client pb.WeatherServiceClient
}

func NewGRPCAdapter(client pb.WeatherServiceClient) *GRPCAdapter {
	return &GRPCAdapter{client: client}
}

func (s *GRPCAdapter) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	req := pb.GetCurrentRequest{
		City: city,
	}
	resp, err := s.client.GetCurrent(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			log.Println(fmt.Errorf("grpc adapter: %v", err))
			return domain.Weather{}, fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		log.Println(fmt.Errorf("grpc adapter: %s", st.Message()))
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
