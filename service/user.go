package service

import (
	"encoding/json"
	"kambing-cup-backend/repository"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type UserService struct {
	conn *pgx.Conn
}

func NewUserService(conn *pgx.Conn) *UserService {
	return &UserService{conn: conn}
}

func (s *UserService) ListUser(w http.ResponseWriter, _ *http.Request) {
	repository := repository.NewUserRepository(s.conn)

	users, err := repository.GetAll()

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
