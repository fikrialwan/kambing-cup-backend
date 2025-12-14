package model

import (
	"time"
)

type Team struct {
	ID        int        `json:"id"`
	SportID   int        `json:"sport_id" validate:"required"`
	Name      string     `json:"name" validate:"required"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}
