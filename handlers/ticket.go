package handlers

import (
	"net/http"

	"server/config"
	"server/models"

	"github.com/gin-gonic/gin"
)

func BookTicket(c *gin.Context){
	var req models.BookRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	db , err := config.DB.Begin()
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"failed to start transaction"})
		return
	}
	var available int
	err = db.QueryRow(
		"SELECT available FROM tickets WHERE id = 1 FOR UPDATE",
	).Scan(&available)

	if err != nil {
		db.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read tickets"})
		return
	}
	if available <= 0 {
		db.Rollback()
		c.JSON(http.StatusBadRequest,gin.H{"error":"no tickets available"})
		return
	}
	_, err = db.Exec("UPDATE tickets SET available = available - 1 WHERE id = 1",)
	if err!=nil{
		db.Rollback()
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to update tickets"})
		return
	}
	_,err = db.Exec("INSERT INTO bookings (name) VALUES (?)",req.Name)
	if err!=nil{
		db.Rollback()
		c.JSON(http.StatusInternalServerError,gin.H{"error":"failed to book"})
		return
	}
	if err := db.Commit(); err!=nil{
		c.JSON(http.StatusOK,gin.H{"error":"failed to commit transaction"})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"message":"Ticket Booked Successfully",
	})

}