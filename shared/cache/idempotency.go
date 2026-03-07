//prevent duplicate requests
package cache

import (
	"context"
	"time"
)

const idempotencyTTL = 24 * time.Hour

// AcquireIdempotencyKey returns true if this is a NEW request (first time seen)
// Returns false if this idempotency key was already processed
func AcquireIdempotencyKey(ctx context.Context, key string) (bool, error) {
	// SET key 1 NX EX 86400 — atomic, only sets if not exists
	ok, err := Client.SetNX(ctx, "idem:"+key, 1, idempotencyTTL).Result()
	return ok, err
}

// MarkIdempotencyComplete stores the response so retries get the cached result
func MarkIdempotencyComplete(ctx context.Context, key string, response []byte) error {
	return Client.Set(ctx, "idem:resp:"+key, response, idempotencyTTL).Err()
}

// GetIdempotencyResponse returns cached response for a duplicate request
func GetIdempotencyResponse(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := Client.Get(ctx, "idem:resp:"+key).Bytes()
	if err != nil {
		return nil, false, nil // key not found = not a duplicate
	}
	return val, true, nil
}