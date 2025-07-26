package router

import (
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func Public(conn *pgx.Conn) http.Handler {
	r := chi.NewRouter()

	tr := repository.NewTournamentRepository(conn)
	ts := service.NewTournamentService(*tr)

	r.Get("/tournament/{slug}", ts.Get)

	return r
}
