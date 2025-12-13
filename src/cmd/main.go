//go:build !test

// Package main provides entry point for the Tether service
// @title Tether API
// @version 1.0.0
// @description A modern, scalable short URL service
// @termsOfService https://github.com/Kosha-Nirman/tether

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:5000
// @BasePath /
// @schemes http https

// @tag.name Health
// @tag.description Health check endpoints

// @tag.name Links
// @tag.description Short link operations

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Kosha-Nirman/tether/docs" // Import generated docs
	"github.com/Kosha-Nirman/tether/src/api/handlers"
	"github.com/Kosha-Nirman/tether/src/api/routes"
	"github.com/Kosha-Nirman/tether/src/internal/repository"
	"github.com/Kosha-Nirman/tether/src/internal/service"
	"github.com/Kosha-Nirman/tether/src/pkg/cache"
	"github.com/Kosha-Nirman/tether/src/pkg/config"
	"github.com/Kosha-Nirman/tether/src/pkg/database"
	"github.com/gin-gonic/gin"
)

func run() error {
	log.Println("🔗 Starting Tether...")

	// ? Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("❌ Failed to load configuration: %w", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// * Create context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("🚀 Starting Tether server on port %d", cfg.Server.Port)

	// ? Initialize Database
	dbConfig := database.Config{
		URI:         cfg.Database.URI,
		Name:        cfg.Database.Name,
		MaxPoolSize: cfg.Database.MaxPoolSize,
		MinPoolSize: cfg.Database.MinPoolSize,
	}

	db, err := database.Connect(ctx, dbConfig)
	if err != nil {
		return fmt.Errorf("❌ Failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("⚠️ Error closing database connection: %v", err)
		}
	}()

	// ? Initialize Cache
	cacheConfig := cache.Config{
		Addr:     cfg.Cache.Addr,
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
		TTL:      cfg.Cache.TTL,
	}

	cc := cache.Connect(cacheConfig)
	defer func() {
		if err := cc.Close(); err != nil {
			log.Printf("⚠️ Error closing cache connection: %v", err)
		}
	}()

	// ? Test Connections in background
	go func() {
		log.Println("🔍 Testing database connection...")
		if err := db.Health(context.Background()); err != nil {
			log.Printf("⚠️ Database health check failed: %v", err)
		} else {
			log.Println("✅ Database connection healthy")
		}

		log.Println("🔍 Testing cache connection...")
		if err := cc.Health(context.Background()); err != nil {
			log.Printf("⚠️ Cache health check failed: %v", err)
		} else {
			log.Println("✅ Cache connection healthy")
		}
	}()

	// * Initialize repositories
	linkRepo := repository.NewLinkRepository(db.Database)

	// * Initialize services
	linkService := service.NewLinkService(linkRepo, cc, cfg)

	// * Initialize handlers
	healthHandler := handlers.NewHealthHandler(db, cc)
	linkHandler := handlers.NewLinkHandler(linkService)

	// * Setup Gin
	r := gin.New()

	// * Add recovery middleware
	r.Use(gin.Recovery())

	// ? Add Logger in development mode
	if cfg.Server.GinMode == "debug" {
		r.Use(gin.Logger())
	}

	// ? Setup all routes following Helix structure
	routes.SetupRoutes(r, cfg, healthHandler, linkHandler)

	// ? Http server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// ? Start server in goroutine
	go func() {
		log.Printf("🌐 Server starting on %s", server.Addr)
		log.Printf("📚 API documentation available at: %s/docs", cfg.Server.BaseURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// * Wait for interrupt signal
	<-ctx.Done()

	log.Println("🛑 Shutting down Tether server...")

	// * Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Tether server shutdown completed")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("❌ %v", err)
	}
}
