// services/db/models/facility.go
package models

import "time"

// Facility represents a facility in the database
type Facility struct {
    ID        int       `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    Name      string    `json:"name"`
    Code      string    `json:"code"`
}

// CreateFacilityParams holds the parameters needed to create a new facility
type CreateFacilityParams struct {
    Name string `json:"name"`
    Code string `json:"code"`
}