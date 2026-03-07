package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitProducer(){
	brokers := brokerAddrs()
	writer = &kafka.Writer{
		Addr:  					kafka.TCP(brokers...),
		Balancer:				&kafka.Hash{},
		BatchSize:				500,
		BatchTimeout:			5 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne, // same key → same partition (ordering per user)
		Async:                  false,            // sync write — we need the offset for tracing
		AllowAutoTopicCreation: false,
		Compression:            kafka.Snappy,
		WriteTimeout:           5 * time.Second,
	}
	log.Println("[Kafka] producer ready, brokers:", brokers)
}

func Publish(ctx context.Context,topic,key string,payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: body,
		Time:  time.Now(),
	})
}

func CloseProducer() {
	if writer != nil {
		_ = writer.Close()
	}
}

func brokerAddrs() []string {
	raw := os.Getenv("KAFKA_BROKERS")
	return strings.Split(raw, ",")
}