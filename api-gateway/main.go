package main

import (
	"log"
	"os"

	"api-gateway/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[API-GATEWAY] no .env file, using system env")
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	rl := 10

	r := gin.New()

	routes.RegisterRoutes(r, rl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("[API-GATEWAY] listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("[API-GATEWAY] failed to start: %v", err)
	}
}
