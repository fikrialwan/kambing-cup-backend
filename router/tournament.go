package router

import (
	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func Tournament(conn *pgx.Conn) http.Handler {
	r := chi.NewRouter()

	tr := repository.NewTournamentRepository(conn)
	ts := service.NewTournamentService(*tr)

	r.Use(middleware.Auth)
	r.Use(middleware.AdminAuth)

	r.Get("/", ts.GetAll)

	return r
}
