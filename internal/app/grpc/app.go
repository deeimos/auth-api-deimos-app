package grpcapp

import (
	authgrpc "auth-api/internal/grpc/auth"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (app *App) Run() {
	if err := app.run(); err != nil {
		panic(err)
	}
}

func (app *App) run() error {
	const op = "grpcApp.Run"

	log := app.log.With(slog.String("op", op), slog.Int("port", app.port))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", app.port))

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC is running", slog.String("addr", listener.Addr().String()))

	if err := app.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (app *App) Stop() {
	const op = "grpc.App"

	app.log.With(slog.String("op", op)).Info("stoping gRPC server", slog.Int("port", app.port))

	app.gRPCServer.GracefulStop()
}
