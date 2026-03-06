package routes

import (
	"gateway/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/book", handlers.BookTicket)
}