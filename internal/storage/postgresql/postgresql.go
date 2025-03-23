package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"auth-api/internal/config"
	"auth-api/internal/domain/models"
	"auth-api/internal/storage"

	"github.com/lib/pq"
)

type s struct {
	db *sql.DB
}

func New(log *slog.Logger, config config.Config) (*s, error) {
	const op = "s.postgresql.New"

	connectData := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connectData)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	log.Info("Connected to PostgreSQL",
		slog.String("host", config.Database.Host),
		slog.Int("port", config.Database.Port),
	)
	return &s{db: db}, nil
}

func (s *s) Stop() error {
	return s.db.Close()
}

func (s *s) CreateUser(ctx context.Context, name string, email string, passHash []byte) (*models.UserModel, error) {
	const query = `
		INSERT INTO users (email, name, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, name, password_hash, created_at`

	createdAt := time.Now().UTC()
	// row := s.db.QueryRowContext(ctx, query, email, name, passHash, createdAt)

	// var user models.UserModel
	// if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
	// 	return nil, fmt.Errorf("CreateUser: %w", err)
	// }
	var user models.UserModel
	err := s.db.QueryRowContext(ctx, query, email, name, passHash, createdAt).
		Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, storage.ErrUserExists
		}
		return nil, fmt.Errorf("CreateUser: %w", err)
	}

	return &user, nil
}

func (s *s) SaveRefreshToken(ctx context.Context, tokenID string, userID string, expiresAt time.Time) error {
	const query = `
		INSERT INTO refresh_tokens (token_id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := s.db.ExecContext(ctx, query, tokenID, userID, expiresAt, time.Now().UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrTokenSaveFailed
	}
	return err
}

func (s *s) RemoveRefreshToken(ctx context.Context, tokenID string) error {
	const query = `DELETE FROM refresh_tokens WHERE token_id = $1`
	_, err := s.db.ExecContext(ctx, query, tokenID)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrTokenNotFound
	}
	return err
}

func (s *s) User(ctx context.Context, email string) (*models.UserModel, error) {
	const query = `SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1`
	row := s.db.QueryRowContext(ctx, query, email)

	var user models.UserModel
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("User: %w", err)
	}

	return &user, nil
}

func (s *s) UserByID(ctx context.Context, userID string) (*models.UserModel, error) {
	const query = `SELECT id, email, name, password_hash, created_at FROM users WHERE id = $1`
	row := s.db.QueryRowContext(ctx, query, userID)

	var user models.UserModel
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserByID: %w", err)
	}

	return &user, nil
}

func (s *s) RefreshToken(ctx context.Context, tokenID string) (string, error) {
	const query = `SELECT user_id FROM refresh_tokens WHERE token_id = $1 AND expires_at > now()`
	row := s.db.QueryRowContext(ctx, query, tokenID)

	var userID string
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUserNotFound
		}
		return "", fmt.Errorf("RefreshToken: %w", err)
	}

	return userID, nil
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
