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
	facilityID, err := strconv.Atoi(c.Params("facilityId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid facility ID",
			"detail": "ID must be a number",
		})
	}

	controllers, err := h.dbService.GetControllersByFacility(c.Context(), facilityID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve controllers",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": controllers,
	})
}

// CreateController handles POST requests to create a new controller
func (h *ControllerHandler) CreateController(c *fiber.Ctx) error {
	var params models.CreateControllerParams
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	// Validate request
	if params.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "name is required",
		})
	}
	if len(params.Initials) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "initials must be exactly 2 characters",
		})
	}
	if params.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "email is required",
		})
	}
	if params.FacilityID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "facility_id must be a positive number",
		})
	}

	controller, err := h.dbService.CreateController(c.Context(), params)
	if err != nil {
		if isDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Controller already exists",
				"detail": "Email or initials already in use at this facility",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create controller",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": controller,
	})
}

// UpdateController handles PUT requests to update a controller
func (h *ControllerHandler) UpdateController(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	var params models.CreateControllerParams
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	controller, err := h.dbService.UpdateController(c.Context(), id, params)
	if err != nil {
		if isDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Controller update conflict",
				"detail": "Email or initials already in use at this facility",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update controller",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": controller,
	})
}

// DeleteController handles DELETE requests to delete a controller
func (h *ControllerHandler) DeleteController(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	err = h.dbService.DeleteController(c.Context(), id)
	if err != nil {
		if isNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Controller not found",
				"detail": fmt.Sprintf("no controller found with ID %d", id),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete controller",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ShowCreateForm renders the controller creation form
func (h *ControllerHandler) ShowCreateForm(c *fiber.Ctx) error {
	return c.Render("controllers/manage", fiber.Map{
		"Title":      "Create New Controller",
		"EditMode":   false,
		"Controller": nil,
	})
}

// ShowEditForm renders the controller edit form with preloaded data
func (h *ControllerHandler) ShowEditForm(c *fiber.Ctx) error {
	controllerID := c.Params("id")

	// Validate the ID
	id, err := strconv.Atoi(controllerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": "ID must be a number",
		})
	}

	// Fetch the controller data
	controller, err := h.dbService.GetControllerByID(c.Context(), id)
	if err != nil {
		if isNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Controller not found",
				"detail": fmt.Sprintf("no controller found with ID %d", id),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve controller",
			"detail": err.Error(),
		})
	}

	return c.Render("controllers/manage", fiber.Map{
		"Title":      "Edit Controller",
		"EditMode":   true,
		"Controller": controller,
	})
}

func (h *ControllerHandler) ShowScheduleForm(c *fiber.Ctx) error {
	return c.Render("controllers/schedule", fiber.Map{})
}

// RegisterRoutes registers all controller routes
func (h *ControllerHandler) RegisterRoutes(app *fiber.App) {
	controllers := app.Group("/controllers")

	// List all controllers
	controllers.Get("/", h.ListControllers)

	// List controllers by facility
	controllers.Get("/facility/:facilityId", h.GetControllersByFacility)

	// Create new controller
	controllers.Post("/", h.CreateController)

	// Update controller
	controllers.Put("/:id", h.UpdateController)

	// Delete controller
	controllers.Delete("/:id", h.DeleteController)

	controllers.Get("/new", h.ShowCreateForm)
	controllers.Get("/edit/:id", h.ShowEditForm)
	controllers.Get("/schedule", h.ShowScheduleForm)
}
