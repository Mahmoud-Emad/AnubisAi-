package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"anubis-backend/database"
	"anubis-backend/models"
	"anubis-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserInfo represents basic user information for API responses
type UserInfo struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	IsAdmin       bool      `json:"is_admin"`
	WalletAddress string    `json:"wallet_address,omitempty"`
	TwinID        *int64    `json:"twin_id,omitempty"`
	Network       string    `json:"network,omitempty"`
	HasWallet     bool      `json:"has_wallet"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserStats represents user activity statistics
type UserStats struct {
	TasksExecuted  int       `json:"tasks_executed"`
	MemoriesStored int       `json:"memories_stored"`
	LastActivity   time.Time `json:"last_activity"`
	AccountCreated time.Time `json:"account_created"`
}

// UserProfileResponse represents the complete user profile response
type UserProfileResponse struct {
	User      UserInfo  `json:"user"`
	Stats     UserStats `json:"stats"`
	Timestamp time.Time `json:"timestamp"`
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get the authenticated user's profile information including stats
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/profile [get]
func GetUserProfile(c *fiber.Ctx) error {
	// Get authenticated user from context (set by AuthMiddleware)
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authentication required",
			"User context not found. Please ensure you are authenticated.")
	}

	authUser, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Invalid user context",
			"Failed to parse user authentication data.")
	}

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Database connection failed",
			"Unable to connect to database.")
	}

	// Fetch user from database
	var user models.User
	if err := db.Where("id = ? AND deleted_at IS NULL", authUser.ID).First(&user).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusNotFound,
			"User not found",
			"The authenticated user could not be found in the database.")
	}

	// Calculate user statistics
	stats, err := calculateUserStats(db, user.ID)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to calculate user statistics",
			fmt.Sprintf("Error calculating stats: %v", err))
	}

	// Convert database user to API response format
	userInfo := UserInfo{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsAdmin:       user.IsAdmin,
		WalletAddress: user.WalletAddress,
		TwinID:        user.TwinID,
		Network:       user.Network,
		HasWallet:     user.HasWallet,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(UserProfileResponse{
		User:      userInfo,
		Stats:     *stats,
		Timestamp: time.Now(),
	})
}

// calculateUserStats computes user activity statistics from the database
func calculateUserStats(db *gorm.DB, userID uuid.UUID) (*UserStats, error) {
	var tasksExecuted int64
	var memoriesStored int64
	var lastActivity time.Time

	// Count completed tasks
	if err := db.Model(&models.TaskExecution{}).
		Where("user_id = ? AND status = 'success'", userID).
		Count(&tasksExecuted).Error; err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Count active memories
	if err := db.Model(&models.UserMemory{}).
		Where("user_id = ? AND is_active = true AND deleted_at IS NULL", userID).
		Count(&memoriesStored).Error; err != nil {
		return nil, fmt.Errorf("failed to count memories: %w", err)
	}

	// Get last activity (most recent task execution or memory creation)
	var lastTaskTime, lastMemoryTime time.Time

	// Get last task execution time
	var lastTask models.TaskExecution
	if err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&lastTask).Error; err == nil {
		lastTaskTime = lastTask.CreatedAt
	}

	// Get last memory creation time
	var lastMemory models.UserMemory
	if err := db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		First(&lastMemory).Error; err == nil {
		lastMemoryTime = lastMemory.CreatedAt
	}

	// Use the most recent activity
	if lastTaskTime.After(lastMemoryTime) {
		lastActivity = lastTaskTime
	} else if !lastMemoryTime.IsZero() {
		lastActivity = lastMemoryTime
	} else {
		// If no activities, use account creation time
		var user models.User
		if err := db.Select("created_at").Where("id = ?", userID).First(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to get user creation time: %w", err)
		}
		lastActivity = user.CreatedAt
	}

	// Get account creation time
	var user models.User
	if err := db.Select("created_at").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get account creation time: %w", err)
	}

	return &UserStats{
		TasksExecuted:  int(tasksExecuted),
		MemoriesStored: int(memoriesStored),
		LastActivity:   lastActivity,
		AccountCreated: user.CreatedAt,
	}, nil
}

// Helper functions for JSON tag conversion
func tagsToString(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(tags)
	return string(data)
}

func tagsFromString(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}
	var tags []string
	json.Unmarshal([]byte(tagsStr), &tags)
	return tags
}

// UpdateUserRequest represents the request body for updating user profile
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Username  string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserRequest true "User update data"
// @Success 200 {object} UserProfileResponse "User profile updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 409 {object} ErrorResponse "Username already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/profile [put]
func UpdateUserProfile(c *fiber.Ctx) error {
	// Get authenticated user from context
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authentication required",
			"User context not found. Please ensure you are authenticated.")
	}

	authUser, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Invalid user context",
			"Failed to parse user authentication data.")
	}

	// Parse request body
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			fmt.Sprintf("Failed to parse request body: %v", err))
	}

	// Validate request (basic validation)
	if req.FirstName == "" && req.LastName == "" && req.Username == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"No update data provided",
			"At least one field (first_name, last_name, username) must be provided.")
	}

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Database connection failed",
			"Unable to connect to database.")
	}

	// Fetch current user from database
	var user models.User
	if err := db.Where("id = ? AND deleted_at IS NULL", authUser.ID).First(&user).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusNotFound,
			"User not found",
			"The authenticated user could not be found in the database.")
	}

	// Check if username is being changed and if it's already taken
	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		if err := db.Where("username = ? AND id != ? AND deleted_at IS NULL", req.Username, user.ID).First(&existingUser).Error; err == nil {
			return NewErrorResponse(c, fiber.StatusConflict,
				"Username already exists",
				"The requested username is already taken by another user.")
		}
	}

	// Update user fields
	updateData := make(map[string]interface{})
	if req.FirstName != "" {
		updateData["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updateData["last_name"] = req.LastName
	}
	if req.Username != "" {
		updateData["username"] = req.Username
	}
	updateData["updated_at"] = time.Now()

	// Perform update
	if err := db.Model(&user).Updates(updateData).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to update user profile",
			fmt.Sprintf("Database update failed: %v", err))
	}

	// Fetch updated user
	if err := db.Where("id = ?", user.ID).First(&user).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to fetch updated user",
			"User was updated but could not be retrieved.")
	}

	// Calculate updated statistics
	stats, err := calculateUserStats(db, user.ID)
	if err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to calculate user statistics",
			fmt.Sprintf("Error calculating stats: %v", err))
	}

	// Convert to API response format
	userInfo := UserInfo{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsAdmin:       user.IsAdmin,
		WalletAddress: user.WalletAddress,
		TwinID:        user.TwinID,
		Network:       user.Network,
		HasWallet:     user.HasWallet,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(UserProfileResponse{
		User:      userInfo,
		Stats:     *stats,
		Timestamp: time.Now(),
	})
}

// UserMemory represents a user memory entry
type UserMemory struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMemoryRequest represents the request body for creating a memory
type CreateMemoryRequest struct {
	Title   string   `json:"title" validate:"required,min=1,max=200"`
	Content string   `json:"content" validate:"required,min=1,max=10000"`
	Tags    []string `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=50"`
}

