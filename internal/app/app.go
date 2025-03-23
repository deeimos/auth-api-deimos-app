package app

import (
	grpcapp "auth-api/internal/app/grpc"
	"auth-api/internal/config"
	"auth-api/internal/services/auth"
	"auth-api/internal/storage/postgresql"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, config config.Config) *App {
	storage, err := postgresql.New(log, config)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage, storage, config.AccessTTL, config.AccessSecret, config.RefreshTTL, config.RefreshSecret)
	grpcApp := grpcapp.New(log, authService, config.GRPCConfig.Port)
	return &App{GRPCServer: grpcApp}
}
