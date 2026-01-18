package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByEmailPassword(email string, password string) (model.User, error)
	GetById(id int) (model.User, error)
	Create(user model.CreateUserRequest) error
	Update(user model.UpdateUserRequest) error
	Delete(id int) error
	SuperadminExists() (bool, error)
	CreateSuperadmin(username, email, password string) error
	GetSuperadminByUsername(username string) (model.User, error)
	UpdateSuperadminEmail(id int, email string) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (u *userRepository) GetAll() ([]model.User, error) {
	var users []model.User
	rows, err := u.pool.Query(context.Background(), "SELECT id, username, email, password, role, created_at, updated_at, deleted_at FROM users WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return users, err
	}

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			log.Print(err.Error())
			return users, err
		}
		users = append(users, user)
	}

	return users, err
}

func (u *userRepository) GetByEmailPassword(email string, password string) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(context.Background(), "SELECT id, username, email, password, role, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND password = $2 AND deleted_at IS NULL", email, password).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	return user, err
}

func (u *userRepository) GetById(id int) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(context.Background(), "SELECT id, username, email, password, role, created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL", id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	return user, err
}

func (u *userRepository) Create(user model.CreateUserRequest) error {
	_, err := u.pool.Exec(context.Background(), "INSERT INTO users (username, email, password, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", user.Username, user.Email, user.Password, user.Role, time.Now(), time.Now())

	return err
}

func (u *userRepository) Update(user model.UpdateUserRequest) error {
	_, err := u.pool.Exec(context.Background(), "UPDATE users SET username = $1, email = $2, password = $3, role = $4, updated_at = $6 WHERE id = $5", user.Username, user.Email, user.Password, user.Role, user.ID, time.Now())

	return err
}

func (u *userRepository) Delete(id int) error {
	_, err := u.pool.Exec(context.Background(), "UPDATE users SET deleted_at = $1 WHERE id = $2", time.Now(), id)

	return err
}

func (u *userRepository) SuperadminExists() (bool, error) {
	var exists bool
	err := u.pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE role = 'SUPERADMIN' AND deleted_at IS NULL)").Scan(&exists)
	return exists, err
}

func (u *userRepository) CreateSuperadmin(username, email, password string) error {
	_, err := u.pool.Exec(context.Background(),
		"INSERT INTO users (username, email, password, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		username, email, password, "SUPERADMIN", time.Now(), time.Now())
	return err
}

func (u *userRepository) GetSuperadminByUsername(username string) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(context.Background(), "SELECT id, username, email, password, role, created_at, updated_at, deleted_at FROM users WHERE username = $1 AND role = 'SUPERADMIN' AND deleted_at IS NULL", username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	return user, err
}

func (u *userRepository) UpdateSuperadminEmail(id int, email string) error {
	_, err := u.pool.Exec(context.Background(), "UPDATE users SET email = $1, updated_at = $2 WHERE id = $3", email, time.Now(), id)
	return err
}