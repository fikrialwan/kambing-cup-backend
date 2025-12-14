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
	HomeID      int         `json:"home_id" validate:"required"`
	AwayID      int         `json:"away_id" validate:"required"`
	HomeScore   *string     `json:"home_score"`
	AwayScore   *string     `json:"away_score"`
	RoundID     int         `json:"round_id" validate:"required"`
	NextRoundID *int        `json:"next_round_id"`
	Round       string      `json:"round" validate:"required"`
	State       MatchState  `json:"state" validate:"required"`
	StartDate   time.Time   `json:"start_date" validate:"required"`
	Winner      *string     `json:"winner"`
	CreatedAt   time.Time   `json:"-"`
	UpdatedAt   time.Time   `json:"-"`
	DeletedAt   *time.Time  `json:"-"`
}
