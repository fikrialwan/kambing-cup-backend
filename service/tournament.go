package service

import (
	"encoding/json"
	"kambing-cup-backend/repository"
	"net/http"
)

type TournamentService struct {
	tournamentRepo *repository.TournamentRepository
}

func NewTournamentService(tournamentRepo repository.TournamentRepository) *TournamentService {
	return &TournamentService{tournamentRepo: &tournamentRepo}
}

func (s *TournamentService) GetAll(w http.ResponseWriter, _ *http.Request) {
	tournaments, err := s.tournamentRepo.GetAll()

	if err != nil {
		http.Error(nil, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(tournaments); err != nil {
		http.Error(nil, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
