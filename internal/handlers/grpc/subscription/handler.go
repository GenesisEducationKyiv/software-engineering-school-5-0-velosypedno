package grpc

import (
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/google/uuid"
)

type subscriptionService interface {
	Activate(token uuid.UUID) error
	Unsubscribe(token uuid.UUID) error
	Subscribe(subInput subsrv.SubscriptionInput) error
}

type SubGrpcServer struct {
	pb.UnimplementedSubscriptionServiceServer

	subSvc subscriptionService
}

func NewSubGrpcServer(subSvc subscriptionService) *SubGrpcServer {
	return &SubGrpcServer{
		subSvc: subSvc,
	}
}
