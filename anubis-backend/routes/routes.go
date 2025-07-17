// Package routes provides HTTP route configuration for the Anubis API.
// This package sets up all API endpoints with appropriate middleware,
// authentication, and documentation integration for optimal performance.
package routes

import (
	"anubis-backend/config"
	"anubis-backend/handlers"
	"anubis-backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	_ "anubis-backend/docs" // Import generated Swagger docs
)

// SetupRoutes configures all API routes with comprehensive middleware stack.
// This function sets up the complete routing configuration including public endpoints,
// authentication routes, and protected user endpoints with proper middleware ordering.
func SetupRoutes(app *fiber.App, cfg *config.Config) {
	// Global middleware stack (order matters for performance and security)
	app.Use(middleware.RequestID())        // Add unique request IDs for tracing
	app.Use(middleware.Logger())           // Request/response logging
	app.Use(middleware.Recovery())         // Panic recovery with graceful error handling
	app.Use(middleware.SecurityHeaders())  // Security headers for protection
	app.Use(middleware.CORS(cfg))          // Cross-origin resource sharing
	app.Use(middleware.RateLimit(cfg))     // Rate limiting for API protection

	// Public routes - no authentication required
	setupPublicRoutes(app)

	// Authentication routes - for user login/registration
	setupAuthRoutes(app)

	// Task execution routes - public for now, can be protected later
	setupTaskRoutes(app)

	// Protected routes - require valid JWT authentication
	setupProtectedRoutes(app)
}

// setupPublicRoutes configures public endpoints that don't require authentication.
func setupPublicRoutes(app *fiber.App) {
	// System information and health monitoring
	app.Get("/health-check", handlers.HealthCheck)
	app.Get("/home", handlers.Home)

	// API documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
}

// setupAuthRoutes configures authentication and user management endpoints.
func setupAuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// User authentication endpoints
	auth.Post("/signin", handlers.SignIn)
	auth.Post("/signup", handlers.SignUp)
	auth.Post("/refresh", handlers.RefreshToken)

	// Password management
	app.Post("/reset-password", handlers.ResetPassword)
}

// setupTaskRoutes configures ThreeFold Grid task execution endpoints.
func setupTaskRoutes(app *fiber.App) {
	// Task discovery and execution
	app.Get("/available-tasks", handlers.AvailableTasks)
	app.Post("/execute-task", handlers.ExecuteTask)
}

// setupProtectedRoutes configures endpoints that require JWT authentication.
func setupProtectedRoutes(app *fiber.App) {
	// Create protected route group with authentication middleware
	protected := app.Group("/", middleware.AuthRequired())

	// User profile management
	protected.Get("/user", handlers.GetUserProfile)
	protected.Put("/user", handlers.UpdateUserProfile)

	// AI memory management
	protected.Get("/user/memories", handlers.GetUserMemories)
	protected.Post("/user/memories", handlers.CreateUserMemory)

	// User settings management
	protected.Get("/user/settings", handlers.GetUserSettings)
	protected.Put("/user/settings", handlers.UpdateUserSetting)
}
