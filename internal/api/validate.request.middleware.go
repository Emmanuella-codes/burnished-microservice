package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func validateRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			var req ProcessCVRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
				c.Abort()
				return
			}
			c.Set("requestBody", req)
		}
		c.Next()
	}
}
