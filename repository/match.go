package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MatchRepository interface {
	Create(ctx context.Context, match model.Match) error
	GetAll(ctx context.Context) ([]model.Match, error)
	GetBySportID(ctx context.Context, sportID int) ([]model.Match, error)
	GetByID(ctx context.Context, id int) (model.Match, error)
	Update(ctx context.Context, match model.Match) error
	Delete(ctx context.Context, id int) error
	DeleteBySportID(ctx context.Context, sportID int) error
}

type matchRepository struct {
	pool *pgxpool.Pool
}

func NewMatchRepository(pool *pgxpool.Pool) MatchRepository {
	return &matchRepository{pool: pool}
}

func (r *matchRepository) Create(ctx context.Context, match model.Match) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO matches (sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", match.SportID, match.HomeID, match.AwayID, match.HomeScore, match.AwayScore, match.RoundID, match.NextRoundID, match.Round, match.State, match.StartDate, match.Winner, match.ImageUrl, time.Now(), time.Now())
	return err
}

func (r *matchRepository) GetAll(ctx context.Context) ([]model.Match, error) {
	var matches []model.Match
	rows, err := r.pool.Query(ctx, "SELECT id, sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, image_url, created_at, updated_at, deleted_at FROM matches WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return matches, err
	}
	defer rows.Close()

	for rows.Next() {
		var match model.Match
		if err := rows.Scan(&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, &match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.Winner, &match.ImageUrl, &match.CreatedAt, &match.UpdatedAt, &match.DeletedAt); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (r *matchRepository) GetBySportID(ctx context.Context, sportID int) ([]model.Match, error) {
	var matches []model.Match
	rows, err := r.pool.Query(ctx, "SELECT id, sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, image_url, created_at, updated_at, deleted_at FROM matches WHERE sport_id = $1 AND deleted_at IS NULL ORDER BY round_id DESC", sportID)
	if err != nil {
		log.Print(err.Error())
		return matches, err
	}
	defer rows.Close()

	for rows.Next() {
		var match model.Match
		if err := rows.Scan(&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, &match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.Winner, &match.ImageUrl, &match.CreatedAt, &match.UpdatedAt, &match.DeletedAt); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (r *matchRepository) GetByID(ctx context.Context, id int) (model.Match, error) {
	var match model.Match
	err := r.pool.QueryRow(ctx, "SELECT id, sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner, image_url, created_at, updated_at, deleted_at FROM matches WHERE id = $1 AND deleted_at IS NULL", id).Scan(&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, &match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.Winner, &match.ImageUrl, &match.CreatedAt, &match.UpdatedAt, &match.DeletedAt)
	return match, err
}

func (r *matchRepository) Update(ctx context.Context, match model.Match) error {
	_, err := r.pool.Exec(ctx, "UPDATE matches SET sport_id = $1, home_id = $2, away_id = $3, home_score = $4, away_score = $5, round_id = $6, next_round_id = $7, round = $8, state = $9, start_date = $10, winner = $11, image_url = $12, updated_at = $13 WHERE id = $14", match.SportID, match.HomeID, match.AwayID, match.HomeScore, match.AwayScore, match.RoundID, match.NextRoundID, match.Round, match.State, match.StartDate, match.Winner, match.ImageUrl, time.Now(), match.ID)
	return err
}

func (r *matchRepository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "UPDATE matches SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (r *matchRepository) DeleteBySportID(ctx context.Context, sportID int) error {
	_, err := r.pool.Exec(ctx, "UPDATE matches SET deleted_at = $1 WHERE sport_id = $2 AND deleted_at IS NULL", time.Now(), sportID)
	return err
}