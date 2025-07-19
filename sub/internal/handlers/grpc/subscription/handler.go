package handlers

import (
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/services/subscription"
	"github.com/google/uuid"
)

type subscriptionService interface {
	Activate(token uuid.UUID) error
	Unsubscribe(token uuid.UUID) error
	Subscribe(subInput subsrv.SubscriptionInput) error
}

type SubGRPCServer struct {
	pb.UnimplementedSubscriptionServiceServer

	subSvc subscriptionService
}

func NewSubGRPCServer(subSvc subscriptionService) *SubGRPCServer {
	return &SubGRPCServer{
		subSvc: subSvc,
	}
}
