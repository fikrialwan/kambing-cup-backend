package main

import (
	"kambing-cup-backend/router"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

func SetupRouter(conn *pgx.Conn) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Mount("/user", router.User(conn))

	return r
}
