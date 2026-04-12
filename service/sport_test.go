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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSportService_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		defer os.RemoveAll("./storage")

		// Mock Tournament GetByID
		mockTournamentRepo.On("GetByID", mock.Anything, 1).Return(model.Tournament{ID: 1}, nil)

		// Mock Sport Check
		mockSportRepo.On("GetByNameAndTournamentWithDeleted", mock.Anything, "Futsal", 1).Return(model.Sport{}, pgx.ErrNoRows)

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
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "Sport created", resp.Message)

		mockTournamentRepo.AssertExpectations(t)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("TournamentNotFound", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		// Mock Tournament GetByID to fail
		mockTournamentRepo.On("GetByID", mock.Anything, 999).Return(model.Tournament{}, pgx.ErrNoRows)

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

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Equal(t, helper.ErrSportTournamentNotFound, resp.ErrorCode)

		mockTournamentRepo.AssertExpectations(t)
		mockSportRepo.AssertNotCalled(t, "Create")
	})

	t.Run("ImageSizeTooLarge", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		defer os.RemoveAll("./storage")

		// Mock Tournament GetByID
		mockTournamentRepo.On("GetByID", mock.Anything, 1).Return(model.Tournament{ID: 1}, nil)

		// Mock Sport Check
		mockSportRepo.On("GetByNameAndTournamentWithDeleted", mock.Anything, "Futsal", 1).Return(model.Sport{}, pgx.ErrNoRows)

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
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Image size must be less than 2MB")

		mockTournamentRepo.AssertExpectations(t)
		mockSportRepo.AssertNotCalled(t, "Create")
	})

	t.Run("GetByID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		expectedSport := model.Sport{ID: 1, Name: "Futsal"}
		mockSportRepo.On("GetByID", mock.Anything, 1).Return(expectedSport, nil)

		req := httptest.NewRequest("GET", "/sport/1", nil)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		w := httptest.NewRecorder()

		svc.GetByID(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
	})
}

func TestSportService_GetAll(t *testing.T) {
	t.Run("SuccessWithoutFilter", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		expectedSports := []model.Sport{{ID: 1, Name: "Futsal"}}
		mockSportRepo.On("GetAll", mock.Anything, 0).Return(expectedSports, nil)

		req := httptest.NewRequest("GET", "/sport", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("SuccessWithFilter", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		expectedSports := []model.Sport{{ID: 1, TournamentID: 1, Name: "Futsal"}}
		mockSportRepo.On("GetAll", mock.Anything, 1).Return(expectedSports, nil)

		req := httptest.NewRequest("GET", "/sport?tournamentId=1", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("InvalidFilter", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		svc := service.NewSportService(mockSportRepo, mockTournamentRepo, ".", nil)

		req := httptest.NewRequest("GET", "/sport?tournamentId=abc", nil)
		w := httptest.NewRecorder()

		svc.GetAll(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp helper.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp.Success)
		assert.Equal(t, helper.ErrBadRequest, resp.ErrorCode)
		assert.Contains(t, resp.Message, "Invalid Tournament ID")
	})
}
