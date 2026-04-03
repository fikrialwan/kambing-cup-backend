package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"kambing-cup-backend/helper"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Login(t *testing.T) {
	// Set SECRET env var for token generation
	os.Setenv("SECRET", "testsecret")

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := service.NewAuthService(mockRepo)

		user := model.User{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
			Role:     "ADMIN",
			Password: "hashedpassword", // In real app, this should be checked against hash
		}
		// Note: The service currently compares plain text password for simplicity in GetByEmailPassword query
		// Adjust if you implemented hashing
		
		mockRepo.On("GetByEmailPassword", mock.Anything, "admin@example.com", "password").Return(user, nil)

		reqBody := service.LoginRequest{
			Email:    "admin@example.com",
			Password: "password",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp helper.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		
		dataMap := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, dataMap["token"])
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := service.NewAuthService(mockRepo)

		mockRepo.On("GetByEmailPassword", mock.Anything, "admin@example.com", "wrong").Return(model.User{}, errors.New("invalid"))

		reqBody := service.LoginRequest{
			Email:    "admin@example.com",
			Password: "wrong",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp helper.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, helper.ErrAuthInvalidCredentials, resp.ErrorCode)
		mockRepo.AssertExpectations(t)
	})
}
