package config

import (
	"kambing-cup-backend/router"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Mount("/user", router.User(pool))
	r.Mount("/auth", router.Auth(pool))
	r.Mount("/tournament", router.Tournament(pool))
	r.Mount("/sport", router.Sport(pool))
	r.Mount("/team", router.Team(pool))
	r.Mount("/match", router.Match(pool))
	r.Mount("/public", router.Public(pool))

	fs := http.FileServer(http.Dir("./storage"))
	r.Handle("/storage/*", http.StripPrefix("/storage/", fs))

	return r
}
