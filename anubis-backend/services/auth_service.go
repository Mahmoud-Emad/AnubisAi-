// Package services provides business logic for authentication and user management.
// This file contains the authentication service with support for decentralized
// identity using TFChain wallets and digital twins.
package services

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"anubis-backend/adapters"
	"anubis-backend/config"
	"anubis-backend/database"
	"anubis-backend/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService provides authentication and user management functionality
type AuthService struct {
	db            *gorm.DB
	tfgridAdapter adapters.TFGridAdapter
	jwtSecret     string
	jwtExpiry     time.Duration
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=32,alpha"`
	LastName  string `json:"last_name" validate:"required,min=2,max=32,alpha"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	Username  string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Mnemonic  string `json:"mnemonic,omitempty"` // Optional - for existing wallet users
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success      bool         `json:"success"`
	Token        string       `json:"token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         *UserProfile `json:"user,omitempty"`
	WalletInfo   *WalletInfo  `json:"wallet_info,omitempty"`
	ExpiresAt    time.Time    `json:"expires_at,omitempty"`
	Message      string       `json:"message,omitempty"`
}

// UserProfile represents user profile information
type UserProfile struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	WalletAddress string    `json:"wallet_address"`
	TwinID        *int64    `json:"twin_id,omitempty"`
	Network       string    `json:"network"`
	HasWallet     bool      `json:"has_wallet"`
	EmailVerified bool      `json:"email_verified"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// WalletInfo represents wallet information for response
type WalletInfo struct {
	Address string `json:"address"`
	Network string `json:"network"`
	HasTwin bool   `json:"has_twin"`
	TwinID  *int64 `json:"twin_id,omitempty"`
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.Config) *AuthService {
	db := database.GetDB()

	// Use TFGrid adapter for production-ready functionality
	tfgridAdapter := adapters.NewTFGridAdapter(cfg.TFGrid.Network)

	return &AuthService{
		db:            db,
		tfgridAdapter: tfgridAdapter,
		jwtSecret:     cfg.JWT.Secret,
		jwtExpiry:     cfg.JWT.Expiry,
	}
}

// Register handles user registration with dual flow support
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateRegistrationInput(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	if err := s.checkUserExists(req.Email, req.Username); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Determine flow based on mnemonic presence
	var walletInfo *adapters.WalletInfo
	var hasWallet bool

	if req.Mnemonic != "" {
		// Flow 1: User has existing wallet
		walletInfo, err = s.tfgridAdapter.DeriveWalletFromMnemonic(req.Mnemonic)
		if err != nil {
			return nil, fmt.Errorf("failed to derive wallet from mnemonic: %w", err)
		}
		hasWallet = true
		log.Printf("User provided existing wallet: %s", walletInfo.Address)
	} else {
		// Flow 2: Generate new wallet for user
		walletInfo, err = s.tfgridAdapter.GenerateWallet()
		if err != nil {
			return nil, fmt.Errorf("failed to generate wallet: %w", err)
		}
		hasWallet = false
		log.Printf("Generated new wallet for user: %s", walletInfo.Address)
	}

	// Check if wallet address is already used
	if err := s.checkWalletExists(walletInfo.Address); err != nil {
		return nil, err
	}

	// Create digital twin
	twinMetadata := adapters.DigitalTwinMetadata{
		Name:        fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		Email:       req.Email,
		Description: "Anubis AI Platform User",
		Attributes: map[string]string{
			"platform": "anubis-ai",
			"version":  "1.0.0",
		},
	}

	digitalTwin, err := s.tfgridAdapter.CreateDigitalTwin(walletInfo.Address, twinMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create digital twin: %w", err)
	}

	// Create user in database
	user := &models.User{
		Email:         req.Email,
		Username:      req.Username,
		Password:      hashedPassword,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		WalletAddress: walletInfo.Address,
		TwinID:        &digitalTwin.ID,
		PublicKey:     walletInfo.PublicKey,
		Network:       walletInfo.Network,
		HasWallet:     hasWallet,
		IsActive:      true,
		EmailVerified: false, // TODO: Implement email verification
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Successfully registered user: %s with wallet: %s and twin: %d",
		user.Email, user.WalletAddress, *user.TwinID)

	// Generate JWT token
	token, expiresAt, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Success: true,
		Token:   token,
		User:    s.userToProfile(user),
		WalletInfo: &WalletInfo{
			Address: walletInfo.Address,
			Network: walletInfo.Network,
			HasTwin: true,
			TwinID:  &digitalTwin.ID,
		},
		ExpiresAt: expiresAt,
		Message:   "Registration successful. Welcome to Anubis AI!",
	}, nil
}

// Login handles user authentication
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validateLoginInput(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, expiresAt, err := s.generateJWT(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("User logged in successfully: %s", user.Email)

	return &AuthResponse{
		Success: true,
		Token:   token,
		User:    s.userToProfile(&user),
		WalletInfo: &WalletInfo{
			Address: user.WalletAddress,
			Network: user.Network,
			HasTwin: user.TwinID != nil,
			TwinID:  user.TwinID,
		},
		ExpiresAt: expiresAt,
		Message:   "Login successful",
	}, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *AuthService) ValidateToken(tokenString string) (*UserProfile, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Get user from database
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	return s.userToProfile(&user), nil
}

// Helper functions

func (s *AuthService) validateRegistrationInput(req *RegisterRequest) error {
	// Validate names (alphabetic only)
	nameRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !nameRegex.MatchString(req.FirstName) {
		return fmt.Errorf("first name must contain only alphabetic characters")
	}
	if !nameRegex.MatchString(req.LastName) {
		return fmt.Errorf("last name must contain only alphabetic characters")
	}

	// Validate name lengths
	if len(req.FirstName) < 2 || len(req.FirstName) > 32 {
		return fmt.Errorf("first name must be between 2 and 32 characters")
	}
	if len(req.LastName) < 2 || len(req.LastName) > 32 {
		return fmt.Errorf("last name must be between 2 and 32 characters")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password strength
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Validate username
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("username must contain only alphanumeric characters")
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Validate mnemonic if provided
	if req.Mnemonic != "" {
		if err := s.tfgridAdapter.ValidateMnemonic(req.Mnemonic); err != nil {
			return fmt.Errorf("invalid mnemonic: %w", err)
		}
	}

	return nil
}

func (s *AuthService) validateLoginInput(req *LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (s *AuthService) checkUserExists(email, username string) error {
	var count int64

	// Check email
	if err := s.db.Model(&models.User{}).Where("email = ? AND deleted_at IS NULL", email).Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("email already exists")
	}

	// Check username
	if err := s.db.Model(&models.User{}).Where("username = ? AND deleted_at IS NULL", username).Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("username already exists")
	}

	return nil
}

func (s *AuthService) checkWalletExists(walletAddress string) error {
	var count int64
	if err := s.db.Model(&models.User{}).Where("wallet_address = ? AND deleted_at IS NULL", walletAddress).Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("wallet address already registered")
	}
	return nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *AuthService) generateJWT(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)

	claims := jwt.MapClaims{
		"user_id":        user.ID.String(),
		"email":          user.Email,
		"wallet_address": user.WalletAddress,
		"exp":            expiresAt.Unix(),
		"iat":            time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *AuthService) userToProfile(user *models.User) *UserProfile {
	return &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		WalletAddress: user.WalletAddress,
		TwinID:        user.TwinID,
		Network:       user.Network,
		HasWallet:     user.HasWallet,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
	}
}
