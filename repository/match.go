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
	_, err := r.pool.Exec(ctx, "INSERT INTO matches (sport_id, home_id, away_id, home_score, away_score, round_id, next_round_id, round, state, start_date, winner_id, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)", match.SportID, match.HomeID, match.AwayID, match.HomeScore, match.AwayScore, match.RoundID, match.NextRoundID, match.Round, match.State, match.StartDate, match.WinnerID, match.ImageUrl, time.Now(), time.Now())
	return err
}

func (r *matchRepository) GetAll(ctx context.Context) ([]model.Match, error) {
	var matches []model.Match
	query := `
		SELECT 
			m.id, m.sport_id, m.home_id, m.away_id, m.home_score, m.away_score, 
			m.round_id, m.next_round_id, m.round, m.state, m.start_date, m.winner_id, m.image_url, 
			m.created_at, m.updated_at,
			t1.name as home_name, t1.company_name as home_company,
			t2.name as away_name, t2.company_name as away_company
		FROM matches m
		LEFT JOIN teams t1 ON m.home_id = t1.id
		LEFT JOIN teams t2 ON m.away_id = t2.id`
	
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		log.Print(err.Error())
		return matches, err
	}
	defer rows.Close()

	for rows.Next() {
		var match model.Match
		if err := rows.Scan(
			&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, 
			&match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.WinnerID, &match.ImageUrl, 
			&match.CreatedAt, &match.UpdatedAt,
			&match.HomeName, &match.HomeCompany, &match.AwayName, &match.AwayCompany,
		); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (r *matchRepository) GetBySportID(ctx context.Context, sportID int) ([]model.Match, error) {
	var matches []model.Match
	query := `
		SELECT 
			m.id, m.sport_id, m.home_id, m.away_id, m.home_score, m.away_score, 
			m.round_id, m.next_round_id, m.round, m.state, m.start_date, m.winner_id, m.image_url, 
			m.created_at, m.updated_at,
			t1.name as home_name, t1.company_name as home_company,
			t2.name as away_name, t2.company_name as away_company
		FROM matches m
		LEFT JOIN teams t1 ON m.home_id = t1.id
		LEFT JOIN teams t2 ON m.away_id = t2.id
		WHERE m.sport_id = $1 
		ORDER BY m.round_id DESC`

	rows, err := r.pool.Query(ctx, query, sportID)
	if err != nil {
		log.Print(err.Error())
		return matches, err
	}
	defer rows.Close()

	for rows.Next() {
		var match model.Match
		if err := rows.Scan(
			&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, 
			&match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.WinnerID, &match.ImageUrl, 
			&match.CreatedAt, &match.UpdatedAt,
			&match.HomeName, &match.HomeCompany, &match.AwayName, &match.AwayCompany,
		); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (r *matchRepository) GetByID(ctx context.Context, id int) (model.Match, error) {
	var match model.Match
	query := `
		SELECT 
			m.id, m.sport_id, m.home_id, m.away_id, m.home_score, m.away_score, 
			m.round_id, m.next_round_id, m.round, m.state, m.start_date, m.winner_id, m.image_url, 
			m.created_at, m.updated_at,
			t1.name as home_name, t1.company_name as home_company,
			t2.name as away_name, t2.company_name as away_company
		FROM matches m
		LEFT JOIN teams t1 ON m.home_id = t1.id
		LEFT JOIN teams t2 ON m.away_id = t2.id
		WHERE m.id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&match.ID, &match.SportID, &match.HomeID, &match.AwayID, &match.HomeScore, &match.AwayScore, 
		&match.RoundID, &match.NextRoundID, &match.Round, &match.State, &match.StartDate, &match.WinnerID, &match.ImageUrl, 
		&match.CreatedAt, &match.UpdatedAt,
		&match.HomeName, &match.HomeCompany, &match.AwayName, &match.AwayCompany,
	)
	return match, err
}

func (r *matchRepository) Update(ctx context.Context, match model.Match) error {
	_, err := r.pool.Exec(ctx, "UPDATE matches SET sport_id = $1, home_id = $2, away_id = $3, home_score = $4, away_score = $5, round_id = $6, next_round_id = $7, round = $8, state = $9, start_date = $10, winner_id = $11, image_url = $12, updated_at = $13 WHERE id = $14", match.SportID, match.HomeID, match.AwayID, match.HomeScore, match.AwayScore, match.RoundID, match.NextRoundID, match.Round, match.State, match.StartDate, match.WinnerID, match.ImageUrl, time.Now(), match.ID)
	return err
}

func (r *matchRepository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM matches WHERE id = $1", id)
	return err
}

func (r *matchRepository) DeleteBySportID(ctx context.Context, sportID int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM matches WHERE sport_id = $1", sportID)
	return err
}