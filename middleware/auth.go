// middleware/auth.go
package middleware

import (
	"fmt"
	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// AuthMiddleware contains the dependencies for the auth middleware
type AuthMiddleware struct {
	Db      *db.Service
	Store   *session.Store
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
	// Initialize Session middleware with storage
	storage := postgres.New(postgres.Config{
		DB:         db.GetPool(),
		Table:      "fiber_storage",
		Reset:      false,
		GCInterval: 10 * time.Second,
	})

	store := session.New(session.Config{
		Storage: storage,
	})

	// Initialize logger
	logger := log.With().Str("middleware", "auth").Logger()

	return &AuthMiddleware{
		Db:      db,
		Store:   store,
		options: options,
		logger:  logger,
	}, nil
}

// Login creates a new session for authenticated users with logging
func (am *AuthMiddleware) Login(c *fiber.Ctx, userID int, facilityID int, role models.Role) error {
	// reqLogger := am.logger.With().
	// 	Str("ip", c.IP()).
	// 	Str("role", role.String()).
	// 	Int("user_id", userID).
	// 	Int("facility_id", facilityID).
	// 	Logger()

	// Get session
	sess, err := am.Store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}

	// Parse login request
	var login struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&login); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// ... your database query here ...
	user, err := db.GetLoginResponse(am.Db, login.Email)
	if err != nil {
		return c.Redirect("/test/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Store user information in session
	log.Info().Msg("Starting to set session values")

	sess.Set("user_id", int64(user.ID))
	log.Debug().
		Int("user_id", user.ID).
		Msg("Set user_id in session")

	sess.Set("role", string(user.Role))
	log.Debug().
		Str("role", string(user.Role)).
		Msg("Set role in session")

	sess.Set("facility_id", user.FacilityID)
	log.Debug().
		Int("facility_id", user.FacilityID).
		Msg("Set facility_id in session")

	log.Info().
		Str("session_id", sess.ID()).
		Msg("Session saved successfully")

		// Must save before redirect!
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("Failed to save session")
		return err
	}

	// Redirect based on role
	switch user.Role {
	case "super":
		return c.Redirect("/super/dashboard")
	case "admin":
		return c.Redirect(fmt.Sprintf("/app/%d/admin/dashboard", user.FacilityID))
	default:
		return c.Redirect(fmt.Sprintf("/app/%d/dashboard", user.FacilityID))
	}
}

// Protected middleware checks if the request has a valid session
func (am *AuthMiddleware) Protected() fiber.Handler {
	const loginRedirect string = "/login"
	return func(c *fiber.Ctx) error {
		log.Info().Msg("Starting protected route middleware check")

		// Authentication middleware
		sess, err := am.Store.Get(c)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", c.Path()).
				Msg("Failed to get session")
			return c.Redirect(loginRedirect)
		}

		log.Debug().
			Str("session_id", sess.ID()).
			Msg("Session retrieved successfully")

		userID := sess.Get("user_id")
		if userID == nil {
			log.Warn().
				Str("session_id", sess.ID()).
				Msg("No user_id found in session")
			return c.Redirect(loginRedirect)
		}

		// Log all session values for debugging
		log.Debug().
			Interface("user_id", userID).
			Interface("role", sess.Get("role")).
			Interface("facility_id", sess.Get("facility_id")).
			Msg("Session values")

		// Add user info to locals for use in handlers
		c.Locals("user_id", userID)
		c.Locals("role", sess.Get("role"))
		c.Locals("facility_id", sess.Get("facility_id"))

		log.Info().
			Interface("user_id", userID).
			Interface("role", sess.Get("role")).
			Interface("facility_id", sess.Get("facility_id")).
			Msg("Authentication successful, proceeding to handler")

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

		sess, err := am.Store.Get(c)
		if err != nil {
			reqLogger.Error().
				Err(err).
				Msg("Failed to get session in admin check")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		userID := sess.Get("user_id")
		role := sess.Get("role")

		if role == "admin" || role == "super" {
			reqLogger.Info().
				Interface("user_id", userID).
				Msg("Admin route accessed")

			return c.Next()

		}

		reqLogger.Warn().
			Interface("user_id", userID).
			Msg("Non-admin user attempted to access admin route")

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}
}

// Logout removes the session with logging
func (am *AuthMiddleware) Logout(c *fiber.Ctx) error {
	sess, err := am.Store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get session",
		})
	}

	if err := sess.Destroy(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to destroy session",
		})
	}

	return c.Redirect("/auth/login")
}
