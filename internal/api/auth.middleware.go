package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		apiKey := os.Getenv("BURNISHED_WEB_API_KEY")
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not configured"})
			c.Abort()
			return
		}
		expectedAuth := "Bearer " + apiKey
		if auth != expectedAuth {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
