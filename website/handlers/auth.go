// website/handlers/login.go
package handlers

import (
	"fmt"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles user authentication
type AuthHandler struct {
	Db     *db.Service
	Auth   *middleware.AuthMiddleware
	Logger zerolog.Logger
}

// NewAuthHandler creates a new login handler
func NewAuthHandler(db *db.Service, auth *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		Db:     db,
		Auth:   auth,
		Logger: log.With().Str("handler", "login").Logger(),
	}
}

// LoginCredentials represents the login form data
type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleLogin processes the login form
func (h *AuthHandler) HandleLogin(c *fiber.Ctx) error {
	// Check form data
	var creds LoginCredentials
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	h.Logger.Debug().
		Str("query", `SELECT id, facility_id, role, password, name 
    FROM controllers 
    WHERE email = $1`).
		Str("email", creds.Email).
		Msg("Executing SQL query")

	// Check database for user
	controller, err := db.GetLoginResponse(h.Db, creds.Email)
	if err != nil {
		h.Logger.Warn().
			Err(err).
			Str("email", creds.Email).
			Msg("Login attempt failed: user not found")
		return c.Redirect("/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Verify password using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(controller.Password), []byte(creds.Password)); err != nil {
		h.Logger.Warn().
			Str("email", creds.Email).
			Msg("Login attempt failed: invalid password")

		return c.Redirect("/login?error=Invalid+credentials", fiber.StatusFound)
	}

	// Use the auth middleware to create new session
	if err := h.Auth.Login(c, controller.ID, controller.FacilityID, controller.Role); err != nil {
		h.Logger.Error().
			Err(err).
			Str("email", creds.Email).
			Int("userID", controller.ID).
			Int("facilityID", controller.FacilityID).
			Str("role", controller.Role.String()).
			Msg("Failed to create session")

		return c.Redirect("/login?error=Server+error", fiber.StatusFound)
	}

	h.Logger.Info().
		Str("email", creds.Email).
		Int("userID", controller.ID).
		Str("role", controller.Role.String()).
		Msg("Login successful")

	// Redirect based on role
	redirectURL := h.getRedirectURL(controller.Role.String(), controller.Code)
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// HandleLogout processes logout requests
func (h *AuthHandler) HandleLogout(c *fiber.Ctx) error {
	return h.Auth.Logout(c)
}

// getRedirectURL returns the appropriate redirect URL based on role
func (h *AuthHandler) getRedirectURL(role string, facility string) string {
	switch role {
	case "super":
		return "/super/dashboard"
	case "admin":
		return fmt.Sprintf("/%s/admin/calendar", facility)
	default:
		return fmt.Sprintf("/%s/calendar", facility)
	}
}

// LoginForm displays the login page
func (h *AuthHandler) LoginForm(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{
		"title": "Login",
		"error": c.Query("error"),
	}, "layouts/base")
}

// LogoutForm displays the login page
func (h *AuthHandler) LogoutForm(c *fiber.Ctx) error {
	return c.Render("pages/logout", fiber.Map{
		"title": "Logout",
		"error": c.Query("error"),
	}, "layouts/base")
}

// RegisterForm displays the register page
func (h *AuthHandler) RegisterForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "RegisterForm").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("retrieving facilities list")

	facilities, err := h.Db.ListFacilities(c.Context())
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("failed to retrieve facilities")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve facilities",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("facility_count", len(facilities)).
		Msg("facilities retrieved successfully")
	
	return c.Render("pages/register", fiber.Map{
		"title": "Register",
		"error": c.Query("error"),
		"facilities": facilities,
	}, "layouts/base")
}
