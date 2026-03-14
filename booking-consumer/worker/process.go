package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"

	sharedcache "shared/cache"
	sharedkafka "shared/kafka"
)

var db *sql.DB

type BookingEvent struct {
	IdempotencyKey string    `json:"idempotency_key"`
	Name           string    `json:"name"`
	UserID         string    `json:"user_id"`
	RequestedAt    time.Time `json:"requested_at"`
}

type BookingResult struct {
	IdempotencyKey string `json:"idempotency_key"`
	BookingID      int64  `json:"booking_id,omitempty"`
	Name           string `json:"name,omitempty"`
	Status         string `json:"status"` // "success" | "failed"
	Reason         string `json:"reason,omitempty"`
}

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Println("[worker] no .env")
	}
	dsn := os.Getenv("BOOKING_DB_DSN")
	if dsn == "" {
		log.Fatal("[worker] BOOKING_DB_DSN not set")
	}
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		log.Fatal("[worker] DB ping failed: ", err)
	}
	log.Println("[worker] booking DB connected")
}

// ProcessBooking is the Kafka message handler — called per message by worker pool
func ProcessBooking(ctx context.Context, msg kafka.Message) error {
	var event BookingEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("[worker] unmarshal error: %v — sending to DLQ", err)
		sharedkafka.PublishDLQ(ctx, msg, "unmarshal_error")
		return nil
	}

	// ── Idempotency: skip if already processed ───────────────────
	// This handles Kafka redelivery (consumer restart, etc.)
	if _, found, _ := sharedcache.GetIdempotencyResponse(ctx, event.IdempotencyKey); found {
		log.Printf("[worker] duplicate event key=%s, skipping", event.IdempotencyKey)
		return nil
	}

	// ── Write booking to DB ──────────────────────────────────────
	result, err := db.ExecContext(ctx,
		"INSERT INTO bookings (name, user_id, created_at) VALUES (?, ?, NOW())",
		event.Name,
		event.UserID,
	)

	bookingResult := BookingResult{IdempotencyKey: event.IdempotencyKey, Name: event.Name}

	if err != nil {
		log.Printf("[worker] DB insert failed key=%s err=%v", event.IdempotencyKey, err)
		// Roll back the Redis counter — booking failed
		_ = sharedcache.AtomicRelease(ctx)

		bookingResult.Status = "failed"
		bookingResult.Reason = "database error"

		// Publish failure result so client polling gets an answer
		_ = publishResult(ctx, bookingResult)
		return err // return error → message will be retried
	}

	id, _ := result.LastInsertId()
	bookingResult.BookingID = id
	bookingResult.Status = "success"

	// ── Cache result for idempotency + client polling ────────────
	respBytes, _ := json.Marshal(bookingResult)
	_ = sharedcache.MarkIdempotencyComplete(ctx, event.IdempotencyKey, respBytes)

	// ── Publish inventory sync event (decrement DB async) ────────
	_ = sharedkafka.Publish(ctx, sharedkafka.TopicInventorySync, event.IdempotencyKey, map[string]any{
		"action":          "decrement",
		"ticket_id":       1,
		"idempotency_key": event.IdempotencyKey,
	})

	log.Printf("[worker] booked id=%d key=%s", id, event.IdempotencyKey)
	return nil
}

func publishResult(ctx context.Context, r BookingResult) error {
	return sharedkafka.Publish(ctx, sharedkafka.TopicBookingResults, r.IdempotencyKey, r)
}
