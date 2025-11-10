package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Kosha-Nirman/tether/src/pkg/config"
	"github.com/gin-gonic/gin"
)

func run() {
	log.Println("🔗 Starting Tether...")

	// ? Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// * Create context for graceful shutdown
	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("🚀 Starting Tether server on port %d", cfg.Server.Port)
}

func main() {
	run()
}
