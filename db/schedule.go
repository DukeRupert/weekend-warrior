// db/facilities.go
package db

import (
	"context"
	"fmt"

	"github.com/dukerupert/weekend-warrior/models"
)

// CreateSchedule creates a new schedule in the database
func (s *Service) CreateSchedule(ctx context.Context, params models.CreateScheduleParams) (*models.Schedule, error) {
	var schedule models.Schedule
	err := s.pool.QueryRow(ctx, `
        INSERT INTO schedules (rdos, anchor, controller_id)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, rdos, anchor, controller_id
    `, params.RDOs, params.Anchor, params.ControllerID).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.RDOs,
		&schedule.Anchor,
		&schedule.ControllerID,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating schedule: %w", err)
	}
	return &schedule, nil
}

// GetSchedule retrieves a schedule by ID from the database
func (s *Service) GetSchedule(ctx context.Context, id int) (*models.Schedule, error) {
	var schedule models.Schedule
	err := s.pool.QueryRow(ctx, `
        SELECT id, created_at, rdos, anchor, controller_id
        FROM schedules
        WHERE id = $1
    `, id).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.RDOs,
		&schedule.Anchor,
		&schedule.ControllerID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting schedule: %w", err)
	}
	return &schedule, nil
}

// GetScheduleByController retrieves a schedule by controller ID from the database
func (s *Service) GetScheduleByController(ctx context.Context, controllerID int) (*models.Schedule, error) {
	var schedule models.Schedule
	err := s.pool.QueryRow(ctx, `
        SELECT id, created_at, rdos, anchor, controller_id
        FROM schedules
        WHERE controller_id = $1
    `, controllerID).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.RDOs,
		&schedule.Anchor,
		&schedule.ControllerID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting schedule by controller: %w", err)
	}
	return &schedule, nil
}

// UpdateSchedule updates an existing schedule in the database
func (s *Service) UpdateSchedule(ctx context.Context, id int, params models.UpdateScheduleParams) (*models.Schedule, error) {
	var schedule models.Schedule
	err := s.pool.QueryRow(ctx, `
        UPDATE schedules
        SET rdos = $1, anchor = $2
        WHERE id = $3
        RETURNING id, created_at, rdos, anchor, controller_id
    `, params.RDOs, params.Anchor, id).Scan(
		&schedule.ID,
		&schedule.CreatedAt,
		&schedule.RDOs,
		&schedule.Anchor,
		&schedule.ControllerID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating schedule: %w", err)
	}
	return &schedule, nil
}

// DeleteSchedule deletes a schedule from the database
func (s *Service) DeleteSchedule(ctx context.Context, id int) error {
	commandTag, err := s.pool.Exec(ctx, `
        DELETE FROM schedules
        WHERE id = $1
    `, id)
	if err != nil {
		return fmt.Errorf("error deleting schedule: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("schedule not found")
	}
	return nil
}
