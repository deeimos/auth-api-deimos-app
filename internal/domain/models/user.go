package models

import "time"

type UserModel struct {
	ID           string
	Email        string
	Name         string
	PasswordHash []byte
	CreatedAt    time.Time
}

type UserInfo struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
}

type UserResponse struct {
	ID           string
	Email        string
	Name         string
	CreatedAt    time.Time
	Token        string
	RefreshToken string
}

type Refresh struct {
	Token        string
	RefreshToken string
}
