//go:build unit

package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/subscription"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcCode(err error) codes.Code {
	s, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return s.Code()
}

func TestSubGRPCServer_Subscribe(t *testing.T) {
	validReq := &pb.SubscribeRequest{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
	}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		var called bool
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			SubscribeFn: func(input subsrv.SubscriptionInput) error {
				called = true
				assert.Equal(t, validReq.Email, input.Email)
				assert.Equal(t, validReq.City, input.City)
				assert.Equal(t, validReq.Frequency, input.Frequency)
				return nil
			},
		})

		// Act
		resp, err := srv.Subscribe(context.Background(), validReq)

		// Assert
		require.NoError(t, err)
		require.True(t, called)
		assert.Equal(t, "Successfully subscribed", resp.Message)
	})

	t.Run("EmptyRequest", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{})

		// Act
		_, err := srv.Subscribe(context.Background(), &pb.SubscribeRequest{})

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpcCode(err))
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{})

		// Act
		_, err := srv.Subscribe(context.Background(), &pb.SubscribeRequest{
			Email:     "invalid-email",
			City:      "Kyiv",
			Frequency: "daily",
		})

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpcCode(err))
	})

	t.Run("InvalidFrequency", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{})

		// Act
		_, err := srv.Subscribe(context.Background(), &pb.SubscribeRequest{
			Email:     "test@example.com",
			City:      "Kyiv",
			Frequency: "weekly",
		})

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpcCode(err))
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			SubscribeFn: func(input subsrv.SubscriptionInput) error {
				return domain.ErrSubAlreadyExists
			},
		})

		// Act
		_, err := srv.Subscribe(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, grpcCode(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			SubscribeFn: func(input subsrv.SubscriptionInput) error {
				return domain.ErrInternal
			},
		})

		// Act
		_, err := srv.Subscribe(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Internal, grpcCode(err))
	})

	t.Run("UnexpectedError", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			SubscribeFn: func(input subsrv.SubscriptionInput) error {
				return errors.New("some unknown error")
			},
		})

		// Act
		_, err := srv.Subscribe(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Internal, grpcCode(err))
	})
}
