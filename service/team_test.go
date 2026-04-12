package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTeamService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		svc := service.NewTeamService(mockRepo)

		mockRepo.On("GetByNameAndSportWithDeleted", mock.Anything, "Team A", 1).Return(model.Team{}, pgx.ErrNoRows)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("model.Team")).Return(nil)

		reqBody := model.Team{
			Name:    "Team A",
			SportID: 1,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/team", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "Team created", resp.Message)
		mockRepo.AssertExpectations(t)
	})
}

func TestTeamService_CreateBulk(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		svc := service.NewTeamService(mockRepo)

		teams := []model.Team{
			{Name: "Team A", SportID: 1},
			{Name: "Team B", SportID: 1},
		}
		mockRepo.On("CreateBulk", mock.Anything, teams).Return(nil)

		body, _ := json.Marshal(teams)
		req := httptest.NewRequest("POST", "/team/bulk", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.CreateBulk(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "Teams created", resp.Message)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RequiredFields", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		svc := service.NewTeamService(mockRepo)

		teams := []model.Team{
			{Name: "", SportID: 1},
		}

		body, _ := json.Marshal(teams)
		req := httptest.NewRequest("POST", "/team/bulk", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.CreateBulk(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		svc := service.NewTeamService(mockRepo)

		expectedTeam := model.Team{ID: 1, Name: "Team A", SportID: 1}
		mockRepo.On("GetByID", mock.Anything, 1).Return(expectedTeam, nil)

		req := httptest.NewRequest("GET", "/team/1", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.GetByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		mockRepo.AssertExpectations(t)
	})
}
