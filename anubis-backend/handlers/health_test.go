package handlers

import (
	"anubis-backend/config"
	"anubis-backend/database"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Initialize database for testing
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Host: ":memory:",
		},
	}
	database.InitDatabase(cfg)

	app := fiber.New()

	app.Get("/health-check", HealthCheck)

	req := httptest.NewRequest("GET", "/health-check", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "1.0.0", response.Version)
	assert.False(t, response.Timestamp.IsZero())
	assert.Contains(t, response.Services, "api")
	assert.Equal(t, "healthy", response.Services["api"])
}

func TestHome(t *testing.T) {
	app := fiber.New()

	app.Get("/home", Home)

	req := httptest.NewRequest("GET", "/home", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response HomeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, "Welcome to Anubis AI Core-Backend API", response.Message)
	assert.Equal(t, "Anubis AI Core-Backend", response.API)
	assert.Equal(t, "1.0.0", response.Version)
	assert.False(t, response.Timestamp.IsZero())
	assert.Contains(t, response.Endpoints, "/health-check")
	assert.Contains(t, response.Endpoints, "/available-tasks")
	assert.Contains(t, response.Endpoints, "/execute-task")
}
