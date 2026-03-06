package routes

import (
	"booking-service/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/bookings", handlers.CreateBooking)
}