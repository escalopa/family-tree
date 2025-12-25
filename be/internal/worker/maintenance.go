package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type MaintenanceWorker struct {
	sessionCleaner    Cleaner
	oauthStateCleaner Cleaner
	interval          time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
	startOnce         sync.Once
	wg                sync.WaitGroup
}

func NewMaintenanceWorker(
	sessionCleaner Cleaner,
	oauthStateCleaner Cleaner,
	interval time.Duration,
) *MaintenanceWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &MaintenanceWorker{
		sessionCleaner:    sessionCleaner,
		oauthStateCleaner: oauthStateCleaner,
		interval:          interval,
		ctx:               ctx,
		cancel:            cancel,
	}
}

func (w *MaintenanceWorker) Start() {
	w.startOnce.Do(func() {
		w.wg.Add(1)
		go w.run()
	})
}

func (w *MaintenanceWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *MaintenanceWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	slog.Info("Maintenance worker started", "cleanup_interval", w.interval.String())
	defer slog.Info("Maintenance worker stopped")

	for {
		select {
		case <-ticker.C:
			w.performCleanup()

		case <-w.ctx.Done():
			slog.Info("Maintenance worker stopped gracefully")
			return
		}
	}
}

func (w *MaintenanceWorker) performCleanup() {
	g, ctx := errgroup.WithContext(w.ctx)

	g.Go(func() error {
		if err := w.sessionCleaner.CleanExpired(ctx); err != nil {
			slog.Error("Maintenance worker: clean expired sessions", "error", err)
			return err
		}
		slog.Debug("Maintenance worker: cleaned expired sessions")
		return nil
	})

	g.Go(func() error {
		if err := w.oauthStateCleaner.CleanExpired(ctx); err != nil {
			slog.Error("Maintenance worker: clean expired OAuth states", "error", err)
			return err
		}
		slog.Debug("Maintenance worker: cleaned expired OAuth states")
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Warn("Maintenance worker: cleanup completed with errors", "error", err)
	}
}
