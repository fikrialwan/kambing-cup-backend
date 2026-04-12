package model

import (
	"time"
)

type MatchState string

const (
	SOON MatchState = "SOON"
	LIVE MatchState = "LIVE"
	DONE MatchState = "DONE"
)

type Match struct {
	ID          int         `json:"id"`
	SportID     int         `json:"sport_id" validate:"required"`
	HomeID      *int        `json:"home_id"`
	AwayID      *int        `json:"away_id"`
	HomeScore   *string     `json:"home_score"`
	AwayScore   *string     `json:"away_score"`
	RoundID     int         `json:"round_id" validate:"required"`
	NextRoundID *int        `json:"next_round_id"`
	Round       string      `json:"round" validate:"required"`
	State       MatchState  `json:"state" validate:"required"`
	StartDate   time.Time   `json:"start_date" validate:"required"`
	WinnerID    *int        `json:"winner_id"`
	ImageUrl    *string     `json:"image_url"`
	HomeName    *string     `json:"home_name,omitzero"`
	HomeCompany *string     `json:"home_company,omitzero"`
	AwayName    *string     `json:"away_name,omitzero"`
	AwayCompany *string     `json:"away_company,omitzero"`
	CreatedAt   time.Time   `json:"-"`
	UpdatedAt   time.Time   `json:"-"`
}
