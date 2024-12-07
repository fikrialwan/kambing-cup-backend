package service

import (
	"encoding/json"
	"fmt"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"log"
	"net/http"
	"path/filepath"
	"time"
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

func (s *TournamentService) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(2 * 1024 * 1024)

	var tournament model.Tournament

	if r.FormValue("name") != "" {
		tournament.Name = r.FormValue("name")
		tournament.Slug = helper.FormatSlug(tournament.Name)
		tournament.IsShow = false
		tournament.IsActive = false
	} else {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")

	if err != nil {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}

	if !helper.IsImage(handler) {
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	fileName := fmt.Sprintf("%s-%d%s", tournament.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

	helper.CheckDirectory("./storage/tournament")

	if err := helper.UploadFile(&file, "./storage/tournament", fileName); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tournament.ImageUrl = fmt.Sprintf("/storage/tournament/%s", fileName)

	if err := s.tournamentRepo.Create(tournament); err != nil {
		log.Print(err.Error())
		helper.DeleteFile(fmt.Sprintf("./storage/tournament/%s", fileName))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Tournament created"))
}
