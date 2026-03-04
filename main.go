package main

import (
	"github.com/gin-gonic/gin"
	"server/config"
	"server/routes"
)

func main(){
	config.Connect()
	r := gin.Default()
	routes.RegisterRoutes(r)
	r.Run(":8000")
}