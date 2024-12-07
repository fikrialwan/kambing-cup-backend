package service

import (
	"encoding/json"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type AuthService struct {
	conn *pgx.Conn
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	ExpIn int64  `json:"exp_in"`
}

func NewAuthService(conn *pgx.Conn) *AuthService {
	return &AuthService{conn: conn}
}

func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	repository := repository.NewUserRepository(s.conn)

	users, err := repository.GetByUsernamePassword(loginRequest.Username, loginRequest.Password)

	if err != nil {
		http.Error(w, "Username or password is incorrect", http.StatusBadRequest)
		return
	}

	token, expIn, err := generateToken(users)

	if err != nil {
		log.Panic(err.Error())
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(LoginResponse{Token: token, ExpIn: expIn}); err != nil {
		log.Panic(err.Error())
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}

func generateToken(user model.User) (s string, expIn int64, err error) {
	expIn = time.Now().Add(time.Hour * 24).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  user.ID,
			"role": user.Role,
			"exp":  expIn,
		})

	if s, err = t.SignedString([]byte(os.Getenv("SECRET"))); err != nil {
		return
	}

	return
}
