// Package main is the entrypoint for ala_service.
//
// This file is part of the servicectl go-webapi scaffold. It's intentionally
// small — the real HTTP handlers live in internal/ala_service/.
//
// What this scaffold gives you out of the box:
//
//   - Two endpoints: /healthz (always 200 if the process is up) and /readyz
//     (200 once any readiness checks pass; 503 otherwise).
//   - Graceful shutdown on SIGINT/SIGTERM with a 10s drain window.
//   - The server listens on $PORT (default 3000) on all interfaces so it
//     works in containers without surprises.
//
// What's intentionally NOT here (and where to add it):
//
//   - Structured logging: drop in slog or zap in internal/ala_service/logger.go.
//   - OpenTelemetry: wrap the http.Server handlers in an otelhttp.NewHandler.
//   - DB connection: open a *sql.DB in main and pass it into the Server.
//   - Auth: middleware in internal/ala_service/middleware.go.
//
// Edit freely. The scaffold is yours.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/henryorsborn/ala_service/internal/ala_service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           ala_service.NewMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Run the server in a goroutine so we can listen for shutdown signals
	// on the main goroutine.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("ala_service listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// Wait for a signal or a fatal server error.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		log.Printf("received signal %s; shutting down", sig)
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
		return
	}

	// Drain in-flight requests for up to 10 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		os.Exit(1)
	}
	log.Printf("shutdown complete")
}