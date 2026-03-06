package routes

import (
	"inventory-service/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine){
	api := r.Group("/api")
	api.POST("/reserve",handlers.Reserve)
}