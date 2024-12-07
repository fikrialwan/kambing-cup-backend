package service

import (
	"encoding/json"
	"kambing-cup-backend/repository"
	"net/http"
	"strconv"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: &userRepo}
}

func (s *UserService) ListUser(w http.ResponseWriter, _ *http.Request) {
	users, err := s.userRepo.GetAll()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-user-id")
	id, err := strconv.Atoi(userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := s.userRepo.GetById(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
