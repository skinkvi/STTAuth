package grpcapp

import (
	authgrpc "STTAuth/internal/grpc/auth"
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

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) SetServer(server *grpc.Server) {
	a.gRPCServer = server
}

func (a *App) Run() error {
	const op = "grpcapp.Run" // типо operation

	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	listerner, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running ", slog.String("addr", listerner.Addr().String()))

	if err = a.gRPCServer.Serve(listerner); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping grpc server", slog.Int("port", a.port))

	// Очень крутая функция она не сразу выключит приложение а сначала завершить все выходящие запросы и не будет принимать новые запросы
	a.gRPCServer.GracefulStop()
}
