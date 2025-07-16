package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Incoming: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
		log.Printf("Status: %d", c.Writer.Status())
	}
}
