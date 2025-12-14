package main

import (
	"log/slog"
	"os"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/pkg/logger"
	"github.com/escalopa/family-tree/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("main: load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg.Server.LogLevel)

	app, err := server.NewApp(cfg)
	if err != nil {
		slog.Error("main: initialize application", "error", err)
		os.Exit(1)
	}

	srv := server.NewServer(app)
	if err := srv.Run(); err != nil {
		slog.Error("main: run server", "error", err)
		os.Exit(1)
	}
}
