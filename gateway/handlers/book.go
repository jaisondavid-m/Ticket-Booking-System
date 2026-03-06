package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type BookRequest struct {
	Name string `json:"name"`
}
func inventoryURL() string {
	if v := os.Getenv("INVENTORY_URL"); v != "" {
		return v
	}
	return "http://localhost:8081"
}

func bookingURL() string {
	if v := os.Getenv("BOOKING_URL"); v != "" {
		return v
	}
	return "http://localhost:8082"
}

func BookTicket(c *gin.Context){
	var req BookRequest
	if err:=c.ShouldBindJSON(&req);err!=nil || req.Name == ""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"name required"})
		return
	}
	//Reserve Ticket (atomic in inventory service)
	reserveResp,err:=http.Post(
		inventoryURL()+"/api/reserve","application/json",nil,
	)
	if err!=nil{
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unreachable"})
		return
	}
	defer reserveResp.Body.Close()

	if reserveResp.StatusCode == http.StatusConflict {
		c.JSON(http.StatusConflict, gin.H{"error": "no tickets available"})
		return
	}
	if reserveResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reservation failed"})
		return
	}

	//Create booking record
	payload, _ := json.Marshal(map[string]string{"name": req.Name})
	bookResp,err:=http.Post(
		bookingURL()+"/api/bookings",
		"application/json",
		bytes.NewBuffer(payload),
	)	
	if err!=nil || bookResp.StatusCode != http.StatusCreated {
		http.Post(inventoryURL()+"/api/release", "application/json", nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking failed, reservation rolled back"})
		return
	}
	defer bookResp.Body.Close()
	c.JSON(http.StatusOK,gin.H{"message":"Ticket Booked Successfully"})
}