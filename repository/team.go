package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository interface {
	Create(team model.Team) error
	GetAll() ([]model.Team, error)
	GetByID(id int) (model.Team, error)
	Update(team model.Team) error
	Delete(id int) error
}

type teamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) TeamRepository {
	return &teamRepository{pool: pool}
}

func (r *teamRepository) Create(team model.Team) error {
	_, err := r.pool.Exec(context.Background(), "INSERT INTO teams (sport_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)", team.SportID, team.Name, time.Now(), time.Now())
	return err
}

func (r *teamRepository) GetAll() ([]model.Team, error) {
	var teams []model.Team
	rows, err := r.pool.Query(context.Background(), "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE deleted_at IS NULL")
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

func (r *teamRepository) GetByID(id int) (model.Team, error) {
	var team model.Team
	err := r.pool.QueryRow(context.Background(), "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE id = $1 AND deleted_at IS NULL", id).Scan(&team.ID, &team.SportID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.DeletedAt)
	return team, err
}

func (r *teamRepository) Update(team model.Team) error {
	_, err := r.pool.Exec(context.Background(), "UPDATE teams SET sport_id = $1, name = $2, updated_at = $3 WHERE id = $4", team.SportID, team.Name, time.Now(), team.ID)
	return err
}

func (r *teamRepository) Delete(id int) error {
	_, err := r.pool.Exec(context.Background(), "UPDATE teams SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}