package repository

import (
	import "blog-api/internal/model"
	"context"
	"database/sql"
	"fmt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email, password, created) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, user.Email, user.Password, user.Created).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, password, created FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, password, created FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Created,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}
