package service_test

import (
	"bytes"
	"context"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTournamentService_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockTournamentRepository)
		svc := service.NewTournamentService(mockRepo, ".")

		expectedTournament := model.Tournament{ID: 1, Name: "AGI 15", Slug: "agi-15"}
		mockRepo.On("GetBySlug", mock.Anything, "agi-15").Return(expectedTournament, nil)

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

func TestTournamentService_Create(t *testing.T) {
	t.Run("SuccessCreate", func(t *testing.T) {
		mockRepo := new(MockTournamentRepository)
		svc := service.NewTournamentService(mockRepo, ".")

		defer os.RemoveAll("./storage")

		mockRepo.On("GetBySlugWithDeleted", mock.Anything, "agi-15").Return(model.Tournament{}, pgx.ErrNoRows)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("model.Tournament")).Return(nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "AGI 15")
		writer.WriteField("total_surah", "114")
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
		header.Set("Content-Type", "image/png")
		part, _ := writer.CreatePart(header)
		part.Write([]byte("fake-image-data"))
		writer.Close()

		req := httptest.NewRequest("POST", "/tournament", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "Tournament created", w.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("SuccessRestore", func(t *testing.T) {
		mockRepo := new(MockTournamentRepository)
		svc := service.NewTournamentService(mockRepo, ".")

		defer os.RemoveAll("./storage")

		deletedAt := time.Now()
		existing := model.Tournament{ID: 1, Name: "AGI 15", Slug: "agi-15", DeletedAt: &deletedAt}
		mockRepo.On("GetBySlugWithDeleted", mock.Anything, "agi-15").Return(existing, nil)
		mockRepo.On("Restore", mock.Anything, mock.AnythingOfType("model.Tournament")).Return(nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "AGI 15")
		writer.WriteField("total_surah", "114")
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
		header.Set("Content-Type", "image/png")
		part, _ := writer.CreatePart(header)
		part.Write([]byte("fake-image-data"))
		writer.Close()

		req := httptest.NewRequest("POST", "/tournament", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Tournament restored", w.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("SlugTaken", func(t *testing.T) {
		mockRepo := new(MockTournamentRepository)
		svc := service.NewTournamentService(mockRepo, ".")

		existing := model.Tournament{ID: 1, Name: "AGI 15", Slug: "agi-15", DeletedAt: nil}
		mockRepo.On("GetBySlugWithDeleted", mock.Anything, "agi-15").Return(existing, nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "AGI 15")
		writer.WriteField("total_surah", "114")
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
		header.Set("Content-Type", "image/png")
		part, _ := writer.CreatePart(header)
		part.Write([]byte("fake-image-data"))
		writer.Close()

		req := httptest.NewRequest("POST", "/tournament", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		svc.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Slug is already taken\n", w.Body.String())
		mockRepo.AssertExpectations(t)
	})
}
