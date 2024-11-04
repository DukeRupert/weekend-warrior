package handlers

import (
	"log"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/models"
	"github.com/gofiber/fiber/v2"
)

// ScheduleHandler handles HTTP requests for schedules
type ScheduleHandler struct {
	dbService *db.Service
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(dbService *db.Service) *ScheduleHandler {
	return &ScheduleHandler{
		dbService: dbService,
	}
}

// CreateSchedule handles POST requests to create a new schedule
func (h *ScheduleHandler) CreateSchedule(c *fiber.Ctx) error {
	log.Println("CreateSchedule() called")

	var params models.CreateScheduleParams
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	schedule, err := h.dbService.CreateSchedule(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create schedule",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": schedule,
	})
}

// GetSchedule handles GET requests to retrieve a schedule by ID
func (h *ScheduleHandler) GetSchedule(c *fiber.Ctx) error {
	log.Println("GetSchedule() called")

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid schedule ID",
			"detail": err.Error(),
		})
	}

	schedule, err := h.dbService.GetSchedule(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve schedule",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// GetScheduleByController handles GET requests to retrieve a schedule by controller ID
func (h *ScheduleHandler) GetScheduleByController(c *fiber.Ctx) error {
	log.Println("GetScheduleByController() called")

	controllerID, err := c.ParamsInt("controller_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid controller ID",
			"detail": err.Error(),
		})
	}

	schedule, err := h.dbService.GetScheduleByController(c.Context(), controllerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve schedule",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// UpdateSchedule handles PUT requests to update an existing schedule
func (h *ScheduleHandler) UpdateSchedule(c *fiber.Ctx) error {
	log.Println("UpdateSchedule() called")

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid schedule ID",
			"detail": err.Error(),
		})
	}

	var params models.UpdateScheduleParams
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	schedule, err := h.dbService.UpdateSchedule(c.Context(), id, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update schedule",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": schedule,
	})
}

// DeleteSchedule handles DELETE requests to remove a schedule
func (h *ScheduleHandler) DeleteSchedule(c *fiber.Ctx) error {
	log.Println("DeleteSchedule() called")

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid schedule ID",
			"detail": err.Error(),
		})
	}

	if err := h.dbService.DeleteSchedule(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete schedule",
			"detail": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterRoutes registers all schedule routes
func (h *ScheduleHandler) RegisterRoutes(app *fiber.App) {
	schedules := app.Group("/schedules")
	// Create new schedule
	schedules.Post("/", h.CreateSchedule)
	// Get schedule by ID
	schedules.Get("/:id", h.GetSchedule)
	// Update schedule by ID
	schedules.Put("/:id", h.UpdateSchedule)
	// Delete schedule by ID
	schedules.Delete("/:id", h.DeleteSchedule)

	// Controller-specific schedule routes
	controllers := app.Group("/controllers")
	// Get schedule by controller ID
	controllers.Get("/:controller_id/schedule", h.GetScheduleByController)
}
