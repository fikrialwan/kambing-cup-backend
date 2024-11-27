package router

import (
	"net/http"

	"kambing-cup-backend/service"

	"github.com/go-chi/chi/v5"
)

func User() http.Handler {
	r := chi.NewRouter()
	s := service.NewUserService()

	r.Get("/", s.ListUser)

	return r
}
