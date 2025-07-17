package api

import (
	"assignment1/service"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, rs *service.RateService) {
	rateHandler := NewRateHandler(rs)
	router.GET("/rate", rateHandler.GetRate)
}
