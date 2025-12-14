package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type SportRepository struct {
	conn *pgx.Conn
}

func NewSportRepository(conn *pgx.Conn) *SportRepository {
	return &SportRepository{conn: conn}
}

func (r *SportRepository) Create(sport model.Sport) error {
	_, err := r.conn.Exec(context.Background(), "INSERT INTO sports (tournament_id, name, slug, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", sport.TournamentID, sport.Name, sport.Slug, sport.ImageUrl, time.Now(), time.Now())
	return err
}

func (r *SportRepository) GetAll() ([]model.Sport, error) {
	var sports []model.Sport
	rows, err := r.conn.Query(context.Background(), "SELECT id, tournament_id, name, slug, image_url, created_at, updated_at, deleted_at FROM sports WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return sports, err
	}
	defer rows.Close()

	for rows.Next() {
		var sport model.Sport
		if err := rows.Scan(&sport.ID, &sport.TournamentID, &sport.Name, &sport.Slug, &sport.ImageUrl, &sport.CreatedAt, &sport.UpdatedAt, &sport.DeletedAt); err != nil {
			log.Print(err.Error())
			return nil, err
		}
		sports = append(sports, sport)
	}

	return sports, nil
}

func (r *SportRepository) GetByID(id int) (model.Sport, error) {
	var sport model.Sport
	err := r.conn.QueryRow(context.Background(), "SELECT id, tournament_id, name, slug, image_url, created_at, updated_at, deleted_at FROM sports WHERE id = $1 AND deleted_at IS NULL", id).Scan(&sport.ID, &sport.TournamentID, &sport.Name, &sport.Slug, &sport.ImageUrl, &sport.CreatedAt, &sport.UpdatedAt, &sport.DeletedAt)
	return sport, err
}

func (r *SportRepository) Update(sport model.Sport) error {
	_, err := r.conn.Exec(context.Background(), "UPDATE sports SET tournament_id = $1, name = $2, slug = $3, image_url = $4, updated_at = $5 WHERE id = $6", sport.TournamentID, sport.Name, sport.Slug, sport.ImageUrl, time.Now(), sport.ID)
	return err
}

func (r *SportRepository) Delete(id int) error {
	_, err := r.conn.Exec(context.Background(), "UPDATE sports SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}
