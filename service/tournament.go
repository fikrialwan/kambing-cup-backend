package service

import (
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

type TournamentService struct {
	tournamentRepo repository.TournamentRepository
	storagePath    string
}

func NewTournamentService(tournamentRepo repository.TournamentRepository, storagePath string) *TournamentService {
	return &TournamentService{tournamentRepo: tournamentRepo, storagePath: storagePath}
}

func (s *TournamentService) GetAll(w http.ResponseWriter, r *http.Request) {
	tournaments, err := s.tournamentRepo.GetAll(r.Context())

	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, tournaments, "", "Tournaments retrieved")
}

func (s *TournamentService) Create(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf("Create tournament took %v", time.Since(start))
	}()

	// Limit the total request size to prevent reading very large files into memory or temp files.
	const maxRequestSize = 3 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			helper.WriteResponse(w, http.StatusRequestEntityTooLarge, false, nil, helper.ErrEntityTooLarge, "Request body too large (max 3MB)")
			return
		}
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Error parsing form")
		return
	}
	log.Printf("ParseMultipartForm took %v", time.Since(start))
	checkpoint := time.Now()

	var tournament model.Tournament

	if r.FormValue("name") != "" {
		tournament.Name = r.FormValue("name")
		tournament.Slug = helper.FormatSlug(tournament.Name)
		tournament.IsShow = false
		tournament.IsActive = false
	} else {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentNameRequired, "Name is required")
		return
	}

	var isDeleted bool
	if existing, err := s.tournamentRepo.GetBySlugWithDeleted(r.Context(), tournament.Slug); err == nil {
		if existing.DeletedAt == nil {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentSlugTaken, "Slug is already taken")
			return
		}
		isDeleted = true
		tournament.ID = existing.ID
	} else if !errors.Is(err, pgx.ErrNoRows) {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}
	log.Printf("GetBySlugWithDeleted took %v", time.Since(checkpoint))
	checkpoint = time.Now()

	if r.FormValue("is_show") != "" {
		tournament.IsShow = r.FormValue("is_show") == "true"
	}

	if r.FormValue("is_active") != "" {
		tournament.IsActive = r.FormValue("is_active") == "true"
	}

	if r.FormValue("total_surah") != "" {
		totalSurah, err := strconv.Atoi(r.FormValue("total_surah"))
		if err != nil {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentTotalSurahInvalid, "Total surah must be a number")
			return
		}
		tournament.TotalSurah = totalSurah
	} else {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentTotalSurahInvalid, "Total surah is required")
		return
	}

	file, handler, err := r.FormFile("image")

	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentImageRequired, "Image is required")
		return
	}

	if !helper.IsImage(handler) {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid image format")
		return
	}

	if !helper.ValidateImageSize(handler, 2*1024*1024) {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Image size must be less than 2MB")
		return
	}

	fileName := fmt.Sprintf("%s-%d%s", tournament.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

	tournamentDir := filepath.Join(s.storagePath, "storage", "tournament")
	helper.CheckDirectory(tournamentDir)

	if err := helper.UploadFile(&file, tournamentDir, fileName); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}
	log.Printf("UploadFile took %v", time.Since(checkpoint))
	checkpoint = time.Now()

	tournament.ImageUrl = fmt.Sprintf("/storage/tournament/%s", fileName)

	if isDeleted {
		if err := s.tournamentRepo.Restore(r.Context(), tournament); err != nil {
			log.Print(err.Error())
			helper.DeleteFile(filepath.Join(tournamentDir, fileName))
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		helper.WriteResponse(w, http.StatusOK, true, nil, "", "Tournament restored")
	} else {
		if err := s.tournamentRepo.Create(r.Context(), tournament); err != nil {
			log.Print(err.Error())
			helper.DeleteFile(filepath.Join(tournamentDir, fileName))
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		helper.WriteResponse(w, http.StatusCreated, true, nil, "", "Tournament created")
	}
	log.Printf("Repository operation took %v", time.Since(checkpoint))
}

func (s *TournamentService) Update(w http.ResponseWriter, r *http.Request) {
	// Limit the total request size to prevent reading very large files into memory or temp files.
	const maxRequestSize = 3 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(maxRequestSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			helper.WriteResponse(w, http.StatusRequestEntityTooLarge, false, nil, helper.ErrEntityTooLarge, "Request body too large (max 3MB)")
			return
		}
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Error parsing form")
		return
	}

	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	file, handler, _ := r.FormFile("image")

	var tournament model.Tournament

	tournament.ID = idInt

	if r.FormValue("name") != "" {
		tournament.Name = r.FormValue("name")
		tournament.Slug = helper.FormatSlug(tournament.Name)
	} else {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentNameRequired, "Name is required")
		return
	}

	if tournamentTemp, err := s.tournamentRepo.GetBySlug(r.Context(), tournament.Slug); err == nil && tournamentTemp.ID != idInt {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentSlugTaken, "Slug is already taken")
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
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrTournamentTotalSurahInvalid, "Total surah must be a number")
			return
		}
		tournament.TotalSurah = totalSurah
	}

	fileName := ""

	if file != nil {
		if !helper.IsImage(handler) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid image format")
			return
		}

		if !helper.ValidateImageSize(handler, 2*1024*1024) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Image size must be less than 2MB")
			return
		}

		fileName = fmt.Sprintf("%s-%d%s", tournament.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

		tournamentDir := filepath.Join(s.storagePath, "storage", "tournament")
		helper.CheckDirectory(tournamentDir)

		if err := helper.UploadFile(&file, tournamentDir, fileName); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}

		tournament.ImageUrl = fmt.Sprintf("/storage/tournament/%s", fileName)
	}

	if tournament.IsActive {
		if err := s.tournamentRepo.DeactivateAllExcept(r.Context(), tournament.ID); err != nil {
			log.Print(err.Error())
			if fileName != "" {
				helper.DeleteFile(filepath.Join(s.storagePath, tournament.ImageUrl))
			}
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	if err := s.tournamentRepo.Update(r.Context(), tournament); err != nil {
		log.Print(err.Error())
		if fileName != "" {
			helper.DeleteFile(filepath.Join(s.storagePath, tournament.ImageUrl))
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Tournament updated")
}

func (s *TournamentService) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idInt, err := strconv.Atoi(id)

	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	if err := s.tournamentRepo.Delete(r.Context(), idInt); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Tournament deleted")
}

func (s *TournamentService) GetActive(w http.ResponseWriter, r *http.Request) {
	tournament, err := s.tournamentRepo.GetActive(r.Context())

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, tournament, "", "Active tournament retrieved")
}

func (s *TournamentService) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	tournament, err := s.tournamentRepo.GetBySlug(r.Context(), slug)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, tournament, "", "Tournament retrieved")
}

func (s *TournamentService) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	tournament, err := s.tournamentRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, tournament, "", "Tournament retrieved")
}