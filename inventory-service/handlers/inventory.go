package handlers

import (
	"inventory-service/config"
	"net/http"

	"github.com/gin-gonic/gin"
)


//Reserve does atomic check + decrement in one query
func Reserve(c *gin.Context){
	result,err := config.DB.Exec("UPDATE tickets SET available = available - 1 WHERE id = 1 AND available > 0",)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"DB error"})	
		return
	}
	rows,_ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "no tickets available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reserved": true})
}

// Release is the compensating transaction — called by gateway if booking fails
func Release(c *gin.Context){
	_,err := config.DB.Exec("UPDATE tickets SET available = available + 1 WHERE id = 1")
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "release failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"released": true})
}