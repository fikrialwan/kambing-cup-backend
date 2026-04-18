package service

import (
	"context"
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
	storagePath    string
	firebaseDb     FirebaseClient
}

func NewSportService(sportRepo repository.SportRepository, tournamentRepo repository.TournamentRepository, storagePath string, firebaseDb FirebaseClient) *SportService {
	return &SportService{sportRepo: sportRepo, tournamentRepo: tournamentRepo, storagePath: storagePath, firebaseDb: firebaseDb}
}

func (s *SportService) SyncToFirebase(ctx context.Context, tournamentID int) error {
	if s.firebaseDb == nil {
		return nil
	}
	tournament, err := s.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		return err
	}

	sports, err := s.sportRepo.GetAll(ctx, tournamentID)
	if err != nil {
		return err
	}

	type FirebaseSport struct {
		Name     string `json:"name"`
		Slug     string `json:"slug"`
		ImageUrl string `json:"imageUrl"`
	}

	fbSports := make([]FirebaseSport, 0)
	for _, sport := range sports {
		fbSports = append(fbSports, FirebaseSport{
			Name:     sport.Name,
			Slug:     sport.Slug,
			ImageUrl: sport.ImageUrl,
		})
	}

	path := fmt.Sprintf("tournaments/%s/sports", tournament.Slug)
	ref := s.firebaseDb.NewRef(path)
	return ref.Set(ctx, fbSports)
}

func (s *SportService) GetAll(w http.ResponseWriter, r *http.Request) {
	tournamentIDStr := r.URL.Query().Get("tournamentId")
	var tournamentID int
	var err error
	if tournamentIDStr != "" {
		tournamentID, err = strconv.Atoi(tournamentIDStr)
		if err != nil {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid Tournament ID")
			return
		}
	}

	sports, err := s.sportRepo.GetAll(r.Context(), tournamentID)
	if err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, sports, "", "Sports retrieved")
}

func (s *SportService) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	sport, err := s.sportRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	helper.WriteResponse(w, http.StatusOK, true, sport, "", "Sport retrieved")
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
			helper.WriteResponse(w, http.StatusRequestEntityTooLarge, false, nil, helper.ErrEntityTooLarge, "Request body too large (max 3MB)")
			return
		}
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Error parsing form")
		return
	}
	log.Printf("ParseMultipartForm took %v", time.Since(start))
	checkpoint := time.Now()

	var sport model.Sport

	tournamentID, err := strconv.Atoi(r.FormValue("tournament_id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid Tournament ID")
		return
	}
	sport.TournamentID = tournamentID

	// Validate tournament exists
	if _, err := s.tournamentRepo.GetByID(r.Context(), tournamentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrSportTournamentNotFound, "Tournament not found")
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}
	log.Printf("Tournament validation took %v", time.Since(checkpoint))
	checkpoint = time.Now()

	if r.FormValue("name") != "" {
		sport.Name = r.FormValue("name")
		sport.Slug = helper.FormatSlug(sport.Name)
	} else {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrSportNameRequired, "Name is required")
		return
	}

	if _, err := s.sportRepo.GetByNameAndTournament(r.Context(), sport.Name, sport.TournamentID); err == nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrSportNameTaken, "Sport name is already taken in this tournament")
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}
	log.Printf("GetByNameAndTournament took %v", time.Since(checkpoint))
	checkpoint = time.Now()

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		if !helper.IsImage(handler) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid image format")
			return
		}

		if !helper.ValidateImageSize(handler, 2*1024*1024) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Image size must be less than 2MB")
			return
		}

		fileName := fmt.Sprintf("%s-%d%s", sport.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

		sportDir := filepath.Join(s.storagePath, "storage", "sport")
		helper.CheckDirectory(sportDir)

		if err := helper.UploadFile(file, sportDir, fileName); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		sport.ImageUrl = fmt.Sprintf("/storage/sport/%s", fileName)
	}
	log.Printf("File handling took %v", time.Since(checkpoint))
	checkpoint = time.Now()

	if err := s.sportRepo.Create(r.Context(), sport); err != nil {
		log.Print(err.Error())
		if sport.ImageUrl != "" {
			helper.DeleteFile(filepath.Join(s.storagePath, sport.ImageUrl))
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}
	helper.WriteResponse(w, http.StatusCreated, true, nil, "", "Sport created")

	go func() {
		if err := s.SyncToFirebase(context.Background(), sport.TournamentID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	log.Printf("Repository operation took %v", time.Since(checkpoint))
}

func (s *SportService) Update(w http.ResponseWriter, r *http.Request) {
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

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	var sport model.Sport
	sport.ID = id

	tournamentID, err := strconv.Atoi(r.FormValue("tournament_id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid Tournament ID")
		return
	}
	sport.TournamentID = tournamentID

	// Validate tournament exists
	if _, err := s.tournamentRepo.GetByID(r.Context(), tournamentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrSportTournamentNotFound, "Tournament not found")
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	if r.FormValue("name") != "" {
		sport.Name = r.FormValue("name")
		sport.Slug = helper.FormatSlug(sport.Name)
	} else {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrSportNameRequired, "Name is required")
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		if !helper.IsImage(handler) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid image format")
			return
		}

		if !helper.ValidateImageSize(handler, 2*1024*1024) {
			helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Image size must be less than 2MB")
			return
		}

		fileName := fmt.Sprintf("%s-%d%s", sport.Slug, time.Now().UnixNano(), filepath.Ext(handler.Filename))

		sportDir := filepath.Join(s.storagePath, "storage", "sport")
		helper.CheckDirectory(sportDir)

		if err := helper.UploadFile(file, sportDir, fileName); err != nil {
			helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
			return
		}
		sport.ImageUrl = fmt.Sprintf("/storage/sport/%s", fileName)
	}

	if err := s.sportRepo.Update(r.Context(), sport); err != nil {
		log.Print(err.Error())
		if sport.ImageUrl != "" {
			helper.DeleteFile(filepath.Join(s.storagePath, sport.ImageUrl))
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), sport.TournamentID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Sport updated")
}

func (s *SportService) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, false, nil, helper.ErrBadRequest, "Invalid ID")
		return
	}

	sport, err := s.sportRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.WriteResponse(w, http.StatusNotFound, false, nil, helper.ErrNotFound, http.StatusText(http.StatusNotFound))
			return
		}
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	if err := s.sportRepo.Delete(r.Context(), id); err != nil {
		helper.WriteResponse(w, http.StatusInternalServerError, false, nil, helper.ErrInternalServer, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		if err := s.SyncToFirebase(context.Background(), sport.TournamentID); err != nil {
			fmt.Printf("Error syncing to Firebase: %v\n", err)
		}
	}()
	helper.WriteResponse(w, http.StatusOK, true, nil, "", "Sport deleted")
}
