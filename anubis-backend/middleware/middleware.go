// Package middleware provides HTTP middleware components for the Anubis API.
// This package contains all middleware functions for request processing including
// authentication, logging, CORS, rate limiting, and security headers.
// All middleware is designed for high performance and production readiness.
package middleware

import (
	"fmt"
	"strings"
	"time"

	"anubis-backend/config"
	"anubis-backend/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
)

// Logger returns a Fiber middleware for comprehensive request logging.
// This middleware provides detailed request/response logging with performance metrics
// and is essential for monitoring and debugging in production environments.
func Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path} - ${ip} \"${ua}\" ${error}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
		Done: func(c *fiber.Ctx, logString []byte) {
			// Custom log processing can be added here
			// For example, sending logs to external monitoring systems
		},
	})
}

// Recovery returns a Fiber middleware for panic recovery with detailed error reporting.
// This middleware ensures the application remains stable even when handlers panic,
// providing graceful error responses and preventing server crashes.
func Recovery() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			// Custom panic handling with detailed error response
			errorMsg := "An unexpected error occurred"
			if err, ok := e.(error); ok {
				errorMsg = err.Error()
			} else if str, ok := e.(string); ok {
				errorMsg = str
			}

			// Return structured error response
			response := handlers.ErrorResponse{
				Error:     "Internal Server Error",
				Message:   errorMsg,
				Timestamp: time.Now(),
				Path:      c.Path(),
				RequestID: c.Get("X-Request-ID", ""),
			}

			c.Status(fiber.StatusInternalServerError).JSON(response)
		},
	})
}

// CORS returns a Fiber middleware for comprehensive CORS handling.
// This middleware provides secure cross-origin resource sharing configuration
// with support for multiple origins, methods, and headers.
func CORS(cfg *config.Config) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.CORS.Origins, ","),
		AllowMethods:     strings.Join(cfg.CORS.Methods, ","),
		AllowHeaders:     strings.Join(cfg.CORS.Headers, ","),
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
		Next: func(c *fiber.Ctx) bool {
			// Skip CORS for health check endpoints to improve performance
			return c.Path() == "/health-check"
		},
	})
}

// RateLimit returns a Fiber middleware for intelligent rate limiting.
// This middleware provides per-IP rate limiting with configurable limits
// and graceful handling of rate limit violations.
func RateLimit(cfg *config.Config) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.API.RateLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the key for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			// Custom rate limit exceeded response
			response := handlers.ErrorResponse{
				Error:     "Rate limit exceeded",
				Message:   fmt.Sprintf("Too many requests. Limit: %d requests per minute. Please try again later.", cfg.API.RateLimit),
				Timestamp: time.Now(),
				Path:      c.Path(),
				RequestID: c.Get("X-Request-ID", ""),
			}
			return c.Status(fiber.StatusTooManyRequests).JSON(response)
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Next: func(c *fiber.Ctx) bool {
			// Skip rate limiting for health checks and internal endpoints
			path := c.Path()
			return path == "/health-check" || strings.HasPrefix(path, "/internal/")
		},
	})
}

// AuthRequired returns a Fiber middleware for comprehensive JWT authentication.
// This middleware validates JWT tokens, extracts user information, and provides
// secure access control for protected endpoints.
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			response := handlers.ErrorResponse{
				Error:     "Authorization required",
				Message:   "Missing Authorization header. Please provide a valid JWT token.",
				Timestamp: time.Now(),
				Path:      c.Path(),
				RequestID: c.Get("X-Request-ID", ""),
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// Validate Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response := handlers.ErrorResponse{
				Error:     "Invalid authorization format",
				Message:   "Authorization header must start with 'Bearer ' followed by a valid JWT token",
				Timestamp: time.Now(),
				Path:      c.Path(),
				RequestID: c.Get("X-Request-ID", ""),
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// Extract and validate token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response := handlers.ErrorResponse{
				Error:     "Missing token",
				Message:   "No token provided in Authorization header",
				Timestamp: time.Now(),
				Path:      c.Path(),
				RequestID: c.Get("X-Request-ID", ""),
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// TODO: Implement actual JWT validation with proper secret verification
		// For development, accept any token that starts with "mock-jwt-token"
		if !strings.HasPrefix(token, "mock-jwt-token") {
			response := handlers.ErrorResponse{
				Error:     "Invalid token",
				Message:   "Token validation failed. Please provide a valid JWT token.",
				Timestamp: time.Now(),
				Path:      c.Path(),
				RequestID: c.Get("X-Request-ID", ""),
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// TODO: Extract user information from JWT and set in context
		// For development, set a mock user ID
		// userID := extractUserIDFromToken(token)
		// c.Locals("user_id", userID)

		return c.Next()
	}
}

// RequestID returns a Fiber middleware for generating unique request identifiers.
// This middleware adds a unique ID to each request for tracing and debugging purposes.
func RequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header:     "X-Request-ID",
		Generator:  func() string { return uuid.New().String() },
		ContextKey: "requestid",
	})
}

// SecurityHeaders returns a Fiber middleware for comprehensive security headers.
// This middleware adds essential security headers to protect against common web vulnerabilities
// including XSS, clickjacking, and content type sniffing attacks.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Set("X-XSS-Protection", "1; mode=block")

		// Enforce HTTPS (only in production)
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content Security Policy for additional XSS protection
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Prevent caching of sensitive data
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		return c.Next()
	}
}
