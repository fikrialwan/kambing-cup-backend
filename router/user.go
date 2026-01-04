package router

import (
	"net/http"

	"kambing-cup-backend/middleware"
	"kambing-cup-backend/repository"
	"kambing-cup-backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func intMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/user" {
			next.ServeHTTP(w, r)
		} else {
			middleware.AdminAuth(next).ServeHTTP(w, r)
		}
	})

}

func User(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	user := repository.NewUserRepository(pool)
	s := service.NewUserService(*user)

	r.Use(middleware.Auth)
	r.Use(intMiddleware)

	r.Get("/", s.GetUser)
	r.Post("/", s.CreateUser)
	r.Put("/{id}", s.UpdateUser)
	r.Delete("/{id}", s.DeleteUser)
	r.Get("/all", s.ListUser)

	return r
}
