package service_test

import (
	"context"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestTournamentService_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTournamentRepository)
		svc := service.NewTournamentService(mockRepo)

		expectedTournament := model.Tournament{ID: 1, Name: "AGI 15", Slug: "agi-15"}
		mockRepo.On("GetBySlug", "agi-15").Return(expectedTournament, nil)

		req := httptest.NewRequest("GET", "/tournament/agi-15", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "agi-15")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		svc.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

// Additional tests for Create, Update (Multipart) would be similar to SportService tests
