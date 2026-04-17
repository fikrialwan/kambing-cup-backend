package router

import (
	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Match(pool *pgxpool.Pool, firebaseDb service.FirebaseClient) http.Handler {
	r := chi.NewRouter()

	mr := repository.NewMatchRepository(pool)
	sr := repository.NewSportRepository(pool)
	tr := repository.NewTournamentRepository(pool)
	ter := repository.NewTeamRepository(pool)

	ms := service.NewMatchService(mr, sr, ter, tr, firebaseDb)
	r.Get("/", ms.GetAll)
	r.Get("/{id}", ms.GetByID)
	r.Get("/{matchId}/history/{teamId}", ms.GetTeamHistoryImages)
	r.Get("/sport/{sportId}/team/{teamId}/history", ms.GetTeamHistoryImagesByTeam)

	// SuperAdmin only routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(middleware.SuperAdminAuth)

		r.Post("/", ms.Create)
		r.Post("/generate", ms.Generate)
		r.Delete("/{id}", ms.Delete)
		r.Delete("/sport/{sportId}", ms.DeleteBySportID)
	})

	// Admin or SuperAdmin routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Use(middleware.AdminOrSuperAdminAuth)

		r.Put("/{id}", ms.Update)
	})

	return r
}
