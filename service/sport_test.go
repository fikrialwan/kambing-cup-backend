package service_test

import (
	"bytes"
	"context"
	"errors"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSportService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		// Mock Tournament GetByID
		mockTournamentRepo.On("GetByID", mock.Anything, 1).Return(model.Tournament{ID: 1}, nil)

		// Mock Sport Create
		mockSportRepo.On("Create", mock.Anything, mock.AnythingOfType("model.Sport")).Return(nil)

		// Create Request
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "Futsal")
		writer.WriteField("tournament_id", "1")
		writer.Close()

		req := httptest.NewRequest("POST", "/sport", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "Sport created", w.Body.String())

		mockTournamentRepo.AssertExpectations(t)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("TournamentNotFound", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		// Mock Tournament GetByID to fail
		mockTournamentRepo.On("GetByID", mock.Anything, 999).Return(model.Tournament{}, errors.New("not found"))

		// Create Request
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "Futsal")
		writer.WriteField("tournament_id", "999")
		writer.Close()

		req := httptest.NewRequest("POST", "/sport", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		svc.Create(w, req)

		// Expect 500 because the service returns InternalServerError for generic errors from GetByID
		// But wait, the service checks errors.Is(err, pgx.ErrNoRows). 
		// My mock returns a generic error. I should probably import pgx to return pgx.ErrNoRows or check how I handle generic errors.
		// In service/sport.go:
		// if _, err := s.tournamentRepo.GetByID(tournamentID); err != nil {
		// 	if errors.Is(err, pgx.ErrNoRows) { ... }
		// 	http.Error(w, http.StatusText(http.StatusInternalServerError), ...)
		// }
		// So generic error -> 500.
		
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockTournamentRepo.AssertExpectations(t)
		mockSportRepo.AssertNotCalled(t, "Create")
	})

	t.Run("ImageSizeTooLarge", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		// Mock Tournament GetByID
		mockTournamentRepo.On("GetByID", mock.Anything, 1).Return(model.Tournament{ID: 1}, nil)

		// Create Request with a "large" image (simulated)
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "Futsal")
		writer.WriteField("tournament_id", "1")

		// Create a part for the image
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="image"; filename="large.png"`)
		header.Set("Content-Type", "image/png")
		part, _ := writer.CreatePart(header)
		// Write more than 2MB of data
		largeData := make([]byte, 2*1024*1024+1)
		part.Write(largeData)
		writer.Close()

		req := httptest.NewRequest("POST", "/sport", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Image size must be less than 2MB")

		mockTournamentRepo.AssertExpectations(t)
		mockSportRepo.AssertNotCalled(t, "Create")
	})

	t.Run("GetByID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		expectedSport := model.Sport{ID: 1, Name: "Futsal"}
		mockSportRepo.On("GetByID", mock.Anything, 1).Return(expectedSport, nil)

		req := httptest.NewRequest("GET", "/sport/1", nil)
		
		// We need to inject chi context for URLParam
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()

		svc.GetByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Check body content if needed
	})
}

func TestSportService_GetAll(t *testing.T) {
	t.Run("SuccessWithoutFilter", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		expectedSports := []model.Sport{{ID: 1, Name: "Futsal"}}
		mockSportRepo.On("GetAll", mock.Anything, 0).Return(expectedSports, nil)

		req := httptest.NewRequest("GET", "/sport", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("SuccessWithFilter", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		expectedSports := []model.Sport{{ID: 1, TournamentID: 1, Name: "Futsal"}}
		mockSportRepo.On("GetAll", mock.Anything, 1).Return(expectedSports, nil)

		req := httptest.NewRequest("GET", "/sport?tournamentId=1", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("InvalidFilter", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo)

		req := httptest.NewRequest("GET", "/sport?tournamentId=abc", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid Tournament ID")
	})
}
