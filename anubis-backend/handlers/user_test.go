package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"anubis-backend/config"
	"anubis-backend/database"
	"anubis-backend/middleware"
	"anubis-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUserTestApp creates a test app with authentication middleware
func setupUserTestApp() (*fiber.App, *services.AuthService) {
	// Create test config
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: ":memory:",
		},
		TFGrid: config.TFGridConfig{
			Network: "test",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing-purposes-only",
			Expiry: 24 * time.Hour,
		},
	}

	// Initialize test database
	database.InitDatabase(cfg)
	database.RunMigrations()

	// Initialize auth service
	authService := services.NewAuthService(cfg)
	InitAuthService(cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes with authentication
	protected := app.Group("/", middleware.AuthMiddleware(authService))
	protected.Get("/user", GetUserProfile)
	protected.Put("/user", UpdateUserProfile)
	protected.Get("/user/memories", GetUserMemories)
	protected.Post("/user/memories", CreateUserMemory)
	protected.Get("/user/settings", GetUserSettings)
	protected.Put("/user/settings", UpdateUserSetting)

	return app, authService
}

// createTestUserAndToken creates a test user and returns a valid JWT token
func createTestUserAndToken(authService *services.AuthService) (string, error) {
	registerReq := services.RegisterRequest{
		FirstName: "Test",
		LastName:  "User",
		Email:     "testuser@example.com",
		Password:  "testpassword123",
		Username:  "testuser",
	}

	_, err := authService.Register(&registerReq)
	if err != nil {
		return "", err
	}

	loginReq := services.LoginRequest{
		Email:    "testuser@example.com",
		Password: "testpassword123",
	}

	response, err := authService.Login(&loginReq)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}

func TestGetUserProfile_Success(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response UserProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.User.ID)
	assert.Equal(t, "testuser@example.com", response.User.Email)
	assert.Equal(t, "testuser", response.User.Username)
	assert.Equal(t, "Test", response.User.FirstName)
	assert.Equal(t, "User", response.User.LastName)
	assert.False(t, response.User.IsAdmin)
}

func TestGetUserProfile_Unauthorized(t *testing.T) {
	app, _ := setupUserTestApp()

	req := httptest.NewRequest("GET", "/user", nil)
	// No Authorization header

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetUserProfile_InvalidToken(t *testing.T) {
	app, _ := setupUserTestApp()

	req := httptest.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateUserProfile_Success(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	updateReq := UpdateUserRequest{
		FirstName: "Updated",
		LastName:  "Name",
		Username:  "updateduser",
	}

	body, err := json.Marshal(updateReq)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/user", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response UserProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Updated", response.User.FirstName)
	assert.Equal(t, "Name", response.User.LastName)
	assert.Equal(t, "updateduser", response.User.Username)
}

func TestUpdateUserProfile_InvalidJSON(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/user", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUserMemories_Success(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/user/memories", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Debug: print response status and body
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Response status: %d, body: %s", resp.StatusCode, string(body))
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response []UserMemory
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Initialize empty slice if nil (JSON unmarshaling behavior)
	if response == nil {
		response = []UserMemory{}
	}

	assert.GreaterOrEqual(t, len(response), 0) // Can be empty initially
}

func TestCreateUserMemory_Success(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	memoryReq := CreateMemoryRequest{
		Title:   "Test Memory",
		Content: "This is a test memory",
		Tags:    []string{"test", "memory"},
	}

	body, err := json.Marshal(memoryReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/user/memories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response UserMemory
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "Test Memory", response.Title)
	assert.Equal(t, "This is a test memory", response.Content)
	assert.Equal(t, []string{"test", "memory"}, response.Tags)
	assert.True(t, response.IsActive)
}

func TestGetUserSettings_Success(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/user/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response []UserSetting
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Initialize empty slice if nil (JSON unmarshaling behavior)
	if response == nil {
		response = []UserSetting{}
	}

	// Settings can be empty initially - no default settings are created
	assert.GreaterOrEqual(t, len(response), 0)
}

func TestUpdateUserSetting_Success(t *testing.T) {
	app, authService := setupUserTestApp()
	token, err := createTestUserAndToken(authService)
	require.NoError(t, err)

	settingReq := UpdateSettingRequest{
		Key:   "theme",
		Value: "dark",
	}

	body, err := json.Marshal(settingReq)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", "/user/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response UserSetting
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "theme", response.Key)
	assert.Equal(t, "dark", response.Value)
}
