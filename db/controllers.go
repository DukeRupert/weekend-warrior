// db/controllers.go
package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dukerupert/weekend-warrior/db/models"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// CreateController creates a new controller in the database
func (s *Service) CreateController(ctx context.Context, params models.CreateControllerParams) (*models.Controller, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("name", params.Name).
		Str("email", params.Email).
		Int("facility_id", params.FacilityID).
		Str("operation", "CreateController").
		Logger()

	logger.Debug().Msg("creating new controller")

	// Hash the password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("failed to hash password")
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	var controller models.Controller
	err = s.pool.QueryRow(ctx, `
        INSERT INTO controllers (name, initials, email, facility_id, password)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, name, initials, email, facility_id
    `, params.Name, params.Initials, params.Email, params.FacilityID, hashedPassword).Scan(
		&controller.ID,
		&controller.CreatedAt,
		&controller.Name,
		&controller.Initials,
		&controller.Email,
		&controller.FacilityID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create controller in database")
		return nil, fmt.Errorf("error creating controller: %w", err)
	}

	logger.Info().
		Int("controller_id", controller.ID).
		Time("created_at", controller.CreatedAt).
		Msg("successfully created controller")

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

// GetControllerByEmail retrieves a controller by their email address
func (s *Service) GetControllerByEmail(ctx context.Context, email string) (*models.Controller, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("email", email).
		Str("operation", "GetControllerByEmail").
		Logger()

	logger.Debug().Msg("retrieving controller by email")

	var controller models.Controller
	err := s.pool.QueryRow(ctx, `
        SELECT id, created_at, name, initials, email, facility_id, password
        FROM controllers
        WHERE email = $1
    `, email).Scan(
		&controller.ID,
		&controller.CreatedAt,
		&controller.Name,
		&controller.Initials,
		&controller.Email,
		&controller.FacilityID,
		&controller.Password,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Debug().Msg("controller not found")
			return nil, fmt.Errorf("controller not found with email %s", email)
		}
		logger.Error().Err(err).Msg("failed to get controller from database")
		return nil, fmt.Errorf("error getting controller: %w", err)
	}

	logger.Debug().
		Int("controller_id", controller.ID).
		Time("created_at", controller.CreatedAt).
		Msg("successfully retrieved controller")

	return &controller, nil
}

// GetControllersByFacilityCode retrieves all controllers for a facility
func (s *Service) GetControllersByFacilityCode(ctx context.Context, facilityCode string) ([]models.Controller, error) {
	rows, err := s.pool.Query(ctx, `
	SELECT 
		c.id,
		c.created_at,
		c.updated_at,
		c.name,
		c.initials,
		c.email,
		c.role,
		c.facility_id
	FROM 
		controllers c
		INNER JOIN facilities f ON c.facility_id = f.id
	WHERE 
		f.code = $1
    `, facilityCode)
	if err != nil {
		return nil, fmt.Errorf("error listing controllers at %s: %w",facilityCode, err)
	}
	defer rows.Close()

	var controllers []models.Controller
	for rows.Next() {
		var c models.Controller
		err := rows.Scan(
			&c.ID,
            &c.CreatedAt,
            &c.UpdatedAt,
            &c.Name,
            &c.Initials,
            &c.Email,
            &c.Role,
			&c.FacilityID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning controller row: %w", err)
		}
		controllers = append(controllers, c)
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
	logger := zerolog.Ctx(ctx).With().
		Int("controller_id", id).
		Str("name", params.Name).
		Str("email", params.Email).
		Int("facility_id", params.FacilityID).
		Str("operation", "UpdateController").
		Logger()

	logger.Debug().Msg("updating controller")

	// Build the query dynamically based on whether a password update is needed
	var query strings.Builder
	var args []interface{}
	query.WriteString(`
		UPDATE controllers
		SET name = $1, initials = $2, email = $3, facility_id = $4`)
	args = append(args, params.Name, params.Initials, params.Email, params.FacilityID)

	// Only update password if it's provided (not empty)
	if params.Password != "" {
		logger.Debug().Msg("updating password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error().Err(err).Msg("failed to hash password")
			return nil, fmt.Errorf("error hashing password: %w", err)
		}
		query.WriteString(`, password = $5`)
		args = append(args, hashedPassword)
		args = append(args, id)
		query.WriteString(`
			WHERE id = $6
			RETURNING id, created_at, name, initials, email, facility_id`)
	} else {
		args = append(args, id)
		query.WriteString(`
			WHERE id = $5
			RETURNING id, created_at, name, initials, email, facility_id`)
	}

	var controller models.Controller
	err := s.pool.QueryRow(ctx, query.String(), args...).Scan(
		&controller.ID,
		&controller.CreatedAt,
		&controller.Name,
		&controller.Initials,
		&controller.Email,
		&controller.FacilityID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Error().Msg("controller not found")
			return nil, fmt.Errorf("controller not found with id %d", id)
		}
		logger.Error().Err(err).Msg("failed to update controller in database")
		return nil, fmt.Errorf("error updating controller: %w", err)
	}

	logger.Info().
		Time("updated_at", time.Now()).
		Bool("password_updated", params.Password != "").
		Msg("successfully updated controller")

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
