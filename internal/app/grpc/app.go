package grpcapp

import (
	authgrpc "auth-api/internal/grpc/auth"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
	addr := fmt.Sprintf(":%d", app.port)

	log.Info("gRPC (h2c) is running", slog.String("addr", addr))

	h2s := &http2.Server{}

	handler := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.gRPCServer.ServeHTTP(w, r)
	}), h2s)

	return http.ListenAndServe(addr, handler)
}

func (app *App) Stop() {
	const op = "grpc.App"

	app.log.With(slog.String("op", op)).Info("stoping gRPC server", slog.Int("port", app.port))

	app.gRPCServer.GracefulStop()
}
