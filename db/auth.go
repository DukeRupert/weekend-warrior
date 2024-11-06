package db

import (
	"context"
	"fmt"
	"time"

	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/jackc/pgx/v5"
)

// loginResponse represents the expected return from a login query
type LoginResponse struct {
	ID         int         `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	Name       string      `json:"name"`
	Initials   string      `json:"initials"`
	Email      string      `json:"email"`
	Password   string      `json:"password"`
	FacilityID int         `json:"facility_id"`
	Role       models.Role `json:"role"`
	Code       string      `json:"code"`
}

func GetLoginResponse(s *Service, email string) (*LoginResponse, error) {
    var response LoginResponse
    
    err := s.QueryRow(context.Background(), `
        SELECT 
            c.id,
            c.created_at,
            c.name,
            c.initials,
            c.email,
            c.password,
            c.facility_id,
            c.role,
            f.code
        FROM controllers c
        JOIN facilities f ON c.facility_id = f.id
        WHERE c.email = $1`,
        email,
    ).Scan(
        &response.ID,
        &response.CreatedAt,
        &response.Name,
        &response.Initials,
        &response.Email,
        &response.Password,
        &response.FacilityID,
        &response.Role,
        &response.Code,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("no user found with email %s", email)
        }
        return nil, fmt.Errorf("error querying database: %w", err)
    }

    return &response, nil
}