// handlers/facilities.go
package handlers

import (
	"fmt"
	"strconv"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FacilityHandler handles HTTP requests for facilities
type FacilityHandler struct {
	dbService *db.Service
	logger    zerolog.Logger
}

// NewFacilityHandler creates a new facility handler
func NewFacilityHandler(dbService *db.Service) *FacilityHandler {
	return &FacilityHandler{
		dbService: dbService,
		logger:    log.With().Str("handler", "facility").Logger(),
	}
}

// CreateFacilityRequest represents the request body for creating a facility
type CreateFacilityRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (h *FacilityHandler) GetUserFacility(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "GetUserFacility").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("retrieving user facility")

	return c.JSON(fiber.Map{
		"data": "Under construction",
	})
}

// GetAll handles GET requests to list all facilities
func (h *FacilityHandler) GetAll(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "ListFacilities").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("retrieving facilities list")

	facilities, err := h.dbService.ListFacilities(c.Context())
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

	return c.Render("pages/super/facilities/page", fiber.Map{
		"title":      "Facilities",
		"error":      c.Query("error"),
		"facilities": facilities,
	}, "layouts/base", "layouts/app")
}

// Create handles POST requests to create a new facility
func (h *FacilityHandler) Create(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "CreateFacility").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing create facility request")

	var req CreateFacilityRequest
	if err := c.BodyParser(&req); err != nil {
		reqLogger.Error().
			Err(err).
			Str("body", string(c.Body())).
			Msg("failed to parse request body")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	// Validation logging
	if req.Name == "" {
		reqLogger.Error().
			Interface("request", req).
			Msg("validation failed: name is required")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "name is required",
		})
	}

	if req.Code == "" {
		reqLogger.Error().
			Interface("request", req).
			Msg("validation failed: code is required")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "code is required",
		})
	}

	if len(req.Code) != 4 {
		reqLogger.Error().
			Str("code", req.Code).
			Interface("request", req).
			Msg("validation failed: invalid code length")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "code must be exactly 4 characters",
		})
	}

	reqLogger.Debug().
		Str("name", req.Name).
		Str("code", req.Code).
		Msg("attempting to create facility")

	facility, err := h.dbService.CreateFacility(c.Context(), models.CreateFacilityParams{
		Name: req.Name,
		Code: req.Code,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			reqLogger.Warn().
				Str("code", req.Code).
				Str("name", req.Name).
				Msg("duplicate facility code detected")

			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Facility code already exists",
				"detail": fmt.Sprintf("code %s is already in use", req.Code),
			})
		}

		reqLogger.Error().
			Err(err).
			Str("name", req.Name).
			Str("code", req.Code).
			Msg("failed to create facility")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create facility",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("facility_id", facility.ID).
		Str("name", facility.Name).
		Str("code", facility.Code).
		Msg("facility created successfully")

	return c.Status(fiber.StatusOK).Render("pages/super/facilities/listItem", fiber.Map{
		"ID":        facility.Name,
		"Name":      facility.Name,
		"CreatedAt": facility.CreatedAt,
		"Code":      facility.Code,
	})
}

