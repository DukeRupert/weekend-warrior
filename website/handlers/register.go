package handlers

import (
	"context"
	"regexp"
	"strings"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/dukerupert/weekend-warrior/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler handles user registration
type RegisterHandler struct {
	db     *db.Service
	auth   *middleware.AuthMiddleware
	logger zerolog.Logger
}

// RegistrationData represents the registration form data
type RegistrationData struct {
	Name     string      `form:"name"`
	Email    string      `form:"email"`
	Initials string      `form:"initials"`
	Password string      `form:"password"`
	Confirm  string      `form:"confirm"`
	Facility int         `form:"facility"`
	Role     models.Role `form:"role"`
}

// ValidationError represents form validation errors
type ValidationError struct {
	Field   string
	Message string
}

// NewRegisterHandler creates a new registration handler
func NewRegisterHandler(db *db.Service, auth *middleware.AuthMiddleware) *RegisterHandler {
	return &RegisterHandler{
		db:     db,
		auth:   auth,
		logger: log.With().Str("handler", "register").Logger(),
	}
}

// RegisterRoutes registers the registration routes
func (h *RegisterHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/register", h.ShowRegisterForm)
	app.Post("/register", h.HandleRegister)
}

// ShowRegisterForm displays the registration page
func (h *RegisterHandler) ShowRegisterForm(c *fiber.Ctx) error {
	// Fetch facilities for the dropdown
	facilities, err := h.db.ListFacilities(context.Background())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to fetch facilities")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	return c.Render("register", fiber.Map{
		"title":      "Register",
		"error":      c.Query("error"),
		"facilities": facilities,
	})
}

// validateRegistration validates the registration data
func (h *RegisterHandler) validateRegistration(data *RegistrationData) []ValidationError {
	var errors []ValidationError

	// Validate name
	if strings.TrimSpace(data.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name is required",
		})
	}

	// Validate email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(data.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Invalid email address",
		})
	}

	// Validate initials
	if len(data.Initials) < 2 || len(data.Initials) > 3 {
		errors = append(errors, ValidationError{
			Field:   "initials",
			Message: "Initials must be 2-3 characters",
		})
	}

	// Validate password
	if len(data.Password) < 8 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must be at least 8 characters",
		})
	}

	if data.Password != data.Confirm {
		errors = append(errors, ValidationError{
			Field:   "confirm",
			Message: "Passwords do not match",
		})
	}

	// Validate facility
	if data.Facility <= 0 {
		errors = append(errors, ValidationError{
			Field:   "facility",
			Message: "Please select a facility",
		})
	}

	return errors
}

// HandleRegister processes the registration form
func (h *RegisterHandler) HandleRegister(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "HandleRegister").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing registration request")

	var data RegistrationData
	if err := c.BodyParser(&data); err != nil {
		reqLogger.Error().
			Err(err).
			Str("body", string(c.Body())).
			Msg("failed to parse registration form data")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	// Log parsed data before validation
	reqLogger.Debug().
		Str("name", data.Name).
		Str("email", data.Email).
		Str("initials", data.Initials).
		Str("role", data.Role.String()).
		Int("facility_id", data.Facility).
		Msg("validating registration data")

	// Validate form data
	if errors := h.validateRegistration(&data); len(errors) > 0 {
		reqLogger.Warn().
			Interface("validation_errors", errors).
			Interface("form_data", data).
			Msg("registration validation failed")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errors,
		})
	}

	// Check if email already exists
	var exists bool
	err := h.db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM controllers WHERE email = $1)",
		data.Email).Scan(&exists)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("email", data.Email).
			Msg("database error checking email existence")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	if exists {
		reqLogger.Info().
			Str("email", data.Email).
			Msg("registration attempted with existing email")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": []ValidationError{{
				Field:   "email",
				Message: "Email already registered",
			}},
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("failed to hash password")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	// Get a connection from the pool for the transaction
	conn, err := h.db.GetPool().Acquire(context.Background())
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("failed to acquire database connection")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}
	defer conn.Release()

	// Start transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("failed to start transaction")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}
	defer tx.Rollback(context.Background())

	// Insert new controller
	var controllerID int
	err = tx.QueryRow(context.Background(),
		`INSERT INTO controllers (name, email, initials, password, facility_id, role) 
		 VALUES ($1, $2, $3, $4, $5, $6) 
		 RETURNING id`,
		data.Name, data.Email, data.Initials, hashedPassword, data.Facility, data.Role).Scan(&controllerID)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Interface("controller_data", map[string]interface{}{
				"name":        data.Name,
				"email":       data.Email,
				"initials":    data.Initials,
				"facility_id": data.Facility,
			}).
			Msg("failed to insert new controller")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create account",
		})
	}

	// Commit transaction
	if err = tx.Commit(context.Background()); err != nil {
		reqLogger.Error().
			Err(err).
			Int("controller_id", controllerID).
			Msg("failed to commit transaction")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	// Log successful registration
	reqLogger.Info().
		Int("controller_id", controllerID).
		Str("email", data.Email).
		Str("name", data.Name).
		Str("initials", data.Initials).
		Int("facility_id", data.Facility).
		Msg("controller registered successfully")

	// Create session and log in the user
	if err := h.auth.Login(c, controllerID, data.Facility, data.Role); err != nil {
		reqLogger.Error().
			Err(err).
			Int("controller_id", controllerID).
			Int("facility_id", data.Facility).
			Msg("failed to create session after registration")

		return c.Redirect("/login", fiber.StatusFound)
	}

	reqLogger.Info().
		Int("controller_id", controllerID).
		Int("facility_id", data.Facility).
		Msg("user logged in after registration")

	// Redirect to dashboard
	return c.Redirect("/api/v1/controllers", fiber.StatusFound)
}
