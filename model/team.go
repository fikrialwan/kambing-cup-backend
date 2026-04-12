package model

import (
	"time"
)

type Team struct {
	ID          int        `json:"id"`
	SportID     int        `json:"sport_id" validate:"required"`
	Name        string     `json:"name" validate:"required"`
	CompanyName string     `json:"company_name"`
	CreatedAt   time.Time  `json:"-"`
	UpdatedAt   time.Time  `json:"-"`
}
