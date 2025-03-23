package storage

import "errors"

var (
	ErrUserExists        = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrTokenNotFound     = errors.New("refresh token not found or expired")
	ErrTokenSaveFailed   = errors.New("failed to save refresh token")
	ErrTokenRemoveFailed = errors.New("failed to remove refresh token")
)
