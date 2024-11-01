// services/db/facilities.go
package db

import (
    "context"
    "fmt"
    "github.com/dukerupert/weekend-warrior/models"
)

// CreateFacility creates a new facility in the database
func (s *Service) CreateFacility(ctx context.Context, params models.CreateFacilityParams) (*models.Facility, error) {
    var facility models.Facility
    
    err := s.pool.QueryRow(ctx, `
        INSERT INTO facilities (name, code)
        VALUES ($1, $2)
        RETURNING id, created_at, name, code
    `, params.Name, params.Code).Scan(
        &facility.ID,
        &facility.CreatedAt,
        &facility.Name,
        &facility.Code,
    )

    if err != nil {
        return nil, fmt.Errorf("error creating facility: %w", err)
    }

    return &facility, nil
}

// GetFacilityByID retrieves a facility by its ID
func (s *Service) GetFacilityByID(ctx context.Context, id int) (*models.Facility, error) {
    var facility models.Facility
    
    err := s.pool.QueryRow(ctx, `
        SELECT id, created_at, name, code
        FROM facilities
        WHERE id = $1
    `, id).Scan(
        &facility.ID,
        &facility.CreatedAt,
        &facility.Name,
        &facility.Code,
    )

    if err != nil {
        return nil, fmt.Errorf("error getting facility: %w", err)
    }

    return &facility, nil
}

// GetFacilityByCode retrieves a facility by its code
func (s *Service) GetFacilityByCode(ctx context.Context, code string) (*models.Facility, error) {
    var facility models.Facility
    
    err := s.pool.QueryRow(ctx, `
        SELECT id, created_at, name, code
        FROM facilities
        WHERE code = $1
    `, code).Scan(
        &facility.ID,
        &facility.CreatedAt,
        &facility.Name,
        &facility.Code,
    )

    if err != nil {
        return nil, fmt.Errorf("error getting facility by code: %w", err)
    }

    return &facility, nil
}

// ListFacilities retrieves all facilities from the database
func (s *Service) ListFacilities(ctx context.Context) ([]models.Facility, error) {
    rows, err := s.pool.Query(ctx, `
        SELECT id, created_at, name, code
        FROM facilities
        ORDER BY name ASC
    `)
    if err != nil {
        return nil, fmt.Errorf("error listing facilities: %w", err)
    }
    defer rows.Close()

    var facilities []models.Facility
    for rows.Next() {
        var facility models.Facility
        err := rows.Scan(
            &facility.ID,
            &facility.CreatedAt,
            &facility.Name,
            &facility.Code,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning facility row: %w", err)
        }
        facilities = append(facilities, facility)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating facility rows: %w", err)
    }

    return facilities, nil
}

// DeleteFacility deletes a facility by its ID
func (s *Service) DeleteFacility(ctx context.Context, id int) error {
    result, err := s.pool.Exec(ctx, `
        DELETE FROM facilities
        WHERE id = $1
    `, id)
    if err != nil {
        return fmt.Errorf("error deleting facility: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("facility with ID %d not found", id)
    }

    return nil
}

// DeleteFacilityByCode deletes a facility by its code
func (s *Service) DeleteFacilityByCode(ctx context.Context, code string) error {
    result, err := s.pool.Exec(ctx, `
        DELETE FROM facilities
        WHERE code = $1
    `, code)
    if err != nil {
        return fmt.Errorf("error deleting facility: %w", err)
    }

    if result.RowsAffected() == 0 {
        return fmt.Errorf("facility with code %s not found", code)
    }

    return nil
}