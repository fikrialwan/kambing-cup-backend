package router

import (
	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Sport(pool *pgxpool.Pool, firebaseDb service.FirebaseClient) http.Handler {
	r := chi.NewRouter()

	sr := repository.NewSportRepository(pool)
	tr := repository.NewTournamentRepository(pool)
	ss := service.NewSportService(sr, tr, ".", firebaseDb)

	r.Get("/", ss.GetAll)
	r.Get("/{id}", ss.GetByID)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(middleware.SuperAdminAuth)

		r.Post("/", ss.Create)
		r.Put("/{id}", ss.Update)
		r.Delete("/{id}", ss.Delete)
	})

	return r
}
