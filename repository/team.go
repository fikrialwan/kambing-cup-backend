package repository

import (
	"context"
	"log"
	"time"

	"kambing-cup-backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository interface {
	Create(ctx context.Context, team model.Team) error
	GetAll(ctx context.Context) ([]model.Team, error)
	GetByID(ctx context.Context, id int) (model.Team, error)
	Update(ctx context.Context, team model.Team) error
	Delete(ctx context.Context, id int) error
	GetByNameAndSport(ctx context.Context, name string, sportID int) (model.Team, error)
	CreateBulk(ctx context.Context, teams []model.Team) error
}

type teamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &teamRepository{pool: pool}
}

func (r *teamRepository) Create(ctx context.Context, team model.Team) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO teams (sport_id, name, company_name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", team.SportID, team.Name, team.CompanyName, time.Now(), time.Now())
	return err
}

func (r *teamRepository) CreateBulk(ctx context.Context, teams []model.Team) error {
	now := time.Now()
	rows := [][]interface{}{}
	for _, team := range teams {
		rows = append(rows, []interface{}{team.SportID, team.Name, team.CompanyName, now, now})
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"teams"},
		[]string{"sport_id", "name", "company_name", "created_at", "updated_at"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (r *teamRepository) GetAll(ctx context.Context) ([]model.Team, error) {
	var teams []model.Team
	rows, err := r.pool.Query(ctx, "SELECT id, sport_id, name, company_name, created_at, updated_at FROM teams")
	if err != nil {
		log.Print(err.Error())
		return teams, err
	}
	defer rows.Close()

	for rows.Next() {
		var team model.Team
		if err := rows.Scan(&team.ID, &team.SportID, &team.Name, &team.CompanyName, &team.CreatedAt, &team.UpdatedAt); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (r *teamRepository) GetByID(ctx context.Context, id int) (model.Team, error) {
	var team model.Team
	err := r.pool.QueryRow(ctx, "SELECT id, sport_id, name, company_name, created_at, updated_at FROM teams WHERE id = $1", id).Scan(&team.ID, &team.SportID, &team.Name, &team.CompanyName, &team.CreatedAt, &team.UpdatedAt)
	return team, err
}

func (r *teamRepository) Update(ctx context.Context, team model.Team) error {
	_, err := r.pool.Exec(ctx, "UPDATE teams SET sport_id = $1, name = $2, company_name = $3, updated_at = $4 WHERE id = $5", team.SportID, team.Name, team.CompanyName, time.Now(), team.ID)
	return err
}

func (r *teamRepository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM teams WHERE id = $1", id)
	return err
}

func (r *teamRepository) GetByNameAndSport(ctx context.Context, name string, sportID int) (model.Team, error) {
	var team model.Team
	err := r.pool.QueryRow(ctx, "SELECT id, sport_id, name, company_name, created_at, updated_at FROM teams WHERE name = $1 AND sport_id = $2", name, sportID).Scan(&team.ID, &team.SportID, &team.Name, &team.CompanyName, &team.CreatedAt, &team.UpdatedAt)
	return team, err
}

