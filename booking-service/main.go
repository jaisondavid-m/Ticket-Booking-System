package main

import (
	"booking-service/config"
	"booking-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Connect()
	r := gin.Default()
	routes.RegisterRoutes(r)
	r.Run(":8082")
}