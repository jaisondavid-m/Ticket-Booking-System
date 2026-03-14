package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"

	sharedkafka "shared/kafka"
)

var db *sql.DB

type InventorySyncEvent struct {
	Action          string `json:"action"` // "decrement" | "release"
	TicketID        int    `json:"ticket_id"`
	IdempotencyKey  string `json:"idempotency_key"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[inventory-consumer] no .env")
	}
	connectDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-quit; cancel() }()

	r := sharedkafka.NewConsumer(sharedkafka.TopicInventorySync, "inventory-consumer-group")
	defer r.Close()

	log.Println("[inventory-consumer] starting")
	sharedkafka.ConsumeWithWorkerPool(ctx, r, 16, processInventory)
}

func processInventory(ctx context.Context, msg kafka.Message) error {
	var event InventorySyncEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("[inventory-consumer] unmarshal error: %v", err)
		return nil
	}

	switch event.Action {
	case "decrement":
		_, err := db.ExecContext(ctx,
			"UPDATE tickets SET available = available - 1 WHERE id = ? AND available > 0",
			event.TicketID,
		)
		if err != nil {
			return err // retry
		}
	case "release":
		_, err := db.ExecContext(ctx,
			"UPDATE tickets SET available = available + 1 WHERE id = ?",
			event.TicketID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func connectDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("[inventory-consumer] no .env")
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set")
	}
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		log.Fatal("[inventory-consumer] DB ping: ", err)
	}
	log.Println("[inventory-consumer] DB connected")
}