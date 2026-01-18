package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SportRepository interface {
	Create(sport model.Sport) error
	GetAll() ([]model.Sport, error)
	GetByID(id int) (model.Sport, error)
	Update(sport model.Sport) error
	Delete(id int) error
}

type sportRepository struct {
	pool *pgxpool.Pool
}

func NewSportRepository(pool *pgxpool.Pool) SportRepository {
	return &sportRepository{pool: pool}
}

func (r *sportRepository) Create(sport model.Sport) error {
	_, err := r.pool.Exec(context.Background(), "INSERT INTO sports (tournament_id, name, slug, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", sport.TournamentID, sport.Name, sport.Slug, sport.ImageUrl, time.Now(), time.Now())
	return err
}

func (r *sportRepository) GetAll() ([]model.Sport, error) {
	var sports []model.Sport
	rows, err := r.pool.Query(context.Background(), "SELECT id, tournament_id, name, slug, image_url, created_at, updated_at, deleted_at FROM sports WHERE deleted_at IS NULL")
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

func (r *sportRepository) GetByID(id int) (model.Sport, error) {
	var sport model.Sport
	err := r.pool.QueryRow(context.Background(), "SELECT id, tournament_id, name, slug, image_url, created_at, updated_at, deleted_at FROM sports WHERE id = $1 AND deleted_at IS NULL", id).Scan(&sport.ID, &sport.TournamentID, &sport.Name, &sport.Slug, &sport.ImageUrl, &sport.CreatedAt, &sport.UpdatedAt, &sport.DeletedAt)
	return sport, err
}

func (r *sportRepository) Update(sport model.Sport) error {
	_, err := r.pool.Exec(context.Background(), "UPDATE sports SET tournament_id = $1, name = $2, slug = $3, image_url = $4, updated_at = $5 WHERE id = $6", sport.TournamentID, sport.Name, sport.Slug, sport.ImageUrl, time.Now(), sport.ID)
	return err
}

func (r *sportRepository) Delete(id int) error {
	_, err := r.pool.Exec(context.Background(), "UPDATE sports SET deleted_at = $1 WHERE id = $2", time.Now(), id)
	return err
}