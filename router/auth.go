package router

import (
	"net/http"

	"kambing-cup-backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Auth(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	s := service.NewAuthService(pool)

	r.Post("/login", s.Login)

	return r
}
