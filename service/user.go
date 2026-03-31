package service

import (
	"encoding/json"
	"errors"
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-user-id")
	id, err := strconv.Atoi(userID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user, err := s.userRepo.GetById(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userReq model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if userReq.Username == "" || userReq.Password == "" || userReq.Role == "" || userReq.Email == "" {
		http.Error(w, "Username, email, password, and role are required", http.StatusBadRequest)
		return
	}

	if !s.checkRole(userReq.Role) {
		http.Error(w, "Role must be ADMIN or SUPERADMIN", http.StatusBadRequest)
		return
	}

	var isDeleted bool
	var existingUser model.User
	if existing, err := s.userRepo.GetByUsernameOrEmailWithDeleted(r.Context(), userReq.Username, userReq.Email); err == nil {
		if existing.DeletedAt == nil {
			http.Error(w, "Username or email is already taken", http.StatusBadRequest)
			return
		}
		isDeleted = true
		existingUser = existing
	} else if !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if isDeleted {
		// Prepare restored user data
		existingUser.Password = userReq.Password
		existingUser.Role = userReq.Role
		if err := s.userRepo.Restore(r.Context(), existingUser); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User restored"))
	} else {
		if err := s.userRepo.Create(r.Context(), userReq); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("User created"))
	}
}

func (s *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user model.UpdateUserRequest
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	user.ID = idInt

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" || user.Role == "" || user.Email == "" {
		http.Error(w, "Username, email, password, and role are required", http.StatusBadRequest)
		return
	}

	if !s.checkRole(user.Role) {
		http.Error(w, "Role must be ADMIN or SUPERADMIN", http.StatusBadRequest)
		return
	}

	_, err = s.userRepo.GetById(r.Context(), user.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := s.userRepo.Update(r.Context(), user); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User updated"))
}

func (s *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := s.userRepo.Delete(r.Context(), idInt); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted"))
}