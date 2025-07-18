// Package middleware provides authentication and security middleware for the Anubis AI Core-Backend API.
// This file contains JWT authentication middleware with support for decentralized identity
// and comprehensive security measures.
package middleware

import (
	"log"
	"strings"

	"anubis-backend/common"
	"anubis-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			response := common.NewErrorResponse(
				"Authorization required",
				"Missing Authorization header. Please provide a valid JWT token.",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// Check Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response := common.NewErrorResponse(
				"Invalid authorization format",
				"Authorization header must start with 'Bearer '. Example: 'Bearer your-jwt-token'",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// Extract and validate token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response := common.NewErrorResponse(
				"Missing token",
				"No token provided in Authorization header",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// Validate JWT token using auth service
		userProfile, err := authService.ValidateToken(token)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			response := common.NewErrorResponse(
				"Invalid token",
				"Token validation failed: "+err.Error(),
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		// Store user profile in context for use by subsequent handlers
		c.Locals("user", userProfile)
		c.Locals("user_id", userProfile.ID)
		c.Locals("wallet_address", userProfile.WalletAddress)

		log.Printf("User authenticated successfully: %s (ID: %s, Wallet: %s)",
			userProfile.Email, userProfile.ID, userProfile.WalletAddress)

		return c.Next()
	}
}

// AdminMiddleware ensures only admin users can access protected routes
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from context (set by AuthMiddleware)
		userProfile := c.Locals("user")
		if userProfile == nil {
			response := common.NewErrorResponse(
				"Unauthorized",
				"Authentication required",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		user, ok := userProfile.(*services.UserProfile)
		if !ok {
			response := common.NewErrorResponse(
				"Internal error",
				"Invalid user context",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(response)
		}

		// Check if user is admin (this would need to be added to UserProfile)
		// For now, we'll check if the user has admin privileges
		// This is a placeholder - you'd need to add IsAdmin field to UserProfile
		if user.Email != "admin@anubis.local" {
			response := common.NewErrorResponse(
				"Forbidden",
				"Admin access required",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusForbidden).JSON(response)
		}

		log.Printf("Admin access granted for user: %s", user.Email)
		return c.Next()
	}
}

// WalletOwnerMiddleware ensures users can only access their own wallet information
func WalletOwnerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from context (set by AuthMiddleware)
		userProfile := c.Locals("user")
		if userProfile == nil {
			response := common.NewErrorResponse(
				"Unauthorized",
				"Authentication required",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		user, ok := userProfile.(*services.UserProfile)
		if !ok {
			response := common.NewErrorResponse(
				"Internal error",
				"Invalid user context",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(response)
		}

		// Extract wallet address from URL parameters or query
		walletAddress := c.Params("wallet_address")
		if walletAddress == "" {
			walletAddress = c.Query("wallet_address")
		}

		// If no specific wallet address is requested, allow access to own wallet
		if walletAddress == "" {
			return c.Next()
		}

		// Check if user owns the specified wallet
		if user.WalletAddress != walletAddress {
			response := common.NewErrorResponse(
				"Forbidden",
				"You can only access your own wallet information",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusForbidden).JSON(response)
		}

		log.Printf("Wallet access granted for user: %s, wallet: %s", user.Email, walletAddress)
		return c.Next()
	}
}

// NetworkValidationMiddleware validates ThreeFold network parameters
func NetworkValidationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract network from query parameters
		network := c.Query("network")
		if network == "" {
			// Default to test network if not specified
			network = "test"
			c.Query("network", network)
		}

		// Validate network
		validNetworks := map[string]bool{
			"main": true,
			"test": true,
			"qa":   true,
			"dev":  true,
		}

		if !validNetworks[network] {
			response := common.NewErrorResponse(
				"Invalid network",
				"Network must be one of: main, test, qa, dev",
				c.Path(),
				c.Get("X-Request-ID", ""),
			)
			return c.Status(fiber.StatusBadRequest).JSON(response)
		}

		// Store validated network in context
		c.Locals("network", network)
		return c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate or extract request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (in production, use UUID)
			requestID = generateRequestID()
		}

		// Set request ID in context and response header
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple implementation using timestamp
	return "req_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
}
