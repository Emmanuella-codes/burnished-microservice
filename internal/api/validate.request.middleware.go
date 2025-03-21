package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func validateRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		// Check content type to determine binding method.
		contentType := c.ContentType()
		if contentType == "application/json" {
				var req ProcessCVRequest
				if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
								"error": "Invalid JSON payload: " + err.Error(),
						})
						c.Abort()
						return
				}
				c.Set("requestBody", req)
		} else if contentType == "multipart/form-data" {
				// for form data, ensure the file is present
				// detailed validation happens in handlers.
				if _, _, err := c.Request.FormFile("cv"); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
								"error": "No CV file provided",
						})
						c.Abort()
						return
				}
		} else {
				c.JSON(http.StatusBadRequest, gin.H{
						"error": "Unsupported content type: " + contentType,
				})
				c.Abort()
				return
		}
			c.Next()
	}
}
