package service

import (
	"encoding/json"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo repository.UserRepository
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	ExpIn int64  `json:"exp_in"`
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid request body")
		return
	}

	if loginRequest.Email == "" || loginRequest.Password == "" {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrAuthRequiredFields, "Email and password are required")
		return
	}

	user, err := s.userRepo.GetByEmailPassword(r.Context(), loginRequest.Email, loginRequest.Password)

	if err != nil {
		log.Default().Println(err.Error())
		helper.WriteResponse(w, http.StatusUnauthorized, false, nil, helper.ErrAuthInvalidCredentials, "Email or password is incorrect")
		return
	}

	token, expIn, err := generateToken(user)

	if err != nil {
		log.Printf("Token generation error: %v", err)
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, LoginResponse{Token: token, ExpIn: expIn}, "", "Login successful")
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