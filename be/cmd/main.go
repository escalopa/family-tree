package main

import (
	"log/slog"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/pkg/logger"
	"github.com/escalopa/family-tree/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration (uses CONFIG_PATH env var)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("main: load config", "error", err)
		panic(err)
	}

	// Setup logger with configured log level
	logger.Setup(cfg.Server.LogLevel)

	// Initialize application
	app, err := server.NewApp(cfg)
	if err != nil {
		slog.Error("main: initialize application", "error", err)
		panic(err)
	}

	// Create and run server
	srv := server.NewServer(app)
	if err := srv.Run(); err != nil {
		slog.Error("main: run server", "error", err)
		panic(err)
	}
}
