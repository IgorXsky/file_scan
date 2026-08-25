package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IgorXsky/file-scan/internal/clamav"
	"github.com/IgorXsky/file-scan/internal/config"
	"github.com/IgorXsky/file-scan/internal/handler"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	client := clamav.NewClient(cfg.ClamAVAddr(), cfg.ClamAVTimeout)

	mux := http.NewServeMux()
	mux.Handle("POST /scan", handler.NewScanHandler(client, cfg.MaxFileSizeBytes))
	mux.Handle("GET /health", handler.NewHealthHandler(client))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ClamAVTimeout + 10*time.Second,
		WriteTimeout: cfg.ClamAVTimeout + 10*time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
