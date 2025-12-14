package model

import (
	"time"
)

type Sport struct {
	ID           int        `json:"id"`
	TournamentID int        `json:"tournament_id" validate:"required"`
	Name         string     `json:"name" validate:"required"`
	Slug         string     `json:"slug"`
	ImageUrl     string     `json:"image_url"`
	CreatedAt    time.Time  `json:"-"`
	UpdatedAt    time.Time  `json:"-"`
	DeletedAt    *time.Time `json:"-"`
}
