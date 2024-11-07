// website/handlers/handler.go
// Handle standard user interactions
package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Db     *db.Service
	Logger zerolog.Logger
}

func NewUserHandler(db *db.Service) *UserHandler {
	return &UserHandler{
		Db:     db,
		Logger: log.With().Str("handler", "user").Logger(),
	}
}

// CreateData represents the registration form data
type CreateData struct {
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

// CreateController handles POST requests to create a new controller
func (h *UserHandler) Create(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "HandleRegister").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing registration request")

	var data CreateData
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
	if errors := h.validateCreate(&data); len(errors) > 0 {
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
	err := h.Db.QueryRow(context.Background(),
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

	// Insert new controller
	var controllerID int
	err = h.Db.QueryRow(context.Background(),
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

	// Log successful registration
	reqLogger.Info().
		Int("controller_id", controllerID).
		Str("email", data.Email).
		Str("name", data.Name).
		Str("initials", data.Initials).
		Int("facility_id", data.Facility).
		Msg("controller registered successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
	})
}

// validateRegistration validates the registration data
func (h *UserHandler) validateCreate(data *CreateData) []ValidationError {
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

// UpdateController handles PUT requests to update a controller
func (h *UserHandler) Update(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "UpdateController").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing update controller request")

	// Parse and validate ID
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", c.Params("id")).
			Msg("invalid controller ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	// Parse request body
	var params models.CreateControllerParams
	if err := c.BodyParser(&params); err != nil {
		reqLogger.Error().
			Err(err).
			Str("body", string(c.Body())).
			Msg("failed to parse request body")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	// Log update attempt with parameters
	reqLogger.Debug().
		Int("controller_id", id).
		Str("name", params.Name).
		Str("initials", params.Initials).
		Str("email", params.Email).
		Int("facility_id", params.FacilityID).
		Msg("attempting to update controller")

	// Perform update
	controller, err := h.Db.UpdateController(c.Context(), id, params)
	if err != nil {
		if isDuplicateKeyError(err) {
			reqLogger.Warn().
				Err(err).
				Int("controller_id", id).
				Str("email", params.Email).
				Str("initials", params.Initials).
				Int("facility_id", params.FacilityID).
				Msg("duplicate controller detected during update")

			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Controller update conflict",
				"detail": "Email or initials already in use at this facility",
			})
		}

		reqLogger.Error().
			Err(err).
			Int("controller_id", id).
			Interface("params", params).
			Msg("failed to update controller")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update controller",
			"detail": err.Error(),
		})
	}

	// Log successful update
	reqLogger.Info().
		Int("controller_id", controller.ID).
		Str("name", controller.Name).
		Str("email", controller.Email).
		Int("facility_id", controller.FacilityID).
		Msg("controller updated successfully")

	return c.JSON(fiber.Map{
		"data": controller,
	})
}

// DeleteController handles DELETE requests to delete a controller
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "DeleteController").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing delete controller request")

	// Parse and validate ID
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", c.Params("id")).
			Msg("invalid controller ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	reqLogger.Debug().
		Int("controller_id", id).
		Msg("attempting to delete controller")

	err = h.Db.DeleteController(c.Context(), id)
	if err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("controller_id", id).
				Msg("controller not found for deletion")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Controller not found",
				"detail": fmt.Sprintf("no controller found with ID %d", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("controller_id", id).
			Msg("failed to delete controller")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete controller",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("controller_id", id).
		Msg("controller deleted successfully")

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ListControllers handles GET requests to list all controllers
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "GetAll").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("retrieving controllers list")

	// Get code parameter as string
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing facility code parameter",
		})
	}

	if len(code) != 4 {
		reqLogger.Error().Msg("code must be exactly 4 characters long")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid id value",
		})
	}

	log.Debug().
		Str("code", code).
		Msg("Successfully parsed facility code")

	controllers, err := h.Db.GetControllersByFacilityCode(c.Context(), code)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("failed to retrieve controllers")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve controllers",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("controller_count", len(controllers)).
		Msg("controllers retrieved successfully")

	return c.Render("pages/admin/controllers/page", fiber.Map{
		"title":       "Controllers",
		"code":        code,
		"error":       c.Query("error"),
		"controllers": controllers,
	}, "layouts/base", "layouts/app")
}

// CreateForm displays the registration page
func (h *UserHandler) CreateForm(c *fiber.Ctx) error {
	return c.Render("pages/admin/controllers/createForm", fiber.Map{
		"Name": "Lothlorien TRACON",
		"Code": "LOTH",
	})
}

// UpdateForm renders the controller edit form with preloaded data
func (h *UserHandler) UpdateForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "EditForm").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing edit form request")

	// Validate the ID
	controllerID := c.Params("id")
	id, err := strconv.Atoi(controllerID)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", controllerID).
			Msg("invalid controller ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	// Log attempt to fetch controller
	reqLogger.Debug().
		Int("controller_id", id).
		Msg("fetching controller data for edit form")

	// Fetch the controller data
	controller, err := h.Db.GetControllerByID(c.Context(), id)
	if err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("controller_id", id).
				Msg("controller not found for edit form")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Controller not found",
				"detail": fmt.Sprintf("no controller found with ID %d", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("controller_id", id).
			Msg("failed to retrieve controller for edit form")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve controller",
			"detail": err.Error(),
		})
	}

	// Log render attempt
	reqLogger.Debug().
		Int("controller_id", id).
		Str("controller_name", controller.Name).
		Str("template", "controllers/manage").
		Msg("rendering controller edit form")

	// Render the form
	err = c.Render("controllers/manage", fiber.Map{
		"Title":      "Assign Schedule",
		"Controller": controller,
	})
	if err != nil {
		reqLogger.Error().
			Err(err).
			Int("controller_id", id).
			Str("template", "controllers/manage").
			Msg("failed to render controller edit form")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to render form",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("controller_id", id).
		Str("controller_name", controller.Name).
		Msg("controller edit form rendered successfully")

	return nil
}

// ScheduleForm renders the controller schedule form
func (h *UserHandler) ScheduleForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.Logger.With().
		Str("method", "ScheduleForm").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().
		Str("template", "controllers/schedule").
		Msg("rendering controller schedule form")

	// Validate the ID
	controllerID := c.Params("id")
	id, err := strconv.Atoi(controllerID)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", controllerID).
			Msg("invalid controller ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	// Log attempt to fetch controller
	reqLogger.Debug().
		Int("controller_id", id).
		Msg("fetching controller data for edit form")

	// Fetch the controller data
	controller, err := h.Db.GetControllerByID(c.Context(), id)
	if err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("controller_id", id).
				Msg("controller not found for edit form")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Controller not found",
				"detail": fmt.Sprintf("no controller found with ID %d", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("controller_id", id).
			Msg("failed to retrieve controller for edit form")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve controller",
			"detail": err.Error(),
		})
	}

	err = c.Render("controllers/schedule", fiber.Map{
		"Title":      "Edit Controller",
		"EditMode":   true,
		"Controller": controller,
	})
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("template", "controllers/schedule").
			Msg("failed to render controller schedule form")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to render schedule form",
			"detail": err.Error(),
		})
	}

	reqLogger.Debug().Msg("controller schedule form rendered successfully")

	return nil
}
