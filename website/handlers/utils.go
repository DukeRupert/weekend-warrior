// handlers/error_helpers.go
package handlers

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

// isNotFoundError checks if the error indicates a resource was not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for pgx "no rows" error
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}

	// Check for our custom "not found" error messages
	errMsg := err.Error()
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "no rows in result set")
}

// isDuplicateKeyError checks if the error indicates a unique constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key value violates unique constraint") ||
		strings.Contains(errMsg, "already exists")
}

// isForeignKeyError checks if the error indicates a foreign key constraint violation
func isForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "foreign key constraint") ||
		strings.Contains(errMsg, "violates foreign key constraint")
}
