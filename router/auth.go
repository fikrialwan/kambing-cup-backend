package router

import (
	"net/http"

	"kambing-cup-backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func Auth(conn *pgx.Conn) http.Handler {
	r := chi.NewRouter()
	s := service.NewAuthService(conn)

	r.Post("/login", s.Login)

	return r
}
