package router

import (
	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Tournament(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	tr := repository.NewTournamentRepository(pool)
	ts := service.NewTournamentService(tr)

	r.Use(middleware.Auth)
	r.Use(middleware.AdminAuth)

	r.Get("/", ts.GetAll)
	r.Get("/{id}", ts.GetByID)
	r.Post("/", ts.Create)
	r.Put("/{id}", ts.Update)
	r.Delete("/{id}", ts.Delete)

	return r
}