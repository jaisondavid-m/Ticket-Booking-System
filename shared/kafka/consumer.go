package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Handler func(ctx context.Context, msg kafka.Message) error

func NewConsumer(topic, groupID string) *kafka.Reader {
	brokers := brokerAddrs()
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6, // 10 MB
		// Avoid ultra-tight fetch polling that amplifies transient network jitter.
		MaxWait:        1 * time.Second,
		CommitInterval: 200 * time.Millisecond,
		ReadBackoffMin: 200 * time.Millisecond,
		ReadBackoffMax: 2 * time.Second,
		Logger:         kafka.LoggerFunc(func(msg string, args ...interface{}) {}),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})
	log.Printf("[Kafka] consumer ready topic=%s group=%s brokers=%v", topic, groupID, brokers)
	return r
}

func ConsumeWithWorkerPool(ctx context.Context, r *kafka.Reader, workers int, h Handler) {
	jobs := make(chan kafka.Message, workers*2)
	for i := 0; i < workers; i++ {
		go func() {
			for msg := range jobs {
				if err := h(ctx, msg); err != nil {
					log.Printf("[Kafka] handler error topic=%s offset=%d err=%v",
						msg.Topic, msg.Offset, err)
					continue
				}
				if err := r.CommitMessages(ctx, msg); err != nil {
					log.Printf("[Kafka] commit error: %v", err)
				}
			}
		}()
	}
	// Read loop
	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("[Kafka] fetch error: %v", err)
			continue
		}
		jobs <- msg
	}
	close(jobs)
}

// Topic Names
const (
	TopicBookingRequests = "booking.requests" // gateway → booking-consumer
	TopicBookingResults  = "booking.results"  // booking-consumer → gateway (SSE/poll)
	TopicInventorySync   = "inventory.sync"   // booking-consumer → inventory-consumer
)

// ---- Dead Letter ----

const TopicBookingDLQ = "booking.dlq"

func PublishDLQ(ctx context.Context, original kafka.Message, reason string) {
	_ = Publish(ctx, TopicBookingDLQ, string(original.Key), map[string]any{
		"original_value": string(original.Value),
		"reason":         reason,
		"timestamp":      time.Now().Unix(),
	})
}

func WorkerCount() int {
	raw := os.Getenv("KAFKA_WORKERS")
	n := 32
	if raw != "" {
		_, err := fmt.Sscanf(raw, "%d", &n)
		if err != nil {
			return 32
		}
	}
	return n
}
