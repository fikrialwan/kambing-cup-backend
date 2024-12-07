package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	conn *pgx.Conn
}

func NewUserRepository(conn *pgx.Conn) *UserRepository {
	return &UserRepository{conn: conn}
}

func (u *UserRepository) GetAll() ([]model.User, error) {
	var users []model.User
	rows, err := u.conn.Query(context.Background(), "SELECT * FROM users WHERE deleted_at IS NULL")
	if err != nil {
		log.Print(err.Error())
		return users, err
	}

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			log.Print(err.Error())
			return users, err
		}
		users = append(users, user)
	}

	return users, err
}

func (u *UserRepository) GetByUsernamePassword(username string, password string) (model.User, error) {
	var user model.User
	err := u.conn.QueryRow(context.Background(), "SELECT * FROM users WHERE username = $1 AND password = $2 AND deleted_at IS NULL", username, password).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	return user, err
}

func (u *UserRepository) GetById(id int) (model.User, error) {
	var user model.User
	err := u.conn.QueryRow(context.Background(), "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", id).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	return user, err
}

func (u *UserRepository) Create(user model.CreateUserRequest) error {
	_, err := u.conn.Exec(context.Background(), "INSERT INTO users (username, password, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", user.Username, user.Password, user.Role, time.Now(), time.Now())

	return err
}

func (u *UserRepository) Update(user model.UpdateUserRequest) error {
	_, err := u.conn.Exec(context.Background(), "UPDATE users SET username = $1, password = $2, role = $3, updated_at = $5 WHERE id = $4", user.Username, user.Password, user.Role, user.ID, time.Now())

	return err
}

func (u *UserRepository) Delete(id int) error {
	_, err := u.conn.Exec(context.Background(), "UPDATE users SET deleted_at = $1 WHERE id = $2", time.Now(), id)

	return err
}
