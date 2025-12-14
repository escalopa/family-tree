package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	app *App
	srv *http.Server
}

func NewServer(app *App) *Server {
	srv := &http.Server{
		Addr:         ":" + app.cfg.Server.Port,
		Handler:      app.engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return &Server{
		app: app,
		srv: srv,
	}
}

func (s *Server) Run() error {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server.Run: start server", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Server.Run: started", "port", s.app.cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server.Run: shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server.Run: shutdown forced", "error", err)
		return err
	}

	s.app.Close()

	slog.Info("Server.Run: exited")
	return nil
}
