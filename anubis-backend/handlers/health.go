// Package handlers provides HTTP request handlers for health monitoring and system status.
// This file contains handlers for health checks and system information endpoints.
package handlers

import (
	"time"

	"anubis-backend/database"

	"github.com/gofiber/fiber/v2"
)

// HealthResponse represents the health check response with detailed service status.
// This provides comprehensive health information for monitoring and alerting systems.
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`                     // Overall system status
	Timestamp time.Time         `json:"timestamp" example:"2024-01-01T12:00:00Z"`     // Health check timestamp
	Version   string            `json:"version" example:"1.0.0"`                      // API version
	Services  map[string]string `json:"services"`                                     // Individual service statuses
	Uptime    string            `json:"uptime,omitempty" example:"2h30m15s"`          // System uptime
	RequestID string            `json:"request_id,omitempty" example:"req_123456789"` // Request identifier
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API and its dependencies including database connectivity and service status
// @Description This endpoint is used by load balancers and monitoring systems to determine if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} ErrorResponse "Service is unhealthy"
// @Router /health-check [get]
func HealthCheck(c *fiber.Ctx) error {
	services := make(map[string]string)

	// Check database health with detailed error reporting
	if err := database.HealthCheck(); err != nil {
		services["database"] = "unhealthy: " + err.Error()

		response := HealthResponse{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Version:   "1.0.0",
			Services:  services,
			RequestID: c.Get("X-Request-ID", ""),
		}

		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}
	services["database"] = "healthy"

	// Add more service checks here as needed
	services["api"] = "healthy"

	// TODO: Add checks for external dependencies like ThreeFold Grid API
	// services["tfgrid"] = "healthy"

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  services,
		RequestID: c.Get("X-Request-ID", ""),
	}

	return c.JSON(response)
}

// HomeResponse represents the home endpoint response with comprehensive API information.
// This provides an overview of the API capabilities and available endpoints.
type HomeResponse struct {
	Message     string    `json:"message" example:"Welcome to Anubis AI Core-Backend API"` // Welcome message
	API         string    `json:"api" example:"Anubis AI Core-Backend"`                    // API name
	Version     string    `json:"version" example:"1.0.0"`                                 // API version
	Timestamp   time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`                // Response timestamp
	Endpoints   []string  `json:"endpoints"`                                               // Available API endpoints
	Environment string    `json:"environment" example:"development"`                       // Current environment
	Framework   string    `json:"framework" example:"Fiber v2.52.0"`                       // Web framework info
	RequestID   string    `json:"request_id,omitempty" example:"req_123456789"`            // Request identifier
}

// Home godoc
// @Summary API home endpoint
// @Description Returns comprehensive information about the Anubis AI Core-Backend API including available endpoints and system info
// @Description This endpoint provides an overview of the API capabilities and serves as a discovery endpoint for clients
// @Tags general
// @Produce json
// @Success 200 {object} HomeResponse "API information retrieved successfully"
// @Router /home [get]
func Home(c *fiber.Ctx) error {
	// Define all available API endpoints with descriptions
	endpoints := []string{
		"GET /health-check - System health status",
		"GET /home - API information and endpoints",
		"GET /available-tasks - List supported ThreeFold Grid tasks",
		"POST /execute-task - Execute ThreeFold Grid tasks",
		"POST /auth/signin - User authentication",
		"POST /auth/signup - User registration",
		"POST /auth/refresh - Refresh JWT token",
		"POST /reset-password - Password reset request",
		"GET /user - User profile information",
		"PUT /user - Update user profile",
		"GET /user/memories - User AI memories",
		"POST /user/memories - Create AI memory",
		"GET /user/settings - User settings",
		"PUT /user/settings - Update user settings",
		"GET /swagger/index.html - API documentation",
	}

	response := HomeResponse{
		Message:     "Welcome to Anubis AI Core-Backend API",
		API:         "Anubis AI Core-Backend",
		Version:     "1.0.0",
		Timestamp:   time.Now(),
		Endpoints:   endpoints,
		Environment: "development", // TODO: Get from config
		Framework:   "Fiber v2.52.0",
		RequestID:   c.Get("X-Request-ID", ""),
	}

	return c.JSON(response)
}
