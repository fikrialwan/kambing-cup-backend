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

type TeamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

func (s *TeamService) GetAll(w http.ResponseWriter, r *http.Request) {
	teams, err := s.teamRepo.GetAll(r.Context())
	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, teams, "", "Teams retrieved")
}

func (s *TeamService) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	team, err := s.teamRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, team, "", "Team retrieved")
}

func (s *TeamService) Create(w http.ResponseWriter, r *http.Request) {
	var team model.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}

	if team.Name == "" || team.SportID == 0 {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTeamRequiredFields, "Name and Sport ID are required")
		return
	}

	var isDeleted bool
	if existing, err := s.teamRepo.GetByNameAndSportWithDeleted(r.Context(), team.Name, team.SportID); err == nil {
		if existing.DeletedAt == nil {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTeamNameTaken, "Team name is already taken in this sport")
			return
		}
		isDeleted = true
		team.ID = existing.ID
	} else if !errors.Is(err, pgx.ErrNoRows) {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	if isDeleted {
		if err := s.teamRepo.Restore(r.Context(), team); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		helper.WriteResponse(w, http.StatusOK, true, nil, "", "Team restored")
	} else {
		if err := s.teamRepo.Create(r.Context(), team); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		helper.WriteResponse(w, http.StatusCreated, true, nil, "", "Team created")
	}
}

func (s *TeamService) CreateBulk(w http.ResponseWriter, r *http.Request) {
	var teams []model.Team
	if err := json.NewDecoder(r.Body).Decode(&teams); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}

	if len(teams) == 0 {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Teams are required")
		return
	}

	for _, team := range teams {
		if team.Name == "" || team.SportID == 0 {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTeamRequiredFields, "Name and Sport ID are required for all teams")
			return
		}
	}

	if err := s.teamRepo.CreateBulk(r.Context(), teams); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusCreated, true, nil, "", "Teams created")
}

func (s *TeamService) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	var team model.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, err.Error())
		return
	}
	team.ID = id

	if err := s.teamRepo.Update(r.Context(), team); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Team updated")
}

func (s *TeamService) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	if err := s.teamRepo.Delete(r.Context(), id); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Team deleted")
}