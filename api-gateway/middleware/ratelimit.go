// api-gateway/middleware/ratelimit.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	sharedcache "shared/cache"
)

// Sliding window rate limit using Redis sorted sets
// Accurate, distributed, handles burst correctly
var rateLimitScript = redis.NewScript(`
	local key     = KEYS[1]
	local now     = tonumber(ARGV[1])
	local window  = tonumber(ARGV[2])
	local limit   = tonumber(ARGV[3])
	local member  = ARGV[4]
	local clearBefore = now - window

	redis.call("ZREMRANGEBYSCORE", key, "-inf", clearBefore)
	local count = redis.call("ZCARD", key)

	if count < limit then
		redis.call("ZADD", key, now, member)
		redis.call("EXPIRE", key, math.ceil(window / 1000))
		return 1   -- allowed
	end
	return 0       -- blocked
`)

func RateLimit(requestsPerSecond int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		ip := c.ClientIP()
		key := fmt.Sprintf("rl:%s", ip)

		nowMs := time.Now().UnixMilli()
		windowMs := int64(1000) // 1 second window
		member := fmt.Sprintf("%d-%s", nowMs, uuid.New().String())

		result, err := rateLimitScript.Run(
			ctx,
			sharedcache.Client,
			[]string{key},
			nowMs,
			windowMs,
			requestsPerSecond,
			member,
		).Int()

		if err != nil || result == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}