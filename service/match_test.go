package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMatchService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockFirebase := new(MockFirebaseClient)

		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTournamentRepo, mockFirebase)

		mockMatchRepo.On("Create", mock.AnythingOfType("model.Match")).Return(nil)

		reqBody := model.Match{
			SportID:   1,
			Round:     "Final",
			State:     model.SOON,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/match", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "Match created", w.Body.String())
		mockMatchRepo.AssertExpectations(t)
	})
}

func TestMatchService_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockFirebase := new(MockFirebaseClient)

		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTournamentRepo, mockFirebase)

		expectedMatch := model.Match{ID: 1, SportID: 1}
		mockMatchRepo.On("GetByID", 1).Return(expectedMatch, nil)

		req := httptest.NewRequest("GET", "/match/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.GetByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockMatchRepo.AssertExpectations(t)
	})
}

func TestMatchService_Generate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockFirebase := new(MockFirebaseClient)
		mockFirebaseRef := new(MockFirebaseRef)

		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTournamentRepo, mockFirebase)

		// Setup mocks
		mockSportRepo.On("GetByID", 1).Return(model.Sport{ID: 1, TournamentID: 1, Slug: "futsal"}, nil)
		mockTournamentRepo.On("GetByID", 1).Return(model.Tournament{ID: 1, Slug: "agi-15"}, nil)
		
		// For Generate, it creates multiple matches. We mock Create to return nil for any match.
		// Since team_count is 4, it generates:
		// 3rd place match (id ?)
		// Final (1)
		// Semis (11, 12)
		// Total 4 matches.
		mockMatchRepo.On("Create", mock.AnythingOfType("model.Match")).Return(nil)

		// Firebase mocks
		// It creates a ref at "agi-15/sports/futsal/matches"
		mockFirebase.On("NewRef", "agi-15/sports/futsal/matches").Return(mockFirebaseRef)
		mockFirebaseRef.On("Set", mock.Anything, mock.Anything).Return(nil)

		reqBody := service.GenerateMatchesRequest{
			TeamCount: 4,
			SportID:   1,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/match/generate", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.Generate(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "Matches generated successfully", w.Body.String())

		mockMatchRepo.AssertExpectations(t)
		mockSportRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockFirebase.AssertExpectations(t)
		mockFirebaseRef.AssertExpectations(t)
	})
}
