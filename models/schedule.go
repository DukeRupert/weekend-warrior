// models/schedule.go
package models

import (
	"time"
)

type Schedule struct {
	ID           int       `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	RDOs         []int     `json:"rdos"`
	Anchor       time.Time `json:"anchor"`
	ControllerID int       `json:"controller_id"`
}

type CreateScheduleParams struct {
	RDOs         []int     `json:"rdos"`
	Anchor       time.Time `json:"anchor"`
	ControllerID int       `json:"controller_id"`
}

type UpdateScheduleParams struct {
	RDOs   []int     `json:"rdos"`
	Anchor time.Time `json:"anchor"`
}
