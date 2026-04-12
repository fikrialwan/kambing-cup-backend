package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type TeamService struct {
	teamRepo       repository.TeamRepository
	sportRepo      repository.SportRepository
	matchRepo      repository.MatchRepository
	tournamentRepo repository.TournamentRepository
	firebaseDb     FirebaseClient
}

func NewTeamService(teamRepo repository.TeamRepository, sportRepo repository.SportRepository, matchRepo repository.MatchRepository, tournamentRepo repository.TournamentRepository, firebaseDb FirebaseClient) *TeamService {
	return &TeamService{
		teamRepo:       teamRepo,
		sportRepo:      sportRepo,
		matchRepo:      matchRepo,
		tournamentRepo: tournamentRepo,
		firebaseDb:     firebaseDb,
	}
}

func (s *TeamService) SyncToFirebase(ctx context.Context, sportID int) error {
	if s.firebaseDb == nil {
		return nil
	}
	sport, err := s.sportRepo.GetByID(ctx, sportID)
	if err != nil {
		return err
	}

	tournament, err := s.tournamentRepo.GetByID(ctx, sport.TournamentID)
	if err != nil {
		return err
	}

	matches, err := s.matchRepo.GetBySportID(ctx, sportID)
	if err != nil {
		return err
	}

	teams, err := s.teamRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	teamMap := make(map[int]model.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// This logic is similar to MatchService.SyncToFirebase
	// but we only need to sync matches for this sport.
	firebaseMatches := make(map[string]interface{})
	for _, curr := range matches {
		var homeName, awayName string
		if curr.HomeID != nil {
			if t, ok := teamMap[*curr.HomeID]; ok {
				homeName = t.Name
			}
		}
		if curr.AwayID != nil {
			if t, ok := teamMap[*curr.AwayID]; ok {
				awayName = t.Name
			}
		}

		canEdit := curr.HomeID == nil || curr.AwayID == nil
		fbMatch := map[string]interface{}{
			"name":                curr.Round + " - Match ",
			"nextMatchId":         "", // This might need more logic if we want to support full bracket
			"startTime":           curr.StartDate.Format("15:04"),
			"state":               string(curr.State),
			"tournamentRoundText": curr.Round,
			"participants": []interface{}{
				nil,
				map[string]interface{}{
					"name":         homeName,
					"resultText":   curr.HomeScore,
					"isWinner":     curr.Winner != nil && homeName != "" && *curr.Winner == homeName,
					"canEditTeams": canEdit,
				},
				map[string]interface{}{
					"name":         awayName,
					"resultText":   curr.AwayScore,
					"isWinner":     curr.Winner != nil && awayName != "" && *curr.Winner == awayName,
					"canEditTeams": canEdit,
				},
			},
		}
		firebaseMatches[strconv.Itoa(curr.ID)] = fbMatch
	}

	path := fmt.Sprintf("%s/sports/%s/matches", tournament.Slug, sport.Slug)
	
	ref := s.firebaseDb.NewRef(path)
	return ref.Set(ctx, firebaseMatches)
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
		go func() {
			if err := s.SyncToFirebase(context.Background(), team.SportID); err != nil {
				fmt.Printf("Error syncing to Firebase: %v\n", err)
			}
		}()
		helper.WriteResponse(w, http.StatusOK, true, nil, "", "Team restored")
	} else {
		if err := s.teamRepo.Create(r.Context(), team); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		go func() {
			if err := s.SyncToFirebase(context.Background(), team.SportID); err != nil {
				fmt.Printf("Error syncing to Firebase: %v\n", err)
			}
		}()
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

	if len(teams) > 0 {
		go func() {
			if err := s.SyncToFirebase(context.Background(), teams[0].SportID); err != nil {
				fmt.Printf("Error syncing to Firebase: %v\n", err)
			}
		}()
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

	go func() {
		if err := s.SyncToFirebase(context.Background(), team.SportID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Team updated")
}

func (s *TeamService) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := s.teamRepo.Delete(r.Context(), id); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), team.SportID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Team deleted")
}