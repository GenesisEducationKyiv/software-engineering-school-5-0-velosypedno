package app

import (
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	subgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/handlers/grpc/subscription"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
)

type PresentationContainer struct {
	Cron        *cron.Cron
	HTTPHandler *gin.Engine
	GRPCSrv     *grpc.Server
}

func NewPresentationContainer(businessContainer *BusinessContainer) (
	*PresentationContainer, error,
) {
	cron, err := newCron(businessContainer.WeathNotifyService)
	if err != nil {
		return nil, err
	}
	grpcSrv := newGRPCServer(businessContainer.SubService)

	return &PresentationContainer{
		Cron:    cron,
		GRPCSrv: grpcSrv,
	}, nil
}

func newCron(notifier weatherNotificationService) (*cron.Cron, error) {
	cron := cron.New()
	_, err := cron.AddFunc("0 * * * *", func() {
		notifier.SendByFreq(domain.FreqHourly)
	})
	if err != nil {
		return nil, err
	}
	_, err = cron.AddFunc("0 7 * * *", func() {
		notifier.SendByFreq(domain.FreqDaily)
	})
	if err != nil {
		return nil, err
	}
	return cron, nil
}

func newGRPCServer(subSvc subscriptionService) *grpc.Server {
	grpcServer := grpc.NewServer()
	pb.RegisterSubscriptionServiceServer(grpcServer, subgrpc.NewSubGRPCServer(subSvc))
	return grpcServer
}
