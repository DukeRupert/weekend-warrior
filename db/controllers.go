// db/controllers.go
package db

import (
	"context"
	"fmt"

	"github.com/dukerupert/weekend-warrior/db/models"
)

// CreateController creates a new controller in the database
func (s *Service) CreateController(ctx context.Context, params models.CreateControllerParams) (*models.Controller, error) {
	var controller models.Controller

	err := s.pool.QueryRow(ctx, `
        INSERT INTO controllers (name, initials, email, facility_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, name, initials, email, facility_id
    `, params.Name, params.Initials, params.Email, params.FacilityID).Scan(
		&controller.ID,
		&controller.CreatedAt,
		&controller.Name,
		&controller.Initials,
		&controller.Email,
		&controller.FacilityID,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating controller: %w", err)
	}

	return &controller, nil
}

// GetControllerByID retrieves a controller by its ID
func (s *Service) GetControllerByID(ctx context.Context, id int) (*models.Controller, error) {
	var controller models.Controller

	err := s.pool.QueryRow(ctx, `
        SELECT id, created_at, name, initials, email, facility_id
        FROM controllers
        WHERE id = $1
    `, id).Scan(
		&controller.ID,
		&controller.CreatedAt,
		&controller.Name,
		&controller.Initials,
		&controller.Email,
		&controller.FacilityID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting controller: %w", err)
	}

	return &controller, nil
}

// GetControllersByFacility retrieves all controllers for a facility
func (s *Service) GetControllersByFacility(ctx context.Context, facilityID int) ([]models.Controller, error) {
	rows, err := s.pool.Query(ctx, `
        SELECT id, created_at, name, initials, email, facility_id
        FROM controllers
        WHERE facility_id = $1
        ORDER BY name ASC
    `, facilityID)
	if err != nil {
		return nil, fmt.Errorf("error listing controllers: %w", err)
	}
	defer rows.Close()

	var controllers []models.Controller
	for rows.Next() {
		var controller models.Controller
		err := rows.Scan(
			&controller.ID,
			&controller.CreatedAt,
			&controller.Name,
			&controller.Initials,
			&controller.Email,
			&controller.FacilityID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning controller row: %w", err)
		}
		controllers = append(controllers, controller)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating controller rows: %w", err)
	}

	return controllers, nil
}

// ListControllers retrieves all controllers
func (s *Service) ListControllers(ctx context.Context) ([]models.Controller, error) {
	rows, err := s.pool.Query(ctx, `
        SELECT id, created_at, name, initials, email, facility_id
        FROM controllers
        ORDER BY name ASC
    `)
	if err != nil {
		return nil, fmt.Errorf("error listing controllers: %w", err)
	}
	defer rows.Close()

	var controllers []models.Controller
	for rows.Next() {
		var controller models.Controller
		err := rows.Scan(
			&controller.ID,
			&controller.CreatedAt,
			&controller.Name,
			&controller.Initials,
			&controller.Email,
			&controller.FacilityID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning controller row: %w", err)
		}
		controllers = append(controllers, controller)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating controller rows: %w", err)
	}

	return controllers, nil
}

// UpdateController updates an existing controller
func (s *Service) UpdateController(ctx context.Context, id int, params models.CreateControllerParams) (*models.Controller, error) {
	var controller models.Controller

	err := s.pool.QueryRow(ctx, `
        UPDATE controllers
        SET name = $1, initials = $2, email = $3, facility_id = $4
        WHERE id = $5
        RETURNING id, created_at, name, initials, email, facility_id
    `, params.Name, params.Initials, params.Email, params.FacilityID, id).Scan(
		&controller.ID,
		&controller.CreatedAt,
		&controller.Name,
		&controller.Initials,
		&controller.Email,
		&controller.FacilityID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating controller: %w", err)
	}

	return &controller, nil
}

// DeleteController deletes a controller by ID
func (s *Service) DeleteController(ctx context.Context, id int) error {
	result, err := s.pool.Exec(ctx, `
        DELETE FROM controllers
        WHERE id = $1
    `, id)
	if err != nil {
		return fmt.Errorf("error deleting controller: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("controller with ID %d not found", id)
	}

	return nil
}
