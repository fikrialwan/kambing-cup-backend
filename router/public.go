package router

import (
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Public(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	tr := repository.NewTournamentRepository(pool)
	ts := service.NewTournamentService(*tr)

	r.Get("/tournament/{slug}", ts.Get)

	return r
}
