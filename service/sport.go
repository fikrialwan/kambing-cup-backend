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
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type SportService struct {
	sportRepo *repository.SportRepository
}

func NewSportService(sportRepo repository.SportRepository) *SportService {
	return &SportService{sportRepo: &sportRepo}
}

func (s *SportService) GetAll(w http.ResponseWriter, r *http.Request) {
	sports, err := s.sportRepo.GetAll()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sports); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *SportService) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	sport, err := s.sportRepo.GetByID(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sport); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *SportService) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(2 * 1024 * 1024)

	var sport model.Sport
	
	tournamentID, err := strconv.Atoi(r.FormValue("tournament_id"))
	if err != nil {
		http.Error(w, "Invalid Tournament ID", http.StatusBadRequest)
		return
	}
	sport.TournamentID = tournamentID


	if r.FormValue("name") != "" {
		sport.Name = r.FormValue("name")
		sport.Slug = helper.FormatSlug(sport.Name)
	} else {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		if !helper.IsImage(handler) {
			http.Error(w, "Invalid image format", http.StatusBadRequest)
			return
		}

		fileName := fmt.Sprintf("%s-%d%s", sport.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

		helper.CheckDirectory("./storage/sport")

		if err := helper.UploadFile(&file, "./storage/sport", fileName); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		sport.ImageUrl = fmt.Sprintf("/storage/sport/%s", fileName)
	}

	if err := s.sportRepo.Create(sport); err != nil {
		log.Print(err.Error())
		if sport.ImageUrl != "" {
			helper.DeleteFile(fmt.Sprintf(".%s", sport.ImageUrl))
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Sport created"))
}

func (s *SportService) Update(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(2 * 1024 * 1024)

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var sport model.Sport
	sport.ID = id

	tournamentID, err := strconv.Atoi(r.FormValue("tournament_id"))
	if err != nil {
		http.Error(w, "Invalid Tournament ID", http.StatusBadRequest)
		return
	}
	sport.TournamentID = tournamentID

	if r.FormValue("name") != "" {
		sport.Name = r.FormValue("name")
		sport.Slug = helper.FormatSlug(sport.Name)
	} else {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	
	file, handler, err := r.FormFile("image")
	if err == nil {
		if !helper.IsImage(handler) {
			http.Error(w, "Invalid image format", http.StatusBadRequest)
			return
		}

		fileName := fmt.Sprintf("%s-%d%s", sport.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

		helper.CheckDirectory("./storage/sport")

		if err := helper.UploadFile(&file, "./storage/sport", fileName); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		sport.ImageUrl = fmt.Sprintf("/storage/sport/%s", fileName)
	}

	if err := s.sportRepo.Update(sport); err != nil {
		log.Print(err.Error())
		if sport.ImageUrl != "" {
			helper.DeleteFile(fmt.Sprintf(".%s", sport.ImageUrl))
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sport updated"))
}

func (s *SportService) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := s.sportRepo.Delete(id); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sport deleted"))
}
