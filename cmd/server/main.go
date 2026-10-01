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
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/henryorsborn/ala_service/internal/ala_service"
)

func main() {
	db, err := ala_service.OpenStore()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	h := &ala_service.Handler{DB: db}

	r := gin.Default()

	// CORS middleware: allow the React dev server (Vite on 5173, CRA on 3000)
	// to call this API. In production this should be restricted to the actual
	// frontend origin, not wildcarded.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/v1/aliases", h.CreateAlias)
	r.GET("/v1/aliases", h.GetAliases)
	r.GET("/:alias_url", h.Redirect)

	log.Println("url-shortener listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
