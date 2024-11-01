// handlers/facilities.go
package handlers

import (
	"fmt"
    "strconv"
    "github.com/gofiber/fiber/v2"
    "github.com/dukerupert/weekend-warrior/db"
	"github.com/dukerupert/weekend-warrior/models"
)

// FacilityHandler handles HTTP requests for facilities
type FacilityHandler struct {
    dbService *db.Service
}

// NewFacilityHandler creates a new facility handler
func NewFacilityHandler(dbService *db.Service) *FacilityHandler {
    return &FacilityHandler{
        dbService: dbService,
    }
}

// CreateFacilityRequest represents the request body for creating a facility
type CreateFacilityRequest struct {
    Name string `json:"name"`
    Code string `json:"code"`
}

// ListFacilities handles GET requests to list all facilities
func (h *FacilityHandler) ListFacilities(c *fiber.Ctx) error {
    facilities, err := h.dbService.ListFacilities(c.Context())
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve facilities",
            "detail": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "data": facilities,
    })
}

func (h *FacilityHandler) CreateFacility(c *fiber.Ctx) error {
    var req CreateFacilityRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
            "detail": err.Error(),
        })
    }

    // Validate request
    if req.Name == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request",
            "detail": "name is required",
        })
    }
    if req.Code == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request",
            "detail": "code is required",
        })
    }
    if len(req.Code) != 4 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request",
            "detail": "code must be exactly 4 characters",
        })
    }

    facility, err := h.dbService.CreateFacility(c.Context(), models.CreateFacilityParams{
        Name: req.Name,
        Code: req.Code,
    })
    if err != nil {
        // Check for unique code violation
        if isDuplicateKeyError(err) {
            return c.Status(fiber.StatusConflict).JSON(fiber.Map{
                "error": "Facility code already exists",
                "detail": fmt.Sprintf("code %s is already in use", req.Code),
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to create facility",
            "detail": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "data": facility,
    })
}

func (h *FacilityHandler) DeleteFacility(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid facility ID",
            "detail": "ID must be a number",
        })
    }

    err = h.dbService.DeleteFacility(c.Context(), id)
    if err != nil {
        // Check if facility wasn't found
        if err.Error() == fmt.Sprintf("facility with ID %d not found", id) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Facility not found",
                "detail": fmt.Sprintf("no facility found with ID %d", id),
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to delete facility",
            "detail": err.Error(),
        })
    }

    return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *FacilityHandler) DeleteFacilityByCode(c *fiber.Ctx) error {
    code := c.Params("code")
    if len(code) != 4 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid facility code",
            "detail": "code must be exactly 4 characters",
        })
    }

    err := h.dbService.DeleteFacilityByCode(c.Context(), code)
    if err != nil {
        // Check if facility wasn't found
        if err.Error() == fmt.Sprintf("facility with code %s not found", code) {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Facility not found",
                "detail": fmt.Sprintf("no facility found with code %s", code),
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to delete facility",
            "detail": err.Error(),
        })
    }

    return c.Status(fiber.StatusNoContent).Send(nil)
}

// Helper function to check for unique constraint violations
func isDuplicateKeyError(err error) bool {
    return err != nil && err.Error() != "" &&
        (err.Error() == "duplicate key value violates unique constraint" ||
            err.Error() == "facility code already exists")
}

// RegisterRoutes registers all facility routes
func (h *FacilityHandler) RegisterRoutes(app *fiber.App) {
	facilities := app.Group("/facilities")
    // List all facilities
    facilities.Get("/", h.ListFacilities)
    // Create new facility
    facilities.Post("/", h.CreateFacility)
    // Delete facility by ID
    facilities.Delete("/:id", h.DeleteFacility)
    // Delete facility by code
    facilities.Delete("/code/:code", h.DeleteFacilityByCode)
}