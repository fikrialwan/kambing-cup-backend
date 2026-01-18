package router

import (
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Auth(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	
	userRepo := repository.NewUserRepository(pool)
	s := service.NewAuthService(userRepo)

	r.Post("/login", s.Login)

	return r
}