// GetUserMemories godoc
// @Summary Get user memories
// @Description Get all memories for the authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} UserMemory "User memories retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/memories [get]
func GetUserMemories(c *fiber.Ctx) error {
	// Get authenticated user from context
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authentication required",
			"User context not found. Please ensure you are authenticated.")
	}

	authUser, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Invalid user context",
			"Failed to parse user authentication data.")
	}

	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Database connection failed",
			"Unable to connect to database.")
	}

	// Fetch user memories from database
	var memories []models.UserMemory
	if err := db.Where("user_id = ? AND deleted_at IS NULL", authUser.ID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to fetch memories",
			fmt.Sprintf("Database query failed: %v", err))
	}

	// Convert to API response format
	var responseMemories []UserMemory
	for _, memory := range memories {
		responseMemories = append(responseMemories, UserMemory{
			ID:        memory.ID,
			UserID:    memory.UserID,
			Title:     memory.Title,
			Content:   memory.Content,
			Tags:      tagsFromString(memory.Tags),
			IsActive:  memory.IsActive,
			CreatedAt: memory.CreatedAt,
			UpdatedAt: memory.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(responseMemories)
}

// CreateUserMemory godoc
// @Summary Create user memory
// @Description Create a new memory for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateMemoryRequest true "Memory data"
// @Success 201 {object} UserMemory "Memory created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/memories [post]
func CreateUserMemory(c *fiber.Ctx) error {
	// Get authenticated user from context
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authentication required",
			"User context not found. Please ensure you are authenticated.")
	}

	authUser, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Invalid user context",
			"Failed to parse user authentication data.")
	}

	// Parse request body
	var req CreateMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			fmt.Sprintf("Failed to parse request body: %v", err))
	}

	// Basic validation
	if req.Title == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Title is required",
			"Memory title cannot be empty.")
	}
	if req.Content == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Content is required",
			"Memory content cannot be empty.")
	}

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Database connection failed",
			"Unable to connect to database.")
	}

	// Create new memory
	memory := models.UserMemory{
		ID:        uuid.New(),
		UserID:    authUser.ID,
		Title:     req.Title,
		Content:   req.Content,
		Tags:      tagsToString(req.Tags),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := db.Create(&memory).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to create memory",
			fmt.Sprintf("Database insert failed: %v", err))
	}

	// Return created memory
	responseMemory := UserMemory{
		ID:        memory.ID,
		UserID:    memory.UserID,
		Title:     memory.Title,
		Content:   memory.Content,
		Tags:      tagsFromString(memory.Tags),
		IsActive:  memory.IsActive,
		CreatedAt: memory.CreatedAt,
		UpdatedAt: memory.UpdatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(responseMemory)
}

