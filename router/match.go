package router

import (
	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"firebase.google.com/go/v4/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Match(pool *pgxpool.Pool, firebaseClient *db.Client) http.Handler {
	r := chi.NewRouter()

	mr := repository.NewMatchRepository(pool)
	sr := repository.NewSportRepository(pool)
	tr := repository.NewTournamentRepository(pool)
	
	fbWrapper := &service.RealFirebaseClient{Client: firebaseClient}
	ms := service.NewMatchService(mr, sr, tr, fbWrapper)

	r.Get("/", ms.GetAll)
	r.Get("/{id}", ms.GetByID)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(middleware.AdminAuth)

		r.Post("/", ms.Create)
		r.Post("/generate", ms.Generate)
		r.Put("/{id}", ms.Update)
		r.Delete("/{id}", ms.Delete)
	})

	return r
}