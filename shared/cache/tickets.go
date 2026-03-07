package cache

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

const TicketKey = "tickets:available"

var ErrNoTickets = errors.New("no tickets available")

// SeedTicketCount — call at startup to sync DB count into Redis
func SeedTicketCount(ctx context.Context, count int64) error {
	return Client.SetNX(ctx, TicketKey, count, 0).Err()
}

// AtomicReserve decrements counter atomically — O(1)
// Returns ErrNoTickets if count is already 0
var reserveScript = redis.NewScript(`
	local current = redis.call("GET", KEYS[1])
	if current == false then
		return redis.error_reply("KEY_MISSING")
	end
	local n = tonumber(current)
	if n <= 0 then
		return -1
	end
	return redis.call("DECRBY", KEYS[1], 1)
`)

func AtomicReserve(ctx context.Context)(int64,error){
	result, err := reserveScript.Run(ctx, Client, []string{TicketKey}).Int64()
	if err != nil {
		return 0, err
	}
	if result < 0 {
		return 0, ErrNoTickets
	}
	return result, nil
}

// AtomicRelease — compensating transaction on booking failure
func AtomicRelease(ctx context.Context) error {
	return Client.Incr(ctx, TicketKey).Err()
}

func GetAvailableCount(ctx context.Context) (int64, error) {
	return Client.Get(ctx, TicketKey).Int64()
}