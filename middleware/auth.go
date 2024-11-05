// middleware/auth.go
package middleware

import (
	"context"
	"time"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware contains the dependencies for the auth middleware
type AuthMiddleware struct {
	db      *db.Service
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
func NewAuthMiddleware(db *db.Service, options SessionOptions) (*AuthMiddleware, error) {
	// Create session store
	store := session.New()

	// Initialize logger
	logger := log.With().Str("middleware", "auth").Logger()

	return &AuthMiddleware{
		db:      db,
		store:   store,
		options: options,
		logger:  logger,
	}, nil
}

// Login creates a new session for authenticated users with logging
func (am *AuthMiddleware) Login(c *fiber.Ctx, userID int, facilityID int, role models.Role) error {
	reqLogger := am.logger.With().
		Str("ip", c.IP()).
		Str("role", role.String()).
		Int("user_id", userID).
		Int("facility_id", facilityID).
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
	sess.Set("role", role.String())

	// Store session in database
	sessionID := sess.ID()
	expiresAt := time.Now().Add(am.options.Expiration)

	_, err = am.db.Exec(context.Background(),
		`INSERT INTO sessions (
        id,
        user_id,
        created_at,
        expires_at,
        ip_address,
        user_agent,
        is_active
    ) VALUES ($1, $2, NOW(), $3, $4, $5, true)`,
		sessionID,
		sess.Get("user_id"), // Make sure you're storing user_id in your session
		expiresAt,
		c.IP(), // Assuming you have access to the Fiber context 'c'
		c.Get("User-Agent"),
	)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("session_id", sessionID).
			Int("user_id", sess.Get("user_id").(int)). // Type assertion needed
			Str("ip", c.IP()).
			Msg("Failed to store session in database")
		return err
	}

	// Set a new cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",                            // Cookie is valid for all paths
		Expires:  time.Now().Add(24 * time.Hour), // Expires in 24 hours
		Secure:   true,                           // Only sent over HTTPS
		HTTPOnly: true,                           // Not accessible via JavaScript
		SameSite: "Strict",                       // Strict same-site policy
	})

	reqLogger.Info().
		Str("session_id", sessionID).
		Time("expires_at", expiresAt).
		Msg("User logged in successfully")

	return nil
}

// Protected middleware checks if the request has a valid session
func (am *AuthMiddleware) Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqLogger := am.logger.With().
			Str("ip", c.IP()).
			Str("path", c.Path()).
			Str("method", c.Method()).
			Logger()

		// Check session ID from cookie
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			reqLogger.Debug().
				Msg("No session cookie found, redirecting to login")
			return c.Redirect("/login")
		}

		reqLogger.Debug().
			Str("session_id", sessionID).
			Msg("Validating session")

		// Validate session in database
		var userID int
		var isActive bool
		err := am.db.QueryRow(context.Background(), `
       SELECT user_id, is_active 
       FROM sessions 
       WHERE id = $1 
           AND expires_at > NOW() 
           AND is_active = true`,
			sessionID,
		).Scan(&userID, &isActive)
		if err != nil {
			if err == pgx.ErrNoRows {
				reqLogger.Info().
					Str("session_id", sessionID).
					Msg("Invalid session ID, redirecting to login")
			} else {
				reqLogger.Error().
					Err(err).
					Str("session_id", sessionID).
					Msg("Database error while validating session")
			}
			c.ClearCookie("session_id")
			return c.Redirect("/login")
		}

		if !isActive {
			reqLogger.Info().
				Str("session_id", sessionID).
				Int("user_id", userID).
				Msg("Inactive session, redirecting to login")
			c.ClearCookie("session_id")
			return c.Redirect("/login")
		}

		reqLogger.Debug().
			Str("session_id", sessionID).
			Int("user_id", userID).
			Msg("Session validated successfully")

		// Store user ID in context for route handlers
		c.Locals("user_id", userID)
		return c.Next()
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
