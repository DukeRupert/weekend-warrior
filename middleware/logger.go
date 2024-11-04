// middleware/logger.go
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Logger returns a middleware that logs HTTP requests
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Store request ID if present
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = c.GetRespHeader("X-Request-ID")
		}

		// Handle panic recovery
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Str("request_id", requestID).
					Str("method", c.Method()).
					Str("path", c.Path()).
					Int("status", fiber.StatusInternalServerError).
					Dur("latency", time.Since(start)).
					Interface("error", err).
					Msg("panic recovered")
			}
		}()

		// Process request
		err := c.Next()

		// Build log entry
		logEvent := log.Info()
		if err != nil {
			logEvent = log.Error().Err(err)
		}

		// Add request details
		logEvent.
			Str("request_id", requestID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Int("status", c.Response().StatusCode()).
			Str("user_agent", c.Get("User-Agent")).
			Dur("latency", time.Since(start)).
			Msg("request processed")

		return err
	}
}
