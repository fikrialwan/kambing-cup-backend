package router

import (
	"net/http"

	"kambing-cup-backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func User(conn *pgx.Conn) http.Handler {
	r := chi.NewRouter()
	s := service.NewUserService(conn)

	r.Get("/", s.ListUser)

	return r
}
