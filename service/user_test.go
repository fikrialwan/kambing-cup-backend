package service_test

import (
	"bytes"
	"encoding/json"
	"kambing-cup-backend/model"
	"kambing-cup-backend/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := service.NewUserService(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("model.CreateUserRequest")).Return(nil)

		reqBody := model.CreateUserRequest{
			Username: "testuser",
			Password: "password",
			Role:     "ADMIN",
			Email:    "test@example.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/user", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.CreateUser(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "User created", w.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidRole", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := service.NewUserService(mockRepo)

		reqBody := model.CreateUserRequest{
			Username: "testuser",
			Password: "password",
			Role:     "USER", // Invalid
			Email:    "test@example.com",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/user", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		svc.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestUserService_GetUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		svc := service.NewUserService(mockRepo)

		expectedUser := model.User{ID: 1, Username: "testuser"}
		mockRepo.On("GetById", 1).Return(expectedUser, nil)

		req := httptest.NewRequest("GET", "/user", nil)
		req.Header.Set("x-user-id", "1")
		w := httptest.NewRecorder()

		svc.GetUser(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
