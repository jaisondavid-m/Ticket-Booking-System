package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attach a trace ID so you can follow one request across all services
		traceID := uuid.New().String()
		c.Request.Header.Set("X-Trace-ID", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)

		start := time.Now()
		c.Next()

		log.Printf(
			"[GATEWAY] trace=%s method=%s path=%s status=%d latency=%s ip=%s",
			traceID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			c.ClientIP(),
		)
	}
}