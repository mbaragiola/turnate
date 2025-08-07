package middleware

import (
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware adds request timeout handling
func TimeoutMiddleware(duration time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(duration),
		timeout.WithResponse(func(c *gin.Context) {
			c.JSON(408, gin.H{
				"error":   "Request timeout",
				"message": "The request took too long to process",
			})
		}),
	)
}