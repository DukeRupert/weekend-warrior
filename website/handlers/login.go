// website/handlers/login.go
package handlers

import (
	"context"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	return c.Render("login", fiber.Map{
		"title": "Login",
		"error": c.Query("error"),
	})
}

// LoginCredentials represents the login form data
type LoginCredentials struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

// HandleLogin processes the login form
func (h *LoginHandler) HandleLogin(c *fiber.Ctx) error {
	var creds LoginCredentials
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Query the controller by email
	var controller struct {
		ID         int
		FacilityID int
		IsAdmin    bool
		Password   string // This should be a hashed password in your DB
	}

	err := h.db.QueryRow(context.Background(),
		`SELECT id, facility_id, is_admin, password 
		FROM controllers 
		WHERE email = $1`,
		creds.Email).Scan(&controller.ID, &controller.FacilityID, &controller.IsAdmin, &controller.Password)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("email", creds.Email).
			Msg("Login attempt failed: user not found")

		return c.Redirect("/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Here you should properly verify the password hash
	// This is just a placeholder - implement proper password verification!
	if controller.Password != creds.Password {
		h.logger.Warn().
			Str("email", creds.Email).
			Msg("Login attempt failed: invalid password")

		return c.Redirect("/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Create session
	if err := h.auth.Login(c, controller.ID, controller.FacilityID, controller.IsAdmin); err != nil {
		h.logger.Error().
			Err(err).
			Str("email", creds.Email).
			Msg("Failed to create session")

		return c.Redirect("/login?error=Server+error", fiber.StatusFound)
	}

	// Redirect to dashboard or original requested URL
	return c.Redirect("/api/v1/controllers", fiber.StatusFound)
}

// HandleLogout processes logout requests
func (h *LoginHandler) HandleLogout(c *fiber.Ctx) error {
	if err := h.auth.Logout(c); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to logout")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	return c.Redirect("/login", fiber.StatusFound)
}
