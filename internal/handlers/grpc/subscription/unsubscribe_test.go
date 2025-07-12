//go:build unit

package handlers_test

import (
	"context"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/subscription"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestSubGRPCServer_Unsubscribe(t *testing.T) {
	validToken := uuid.New()
	validReq := &pb.UnsubscribeRequest{Token: validToken.String()}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		var called bool
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			UnsubscribeFn: func(u uuid.UUID) error {
				called = true
				assert.Equal(t, validToken, u)
				return nil
			},
		})

		// Act
		resp, err := srv.Unsubscribe(context.Background(), validReq)

		// Assert
		require.NoError(t, err)
		require.True(t, called)
		assert.Equal(t, "successfully unsubscribed", resp.Message)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{})

		// Act
		_, err := srv.Unsubscribe(context.Background(), &pb.UnsubscribeRequest{Token: "invalid"})

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpcCode(err))
	})

	t.Run("NotFound", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			UnsubscribeFn: func(u uuid.UUID) error {
				return domain.ErrSubNotFound
			},
		})

		// 	Act
		_, err := srv.Unsubscribe(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, grpcCode(err))
	})

	t.Run("InternalError", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			UnsubscribeFn: func(u uuid.UUID) error {
				return domain.ErrInternal
			},
		})

		// Act
		_, err := srv.Unsubscribe(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Internal, grpcCode(err))
	})

	t.Run("UnexpectedError", func(t *testing.T) {
		// Arrange
		srv := subgrpc.NewSubGRPCServer(&mockSubService{
			UnsubscribeFn: func(u uuid.UUID) error {
				return domain.ErrInternal
			},
		})

		// Act
		_, err := srv.Unsubscribe(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		assert.Equal(t, codes.Internal, grpcCode(err))
	})
}
