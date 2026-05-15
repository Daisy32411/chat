package repository

import (
	"context"
	"database/sql"
	"strings"

	"chat-app/internal/models"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, username, email, created_at
    `, username, email, passwordHash).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByUsernameOrEmail(ctx context.Context, value string) (*models.User, string, error) {
	var u models.User
	var passwordHash string
	err := r.db.QueryRowContext(ctx, `
        SELECT id, username, email, password_hash, created_at
        FROM users
        WHERE username = $1 OR email = $1
    `, value).Scan(&u.ID, &u.Username, &u.Email, &passwordHash, &u.CreatedAt)
	if err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, `
        SELECT id, username, email, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Search(ctx context.Context, query string, limit int) ([]models.User, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []models.User{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT id, username, email, created_at
        FROM users
        WHERE username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'
        ORDER BY username
        LIMIT $2
    `, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
