package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository interface {
	Create(ctx context.Context, team model.Team) error
	GetAll(ctx context.Context) ([]model.Team, error)
	GetByID(ctx context.Context, id int) (model.Team, error)
	Update(ctx context.Context, team model.Team) error
	Delete(ctx context.Context, id int) error
	GetByNameAndSportWithDeleted(ctx context.Context, name string, sportID int) (model.Team, error)
	Restore(ctx context.Context, team model.Team) error
}

type teamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &teamRepository{pool: pool}
}

func (r *teamRepository) Create(ctx context.Context, team model.Team) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO teams (sport_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)", team.SportID, team.Name, time.Now(), time.Now())
	return err
}

func (r *teamRepository) GetAll(ctx context.Context) ([]model.Team, error) {
	var teams []model.Team
	rows, err := r.pool.Query(ctx, "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return teams, err
	}
	defer rows.Close()

	for rows.Next() {
		var team model.Team
		if err := rows.Scan(&team.ID, &team.SportID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.DeletedAt); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (r *teamRepository) GetByID(ctx context.Context, id int) (model.Team, error) {
	var team model.Team
	err := r.pool.QueryRow(ctx, "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE id = $1 AND deleted_at IS NULL", id).Scan(&team.ID, &team.SportID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.DeletedAt)
	return team, err
}

func (r *teamRepository) Update(ctx context.Context, team model.Team) error {
	_, err := r.pool.Exec(ctx, "UPDATE teams SET sport_id = $1, name = $2, updated_at = $3 WHERE id = $4", team.SportID, team.Name, time.Now(), team.ID)
	return err
}

func (r *teamRepository) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "UPDATE teams SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (r *teamRepository) GetByNameAndSportWithDeleted(ctx context.Context, name string, sportID int) (model.Team, error) {
	var team model.Team
	err := r.pool.QueryRow(ctx, "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE name = $1 AND sport_id = $2", name, sportID).Scan(&team.ID, &team.SportID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.DeletedAt)
	return team, err
}

func (r *teamRepository) Restore(ctx context.Context, team model.Team) error {
	_, err := r.pool.Exec(ctx, "UPDATE teams SET updated_at = $1, deleted_at = NULL WHERE id = $2", time.Now(), team.ID)
	return err
}