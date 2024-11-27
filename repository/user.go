package repository

import (
	"context"
	"kambing-cup-backend/model"
	"log"

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
