package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type TeamRepository struct {
	conn *pgx.Conn
}

func NewTeamRepository(conn *pgx.Conn) *TeamRepository {
	return &TeamRepository{conn: conn}
}

func (r *TeamRepository) Create(team model.Team) error {
	_, err := r.conn.Exec(context.Background(), "INSERT INTO teams (sport_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)", team.SportID, team.Name, time.Now(), time.Now())
	return err
}

func (r *TeamRepository) GetAll() ([]model.Team, error) {
	var teams []model.Team
	rows, err := r.conn.Query(context.Background(), "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE deleted_at IS NULL")
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

func (r *TeamRepository) GetByID(id int) (model.Team, error) {
	var team model.Team
	err := r.conn.QueryRow(context.Background(), "SELECT id, sport_id, name, created_at, updated_at, deleted_at FROM teams WHERE id = $1 AND deleted_at IS NULL", id).Scan(&team.ID, &team.SportID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.DeletedAt)
	return team, err
}

func (r *TeamRepository) Update(team model.Team) error {
	_, err := r.conn.Exec(context.Background(), "UPDATE teams SET sport_id = $1, name = $2, updated_at = $3 WHERE id = $4", team.SportID, team.Name, time.Now(), team.ID)
	return err
}

func (r *TeamRepository) Delete(id int) error {
	_, err := r.conn.Exec(context.Background(), "UPDATE teams SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}
