package config

import (
	"kambing-cup-backend/router"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Configure CORS
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsStr != "" {
		allowedOrigins = strings.Split(allowedOriginsStr, ";")
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(corsMiddleware.Handler)

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
