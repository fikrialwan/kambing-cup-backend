package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMatchService_GetAll(t *testing.T) {
	t.Run("Success without sportId", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		svc := service.NewMatchService(mockMatchRepo, nil, nil, nil, nil)

		expectedMatches := []model.Match{{ID: 1}, {ID: 2}}
		mockMatchRepo.On("GetAll", mock.Anything).Return(expectedMatches, nil)

		req := httptest.NewRequest("GET", "/match", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		mockMatchRepo.AssertExpectations(t)
	})

	t.Run("Success with sportId", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		svc := service.NewMatchService(mockMatchRepo, nil, nil, nil, nil)

		expectedMatches := []model.Match{{ID: 1, SportID: 1}}
		mockMatchRepo.On("GetBySportID", mock.Anything, 1).Return(expectedMatches, nil)

		req := httptest.NewRequest("GET", "/match?sportId=1", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		mockMatchRepo.AssertExpectations(t)
	})

	t.Run("Invalid sportId", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		svc := service.NewMatchService(mockMatchRepo, nil, nil, nil, nil)

		req := httptest.NewRequest("GET", "/match?sportId=invalid", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Equal(t, helper.ErrBadRequest, resp.ErrorCode)
	})
}

func TestMatchService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTeamRepo, mockTournamentRepo, nil)

		mockMatchRepo.On("Create", mock.Anything, mock.AnythingOfType("model.Match")).Return(nil)

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
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "Match created", resp.Message)
		mockMatchRepo.AssertExpectations(t)
	})
}

func TestMatchService_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTeamRepo, mockTournamentRepo, nil)

		expectedMatch := model.Match{ID: 1, SportID: 1}
		mockMatchRepo.On("GetByID", mock.Anything, 1).Return(expectedMatch, nil)

		req := httptest.NewRequest("GET", "/match/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.GetByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		mockMatchRepo.AssertExpectations(t)
	})
}

func TestMatchService_Generate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockFirebase := new(MockFirebaseClient)
		mockFirebaseRef := new(MockFirebaseRef)

		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTeamRepo, mockTournamentRepo, mockFirebase)

		// Setup mocks
		mockSportRepo.On("GetByID", mock.Anything, 1).Return(model.Sport{ID: 1, TournamentID: 1, Slug: "futsal"}, nil)
		mockTournamentRepo.On("GetByID", mock.Anything, 1).Return(model.Tournament{ID: 1, Slug: "agi-15"}, nil)
		
		mockMatchRepo.On("Create", mock.Anything, mock.AnythingOfType("model.Match")).Return(nil)

		// Firebase mocks
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
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "Matches generated successfully", resp.Message)

		mockMatchRepo.AssertExpectations(t)
		mockSportRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockFirebase.AssertExpectations(t)
		mockFirebaseRef.AssertExpectations(t)
	})
}

func TestMatchService_Update(t *testing.T) {
	t.Run("Success SOON to LIVE", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTeamRepo, mockTournamentRepo, nil)

		defer os.RemoveAll("./storage")

		id := 1
		existingMatch := model.Match{ID: id, State: model.SOON, SportID: 1}
		mockMatchRepo.On("GetByID", mock.Anything, id).Return(existingMatch, nil)
		mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m model.Match) bool {
			return m.State == model.LIVE && m.ImageUrl != nil
		})).Return(nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("state", "LIVE")
		
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
		header.Set("Content-Type", "image/png")
		part, _ := writer.CreatePart(header)
		part.Write([]byte("fake-image-data"))
		writer.Close()

		req := httptest.NewRequest("PUT", "/match/1", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockMatchRepo.AssertExpectations(t)
	})

	t.Run("Fail SOON to LIVE without Image", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		svc := service.NewMatchService(mockMatchRepo, nil, nil, nil, nil)

		id := 1
		existingMatch := model.Match{ID: id, State: model.SOON}
		mockMatchRepo.On("GetByID", mock.Anything, id).Return(existingMatch, nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("state", "LIVE")
		writer.Close()

		req := httptest.NewRequest("PUT", "/match/1", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.Update(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "305", resp.ErrorCode)
	})

	t.Run("Success LIVE to DONE", func(t *testing.T) {
		mockMatchRepo := new(MockMatchRepository)
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewMatchService(mockMatchRepo, mockSportRepo, mockTeamRepo, mockTournamentRepo, nil)

		id := 1
		winner := "Team A"
		existingMatch := model.Match{ID: id, State: model.LIVE, SportID: 1, HomeID: helper.IntPtr(10), RoundID: 11, NextRoundID: helper.IntPtr(2)}
		mockMatchRepo.On("GetByID", mock.Anything, id).Return(existingMatch, nil)
		mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m model.Match) bool {
			return m.State == model.DONE && *m.Winner == winner
		})).Return(nil)

		// Next round update mocks
		mockTeamRepo.On("GetByID", mock.Anything, 10).Return(model.Team{ID: 10, Name: "Team A"}, nil)
		nextMatch := model.Match{ID: 2, SportID: 1}
		mockMatchRepo.On("GetByID", mock.Anything, 2).Return(nextMatch, nil)
		mockMatchRepo.On("Update", mock.Anything, mock.MatchedBy(func(m model.Match) bool {
			return m.ID == 2 && m.HomeID != nil && *m.HomeID == 10
		})).Return(nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("state", "DONE")
		writer.WriteField("winner", winner)
		writer.Close()

		req := httptest.NewRequest("PUT", "/match/1", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockMatchRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})
}
