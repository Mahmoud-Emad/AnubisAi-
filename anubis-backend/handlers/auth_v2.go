// Package handlers provides HTTP request handlers for decentralized authentication.
// This file contains handlers for user authentication, registration, and token management
// with support for TFChain wallets and digital twins while maintaining data sovereignty.
package handlers

import (
	"log"
	"strings"

	"anubis-backend/config"
	"anubis-backend/services"

	"github.com/gofiber/fiber/v2"
)

// Global auth service instance
var authService *services.AuthService

// InitAuthService initializes the authentication service
func InitAuthService(cfg *config.Config) {
	authService = services.NewAuthService(cfg)
	log.Println("Authentication service initialized successfully")
}

// RegisterV2 godoc
// @Summary Register a new user with decentralized identity
// @Description Register a new user with support for existing TFChain wallets or auto-generated wallets
// @Description Flow 1: User provides mnemonic for existing wallet
// @Description Flow 2: System generates new wallet for Web2 users
// @Tags auth
// @Accept json
// @Produce json
// @Param request body services.RegisterRequest true "Registration request with optional mnemonic"
// @Success 201 {object} services.AuthResponse "Registration successful"
// @Failure 400 {object} ErrorResponse "Invalid request or validation errors"
// @Failure 409 {object} ErrorResponse "User or wallet already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func RegisterV2(c *fiber.Ctx) error {
	var req services.RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			"Failed to parse JSON request body: "+err.Error())
	}

	// Log registration attempt (without sensitive data)
	flowType := "auto-generated wallet"
	if req.Mnemonic != "" {
		flowType = "existing wallet"
	}
	log.Printf("Registration attempt for %s using %s", req.Email, flowType)

	// Call auth service
	response, err := authService.Register(&req)
	if err != nil {
		log.Printf("Registration failed for %s: %v", req.Email, err)

		// Determine appropriate status code based on error
		statusCode := fiber.StatusInternalServerError
		if strings.Contains(err.Error(), "validation failed") ||
			strings.Contains(err.Error(), "invalid") {
			statusCode = fiber.StatusBadRequest
		} else if strings.Contains(err.Error(), "already exists") {
			statusCode = fiber.StatusConflict
		}

		return NewErrorResponse(c, statusCode,
			"Registration failed",
			err.Error())
	}

	log.Printf("User registered successfully: %s with wallet: %s",
		response.User.Email, response.User.WalletAddress)

	return c.Status(fiber.StatusCreated).JSON(response)
}

// LoginV2 godoc
// @Summary Authenticate user with email and password
// @Description Authenticate user and return JWT token with wallet information
// @Description Supports users with both existing and auto-generated wallets
// @Tags auth
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "Login credentials"
// @Success 200 {object} services.AuthResponse "Authentication successful"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 401 {object} ErrorResponse "Invalid credentials or inactive account"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func LoginV2(c *fiber.Ctx) error {
	var req services.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			"Failed to parse JSON request body: "+err.Error())
	}

	// Log login attempt
	log.Printf("Login attempt for: %s", req.Email)

	// Call auth service
	response, err := authService.Login(&req)
	if err != nil {
		log.Printf("Login failed for %s: %v", req.Email, err)

		// Determine appropriate status code
		statusCode := fiber.StatusInternalServerError
		if strings.Contains(err.Error(), "validation failed") {
			statusCode = fiber.StatusBadRequest
		} else if strings.Contains(err.Error(), "invalid email or password") ||
			strings.Contains(err.Error(), "deactivated") {
			statusCode = fiber.StatusUnauthorized
		}

		return NewErrorResponse(c, statusCode,
			"Authentication failed",
			err.Error())
	}

	log.Printf("User logged in successfully: %s", response.User.Email)

	return c.JSON(response)
}

// ValidateTokenV2 godoc
// @Summary Validate JWT token
// @Description Validate JWT token and return user information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} services.UserProfile "Token is valid"
// @Failure 401 {object} ErrorResponse "Invalid or expired token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/validate [get]
func ValidateTokenV2(c *fiber.Ctx) error {
	// Extract token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authorization required",
			"Missing Authorization header")
	}

	// Check Bearer format
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Invalid authorization format",
			"Authorization header must start with 'Bearer '")
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Missing token",
			"No token provided in Authorization header")
	}

	// Validate token
	userProfile, err := authService.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Invalid token",
			err.Error())
	}

	return c.JSON(fiber.Map{
		"valid": true,
		"user":  userProfile,
	})
}

// GetWalletInfo godoc
// @Summary Get wallet information for authenticated user
// @Description Retrieve wallet and digital twin information for the authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} services.WalletInfo "Wallet information retrieved"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/wallet [get]
func GetWalletInfo(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Unauthorized",
			"User not found in context")
	}

	user, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Internal error",
			"Invalid user context")
	}

	walletInfo := &services.WalletInfo{
		Address: user.WalletAddress,
		Network: user.Network,
		HasTwin: user.TwinID != nil,
		TwinID:  user.TwinID,
	}

	return c.JSON(walletInfo)
}

// RefreshTokenV2 godoc
// @Summary Refresh JWT token
// @Description Generate a new JWT token for authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} services.AuthResponse "Token refreshed successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func RefreshTokenV2(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Unauthorized",
			"User not found in context")
	}

	user, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Internal error",
			"Invalid user context")
	}

	// For now, return a success message
	// In a full implementation, you would generate a new token
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Token refresh functionality will be implemented",
		"user":    user,
	})
}

// LogoutV2 godoc
// @Summary Logout user
// @Description Logout user (client-side token removal)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} SuccessResponse "Logout successful"
// @Router /auth/logout [post]
func LogoutV2(c *fiber.Ctx) error {
	// In a stateless JWT system, logout is typically handled client-side
	// by removing the token. For enhanced security, you could implement
	// a token blacklist or use shorter-lived tokens with refresh tokens.

	return NewSuccessResponse(c, nil, "Logout successful. Please remove the token from client storage.")
}

// GetNetworkInfo godoc
// @Summary Get ThreeFold network information
// @Description Retrieve current ThreeFold network configuration and status
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "Network information"
// @Router /auth/network [get]
func GetNetworkInfo(c *fiber.Ctx) error {
	// This would typically come from the TFGrid adapter
	// For now, return mock network information
	networkInfo := fiber.Map{
		"network":      "test",
		"chain_id":     "tfchain-testnet",
		"rpc_endpoint": "wss://tfchain.test.grid.tf",
		"ws_endpoint":  "wss://tfchain.test.grid.tf",
		"status":       "active",
	}

	return c.JSON(networkInfo)
}
