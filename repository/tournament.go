package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TournamentRepository struct {
	pool *pgxpool.Pool
}

func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{pool: pool}
}

func (T *TournamentRepository) GetAll() ([]model.Tournament, error) {
	var tournaments []model.Tournament
	rows, err := T.pool.Query(context.Background(), "SELECT * FROM tournaments WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return tournaments, err
	}

	for rows.Next() {
		var tournament model.Tournament
		if err := rows.Scan(&tournament.ID, &tournament.Name, &tournament.Slug, &tournament.IsShow, &tournament.IsActive, &tournament.ImageUrl, &tournament.CreatedAt, &tournament.UpdatedAt, &tournament.DeletedAt); err != nil {
			log.Print(err.Error())
			return tournaments, err
		}
		tournaments = append(tournaments, tournament)
	}

	return tournaments, err
}

func (T *TournamentRepository) Create(tournament model.Tournament) error {
	_, err := T.pool.Exec(context.Background(), "INSERT INTO tournaments (name, slug, is_show, is_active, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, tournament.ImageUrl, time.Now(), time.Now())

	return err
}

func (T *TournamentRepository) Update(tournament model.Tournament) error {
	if tournament.ImageUrl == "" {
		_, err := T.pool.Exec(context.Background(), "UPDATE tournaments SET name = $1, slug = $2, is_show = $3, is_active = $4, updated_at = $5 WHERE id = $6", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, time.Now(), tournament.ID)
		return err
	}

	_, err := T.pool.Exec(context.Background(), "UPDATE tournaments SET name = $1, slug = $2, is_show = $3, is_active = $4, image_url = $5, updated_at = $6 WHERE id = $7", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, tournament.ImageUrl, time.Now(), tournament.ID)

	return err
}

func (T *TournamentRepository) Delete(id int) error {
	_, err := T.pool.Exec(context.Background(), "UPDATE tournaments SET deleted_at = $1 WHERE id = $2", time.Now(), id)

	return err
}

func (T *TournamentRepository) GetBySlug(slug string) (model.Tournament, error) {
	var tournament model.Tournament
	err := T.pool.QueryRow(context.Background(), "SELECT * FROM tournaments WHERE slug = $1 AND deleted_at IS NULL", slug).Scan(&tournament.ID, &tournament.Name, &tournament.Slug, &tournament.IsShow, &tournament.IsActive, &tournament.ImageUrl, &tournament.CreatedAt, &tournament.UpdatedAt, &tournament.DeletedAt)
	return tournament, err
}
