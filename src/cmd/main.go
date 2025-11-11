package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	// ? Test Connections in background
	go func() {
		log.Println("🔍 Testing database connection...")
		if err := db.Health(context.Background()); err != nil {
			log.Printf("⚠️ Database health check failed: %v", err)
		} else {
			log.Println("✅ Database connection healthy")
		}
	}()

	// * Setup Gin
	r := gin.New()

	// ? Add Logger in development mode
	if cfg.Server.GinMode == "debug" {
		r.Use(gin.Logger())
	}

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

	log.Println("🛑 Shutting down LinkForge server...")

	// * Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ LinkForge server shutdown completed")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("❌ %v", err)
	}
}
