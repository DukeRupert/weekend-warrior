// website/handlers/login.go
package handlers

import (
	"context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
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
	return c.Render("login", fiber.Map{
		"title": "Login",
		"error": c.Query("error"),
	})
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

    // Query the controller by email
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

    // Create session with role-based permissions
    sessionData := SessionData{
        UserID:     controller.ID,
        FacilityID: controller.FacilityID,
        Role:       controller.Role,
        Name:       controller.Name,
    }

    // Create a new session
    session, err := h.createSession(c, sessionData)
    if err != nil {
        h.logger.Error().
            Err(err).
            Str("email", creds.Email).
            Interface("sessionData", sessionData). // Log the session data for debugging
            Msg("Failed to create session")

        return c.Redirect("/login?error=Server+error", fiber.StatusFound)
    }

    // Set the session cookie
    c.Cookie(&fiber.Cookie{
        Name:     "session_id",
        Value:    session.ID,
        Expires:  session.ExpiresAt,
        HTTPOnly: true,
        Secure:   true,
        SameSite: "Lax",
    })

    // Redirect based on role
    redirectURL := h.getRedirectURL(controller.Role)
    return c.Redirect(redirectURL, fiber.StatusFound)
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

func (h *LoginHandler) createSession(c *fiber.Ctx, data SessionData) (*Session, error) {
    session := Session{
        ID:        uuid.New().String(),
        ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
    }

    // Serialize session data
    sessionBytes, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal session data: %w", err)
    }

    // Store session in database - explicitly convert to []byte
    _, err = h.db.Exec(context.Background(),
        `INSERT INTO sessions (id, data, expires_at, user_id, ip_address, user_agent)
         VALUES ($1, $2::bytea, $3, $4, $5, $6)`,
        session.ID,
        sessionBytes,
        session.ExpiresAt,
        data.UserID,
        c.IP(),
        c.Get("User-Agent"),
    )
    if err != nil {
        h.logger.Error().
            Err(err).
            Str("sessionID", session.ID).
            Int("userID", data.UserID).
            Msg("Failed to store session in database")
        return nil, fmt.Errorf("failed to store session: %w", err)
    }

    return &session, nil
}

// getRedirectURL returns the appropriate redirect URL based on role
func (h *LoginHandler) getRedirectURL(role string) string {
    switch role {
    case "super":
        return "/api/v1/admin/dashboard"
    case "admin":
        return "/api/v1/facility/dashboard"
    default:
        return "/api/v1/dashboard"
    }
}

// Session represents a user session
type Session struct {
    ID        string    `json:"id"`
    ExpiresAt time.Time `json:"expires_at"`
}
