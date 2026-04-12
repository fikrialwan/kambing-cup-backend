package service

import (
	"encoding/json"
	"errors"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) checkRole(role string) bool {
	return role == "ADMIN" || role == "SUPERADMIN"
}

func (s *UserService) ListUser(w http.ResponseWriter, r *http.Request) {
	users, err := s.userRepo.GetAll(r.Context())

	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, users, "", "Users retrieved")
}

func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-user-id")
	id, err := strconv.Atoi(userID)

	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, "Internal Server Error: Missing x-user-id")
		return
	}

	user, err := s.userRepo.GetById(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, user, "", "User retrieved")
}

func (s *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userReq model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}

	if userReq.Username == "" || userReq.Password == "" || userReq.Role == "" || userReq.Email == "" {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrUserRequiredFields, "Username, email, password, and role are required")
		return
	}

	if !s.checkRole(userReq.Role) {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrUserInvalidRole, "Role must be ADMIN or SUPERADMIN")
		return
	}

	if _, err := s.userRepo.GetByUsernameOrEmail(r.Context(), userReq.Username, userReq.Email); err == nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrUserAlreadyExists, "Username or email is already taken")
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	if err := s.userRepo.Create(r.Context(), userReq); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}
	helper.WriteResponse(w, http.StatusCreated, true, nil, "", "User created")
}

func (s *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user model.UpdateUserRequest
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	user.ID = idInt

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}

	if user.Username == "" || user.Password == "" || user.Role == "" || user.Email == "" {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrUserRequiredFields, "Username, email, password, and role are required")
		return
	}

	if !s.checkRole(user.Role) {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrUserInvalidRole, "Role must be ADMIN or SUPERADMIN")
		return
	}

	_, err = s.userRepo.GetById(r.Context(), user.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	if err := s.userRepo.Update(r.Context(), user); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, nil, "", "User updated")
}

func (s *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	if err := s.userRepo.Delete(r.Context(), idInt); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, nil, "", "User deleted")
}