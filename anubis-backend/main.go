// @title Anubis AI Core-Backend API
// @version 1.0.0
// @description A comprehensive core backend API for the Anubis AI platform with ThreeFold Grid integration
// @description This API provides endpoints for user management, task execution, AI memory management, and platform services.

// @contact.name Anubis AI Support
// @contact.email support@anubis.ai

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Authorization: Bearer {token}"

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"anubis-backend/config"
	"anubis-backend/database"
	"anubis-backend/routes"
	"anubis-backend/services"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration with environment-specific settings
	cfg := config.LoadConfig()

	// Create Fiber app with production-ready configuration
	app := fiber.New(fiber.Config{
		AppName:      "Anubis AI Core-Backend API v1.0.0",
		ServerHeader: "Anubis-API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Custom error handler for consistent error responses
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error":     "Internal Server Error",
				"message":   err.Error(),
				"timestamp": fiber.Map{},
				"path":      c.Path(),
			})
		},
		// Performance optimizations for production
		Prefork:       cfg.Env == "production",
		CaseSensitive: true,
		StrictRouting: true,
		// Security settings
		DisableStartupMessage: cfg.Env == "production",
	})

	// Initialize database with comprehensive error handling
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run database migrations with validation
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Seed database in development environment
	if cfg.Env == "development" {
		if err := database.SeedDatabase(); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	// Initialize services with dependency injection
	if err := services.InitTaskService(cfg); err != nil {
		log.Fatalf("Failed to initialize task service: %v", err)
	}

	// Setup routes with middleware stack
	routes.SetupRoutes(app, cfg)

	// Setup graceful shutdown handling
	setupGracefulShutdown(app)

	// Start server with comprehensive logging
	log.Printf("🚀 Starting Anubis AI Core-Backend API server on port %s", cfg.Port)
	log.Printf("📊 Environment: %s", cfg.Env)
	log.Printf("📚 Swagger documentation available at: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Printf("🏠 API home page: http://localhost:%s/home", cfg.Port)
	log.Printf("❤️  Health check: http://localhost:%s/health-check", cfg.Port)

	// Start the Fiber server
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// setupGracefulShutdown configures graceful shutdown handling for the Fiber application.
// This ensures proper cleanup of resources and connections when the server is terminated.
func setupGracefulShutdown(app *fiber.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Received shutdown signal, gracefully shutting down...")

		// Shutdown the Fiber server gracefully
		if err := app.Shutdown(); err != nil {
			log.Printf("❌ Error during server shutdown: %v", err)
		}

		// Close database connections
		if err := database.CloseDatabase(); err != nil {
			log.Printf("❌ Error closing database: %v", err)
		}

		log.Println("✅ Server shutdown complete")
		os.Exit(0)
	}()
}
