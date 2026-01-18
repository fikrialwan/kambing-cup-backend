package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TournamentRepository interface {
	GetAll() ([]model.Tournament, error)
	Create(tournament model.Tournament) error
	Update(tournament model.Tournament) error
	Delete(id int) error
	GetBySlug(slug string) (model.Tournament, error)
	GetByID(id int) (model.Tournament, error)
}

type tournamentRepository struct {
	pool *pgxpool.Pool
}

func NewTournamentRepository(pool *pgxpool.Pool) TournamentRepository {
	return &tournamentRepository{pool: pool}
}

func (T *tournamentRepository) GetAll() ([]model.Tournament, error) {
	var tournaments []model.Tournament
	rows, err := T.pool.Query(context.Background(), "SELECT id, name, slug, is_show, is_active, image_url, total_surah, created_at, updated_at, deleted_at FROM tournaments WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return tournaments, err
	}

	for rows.Next() {
		var tournament model.Tournament
		if err := rows.Scan(&tournament.ID, &tournament.Name, &tournament.Slug, &tournament.IsShow, &tournament.IsActive, &tournament.ImageUrl, &tournament.TotalSurah, &tournament.CreatedAt, &tournament.UpdatedAt, &tournament.DeletedAt); err != nil {
			log.Print(err.Error())
			return tournaments, err
		}
		tournaments = append(tournaments, tournament)
	}

	return tournaments, err
}

func (T *tournamentRepository) Create(tournament model.Tournament) error {
	_, err := T.pool.Exec(context.Background(), "INSERT INTO tournaments (name, slug, is_show, is_active, image_url, total_surah, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, tournament.ImageUrl, tournament.TotalSurah, time.Now(), time.Now())

	return err
}

func (T *tournamentRepository) Update(tournament model.Tournament) error {
	if tournament.ImageUrl == "" {
		_, err := T.pool.Exec(context.Background(), "UPDATE tournaments SET name = $1, slug = $2, is_show = $3, is_active = $4, total_surah = $5, updated_at = $6 WHERE id = $7", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, tournament.TotalSurah, time.Now(), tournament.ID)
		return err
	}

	_, err := T.pool.Exec(context.Background(), "UPDATE tournaments SET name = $1, slug = $2, is_show = $3, is_active = $4, image_url = $5, total_surah = $6, updated_at = $7 WHERE id = $8", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, tournament.ImageUrl, tournament.TotalSurah, time.Now(), tournament.ID)

	return err
}

func (T *tournamentRepository) Delete(id int) error {
	_, err := T.pool.Exec(context.Background(), "UPDATE tournaments SET deleted_at = $1 WHERE id = $2", time.Now(), id)

	return err
}

func (T *tournamentRepository) GetBySlug(slug string) (model.Tournament, error) {
	var tournament model.Tournament
	err := T.pool.QueryRow(context.Background(), "SELECT id, name, slug, is_show, is_active, image_url, total_surah, created_at, updated_at, deleted_at FROM tournaments WHERE slug = $1 AND deleted_at IS NULL", slug).Scan(&tournament.ID, &tournament.Name, &tournament.Slug, &tournament.IsShow, &tournament.IsActive, &tournament.ImageUrl, &tournament.TotalSurah, &tournament.CreatedAt, &tournament.UpdatedAt, &tournament.DeletedAt)
	return tournament, err
}

func (T *tournamentRepository) GetByID(id int) (model.Tournament, error) {
	var tournament model.Tournament
	err := T.pool.QueryRow(context.Background(), "SELECT id, name, slug, is_show, is_active, image_url, total_surah, created_at, updated_at, deleted_at FROM tournaments WHERE id = $1 AND deleted_at IS NULL", id).Scan(&tournament.ID, &tournament.Name, &tournament.Slug, &tournament.IsShow, &tournament.IsActive, &tournament.ImageUrl, &tournament.TotalSurah, &tournament.CreatedAt, &tournament.UpdatedAt, &tournament.DeletedAt)
	return tournament, err
}