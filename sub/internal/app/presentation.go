package app

import (
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	subgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/handlers/grpc"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
)

type PresentationContainer struct {
	Cron    *cron.Cron
	HTTPSrv *http.Server
	GRPCSrv *grpc.Server
}

func NewPresentationContainer(cfg *config.Config, businessContainer *BusinessContainer, logFactory *logging.LoggerFactory) (
	*PresentationContainer, error,
) {
	cron, err := newCron(businessContainer.WeathNotifyService)
	if err != nil {
		return nil, err
	}
	grpcSrv := newGRPCServer(businessContainer.SubService, logFactory)
	httpSrv := newHTTPServer(cfg)

	return &PresentationContainer{
		Cron:    cron,
		GRPCSrv: grpcSrv,
		HTTPSrv: httpSrv,
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

func newGRPCServer(subSvc subscriptionService, logFactory *logging.LoggerFactory) *grpc.Server {
	grpcServer := grpc.NewServer()
	grpcLogger := logFactory.ForPackage("handlers/grpc")
	pb.RegisterSubscriptionServiceServer(grpcServer, subgrpc.NewSubGRPCServer(grpcLogger, subSvc))
	return grpcServer
}

func newHTTPServer(cfg *config.Config) *http.Server {
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	httpSrv := http.Server{
		Addr:        ":" + cfg.HTTPPort,
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	return &httpSrv
}
