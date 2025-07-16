package handlers

import (
	"assignment1/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterRoutes(router *gin.Engine, rs *service.RateService) {
	router.GET("/rate/:pair", func(c *gin.Context) {
		pair := c.Param("pair")
		data, err := rs.GetRate(pair)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rate not found or fetch failed"})
			return
		}
		c.JSON(http.StatusOK, data)
	})
}
