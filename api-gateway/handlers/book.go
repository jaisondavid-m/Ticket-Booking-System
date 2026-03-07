package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	sharedcache "shared/cache"
)

type BookRequest struct {
	Name           string `json:"name"`
	IdempotencyKey string `json:"idempotency_key"`
}

func BookTicket(c *gin.Context) {
	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}

	ctx := context.Background()

	// ── Idempotency check ─────────────────────────────────────────
	if req.IdempotencyKey != "" {
		if cached, found, _ := sharedcache.GetIdempotencyResponse(ctx, req.IdempotencyKey); found {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
	}

	// ── Redis atomic reserve (skip DB entirely for availability check)
	_, err := sharedcache.AtomicReserve(ctx)
	if err == sharedcache.ErrNoTickets {
		c.JSON(http.StatusConflict, gin.H{"error": "no tickets available"})
		return
	}
	if err != nil {
		// Redis down — fall back to inventory service
		if !fallbackReserve(c) {
			return
		}
	}

	// ── Create booking record ──────────────────────────────────────
	payload, _ := json.Marshal(map[string]string{"name": req.Name})
	bookResp, err := http.Post(
		bookingURL()+"/api/bookings",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil || bookResp.StatusCode != http.StatusCreated {
		// Rollback: increment Redis counter back
		sharedcache.AtomicRelease(ctx)
		// Also tell inventory service to release (keeps DB in sync)
		http.Post(inventoryURL()+"/api/release", "application/json", nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking failed, reservation rolled back"})
		return
	}
	defer bookResp.Body.Close()

	resp := gin.H{"message": "Ticket Booked Successfully"}

	// Cache idempotency response
	if req.IdempotencyKey != "" {
		respBytes, _ := json.Marshal(resp)
		sharedcache.MarkIdempotencyComplete(ctx, req.IdempotencyKey, respBytes)
	}

	c.JSON(http.StatusOK, resp)
}

// fallbackReserve hits inventory service when Redis is unavailable
func fallbackReserve(c *gin.Context) bool {
	reserveResp, err := http.Post(inventoryURL()+"/api/reserve", "application/json", nil)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "inventory service unreachable"})
		return false
	}
	defer reserveResp.Body.Close()
	if reserveResp.StatusCode == http.StatusConflict {
		c.JSON(http.StatusConflict, gin.H{"error": "no tickets available"})
		return false
	}
	if reserveResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reservation failed"})
		return false
	}
	return true
}

func inventoryURL() string {
	if v := os.Getenv("INVENTORY_URL"); v != "" {
		return v
	}
	return "http://nginx"
}

func bookingURL() string {
	if v := os.Getenv("BOOKING_URL"); v != "" {
		return v
	}
	return "http://nginx"
}