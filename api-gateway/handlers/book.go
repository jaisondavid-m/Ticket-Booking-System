package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	sharedcache "shared/cache"
	sharedkafka "shared/kafka"
)

type BookRequest struct {
	Name           string `json:"name"`
	IdempotencyKey string `json:"idempotency_key"`
}

// BookingEvent is what goes onto Kafka
type BookingEvent struct {
	IdempotencyKey string    `json:"idempotency_key"`
	Name           string    `json:"name"`
	UserID         string    `json:"user_id"`
	RequestedAt    time.Time `json:"requested_at"`
}

func BookTicket(c *gin.Context) {
	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}

	ctx := context.Background()

	// ── Idempotency: return cached result for duplicate requests ──
	if req.IdempotencyKey != "" {
		if cached, found, _ := sharedcache.GetIdempotencyResponse(ctx, req.IdempotencyKey); found {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
	}

	// ── Redis atomic reserve — fast gate ──────────────────
	// This is the ONLY availability check. Kafka consumer trusts this.
	_, err := sharedcache.AtomicReserve(ctx)
	if err == sharedcache.ErrNoTickets {
		c.JSON(http.StatusConflict, gin.H{"error": "no tickets available"})
		return
	}
	if err != nil {
		// Redis down — fail safe (don't oversell)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
		return
	}

	// ── Publish to Kafka — fire and forget from client perspective ─
	userID := c.Request.Header.Get("X-User-ID")
	idempKey := req.IdempotencyKey
	if idempKey == "" {
		idempKey = userID + ":" + time.Now().Format(time.RFC3339Nano)
	}

	event := BookingEvent{
		IdempotencyKey: idempKey,
		Name:           req.Name,
		UserID:         userID,
		RequestedAt:    time.Now(),
	}

	// Key by idempotency key → same user's requests go to same partition (ordering)
	if err := sharedkafka.Publish(ctx, sharedkafka.TopicBookingRequests, idempKey, event); err != nil {
		// Kafka publish failed — roll back the Redis decrement
		sharedcache.AtomicRelease(ctx)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue booking"})
		return
	}

	// Return 202 Accepted — client polls /api/booking/status/:key for result
	c.JSON(http.StatusAccepted, gin.H{
		"message":         "booking queued",
		"idempotency_key": idempKey,
		"poll_url":        "/api/booking/status/" + idempKey,
	})
}

// BookingStatus — client polls this after receiving 202
func BookingStatus(c *gin.Context) {
	key := c.Param("key")
	ctx := context.Background()

	cached, found, _ := sharedcache.GetIdempotencyResponse(ctx, key)
	if !found {
		c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
		return
	}
	
	var result map[string]any
	_ = json.Unmarshal(cached, &result)
	c.JSON(http.StatusOK, result)
}