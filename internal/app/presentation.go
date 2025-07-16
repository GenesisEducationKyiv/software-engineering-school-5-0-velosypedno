package app

import (
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/subscription"
	weathgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/weather"
	subhttp "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/http/subscription"
	weathhttp "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/http/weather"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
)

const weatherRequestTimeout = 5 * time.Second

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
	httpHandler := newHTTPHandler(businessContainer.WeathService, businessContainer.SubService)
	grpcSrv := newGRPCServer(businessContainer.WeathService, businessContainer.SubService)

	return &PresentationContainer{
		Cron:        cron,
		HTTPHandler: httpHandler,
		GRPCSrv:     grpcSrv,
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

func newHTTPHandler(weathSvc weatherService, subSvc subscriptionService) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/weather", weathhttp.NewWeatherGETHandler(weathSvc, weatherRequestTimeout))
		api.POST("/subscribe", subhttp.NewSubscribePOSTHandler(subSvc))
		api.GET("/confirm/:token", subhttp.NewConfirmGETHandler(subSvc))
		api.GET("/unsubscribe/:token", subhttp.NewUnsubscribeGETHandler(subSvc))
	}
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return router
}

func newGRPCServer(weathSvc weatherService, subSvc subscriptionService) *grpc.Server {
	grpcServer := grpc.NewServer()
	pb.RegisterSubscriptionServiceServer(grpcServer, subgrpc.NewSubGRPCServer(subSvc))
	pb.RegisterWeatherServiceServer(grpcServer, weathgrpc.NewWeatherGRPCServer(weathSvc, weatherRequestTimeout))
	return grpcServer
}
