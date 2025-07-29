//go:build unit

package handlers_test

import (
	"context"
	"testing"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	handlers "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/handlers/grpc"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/services/subscription"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockSubService struct {
	ActivateFn    func(uuid.UUID) error
	UnsubscribeFn func(uuid.UUID) error
	SubscribeFn   func(subsrv.SubscriptionInput) error
}

func (m *mockSubService) Activate(token uuid.UUID) error {
	if m.ActivateFn != nil {
		return m.ActivateFn(token)
	}
	return nil
}

func (m *mockSubService) Unsubscribe(token uuid.UUID) error {
	if m.UnsubscribeFn != nil {
		return m.UnsubscribeFn(token)
	}
	return nil
}

func (m *mockSubService) Subscribe(sub subsrv.SubscriptionInput) error {
	if m.SubscribeFn != nil {
		return m.SubscribeFn(sub)
	}
	return nil
}
func TestSubGRPCServer_Confirm(t *testing.T) {
	validToken := uuid.New()
	validReq := &pb.ConfirmRequest{Token: validToken.String()}

	t.Run("Success", func(t *testing.T) {
		// Arrange
		srv := handlers.NewSubGRPCServer(
			&mockSubService{
				ActivateFn: func(u uuid.UUID) error {
					if u != validToken {
						t.Errorf("unexpected token: got %v, want %v", u, validToken)
					}
					return nil
				},
			},
		)

		// Act
		resp, err := srv.Confirm(context.Background(), validReq)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "successfully confirmed", resp.Message)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Arrange
		srv := handlers.NewSubGRPCServer(&mockSubService{})

		// Act
		_, err := srv.Confirm(context.Background(), &pb.ConfirmRequest{Token: "invalid"})

		// Assert
		require.Error(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("NotFound", func(t *testing.T) {
		// Arrange
		srv := handlers.NewSubGRPCServer(
			&mockSubService{
				ActivateFn: func(u uuid.UUID) error {
					return domain.ErrSubNotFound
				},
			},
		)

		// Act
		_, err := srv.Confirm(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
	})

	t.Run("InternalError", func(t *testing.T) {
		// Arrange
		srv := handlers.NewSubGRPCServer(
			&mockSubService{
				ActivateFn: func(u uuid.UUID) error {
					return domain.ErrInternal
				},
			},
		)

		// Act
		_, err := srv.Confirm(context.Background(), validReq)

		// Assert
		require.Error(t, err)
		s, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code())
	})
}
