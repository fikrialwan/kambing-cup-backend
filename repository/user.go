package repository

import (
	"kambing-cup-backend/model"
	"time"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (u *UserRepository) GetAll() []*model.User {
	return []*model.User{{
		ID:        1,
		Username:  "admin",
		Password:  "admin",
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: time.Now(),
	}}
}