// UpdateFacility handles POST requests to create a new facility
func (h *FacilityHandler) Update(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "Update Facility").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing update facility request")

	// Get id parameter as string
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing id parameter",
		})
	}

	// Convert to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("id_param", idStr).
			Msg("Invalid id parameter format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid id format",
		})
	}

	// Optional: Check if id is positive
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid id value",
		})
	}

	log.Debug().
		Int("id", id).
		Msg("Successfully parsed facility ID")

	var req models.UpdateFacilityParams
	if err := c.BodyParser(&req); err != nil {
		reqLogger.Error().
			Err(err).
			Str("body", string(c.Body())).
			Msg("failed to parse request body")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
	}

	// Validation logging
	if req.Name == "" {
		reqLogger.Error().
			Interface("request", req).
			Msg("validation failed: name is required")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "name is required",
		})
	}

	if req.Code == "" {
		reqLogger.Error().
			Interface("request", req).
			Msg("validation failed: code is required")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "code is required",
		})
	}

	if len(req.Code) != 4 {
		reqLogger.Error().
			Str("code", req.Code).
			Interface("request", req).
			Msg("validation failed: invalid code length")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid request",
			"detail": "code must be exactly 4 characters",
		})
	}

	reqLogger.Debug().
		Str("name", req.Name).
		Str("code", req.Code).
		Msg("attempting to update facility")

	facility, err := h.dbService.UpdateFacility(c.Context(), models.UpdateFacilityParams{
		ID: id,
		Name: req.Name,
		Code: req.Code,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			reqLogger.Warn().
				Str("code", req.Code).
				Str("name", req.Name).
				Msg("duplicate facility code detected")

			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Facility code already exists",
				"detail": fmt.Sprintf("code %s is already in use", req.Code),
			})
		}

		reqLogger.Error().
			Err(err).
			Str("name", req.Name).
			Str("code", req.Code).
			Msg("failed to create facility")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create facility",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("facility_id", facility.ID).
		Str("name", facility.Name).
		Str("code", facility.Code).
		Msg("facility created successfully")

	return c.Status(fiber.StatusOK).Render("pages/super/facilities/listItem", fiber.Map{
		"ID":        facility.Name,
		"Name":      facility.Name,
		"CreatedAt": facility.CreatedAt,
		"Code":      facility.Code,
	})
}

// Delete handles DELETE requests to delete a facility
func (h *FacilityHandler) Delete(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "DeleteFacility").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	reqLogger.Info().Msg("processing delete facility request")

	// Parse and validate ID
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		reqLogger.Error().
			Err(err).
			Str("id_raw", c.Params("id")).
			Msg("invalid facility ID format")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Invalid facility ID",
			"detail": "ID must be a number",
		})
	}

	reqLogger.Debug().
		Int("facility_id", id).
		Msg("attempting to delete facility")

	err = h.dbService.DeleteFacility(c.Context(), id)
	if err != nil {
		if err.Error() == fmt.Sprintf("facility with id %s not found", id) {
			reqLogger.Warn().
				Int("facility_id", id).
				Msg("facility not found for deletion")

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":  "Facility not found",
				"detail": fmt.Sprintf("no facility found with id %s", id),
			})
		}

		reqLogger.Error().
			Err(err).
			Int("facility_id", id).
			Msg("failed to delete facility")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to delete facility",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Int("facility_id", id).
		Msg("facility deleted successfully")

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// CreateForm renders the facility creation form
func (h *FacilityHandler) CreateForm(c *fiber.Ctx) error {
	return c.Render("pages/super/facilities/createForm", fiber.Map{
		"Name": "Lothlorien TRACON",
		"Code": "LOTH",
	})
}

// UpdateForm returns form to edit a facility
func (h *FacilityHandler) UpdateForm(c *fiber.Ctx) error {
	// Create request-specific logger
	reqLogger := h.logger.With().
		Str("method", "EditFacility").
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Logger()

	// Get id parameter as string
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing id parameter",
		})
	}

	// Convert to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("id_param", idStr).
			Msg("Invalid id parameter format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid id format",
		})
	}

	// Optional: Check if id is positive
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid id value",
		})
	}

	log.Debug().
		Int("id", id).
		Msg("Successfully parsed facility ID")

	reqLogger.Info().Msg("retrieving facility")

	facility, err := h.dbService.GetFacilityByID(c.Context(), id)
	if err != nil {
		reqLogger.Error().
			Err(err).
			Msg("failed to retrieve facility")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to retrieve facility",
			"detail": err.Error(),
		})
	}

	reqLogger.Info().
		Str("facility code", facility.Code).
		Msg("facility retrieved successfully")

	return c.Render("pages/super/facilities/updateForm", fiber.Map{
		"ID":        facility.ID,
		"Name":      facility.Name,
		"CreatedAt": facility.CreatedAt,
		"Code":      facility.Code,
	})
}

// GetFacilityControllers returns all controllers for a facility code
func (h *FacilityHandler) GetFacilityControllers(c *fiber.Ctx) error {
	return c.Render("facilities/create", fiber.Map{
		"Title": "Facility Controller List",
	})
}
