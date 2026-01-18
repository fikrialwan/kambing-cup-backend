package model

import (
	"time"
)

type Tournament struct {
	ID        int        `json:"id"`
	Name      string     `json:"name" validate:"required"`
	Slug      string     `json:"slug" validate:"required"`
	IsShow    bool       `json:"is_show" validate:"required"`
	IsActive  bool       `json:"is_active" validate:"required"`
	ImageUrl  string     `json:"image_url" validate:"required"`
	TotalSurah int       `json:"total_surah" validate:"required"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}
