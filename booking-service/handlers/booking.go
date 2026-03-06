package handlers

import (
	"booking-service/config"
	"booking-service/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateBooking(c *gin.Context){
	var req models.CreateBookingRequest
	if err:=c.ShouldBindJSON(&req);err!=nil || req.Name==""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"name Required"})
		return
	}
	result , err := config.DB.Exec("INSERT INTO bookings (name) VALUES (?)",req.Name)
	if err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking insert failed"})
		return
	}
	id,_:=result.LastInsertId()
	c.JSON(http.StatusCreated,gin.H{
		"booking_id":id,
		"name":req.Name,
	})
}