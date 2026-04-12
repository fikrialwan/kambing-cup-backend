package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetAll(ctx context.Context) ([]model.User, error)
	GetByEmailPassword(ctx context.Context, email string, password string) (model.User, error)
	GetById(ctx context.Context, id int) (model.User, error)
	Create(ctx context.Context, user model.CreateUserRequest) error
	Update(ctx context.Context, user model.UpdateUserRequest) error
	Delete(ctx context.Context, id int) error
	SuperadminExists(ctx context.Context) (bool, error)
	CreateSuperadmin(ctx context.Context, username, email, password string) error
	GetSuperadminByUsername(ctx context.Context, username string) (model.User, error)
	UpdateSuperadminEmail(ctx context.Context, id int, email string) error
	GetByUsernameOrEmail(ctx context.Context, username, email string) (model.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (u *userRepository) GetAll(ctx context.Context) ([]model.User, error) {
	var users []model.User
	rows, err := u.pool.Query(ctx, "SELECT id, username, email, password, role, created_at, updated_at FROM users")
	if err != nil {
		log.Print(err.Error())
		return users, err
	}

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Print(err.Error())
			return users, err
		}
		users = append(users, user)
	}

	return users, err
}

func (u *userRepository) GetByEmailPassword(ctx context.Context, email string, password string) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(ctx, "SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE email = $1 AND password = $2", email, password).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (u *userRepository) GetById(ctx context.Context, id int) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(ctx, "SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (u *userRepository) Create(ctx context.Context, user model.CreateUserRequest) error {
	_, err := u.pool.Exec(ctx, "INSERT INTO users (username, email, password, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", user.Username, user.Email, user.Password, user.Role, time.Now(), time.Now())

	return err
}

func (u *userRepository) Update(ctx context.Context, user model.UpdateUserRequest) error {
	_, err := u.pool.Exec(ctx, "UPDATE users SET username = $1, email = $2, password = $3, role = $4, updated_at = $6 WHERE id = $5", user.Username, user.Email, user.Password, user.Role, user.ID, time.Now())

	return err
}

func (u *userRepository) Delete(ctx context.Context, id int) error {
	_, err := u.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)

	return err
}

func (u *userRepository) SuperadminExists(ctx context.Context) (bool, error) {
	var exists bool
	err := u.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE role = 'SUPERADMIN')").Scan(&exists)
	return exists, err
}

func (u *userRepository) CreateSuperadmin(ctx context.Context, username, email, password string) error {
	_, err := u.pool.Exec(ctx,
		"INSERT INTO users (username, email, password, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		username, email, password, "SUPERADMIN", time.Now(), time.Now())
	return err
}

func (u *userRepository) GetSuperadminByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(ctx, "SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE username = $1 AND role = 'SUPERADMIN'", username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (u *userRepository) UpdateSuperadminEmail(ctx context.Context, id int, email string) error {
	_, err := u.pool.Exec(ctx, "UPDATE users SET email = $1, updated_at = $2 WHERE id = $3", email, time.Now(), id)
	return err
}

func (u *userRepository) GetByUsernameOrEmail(ctx context.Context, username, email string) (model.User, error) {
	var user model.User
	err := u.pool.QueryRow(ctx, "SELECT id, username, email, password, role, created_at, updated_at FROM users WHERE username = $1 OR email = $2", username, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}