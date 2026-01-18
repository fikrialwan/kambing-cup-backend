package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MatchRepository interface {
	Create(match model.Match) error
	GetAll() ([]model.Match, error)
	GetByID(id int) (model.Match, error)
	Update(match model.Match) error
	Delete(id int) error
}

type matchRepository struct {
	pool *pgxpool.Pool
}

func NewMatchRepository(pool *pgxpool.Pool) MatchRepository {
	return &matchRepository{pool: pool}
}

func (r *matchRepository) Create(match model.Match) error {
	_, err := r.pool.Exec(context.Background(), "INSERT INTO matches (sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)", match.SportID, match.HomeID, match.AwayID, match.HomeScore, match.AwayScore, match.RoundID, match.NextRoundID, match.Round, match.State, match.StartDate, match.Winner, time.Now(), time.Now())
	return err
}

func (r *matchRepository) GetAll() ([]model.Match, error) {
	var matches []model.Match
	rows, err := r.pool.Query(context.Background(), "SELECT id, sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, created_at, updated_at, deleted_at FROM matches WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return matches, err
	}
	defer rows.Close()

	for rows.Next() {
		var match model.Match
		if err := rows.Scan(&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, &match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.Winner, &match.CreatedAt, &match.UpdatedAt, &match.DeletedAt); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (r *matchRepository) GetByID(id int) (model.Match, error) {
	var match model.Match
	err := r.pool.QueryRow(context.Background(), "SELECT id, sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, created_at, updated_at, deleted_at FROM matches WHERE id = $1 AND deleted_at IS NULL", id).Scan(&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, &match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.Winner, &match.CreatedAt, &match.UpdatedAt, &match.DeletedAt)
	return match, err
}

func (r *matchRepository) Update(match model.Match) error {
	_, err := r.pool.Exec(context.Background(), "UPDATE matches SET sport_id = $1, home_id = $2, away_id = $3, home_score = $4, away_score = $5, round_id = $6, next_round_id = $7, round = $8, state = $9, start_date = $10, winner = $11, updated_at = $12 WHERE id = $13", match.SportID, match.HomeID, match.AwayID, match.HomeScore, match.AwayScore, match.RoundID, match.NextRoundID, match.Round, match.State, match.StartDate, match.Winner, time.Now(), match.ID)
	return err
}

func (r *matchRepository) Delete(id int) error {
	_, err := r.pool.Exec(context.Background(), "UPDATE matches SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}