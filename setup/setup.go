package setup

import (
	"assignment1/api"
	"assignment1/cache"
	"assignment1/config"
	"assignment1/db"
	ratepb "assignment1/grpc/proto"
	"assignment1/middleware"
	"assignment1/provider"
	"assignment1/service"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	Config  *config.Config
	Router  *gin.Engine
	Service *service.RateService
}

func Initialize() *App {
	cfg := config.Load()

	db.Connect(cfg.DBUrl)

	prov := &provider.OpenExchangeProvider{
		URL:     cfg.ExchangeURL,
		AppId:   cfg.ExchangeAppId,
		Adapter: &provider.OpenExchangeAdapter{},
	}

	// For in-memory:
	//c := cache.NewInMemoryCache()

	// For Redis:
	c := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	svc := &service.RateService{
		Provider:            prov,
		Expiry:              time.Duration(cfg.CacheExpiry) * time.Second,
		Cache:               c,
		BackgroundTaskTimer: cfg.BackgroundTaskTimer,
		GlobalBaseCurrency:  cfg.GlobalBaseCurrency,
	}

	r := gin.Default()
	r.Use(middleware.Logger())

	api.RegisterRoutes(r, svc)

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		ratepb.RegisterRateServiceServer(grpcServer, api.NewRateGRPCServer(svc))
		log.Printf("gRPC server listening on %s", ":"+cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	return &App{
		Config:  cfg,
		Router:  r,
		Service: svc,
	}
}
