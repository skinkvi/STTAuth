package app

import (
	grpcapp "STTAuth/internal/app/grpc"
	"STTAuth/internal/config"
	"STTAuth/internal/services/auth"
	"STTAuth/internal/storage/postgre"

	"log/slog"

	_ "github.com/lib/pq"
)

type App struct {
	GRPCSrv *grpcapp.App
	Storage *postgre.Storage
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	storage, err := postgre.NewPostgreStorage(log, cfg.Storage.Postgres.URL)
	if err != nil {
		return nil, err
	}
	authService := auth.New(log, storage, storage, storage, cfg.TokenTTL)
	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)
	return &App{
		GRPCSrv: grpcApp,
		Storage: storage,
	}, nil
}
