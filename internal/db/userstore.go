package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrInvalidUserData  = errors.New("invalid user data")
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, email, name string) (*User, error) {
	if email == "" || name == "" {
		return nil, ErrInvalidUserData
	}

	var user User
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, name, status, created_at, updated_at)
		VALUES ($1, $2, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, email, name, status, created_at, updated_at`,
		email, name,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation on email
		if isPgUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, email, name, status, created_at, updated_at
		FROM users
		WHERE id = $1 AND status != 'deleted'`,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) Update(ctx context.Context, id int, email, name string) (*User, error) {
	if email == "" || name == "" {
		return nil, ErrInvalidUserData
	}

	var user User
	err := s.db.QueryRowContext(
		ctx,
		`UPDATE users
		SET email = $1, name = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND status != 'deleted'
		RETURNING id, email, name, status, created_at, updated_at`,
		email, name, id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) Delete(ctx context.Context, id int) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE users 
		SET status = 'deleted', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status != 'deleted'`,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *UserStore) List(ctx context.Context, limit, offset int) ([]User, error) {
	if limit <= 0 {
		limit = 50 // default limit
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, email, name, status, created_at, updated_at
		FROM users
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, email, name, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND status != 'deleted'`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Helper function to check for postgres unique violation
func isPgUniqueViolation(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "23505"
}

