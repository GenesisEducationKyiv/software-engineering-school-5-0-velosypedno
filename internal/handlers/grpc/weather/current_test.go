package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	weathgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/weather"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockWeatherService struct {
	GetCurrentFn func(ctx context.Context, city string) (domain.Weather, error)
}

func (m *mockWeatherService) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	if m.GetCurrentFn != nil {
		return m.GetCurrentFn(ctx, city)
	}
	return domain.Weather{}, nil
}

func grpcCode(err error) codes.Code {
	s, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return s.Code()
}

func TestWeatherGRPCServer_GetCurrent(t *testing.T) {
	city := "Kyiv"
	validReq := &pb.GetCurrentRequest{City: city}
	expectedWeather := domain.Weather{
		Temperature: 21.3,
		Humidity:    50.0,
		Description: "clear",
	}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		srv := weathgrpc.NewWeatherGRPCServer(&mockWeatherService{
			GetCurrentFn: func(ctx context.Context, c string) (domain.Weather, error) {
				require.Equal(t, city, c)
				return expectedWeather, nil
			},
		}, 2*time.Millisecond)

		// Act
		resp, err := srv.GetCurrent(context.Background(), validReq)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, float32(expectedWeather.Temperature), resp.Temperature)
		assert.Equal(t, float32(expectedWeather.Humidity), resp.Humidity)
		assert.Equal(t, expectedWeather.Description, resp.Description)
	})

	t.Run("CityNotFound", func(t *testing.T) {
		// Arrange
		srv := weathgrpc.NewWeatherGRPCServer(&mockWeatherService{
			GetCurrentFn: func(ctx context.Context, c string) (domain.Weather, error) {
				return domain.Weather{}, domain.ErrCityNotFound
			},
		}, 2*time.Millisecond)

		// Act
		_, err := srv.GetCurrent(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, grpcCode(err))
	})

	t.Run("WeatherUnavailable", func(t *testing.T) {
		// Arrange
		srv := weathgrpc.NewWeatherGRPCServer(&mockWeatherService{
			GetCurrentFn: func(ctx context.Context, c string) (domain.Weather, error) {
				return domain.Weather{}, domain.ErrWeatherUnavailable
			},
		}, 2*time.Millisecond)

		// Act
		_, err := srv.GetCurrent(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Unavailable, grpcCode(err))
	})

	t.Run("ProviderUnreliable", func(t *testing.T) {
		// Arrange
		srv := weathgrpc.NewWeatherGRPCServer(&mockWeatherService{
			GetCurrentFn: func(ctx context.Context, c string) (domain.Weather, error) {
				return domain.Weather{}, domain.ErrProviderUnreliable
			},
		}, 2*time.Millisecond)

		// Act
		_, err := srv.GetCurrent(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Unavailable, grpcCode(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		// Arrange
		srv := weathgrpc.NewWeatherGRPCServer(&mockWeatherService{
			GetCurrentFn: func(ctx context.Context, c string) (domain.Weather, error) {
				return domain.Weather{}, domain.ErrInternal
			},
		}, 2*time.Millisecond)

		// Act
		_, err := srv.GetCurrent(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Internal, grpcCode(err))
	})

	t.Run("UnknownError", func(t *testing.T) {
		srv := weathgrpc.NewWeatherGRPCServer(&mockWeatherService{
			GetCurrentFn: func(ctx context.Context, c string) (domain.Weather, error) {
				return domain.Weather{}, errors.New("unknown failure")
			},
		}, 2*time.Millisecond)

		_, err := srv.GetCurrent(context.Background(), validReq)
		require.Error(t, err)
		assert.Equal(t, codes.Internal, grpcCode(err))
	})
}
