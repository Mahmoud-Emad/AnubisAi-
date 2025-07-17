package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SignInRequest represents the sign-in request with comprehensive validation.
type SignInRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
}

// SignInResponse represents the sign-in response.
type SignInResponse struct {
	Success      bool      `json:"success"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	User         UserInfo  `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserInfo represents basic user information.
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsAdmin   bool      `json:"is_admin"`
}

// SignIn godoc
// @Summary User sign in
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Sign in credentials"
// @Success 200 {object} SignInResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/signin [post]
func SignIn(c *fiber.Ctx) error {
	var req SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			"Failed to parse JSON request body: "+err.Error())
	}

	if req.Email == "admin@anubis.local" && req.Password == "admin123" {
		response := SignInResponse{
			Success:      true,
			Token:        "mock-jwt-token-" + uuid.New().String(),
			RefreshToken: "mock-refresh-token-" + uuid.New().String(),
			User: UserInfo{
				ID:        uuid.New(),
				Email:     req.Email,
				Username:  "admin",
				FirstName: "Admin",
				LastName:  "User",
				IsAdmin:   true,
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	return NewErrorResponse(c, fiber.StatusUnauthorized,
		"Authentication failed",
		"Invalid email or password")
}

// SignUpRequest represents the sign-up request.
type SignUpRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// SignUpResponse represents the sign-up response.
type SignUpResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	User    UserInfo `json:"user"`
}

// SignUp godoc
// @Summary User sign up
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Sign up information"
// @Success 201 {object} SignUpResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/signup [post]
func SignUp(c *fiber.Ctx) error {
	var req SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			"Failed to parse JSON request body: "+err.Error())
	}

	if req.Email == "admin@anubis.local" {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Error:     "User already exists",
			Message:   "A user with this email already exists",
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	newUser := UserInfo{
		ID:        uuid.New(),
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsAdmin:   false,
	}

	return c.Status(fiber.StatusCreated).JSON(SignUpResponse{
		Success: true,
		Message: "User registered successfully",
		User:    newUser,
	})
}

// ResetPasswordRequest represents the password reset request.
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordResponse represents the password reset response.
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetPassword godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Password reset request"
// @Success 200 {object} ResetPasswordResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /reset-password [post]
func ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:     "Invalid request format",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	if req.Email != "admin@anubis.local" {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:     "User not found",
			Message:   "No user found with this email address",
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(ResetPasswordResponse{
		Success: true,
		Message: "Password reset email sent successfully",
	})
}

// RefreshTokenRequest represents the refresh token request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the refresh token response.
type RefreshTokenResponse struct {
	Success   bool      `json:"success"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Get a new JWT token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} RefreshTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:     "Invalid request format",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(RefreshTokenResponse{
		Success:   true,
		Token:     "new-mock-jwt-token-" + uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
}
