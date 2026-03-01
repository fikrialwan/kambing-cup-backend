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
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type SportService struct {
	sportRepo      repository.SportRepository
	tournamentRepo repository.TournamentRepository
}

func NewSportService(sportRepo repository.SportRepository, tournamentRepo repository.TournamentRepository) *SportService {
	return &SportService{sportRepo: sportRepo, tournamentRepo: tournamentRepo}
}

func (s *SportService) GetAll(w http.ResponseWriter, r *http.Request) {
	sports, err := s.sportRepo.GetAll(r.Context())
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

	sport, err := s.sportRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sport); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *SportService) Create(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf("Create sport took %v", time.Since(start))
	}()

	// Limit the total request size to prevent reading very large files into memory or temp files.
	const maxRequestSize = 3 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			http.Error(w, "Request body too large (max 3MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	log.Printf("ParseMultipartForm took %v", time.Since(start))
	checkpoint := time.Now()

	var sport model.Sport

	tournamentID, err := strconv.Atoi(r.FormValue("tournament_id"))
	if err != nil {
		http.Error(w, "Invalid Tournament ID", http.StatusBadRequest)
		return
	}
	sport.TournamentID = tournamentID

	// Validate tournament exists
	if _, err := s.tournamentRepo.GetByID(r.Context(), tournamentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Tournament not found", http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("Tournament validation took %v", time.Since(checkpoint))
	checkpoint = time.Now()

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

		if !helper.ValidateImageSize(handler, 2*1024*1024) {
			http.Error(w, "Image size must be less than 2MB", http.StatusBadRequest)
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
	log.Printf("File handling took %v", time.Since(checkpoint))
	checkpoint = time.Now()

	if err := s.sportRepo.Create(r.Context(), sport); err != nil {
		log.Print(err.Error())
		if sport.ImageUrl != "" {
			helper.DeleteFile(fmt.Sprintf(".%s", sport.ImageUrl))
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("Repository Create took %v", time.Since(checkpoint))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Sport created"))
}

func (s *SportService) Update(w http.ResponseWriter, r *http.Request) {
	// Limit the total request size to prevent reading very large files into memory or temp files.
	const maxRequestSize = 3 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			http.Error(w, "Request body too large (max 3MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

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

	// Validate tournament exists
	if _, err := s.tournamentRepo.GetByID(r.Context(), tournamentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Tournament not found", http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

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

		if !helper.ValidateImageSize(handler, 2*1024*1024) {
			http.Error(w, "Image size must be less than 2MB", http.StatusBadRequest)
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

	if err := s.sportRepo.Update(r.Context(), sport); err != nil {
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

	if err := s.sportRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sport deleted"))
}