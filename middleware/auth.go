// middleware/auth.go
package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware contains the dependencies for the auth middleware
type AuthMiddleware struct {
	db      *pgxpool.Pool
	store   *session.Store
	options SessionOptions
	logger  zerolog.Logger
}

// SessionOptions contains configuration for the session middleware
type SessionOptions struct {
	CookieName     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
	Expiration     time.Duration
}

// DefaultSessionOptions returns default session configuration
func DefaultSessionOptions() SessionOptions {
	return SessionOptions{
		CookieName:     "session_id",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
	}
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware(db *pgxpool.Pool, options SessionOptions) (*AuthMiddleware, error) {
	// Create session store
	store := session.New(session.Config{
		KeyLookup:      "cookie:" + options.CookieName,
		CookieSecure:   options.CookieSecure,
		CookieHTTPOnly: options.CookieHTTPOnly,
		Expiration:     options.Expiration,
	})

	// Create tables if they don't exist
	if err := createSessionTable(db); err != nil {
		return nil, err
	}

	// Initialize logger
	logger := log.With().Str("middleware", "auth").Logger()

	return &AuthMiddleware{
		db:      db,
		store:   store,
		options: options,
		logger:  logger,
	}, nil
}

// createSessionTable creates the sessions table if it doesn't exist
func createSessionTable(db *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(64) PRIMARY KEY,
			data BYTEA NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`
	_, err := db.Exec(context.Background(), query)
	return err
}

// Protected middleware checks if the request has a valid session
func (am *AuthMiddleware) Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()
		path := c.Path()
		method := c.Method()

		// Create logger with request context
		reqLogger := am.logger.With().
			Str("path", path).
			Str("method", method).
			Str("ip", c.IP()).
			Str("request_id", c.Get("X-Request-ID")).
			Logger()

		sess, err := am.store.Get(c)
		if err != nil {
			reqLogger.Error().
				Err(err).
				Msg("Failed to get session")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		// Check if user data exists in session
		userID := sess.Get("user_id")
		if userID == nil {
			reqLogger.Warn().
				Msg("No user ID in session, redirecting to login")

			return c.Redirect("/login", fiber.StatusFound)
		}

		facilityID := sess.Get("facility_id")
		isAdmin := sess.Get("is_admin")

		// Add user data to context for use in handlers
		c.Locals("user_id", userID)
		c.Locals("facility_id", facilityID)
		c.Locals("is_admin", isAdmin)

		// Process the request
		err = c.Next()

		// Log the completed request
		logEvent := reqLogger.Info().
			Int("user_id", userID.(int)).
			Int("facility_id", facilityID.(int)).
			Bool("is_admin", isAdmin.(bool)).
			Int("status_code", c.Response().StatusCode()).
			Dur("duration_ms", time.Since(startTime))

		if err != nil {
			logEvent.Err(err)
		}

		logEvent.Msg("Protected route accessed")

		return err
	}
}

// AdminOnly middleware checks if the user is an admin with logging
func (am *AuthMiddleware) AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqLogger := am.logger.With().
			Str("path", c.Path()).
			Str("method", c.Method()).
			Str("ip", c.IP()).
			Str("request_id", c.Get("X-Request-ID")).
			Logger()

		sess, err := am.store.Get(c)
		if err != nil {
			reqLogger.Error().
				Err(err).
				Msg("Failed to get session in admin check")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		isAdmin := sess.Get("is_admin")
		userID := sess.Get("user_id")

		if isAdmin == nil || !isAdmin.(bool) {
			reqLogger.Warn().
				Interface("user_id", userID).
				Msg("Non-admin user attempted to access admin route")

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		reqLogger.Info().
			Interface("user_id", userID).
			Msg("Admin route accessed")

		return c.Next()
	}
}

// CleanupSessions removes expired sessions from the database with logging
func (am *AuthMiddleware) CleanupSessions() error {
	startTime := time.Now()

	result, err := am.db.Exec(context.Background(),
		"DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP")
	if err != nil {
		am.logger.Error().
			Err(err).
			Msg("Failed to cleanup expired sessions")
		return err
	}

	rowsAffected := result.RowsAffected()
	am.logger.Info().
		Int64("sessions_removed", rowsAffected).
		Dur("duration_ms", time.Since(startTime)).
		Msg("Expired sessions cleaned up")

	return nil
}

// Login creates a new session for authenticated users with logging
func (am *AuthMiddleware) Login(c *fiber.Ctx, userID int, facilityID int, isAdmin bool) error {
	reqLogger := am.logger.With().
		Str("ip", c.IP()).
		Int("user_id", userID).
		Int("facility_id", facilityID).
		Bool("is_admin", isAdmin).
		Logger()

	sess, err := am.store.Get(c)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("Failed to create session during login")
		return err
	}

	// Store user data in session
	sess.Set("user_id", userID)
	sess.Set("facility_id", facilityID)
	sess.Set("is_admin", isAdmin)

	// Save session
	if err := sess.Save(); err != nil {
		reqLogger.Error().
			Err(err).
			Msg("Failed to save session during login")
		return err
	}

	// Store session in database
	sessionID := sess.ID()
	expiresAt := time.Now().Add(am.options.Expiration)

	_, err = am.db.Exec(context.Background(),
		"INSERT INTO sessions (id, data, expires_at) VALUES ($1, $2, $3)",
		sessionID, sess.Get("data"), expiresAt)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("Failed to store session in database")
		return err
	}

	reqLogger.Info().
		Str("session_id", sessionID).
		Time("expires_at", expiresAt).
		Msg("User logged in successfully")

	return nil
}

// Logout removes the session with logging
func (am *AuthMiddleware) Logout(c *fiber.Ctx) error {
	sess, err := am.store.Get(c)
	if err != nil {
		am.logger.Error().
			Err(err).
			Str("ip", c.IP()).
			Msg("Failed to get session during logout")
		return err
	}

	userID := sess.Get("user_id")
	sessionID := sess.ID()

	reqLogger := am.logger.With().
		Str("ip", c.IP()).
		Str("session_id", sessionID).
		Interface("user_id", userID).
		Logger()

	// Delete session from database
	_, err = am.db.Exec(context.Background(), "DELETE FROM sessions WHERE id = $1", sessionID)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("Failed to delete session from database during logout")
		return err
	}

	// Destroy session
	if err := sess.Destroy(); err != nil {
		reqLogger.Error().
			Err(err).
			Msg("Failed to destroy session during logout")
		return err
	}

	reqLogger.Info().Msg("User logged out successfully")
	return nil
}
