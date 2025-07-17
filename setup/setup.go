package setup

import (
	"assignment1/api"
	"assignment1/cache"
	"assignment1/config"
	"assignment1/db"
	"assignment1/middleware"
	"assignment1/provider"
	"assignment1/service"
	"github.com/gin-gonic/gin"
	"time"
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
		URL:   cfg.ExchangeURL,
		AppId: cfg.ExchangeAppId,
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

	return &App{
		Config:  cfg,
		Router:  r,
		Service: svc,
	}
}
