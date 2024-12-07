package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type TournamentRepository struct {
	conn *pgx.Conn
}

func NewTournamentRepository(conn *pgx.Conn) *TournamentRepository {
	return &TournamentRepository{conn: conn}
}

func (T *TournamentRepository) GetAll() ([]model.Tournament, error) {
	var tournaments []model.Tournament
	rows, err := T.conn.Query(context.Background(), "SELECT * FROM tournaments WHERE deleted_at IS NULL")
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
	_, err := T.conn.Exec(context.Background(), "INSERT INTO tournaments (name, slug, is_show, is_active, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", tournament.Name, tournament.Slug, tournament.IsShow, tournament.IsActive, tournament.ImageUrl, time.Now(), time.Now())

	return err
}
