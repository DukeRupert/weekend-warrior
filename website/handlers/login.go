// website/handlers/login.go
package handlers

import (
	"context"
	"time"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/dukerupert/weekend-warrior/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler handles user authentication
type LoginHandler struct {
	db     *db.Service
	auth   *middleware.AuthMiddleware
	logger zerolog.Logger
}

// NewLoginHandler creates a new login handler
func NewLoginHandler(db *db.Service, auth *middleware.AuthMiddleware) *LoginHandler {
	return &LoginHandler{
		db:     db,
		auth:   auth,
		logger: log.With().Str("handler", "login").Logger(),
	}
}

// RegisterRoutes registers the login routes
func (h *LoginHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/login", h.ShowLoginForm)
	app.Post("/login", h.HandleLogin)
	app.Post("/logout", h.HandleLogout)
}

// ShowLoginForm displays the login page
func (h *LoginHandler) ShowLoginForm(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{
		"title": "Login",
		"error": c.Query("error"),
	}, "layouts/base")
}

// LoginCredentials represents the login form data
type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SessionData represents the data stored in a session
type SessionData struct {
	UserID     int    `json:"user_id"`
	FacilityID int    `json:"facility_id"`
	Role       string `json:"role"`
	Name       string `json:"name"`
}

// HandleLogin processes the login form
func (h *LoginHandler) HandleLogin(c *fiber.Ctx) error {
	var creds LoginCredentials
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	h.logger.Debug().
		Str("query", `SELECT id, facility_id, role, password, name 
    FROM controllers 
    WHERE email = $1`).
		Str("email", creds.Email).
		Msg("Executing SQL query")

	var controller models.Controller
	err := h.db.QueryRow(context.Background(),
		`SELECT id, facility_id, role, password, name 
     FROM controllers 
     WHERE email = $1`,
		creds.Email).Scan(
		&controller.ID,
		&controller.FacilityID,
		&controller.Role,
		&controller.Password,
		&controller.Name,
	)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("email", creds.Email).
			Msg("Login attempt failed: user not found")
		return c.Redirect("/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Verify password using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(controller.Password), []byte(creds.Password)); err != nil {
		h.logger.Warn().
			Str("email", creds.Email).
			Msg("Login attempt failed: invalid password")

		return c.Redirect("/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Use the auth middleware's Login method instead of creating our own session
	if err := h.auth.Login(c, controller.ID, controller.FacilityID, controller.Role); err != nil {
		h.logger.Error().
			Err(err).
			Str("email", creds.Email).
			Int("userID", controller.ID).
			Int("facilityID", controller.FacilityID).
			Str("role", controller.Role.String()).
			Msg("Failed to create session")

		return c.Redirect("/login?error=Server+error", fiber.StatusFound)
	}

	h.logger.Info().
		Str("email", creds.Email).
		Int("userID", controller.ID).
		Str("role", controller.Role.String()).
		Msg("Login successful")

	// Redirect based on role
	redirectURL := h.getRedirectURL(controller.Role.String())
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// HandleLogout processes logout requests
func (h *LoginHandler) HandleLogout(c *fiber.Ctx) error {
	return h.auth.Logout(c)
}

// getRedirectURL returns the appropriate redirect URL based on role
func (h *LoginHandler) getRedirectURL(role string) string {
	switch role {
	case "super":
		return "/app/v1/super/dashboard"
	case "admin":
		return "/app/v1/admin/dashboard"
	default:
		return "/app/v1/dashboard"
	}
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}
