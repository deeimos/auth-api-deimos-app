package main

import (
	"auth-api/internal/app"
	"auth-api/internal/config"
	"os"
	"os/signal"
	"syscall"

	"log/slog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	config := config.MustLoad()

	log := setupLogger(config.Env)
	log.Info("Starting server", slog.String("env", config.Env))
	log.Debug("debug messages enabled")

	// if err := db.InitDB(config, log); err != nil {
	// 	log.Error("Failed to connect to PostgreSQL", slog.String("error", err.Error()))
	// 	os.Exit(1)
	// }

	// log.Info("PostgreSQL connection established")

	application := app.New(log, *config)
	go application.GRPCServer.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	stoped := <-stop

	log.Info("stoping application", slog.String("signal", stoped.String()))
	application.GRPCServer.Stop()
	log.Info("application stoped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
