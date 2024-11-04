// handlers/controllers.go
package handlers

import (
	"fmt"
	"strconv"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ControllerHandler struct {
	dbService *db.Service
	logger    zerolog.Logger
}

func NewControllerHandler(dbService *db.Service) *ControllerHandler {
	return &ControllerHandler{
		dbService: dbService,
		logger:    log.With().Str("handler", "controller").Logger(),
	}
}

// ListControllers handles GET requests to list all controllers
func (h *ControllerHandler) ListControllers(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "ListControllers").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("retrieving controllers list")

	controllers, err := h.dbService.ListControllers(c.Context())
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

	return c.JSON(fiber.Map{
		"data": controllers,
	})
}

// GetControllersByFacility handles GET requests to list controllers by facility
func (h *ControllerHandler) GetControllersByFacility(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "GetControllersByFacility").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	facilityID, err := strconv.Atoi(c.Params("facilityId"))
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("facility_id_raw", c.Params("facilityId")).
			Msg("invalid facility ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid facility ID",
			"detail": "ID must be a number",
		})
	}

	reqLogger.Debug().
		Int("facility_id", facilityID).
		Msg("retrieving controllers for facility")

	controllers, err := h.dbService.GetControllersByFacility(c.Context(), facilityID)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Int("facility_id", facilityID).
			Msg("failed to retrieve controllers for facility")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve controllers",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("facility_id", facilityID).
		Int("controller_count", len(controllers)).
		Msg("successfully retrieved controllers for facility")

	return c.JSON(fiber.Map{
		"data": controllers,
	})
}

// CreateController handles POST requests to create a new controller
func (h *ControllerHandler) CreateController(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "CreateController").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing create controller request")

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

	// Validation logging with detailed context
	if params.Name == "" {
		reqLogger.Error().
			Interface("params", params).
			Msg("validation failed: name is required")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "name is required",
		})
	}

	if len(params.Initials) != 2 {
		reqLogger.Error().
			Str("initials", params.Initials).
			Interface("params", params).
			Msg("validation failed: invalid initials length")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "initials must be exactly 2 characters",
		})
	}

	if params.Email == "" {
		reqLogger.Error().
			Interface("params", params).
			Msg("validation failed: email is required")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "email is required",
		})
	}

	if params.FacilityID <= 0 {
		reqLogger.Error().
			Int("facility_id", params.FacilityID).
			Interface("params", params).
			Msg("validation failed: invalid facility ID")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "facility_id must be a positive number",
		})
	}

	// Log validated parameters before database operation
	reqLogger.Debug().
		Str("name", params.Name).
		Str("initials", params.Initials).
		Str("email", params.Email).
		Int("facility_id", params.FacilityID).
		Msg("attempting to create controller")

	controller, err := h.dbService.CreateController(c.Context(), params)
	if err != nil {
		if isDuplicateKeyError(err) {
			reqLogger.Warn().
				Err(err).
				Str("email", params.Email).
				Str("initials", params.Initials).
				Int("facility_id", params.FacilityID).
				Msg("duplicate controller detected")

			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Controller already exists",
				"detail": "Email or initials already in use at this facility",
			})
		}

		reqLogger.Error().
			Err(err).
			Interface("params", params).
			Msg("failed to create controller")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create controller",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("controller_id", controller.ID).
		Str("name", controller.Name).
		Str("email", controller.Email).
		Int("facility_id", controller.FacilityID).
		Msg("controller created successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": controller,
	})
}

// UpdateController handles PUT requests to update a controller
func (h *ControllerHandler) UpdateController(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
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
	controller, err := h.dbService.UpdateController(c.Context(), id, params)
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
func (h *ControllerHandler) DeleteController(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
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

	err = h.dbService.DeleteController(c.Context(), id)
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

// ShowCreateForm renders the controller creation form
func (h *ControllerHandler) ShowCreateForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "ShowCreateForm").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().
		Str("template", "controllers/manage").
		Str("title", "Create New Controller").
		Bool("edit_mode", false).
		Msg("rendering controller creation form")

	err := c.Render("controllers/manage", fiber.Map{
		"Title":      "Create New Controller",
		"EditMode":   false,
		"Controller": nil,
	})
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("template", "controllers/manage").
			Msg("failed to render controller creation form")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to render form",
			"detail": err.Error(),
		})
	}

	reqLogger.Debug().Msg("controller creation form rendered successfully")

	return nil
}

// ShowEditForm renders the controller edit form with preloaded data
func (h *ControllerHandler) ShowEditForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "ShowEditForm").
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
	controller, err := h.dbService.GetControllerByID(c.Context(), id)
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
		"Title":      "Edit Controller",
		"EditMode":   true,
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

// ShowScheduleForm renders the controller schedule form
func (h *ControllerHandler) ShowScheduleForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "ShowScheduleForm").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().
		Str("template", "controllers/schedule").
		Msg("rendering controller schedule form")

	err := c.Render("controllers/schedule", fiber.Map{})
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

// RegisterRoutes registers all controller routes
func (h *ControllerHandler) RegisterRoutes(app *fiber.App) {
	controllers := app.Group("/controllers")

	// List all controllers
	controllers.Get("/", h.ListControllers)
	controllers.Post("/", h.CreateController)
	controllers.Put("/:id", h.UpdateController)
	controllers.Delete("/:id", h.DeleteController)

	// List controllers by facility
	controllers.Get("/facility/:facilityId", h.GetControllersByFacility)

	// Create new controller
	controllers.Get("/new", h.ShowCreateForm)

	// Update existing controller
	controllers.Get("/edit/:id", h.ShowEditForm)

	// Assign schedule to controller
	controllers.Get("/schedule", h.ShowScheduleForm)
}
