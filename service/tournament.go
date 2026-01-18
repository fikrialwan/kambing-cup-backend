package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/repository"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(tournaments); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	if _, err := s.tournamentRepo.GetBySlug(tournament.Slug); err == nil {
		http.Error(w, "Slug is already taken", http.StatusBadRequest)
		return
	}

	if r.FormValue("is_show") != "" {
		tournament.IsShow = r.FormValue("is_show") == "true"
	}

	if r.FormValue("is_active") != "" {
		tournament.IsActive = r.FormValue("is_active") == "true"
	}

	if r.FormValue("total_surah") != "" {
		totalSurah, err := strconv.Atoi(r.FormValue("total_surah"))
		if err != nil {
			http.Error(w, "Total surah must be a number", http.StatusBadRequest)
			return
		}
		tournament.TotalSurah = totalSurah
	} else {
		http.Error(w, "Total surah is required", http.StatusBadRequest)
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

func (s *TournamentService) Update(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(2 * 1024 * 1024)

	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	file, handler, _ := r.FormFile("image")

	var tournament model.Tournament

	tournament.ID = idInt

	if r.FormValue("name") != "" {
		tournament.Name = r.FormValue("name")
		tournament.Slug = helper.FormatSlug(tournament.Name)
	} else {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if tournamentTemp, err := s.tournamentRepo.GetBySlug(tournament.Slug); err == nil && tournamentTemp.ID != idInt {
		http.Error(w, "Slug is already taken", http.StatusBadRequest)
		return
	}

	if r.FormValue("is_show") != "" {
		tournament.IsShow = r.FormValue("is_show") == "true"
	}

	if r.FormValue("is_active") != "" {
		tournament.IsActive = r.FormValue("is_active") == "true"
	}

	if r.FormValue("total_surah") != "" {
		totalSurah, err := strconv.Atoi(r.FormValue("total_surah"))
		if err != nil {
			http.Error(w, "Total surah must be a number", http.StatusBadRequest)
			return
		}
		tournament.TotalSurah = totalSurah
	}

	fileName := ""

	if file != nil {
		if !helper.IsImage(handler) {
			http.Error(w, "Invalid image format", http.StatusBadRequest)
			return
		}

		fileName = fmt.Sprintf("%s-%d%s", tournament.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

		helper.CheckDirectory("./storage/tournament")

		if err := helper.UploadFile(&file, "./storage/tournament", fileName); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		tournament.ImageUrl = fmt.Sprintf("/storage/tournament/%s", fileName)
	}

	if err := s.tournamentRepo.Update(tournament); err != nil {
		log.Print(err.Error())
		if fileName != "" {
			helper.DeleteFile(fmt.Sprintf("./storage/tournament/%s", fileName))
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Tournament updated"))
}

func (s *TournamentService) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := s.tournamentRepo.Delete(idInt); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Tournament deleted"))
}

func (s *TournamentService) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	tournament, err := s.tournamentRepo.GetBySlug(slug)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(tournament); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
