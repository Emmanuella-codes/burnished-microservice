package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func rateLimitMiddleware() gin.HandlerFunc {
	// set limit to 9 requests per minute
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:	9,
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	return mgin.NewMiddleware(instance)
}
