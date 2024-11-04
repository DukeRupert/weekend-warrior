// db/models/controller.go
package models

import "time"

// Controller represents a controller in the database
type Controller struct {
	ID         int       `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Name       string    `json:"name"`
	Initials   string    `json:"initials"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // Hashed password
	FacilityID int       `json:"facility_id"`
	Role       string    `json:"role"`
}

// CreateControllerParams holds the parameters needed to create a new controller
type CreateControllerParams struct {
	Name       string `json:"name"`
	Initials   string `json:"initials"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	FacilityID int    `json:"facility_id"`
	Role       string `json:"role" validate:"required,oneof=super admin user"`
}
