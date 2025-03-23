package auth

import (
	"auth-api/internal/domain/models"
	"auth-api/internal/lib/jwt"
	"auth-api/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log           *slog.Logger
	usrSaver      UserSaver
	usrProvider   UserProvider
	accessTTL     time.Duration
	accessSecret  string
	refreshTTL    time.Duration
	refreshSecret string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrAlreadyExist       = errors.New("user already")
)

type UserSaver interface {
	CreateUser(ctx context.Context, name string, email string, passHash []byte) (*models.UserModel, error)
	SaveRefreshToken(ctx context.Context, tokenID string, userID string, expiresAt time.Time) error
	RemoveRefreshToken(ctx context.Context, tokenID string) error
}

type UserProvider interface {
	User(ctx context.Context, email string) (*models.UserModel, error)
	UserByID(ctx context.Context, userID string) (*models.UserModel, error)
	RefreshToken(ctx context.Context, tokenID string) (userID string, err error)
}

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, accessTTL time.Duration, accessSecret string, refreshTTL time.Duration, refreshSecret string) *Auth {
	return &Auth{
		log:           log,
		usrSaver:      userSaver,
		usrProvider:   userProvider,
		accessTTL:     accessTTL,
		accessSecret:  accessSecret,
		refreshTTL:    refreshTTL,
		refreshSecret: refreshSecret,
	}
}

func (auth *Auth) Register(ctx context.Context, name string, email string, password string) (*models.UserResponse, error) {
	const op = "auth.RegisterUser"

	log := auth.log.With(slog.String("op", op))

	log.Info("Creating user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed generate password hash", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	user, err := auth.usrSaver.CreateUser(ctx, name, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("user already exists", err)
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		log.Error("failed to save user", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return auth.createTokens(ctx, user)
}

func (auth *Auth) Login(ctx context.Context, email string, password string) (*models.UserResponse, error) {
	const op = "auth.Login"

	log := auth.log.With(slog.String("op", op))

	log.Info("Get user from db")
	user, err := auth.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", err)
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Error("invalid credentials", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logined")
	return auth.createTokens(ctx, user)
}

func (auth *Auth) Refresh(ctx context.Context, refreshToken string) (*models.Refresh, error) {
	const op = "auth.Refresh"
	log := auth.log.With(slog.String("op", op))

	claims, err := jwt.ParseRefreshToken(refreshToken, auth.refreshSecret)
	if err != nil {
		log.Error("invalid refresh token", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	userID, err := auth.usrProvider.RefreshToken(ctx, claims.TokenID)
	if err != nil {
		log.Error("refresh token not found", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	user, err := auth.usrProvider.UserByID(ctx, userID)
	if err != nil {
		log.Error("user not found", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	if err := auth.usrSaver.RemoveRefreshToken(ctx, claims.TokenID); err != nil {
		log.Error("failed to remove old refresh token", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	data, err := auth.createTokens(ctx, user)
	if err != nil {
		log.Error("failed to generate tokens", err)
		return nil, err
	}
	return &models.Refresh{
		Token:        data.Token,
		RefreshToken: data.RefreshToken,
	}, nil
}

func (auth *Auth) GetUser(ctx context.Context, token string) (*models.UserInfo, error) {
	const op = "auth.GetUser"

	log := auth.log.With(slog.String("op", op))

	claims, err := jwt.ParseAccessToken(token, auth.accessSecret)
	if err != nil {
		log.Error("failed to parse access token", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	user, err := auth.usrProvider.User(ctx, claims.UserID)
	if err != nil {
		log.Error("failed to get user by access token", err)
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	return &models.UserInfo{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (auth *Auth) createTokens(ctx context.Context, user *models.UserModel) (*models.UserResponse, error) {
	token, err := jwt.NewAccessToken(*user, auth.accessSecret, auth.accessTTL)
	if err != nil {
		return nil, err
	}

	tokenID := uuid.New().String()
	refresh, err := jwt.NewRefreshToken(user.ID, tokenID, auth.refreshSecret, auth.refreshTTL)
	if err != nil {
		return nil, err
	}

	if err := auth.usrSaver.SaveRefreshToken(ctx, tokenID, user.ID, time.Now().Add(auth.refreshTTL)); err != nil {
		return nil, err
	}

	return &models.UserResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		Token:        token,
		RefreshToken: refresh,
	}, nil
}
