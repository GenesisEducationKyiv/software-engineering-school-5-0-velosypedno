//go:build integration

package api_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/test/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestGetCurrentWeatherGRPCHandler(main *testing.T) {
	// setup fake weather APIs
	freeWeatherAPI := mock.NewFreeWeatherAPI()
	defer freeWeatherAPI.Close()
	tomorrowAPI := mock.NewTomorrowAPI()
	defer tomorrowAPI.Close()
	vcAPI := mock.NewVisualCrossingAPI()
	defer vcAPI.Close()

	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg.FreeWeather.URL = freeWeatherAPI.URL
	cfg.TomorrowWeather.URL = tomorrowAPI.URL
	cfg.VisualCrossing.URL = vcAPI.URL
	fmt.Println(cfg)

	// start App
	logFactory := logging.NewFakeFactory()
	a := app.New(cfg, logFactory)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.Run(ctx)

	// setup grpc client
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(cfg.GRPCSrv.Addr(), opt)
	require.NoError(main, err)
	weathClient := pb.NewWeatherServiceClient(conn)

	main.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		req := &pb.GetCurrentRequest{
			City: "Kyiv",
		}

		resp, err := weathClient.GetCurrent(ctx, req)
		require.NoError(t, err, "Expected no error for valid city")
		require.NotNil(t, resp, "Expected non-nil response")
	})

	main.Run("InvalidCity", func(t *testing.T) {
		ctx := context.Background()

		req := &pb.GetCurrentRequest{
			City: mock.CityDoesNotExist,
		}

		resp, err := weathClient.GetCurrent(ctx, req)
		require.Error(t, err, "Expected error for invalid city")
		require.Nil(t, resp, "Expected nil response for invalid city")

		st, ok := status.FromError(err)
		require.True(t, ok, "Expected gRPC status error")
		require.Equal(t, codes.NotFound, st.Code(), "Expected NotFound status code")
	})
}
