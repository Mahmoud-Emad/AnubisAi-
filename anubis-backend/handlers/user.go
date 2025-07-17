package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	User      UserInfo  `json:"user"`
	Stats     UserStats `json:"stats"`
	Timestamp time.Time `json:"timestamp"`
}

type UserStats struct {
	TasksExecuted  int       `json:"tasks_executed"`
	MemoriesStored int       `json:"memories_stored"`
	LastActivity   time.Time `json:"last_activity"`
	AccountCreated time.Time `json:"account_created"`
}

func GetUserProfile(c *fiber.Ctx) error {
	mockUser := UserInfo{
		ID:        uuid.New(),
		Email:     "user@anubis.local",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
	}

	mockStats := UserStats{
		TasksExecuted:  42,
		MemoriesStored: 15,
		LastActivity:   time.Now().Add(-2 * time.Hour),
		AccountCreated: time.Now().Add(-30 * 24 * time.Hour),
	}

	return c.Status(fiber.StatusOK).JSON(UserProfileResponse{
		User:      mockUser,
		Stats:     mockStats,
		Timestamp: time.Now(),
	})
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

func UpdateUserProfile(c *fiber.Ctx) error {
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:     "Invalid request format",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	updatedUser := UserInfo{
		ID:        uuid.New(),
		Email:     "user@anubis.local",
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsAdmin:   false,
	}

	if req.Username == "" {
		updatedUser.Username = "testuser"
	}
	if req.FirstName == "" {
		updatedUser.FirstName = "Test"
	}
	if req.LastName == "" {
		updatedUser.LastName = "User"
	}

	mockStats := UserStats{
		TasksExecuted:  42,
		MemoriesStored: 15,
		LastActivity:   time.Now(),
		AccountCreated: time.Now().Add(-30 * 24 * time.Hour),
	}

	return c.Status(fiber.StatusOK).JSON(UserProfileResponse{
		User:      updatedUser,
		Stats:     mockStats,
		Timestamp: time.Now(),
	})
}

type UserMemory struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserMemoriesResponse struct {
	Memories   []UserMemory   `json:"memories"`
	Pagination PaginationMeta `json:"pagination"`
	Timestamp  time.Time      `json:"timestamp"`
}

func GetUserMemories(c *fiber.Ctx) error {
	mockMemories := []UserMemory{
		{
			ID:        uuid.New(),
			Title:     "ThreeFold Farm Preferences",
			Content:   "User prefers farms in Belgium and Austria for better latency",
			Tags:      []string{"preferences", "location", "farms"},
			IsActive:  true,
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Title:     "Deployment History",
			Content:   "Successfully deployed 3 VMs on Freefarm last month",
			Tags:      []string{"deployment", "history", "vms"},
			IsActive:  true,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
	}

	pagination := PaginationMeta{
		Page:       1,
		PageSize:   10,
		TotalCount: len(mockMemories),
		TotalPages: 1,
	}

	return c.Status(fiber.StatusOK).JSON(UserMemoriesResponse{
		Memories:   mockMemories,
		Pagination: pagination,
		Timestamp:  time.Now(),
	})
}

type CreateMemoryRequest struct {
	Title   string   `json:"title" validate:"required,max=200"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags,omitempty"`
}

func CreateUserMemory(c *fiber.Ctx) error {
	var req CreateMemoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:     "Invalid request format",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	newMemory := UserMemory{
		ID:        uuid.New(),
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return c.Status(fiber.StatusCreated).JSON(newMemory)
}

type UserSetting struct {
	ID    uuid.UUID `json:"id"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

type UserSettingsResponse struct {
	Settings  []UserSetting `json:"settings"`
	Timestamp time.Time     `json:"timestamp"`
}

func GetUserSettings(c *fiber.Ctx) error {
	mockSettings := []UserSetting{
		{ID: uuid.New(), Key: "preferred_network", Value: "main"},
		{ID: uuid.New(), Key: "default_page_size", Value: "10"},
		{ID: uuid.New(), Key: "theme", Value: "dark"},
	}

	return c.Status(fiber.StatusOK).JSON(UserSettingsResponse{
		Settings:  mockSettings,
		Timestamp: time.Now(),
	})
}

type UpdateSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func UpdateUserSetting(c *fiber.Ctx) error {
	var req UpdateSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:     "Invalid request format",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Path:      c.Path(),
		})
	}

	updatedSetting := UserSetting{
		ID:    uuid.New(),
		Key:   req.Key,
		Value: req.Value,
	}

	return c.Status(fiber.StatusOK).JSON(updatedSetting)
}
