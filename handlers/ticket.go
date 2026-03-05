package handlers

import (
	"net/http"
	"server/config"
	"server/models"

	"github.com/gin-gonic/gin"
)

func BookTicket(c *gin.Context) {
	var req models.BookRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	var available int
	err := config.DB.QueryRow(
		"SELECT available FROM tickets WHERE id = 1",
	).Scan(&available)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read tickets"})
		return
	}
	if available <= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "no tickets available"})
		return
	}
	_, err = config.DB.Exec(
		"UPDATE tickets SET available = available - 1 WHERE id = 1",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed update"})
		return
	}
	_, err = config.DB.Exec(
		"INSERT INTO bookings (name) VALUES (?)",
		req.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket booked",
	})
}