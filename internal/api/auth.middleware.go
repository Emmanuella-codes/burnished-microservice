package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
			auth := c.GetHeader("Authorization")
			expectedAuth := "Bearer " + os.Getenv("BURNISHED_WEB_API_KEY")
			if auth != expectedAuth {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
					c.Abort()
					return
			}
			c.Next()
	}
}
