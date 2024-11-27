package service

import (
	"encoding/json"
	"kambing-cup-backend/repository"
	"net/http"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) ListUser(w http.ResponseWriter, _ *http.Request) {
	repository := repository.NewUserRepository()

	users := repository.GetAll()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
