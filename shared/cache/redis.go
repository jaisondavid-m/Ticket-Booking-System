package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     200,
		MinIdleConns: 50,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	initial := int64(100)
	if raw := os.Getenv("TICKETS_INITIAL_COUNT"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed >= 0 {
			initial = parsed
		}
	}
	if err := SeedTicketCount(context.Background(), initial); err != nil {
		log.Printf("[redis] ticket seed skipped: %v", err)
	}
}
