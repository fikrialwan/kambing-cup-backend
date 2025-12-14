package router

import (
	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func Match(conn *pgx.Conn) http.Handler {
	r := chi.NewRouter()

	mr := repository.NewMatchRepository(conn)
	ms := service.NewMatchService(*mr)

	r.Get("/", ms.GetAll)
	r.Get("/{id}", ms.GetByID)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(middleware.AdminAuth)

		r.Post("/", ms.Create)
		r.Put("/{id}", ms.Update)
		r.Delete("/{id}", ms.Delete)
	})

	return r
}
