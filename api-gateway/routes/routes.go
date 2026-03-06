package routes

import (
	"api-gateway/handlers"
	"api-gateway/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, rl *middleware.RateLimiter) {

	r.Use(middleware.Logger())
	r.Use(middleware.RateLimit(rl))

	api := r.Group("/api")
	api.Use(middleware.Auth())
	{
		api.POST("/book", handlers.BookTicket)
	}
}
