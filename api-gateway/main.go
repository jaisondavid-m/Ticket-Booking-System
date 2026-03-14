package main

import (
	"log"
	"os"
	"time"

	"api-gateway/routes"

	sharedcache "shared/cache"
	sharedkafka "shared/kafka"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[API-GATEWAY] no .env file, using system env")
	}

	sharedcache.InitRedis()
	sharedkafka.InitProducer()
	defer sharedkafka.CloseProducer()

	// Create Kafka topics in background (Kafka may still be starting)
	go func() {
		for i := 0; i < 10; i++ {
			if err := sharedkafka.CreateTopics("kafka1:9092"); err == nil {
				return
			}
			time.Sleep(3 * time.Second)
		}
		log.Println("[API-GATEWAY] warn: could not create Kafka topics")
	}()

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	rl := 10000

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
