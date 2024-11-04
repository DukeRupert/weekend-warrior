// website/handlers/auth.go
package handlers

import (
	"errors"

	"github.com/dukerupert/weekend-warrior/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the structure of the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse represents the structure of the successful login response
type LoginResponse struct {
	Controller struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		FacilityID uint   `json:"facility_id"`
		Role       string `json:"role"`
	} `json:"controller"`
}

// Custom errors for authentication
var (
	ErrNotFound        = errors.New("record not found")
	ErrInvalidPassword = errors.New("invalid password")
)

// checkPassword compares a plaintext password with a hash
func checkPassword(plaintext, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	return err == nil
}

type AuthHandler struct {
	dbService *db.Service
	logger    zerolog.Logger
}

func NewAuthHandler(dbService *db.Service) *ControllerHandler {
	return &ControllerHandler{
		dbService: dbService,
		logger:    log.With().Str("handler", "controller").Logger(),
	}
}

func (h *AuthHandler) login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx := c.Context()
	controller, err := h.dbService.GetControllerByEmail(ctx, req.Email)
	if err != nil {
		if err == ErrNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if !checkPassword(req.Password, controller.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	sess := c.Locals("session").(*session.Session)
	sess.Set("controller_id", controller.ID)
	sess.Set("facility_id", controller.FacilityID)
	sess.Set("role", controller.Role)

	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Session error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"controller": map[string]interface{}{
			"id":          controller.ID,
			"name":        controller.Name,
			"email":       controller.Email,
			"facility_id": controller.FacilityID,
			"role":        controller.Role,
		},
	})
}
