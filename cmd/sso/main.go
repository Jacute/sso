package main

import (
	"log/slog"
	"os"
	"sso/internal/app"
	"sso/internal/config"

	"github.com/jacute/prettylogger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupPrettyLogger(cfg.Env)
	log.Info(
		"Starting app",
		slog.Any("config", cfg),
	)

	application := app.New(
		log,
		cfg.GRPC.Port,
		cfg.StoragePath,
		cfg.TokenTTL,
	)
	application.GrpcServer.MustRun()
}

func setupPrettyLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			prettylogger.NewColoredHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			prettylogger.NewColoredHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			prettylogger.NewColoredHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
