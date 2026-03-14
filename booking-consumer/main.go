package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"booking-consumer/worker"
	sharedcache "shared/cache"
	sharedkafka "shared/kafka"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[booking-consumer] no .env, using system env")
	}

	// Init Redis
	sharedcache.InitRedis()

	// Init Kafka producer (for publishing results + inventory sync)
	sharedkafka.InitProducer()
	defer sharedkafka.CloseProducer()

	// Connect to booking DB
	worker.ConnectDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("[booking-consumer] shutting down...")
		cancel()
	}()

	// Start consuming — one consumer group, N worker goroutines
	r := sharedkafka.NewConsumer(sharedkafka.TopicBookingRequests, "booking-consumer-group")
	defer r.Close()

	workerCount := sharedkafka.WorkerCount()
	log.Printf("[booking-consumer] starting %d workers", workerCount)
	sharedkafka.ConsumeWithWorkerPool(ctx, r, workerCount, worker.ProcessBooking)
}