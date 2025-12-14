package main

import (
	"log/slog"
	"os"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/escalopa/family-tree/internal/pkg/logger"
	"github.com/escalopa/family-tree/internal/server"

	_ "github.com/escalopa/family-tree/docs" // Swagger docs
)

// @title Family Tree API
// @version 1.0
// @description API for managing family tree members, users, and relationships
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@familytree.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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
