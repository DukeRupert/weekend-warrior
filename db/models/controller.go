// db/models/controller.go
package models

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
	RoleSuper Role = "super"
)

// String implements the Stringer interface
func (r Role) String() string {
	return string(r)
}

// Controller represents a controller in the database
type Controller struct {
	ID         int       `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Name       string    `json:"name"`
	Initials   string    `json:"initials"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // Hashed password
	FacilityID int       `json:"facility_id"`
	Role       Role      `json:"role" validate:"required,oneof=super admin user"`
}

// CreateControllerParams holds the parameters needed to create a new controller
type CreateControllerParams struct {
	Name       string `json:"name"`
	Initials   string `json:"initials"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	FacilityID int    `json:"facility_id"`
	Role       Role   `json:"role" validate:"required,oneof=super admin user"`
}