// UserSetting represents a user setting entry
type UserSetting struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUserSettings godoc
// @Summary Get user settings
// @Description Get all settings for the authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserSetting "User settings retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/settings [get]
func GetUserSettings(c *fiber.Ctx) error {
	// Get authenticated user from context
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authentication required",
			"User context not found. Please ensure you are authenticated.")
	}

	authUser, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Invalid user context",
			"Failed to parse user authentication data.")
	}

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Database connection failed",
			"Unable to connect to database.")
	}

	// Fetch user settings from database
	var settings []models.UserSetting
	if err := db.Where("user_id = ?", authUser.ID).
		Order("key ASC").
		Find(&settings).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to fetch settings",
			fmt.Sprintf("Database query failed: %v", err))
	}

	// Convert to API response format
	var responseSettings []UserSetting
	for _, setting := range settings {
		responseSettings = append(responseSettings, UserSetting{
			ID:        setting.ID,
			UserID:    setting.UserID,
			Key:       setting.Key,
			Value:     setting.Value,
			CreatedAt: setting.CreatedAt,
			UpdatedAt: setting.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(responseSettings)
}

// UpdateSettingRequest represents the request body for updating a setting
type UpdateSettingRequest struct {
	Key   string `json:"key" validate:"required,min=1,max=100"`
	Value string `json:"value" validate:"required,min=1,max=1000"`
}

// UpdateUserSetting godoc
// @Summary Update user setting
// @Description Update or create a setting for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateSettingRequest true "Setting data"
// @Success 200 {object} UserSetting "Setting updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/settings [put]
func UpdateUserSetting(c *fiber.Ctx) error {
	// Get authenticated user from context
	userProfile := c.Locals("user")
	if userProfile == nil {
		return NewErrorResponse(c, fiber.StatusUnauthorized,
			"Authentication required",
			"User context not found. Please ensure you are authenticated.")
	}

	authUser, ok := userProfile.(*services.UserProfile)
	if !ok {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Invalid user context",
			"Failed to parse user authentication data.")
	}

	// Parse request body
	var req UpdateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			fmt.Sprintf("Failed to parse request body: %v", err))
	}

	// Basic validation
	if req.Key == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Key is required",
			"Setting key cannot be empty.")
	}
	if req.Value == "" {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Value is required",
			"Setting value cannot be empty.")
	}

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Database connection failed",
			"Unable to connect to database.")
	}

	// Check if setting already exists
	var existingSetting models.UserSetting
	err := db.Where("user_id = ? AND key = ?", authUser.ID, req.Key).First(&existingSetting).Error

	if err == nil {
		// Update existing setting
		existingSetting.Value = req.Value
		existingSetting.UpdatedAt = time.Now()

		if err := db.Save(&existingSetting).Error; err != nil {
			return NewErrorResponse(c, fiber.StatusInternalServerError,
				"Failed to update setting",
				fmt.Sprintf("Database update failed: %v", err))
		}

		// Return updated setting
		responseSetting := UserSetting{
			ID:        existingSetting.ID,
			UserID:    existingSetting.UserID,
			Key:       existingSetting.Key,
			Value:     existingSetting.Value,
			CreatedAt: existingSetting.CreatedAt,
			UpdatedAt: existingSetting.UpdatedAt,
		}

		return c.Status(fiber.StatusOK).JSON(responseSetting)
	} else {
		// Create new setting
		newSetting := models.UserSetting{
			ID:        uuid.New(),
			UserID:    authUser.ID,
			Key:       req.Key,
			Value:     req.Value,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&newSetting).Error; err != nil {
			return NewErrorResponse(c, fiber.StatusInternalServerError,
				"Failed to create setting",
				fmt.Sprintf("Database insert failed: %v", err))
		}

		// Return created setting
		responseSetting := UserSetting{
			ID:        newSetting.ID,
			UserID:    newSetting.UserID,
			Key:       newSetting.Key,
			Value:     newSetting.Value,
			CreatedAt: newSetting.CreatedAt,
			UpdatedAt: newSetting.UpdatedAt,
		}

		return c.Status(fiber.StatusOK).JSON(responseSetting)
	}
}
