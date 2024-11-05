package handlers

import (
	"fmt"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ScheduleHandler handles HTTP requests for schedules
type ScheduleHandler struct {
	dbService *db.Service
	logger    zerolog.Logger
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(dbService *db.Service) *ScheduleHandler {
	return &ScheduleHandler{
		dbService: dbService,
		logger:    log.With().Str("handler", "schedule").Logger(),
	}
}

// CreateSchedule handles POST requests to create a new schedule
func (h *ScheduleHandler) CreateSchedule(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "CreateSchedule").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing create schedule request")

	var params models.CreateScheduleParams
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

	// Log the parsed parameters
	reqLogger.Debug().
		Interface("controller_id", params.ControllerID).
		Interface("rdos", params.RDOs).
		Time("anchor", params.Anchor).
		Msg("attempting to create schedule")

	schedule, err := h.dbService.CreateSchedule(c.Context(), params)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Interface("params", params).
			Msg("failed to create schedule")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create schedule",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("schedule_id", schedule.ID).
		Int("controller_id", schedule.ControllerID).
		Time("created_at", schedule.CreatedAt).
		Msg("schedule created successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": schedule,
	})
}

// GetSchedule handles GET requests to retrieve a schedule by ID
func (h *ScheduleHandler) GetSchedule(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "GetSchedule").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing get schedule request")

	id, err := c.ParamsInt("id")
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", c.Params("id")).
			Msg("invalid schedule ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid schedule ID",
			"detail": err.Error(),
		})
	}

	reqLogger.Debug().
		Int("schedule_id", id).
		Msg("retrieving schedule")

	schedule, err := h.dbService.GetSchedule(c.Context(), id)
	if err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("schedule_id", id).
				Msg("schedule not found")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Schedule not found",
				"detail": fmt.Sprintf("no schedule found with ID %d", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("schedule_id", id).
			Msg("failed to retrieve schedule")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve schedule",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("schedule_id", id).
		Int("controller_id", schedule.ControllerID).
		Time("created_at", schedule.CreatedAt).
		Msg("schedule retrieved successfully")

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// GetScheduleByController handles GET requests to retrieve a schedule by controller ID
func (h *ScheduleHandler) GetScheduleByController(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "GetScheduleByController").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing get schedule by controller request")

	controllerID, err := c.ParamsInt("id")
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("controller_id_raw", c.Params("id")).
			Msg("invalid controller ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": err.Error(),
		})
	}

	reqLogger.Debug().
		Int("controller_id", controllerID).
		Msg("retrieving schedule for controller")

	schedule, err := h.dbService.GetScheduleByController(c.Context(), controllerID)
	if err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("id", controllerID).
				Msg("no schedule found for controller")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Schedule not found",
				"detail": fmt.Sprintf("no schedule found for controller ID %d", controllerID),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("id", controllerID).
			Msg("failed to retrieve schedule")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve schedule",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("controller_id", controllerID).
		Int("schedule_id", schedule.ID).
		Time("created_at", schedule.CreatedAt).
		Interface("rdos", schedule.RDOs).
		Msg("schedule retrieved successfully")

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// UpdateSchedule handles PUT requests to update an existing schedule
func (h *ScheduleHandler) UpdateSchedule(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "UpdateSchedule").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing update schedule request")

	id, err := c.ParamsInt("id")
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", c.Params("id")).
			Msg("invalid schedule ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid schedule ID",
			"detail": err.Error(),
		})
	}

	var params models.UpdateScheduleParams
	if err := c.BodyParser(&params); err != nil {
		reqLogger.Error().
			Err(err).
			Str("body", string(c.Body())).
			Int("schedule_id", id).
			Msg("failed to parse request body")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	reqLogger.Debug().
		Int("schedule_id", id).
		Interface("rdos", params.RDOs).
		Time("anchor", params.Anchor).
		Msg("attempting to update schedule")

	schedule, err := h.dbService.UpdateSchedule(c.Context(), id, params)
	if err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("schedule_id", id).
				Interface("params", params).
				Msg("schedule not found for update")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Schedule not found",
				"detail": fmt.Sprintf("no schedule found with ID %d", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("schedule_id", id).
			Interface("params", params).
			Msg("failed to update schedule")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update schedule",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("schedule_id", schedule.ID).
		Int("controller_id", schedule.ControllerID).
		Interface("rdos", schedule.RDOs).
		Time("anchor", schedule.Anchor).
		Msg("schedule updated successfully")

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// DeleteSchedule handles DELETE requests to remove a schedule
func (h *ScheduleHandler) DeleteSchedule(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "DeleteSchedule").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing delete schedule request")

	id, err := c.ParamsInt("id")
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", c.Params("id")).
			Msg("invalid schedule ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid schedule ID",
			"detail": err.Error(),
		})
	}

	reqLogger.Debug().
		Int("schedule_id", id).
		Msg("attempting to delete schedule")

	if err := h.dbService.DeleteSchedule(c.Context(), id); err != nil {
		if isNotFoundError(err) {
			reqLogger.Warn().
				Int("schedule_id", id).
				Msg("schedule not found for deletion")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Schedule not found",
				"detail": fmt.Sprintf("no schedule found with ID %d", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("schedule_id", id).
			Msg("failed to delete schedule")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete schedule",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("schedule_id", id).
		Msg("schedule deleted successfully")

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterRoutes registers all schedule routes
func (h *ScheduleHandler) RegisterRoutes(app fiber.Router) {
	schedules := app.Group("api/v1/schedules")
	// Create new schedule
	schedules.Post("/", h.CreateSchedule)
	// Get schedule by ID
	schedules.Get("/:id", h.GetSchedule)
	// Update schedule by ID
	schedules.Put("/:id", h.UpdateSchedule)
	// Delete schedule by ID
	schedules.Delete("/:id", h.DeleteSchedule)

	// Controller-specific schedule routes
	controllers := schedules.Group("/controller")
	// Get schedule by controller ID
	controllers.Get("/:id", h.GetScheduleByController)
}
