package users

import (
	"chariottakehome/internal/database"
	id "chariottakehome/internal/identifier"
	"context"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string) (*User, error)
}

type userRepository struct {
	database *database.DatabasePool
}

func NewRepo(database *database.DatabasePool) UserRepository {
	return &userRepository{database}

}

func (r *userRepository) CreateUser(ctx context.Context, email string) (*User, error) {
	id, err := id.New()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := User{
		Id:        id,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sql, args := prepareInsertUser(user)
	_, err = r.database.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
