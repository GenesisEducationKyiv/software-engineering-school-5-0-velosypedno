//go:build unit

package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/services"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockClient struct {
	subscribeFn   func(ctx context.Context, in *pb.SubscribeRequest) (*pb.SubscribeResponse, error)
	confirmFn     func(ctx context.Context, in *pb.ConfirmRequest) (*pb.ConfirmResponse, error)
	unsubscribeFn func(ctx context.Context, in *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error)
}

func (m *mockClient) Subscribe(ctx context.Context, in *pb.SubscribeRequest, opts ...grpc.CallOption) (*pb.SubscribeResponse, error) {
	return m.subscribeFn(ctx, in)
}
func (m *mockClient) Confirm(ctx context.Context, in *pb.ConfirmRequest, opts ...grpc.CallOption) (*pb.ConfirmResponse, error) {
	return m.confirmFn(ctx, in)
}
func (m *mockClient) Unsubscribe(ctx context.Context, in *pb.UnsubscribeRequest, opts ...grpc.CallOption) (*pb.UnsubscribeResponse, error) {
	return m.unsubscribeFn(ctx, in)
}

func TestGRPCAdapter_Subscribe(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			subscribeFn: func(ctx context.Context, in *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
				require.Equal(t, "test@example.com", in.Email)
				return &pb.SubscribeResponse{}, nil
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Subscribe(services.SubscriptionInput{
			Email:     "test@example.com",
			Frequency: "daily",
			City:      "Kyiv",
		})

		// Assert
		assert.NoError(t, err)
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			subscribeFn: func(ctx context.Context, in *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
				return nil, status.Error(codes.AlreadyExists, "already exists")
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Subscribe(services.SubscriptionInput{})

		// Assert
		assert.ErrorIs(t, err, domain.ErrSubAlreadyExists)
	})

	t.Run("InvalidArgument", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			subscribeFn: func(ctx context.Context, in *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid")
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Subscribe(services.SubscriptionInput{})

		// Assert
		assert.ErrorIs(t, err, domain.ErrSubInvalid)
	})

	t.Run("RawError", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			subscribeFn: func(ctx context.Context, in *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
				return nil, errors.New("connection lost")
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Subscribe(services.SubscriptionInput{})

		// Assert
		assert.ErrorIs(t, err, domain.ErrInternal)
	})
}

func TestGRPCAdapter_Activate(t *testing.T) {
	validToken := uuid.New()

	t.Run("Success", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			confirmFn: func(ctx context.Context, in *pb.ConfirmRequest) (*pb.ConfirmResponse, error) {
				require.Equal(t, validToken.String(), in.Token)
				return &pb.ConfirmResponse{}, nil
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Activate(validToken)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			confirmFn: func(ctx context.Context, in *pb.ConfirmRequest) (*pb.ConfirmResponse, error) {
				return nil, status.Error(codes.NotFound, "not found")
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Activate(validToken)

		// Assert
		assert.ErrorIs(t, err, domain.ErrSubNotFound)
	})
}

func TestGRPCAdapter_Unsubscribe(t *testing.T) {
	validToken := uuid.New()

	t.Run("Success", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			unsubscribeFn: func(ctx context.Context, in *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
				require.Equal(t, validToken.String(), in.Token)
				return &pb.UnsubscribeResponse{}, nil
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Unsubscribe(validToken)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("InternalError", func(t *testing.T) {
		// Arrange
		client := &mockClient{
			unsubscribeFn: func(ctx context.Context, in *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
				return nil, status.Error(codes.Internal, "internal")
			},
		}
		adapter := services.NewGRPCAdapter(zap.NewNop(), client)

		// Act
		err := adapter.Unsubscribe(validToken)

		// Assert
		assert.ErrorIs(t, err, domain.ErrInternal)
	})
}
