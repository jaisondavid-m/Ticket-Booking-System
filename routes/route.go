package routes
import (
	"github.com/gin-gonic/gin"
	"server/handlers"
)

func RegisterRoutes(r *gin.Engine){
	api := r.Group("/api")
	api.POST("/book",handlers.BookTicket)
}