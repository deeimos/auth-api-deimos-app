package jwt

import (
	"auth-api/internal/domain/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type AccessClaims struct {
	UserID string
	Email  string
	Name   string
}

type RefreshClaims struct {
	UserID  string
	TokenID string
}

func NewAccessToken(user models.UserModel, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewRefreshToken(userID string, tokenID string, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"token_id": tokenID,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseAccessToken(tokenStr string, secret string) (*AccessClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	email, okEmail := claims["email"].(string)
	name, okName := claims["name"].(string)

	if !ok || !okEmail || !okName {
		return nil, ErrInvalidToken
	}

	return &AccessClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
	}, nil
}

func ParseRefreshToken(tokenStr string, secret string) (*RefreshClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	tokenID, ok := claims["token_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &RefreshClaims{
		UserID:  userID,
		TokenID: tokenID,
	}, nil
}
