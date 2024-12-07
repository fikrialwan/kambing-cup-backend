package router

import (
	"net/http"

	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func User(conn *pgx.Conn) http.Handler {
	r := chi.NewRouter()

	user := repository.NewUserRepository(conn)
	s := service.NewUserService(*user)

	r.Use(middleware.Auth)
	r.Use(middleware.AdminAuth)

	r.Get("/", s.GetUser)
	r.Get("/all", s.ListUser)

	return r
}
