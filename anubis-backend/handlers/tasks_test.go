package handlers

import (
	"anubis-backend/config"
	"anubis-backend/database"
	"anubis-backend/services"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTaskTestApp creates a test app for task endpoints
func setupTaskTestApp() *fiber.App {
	// Initialize test database
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:       "sqlite",
			SQLitePath: ":memory:",
		},
		TFGrid: config.TFGridConfig{
			Network: "test",
		},
	}
	err := database.InitDatabase(cfg)
	if err != nil {
		panic("Failed to initialize test database: " + err.Error())
	}

	// Run migrations
	err = database.RunMigrations()
	if err != nil {
		panic("Failed to run migrations: " + err.Error())
	}

	// Initialize task service
	err = services.InitTaskService(cfg)
	if err != nil {
		panic("Failed to initialize task service: " + err.Error())
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup task routes
	app.Get("/available-tasks", AvailableTasks)
	app.Post("/execute-task", ExecuteTask)

	return app
}

func TestAvailableTasks_Success(t *testing.T) {
	app := setupTaskTestApp()

	req := httptest.NewRequest("GET", "/available-tasks", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response AvailableTasksResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.Tasks)
	assert.Greater(t, len(response.Tasks), 0)
	assert.Equal(t, len(response.Tasks), response.Count)

	// Verify task structure
	task := response.Tasks[0]
	assert.NotEmpty(t, task.Name)
	assert.NotEmpty(t, task.Description)
	assert.NotEmpty(t, task.Category)
	assert.NotEmpty(t, task.Version)
	assert.NotNil(t, task.Parameters)
	assert.NotNil(t, task.Example)
}

func TestAvailableTasks_WithFilters(t *testing.T) {
	app := setupTaskTestApp()

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "Filter by category",
			query:    "?category=ai",
			expected: 1, // Should have at least 1 AI task
		},
		{
			name:     "Filter by difficulty",
			query:    "?difficulty=1",
			expected: 1, // Should have at least 1 easy task
		},
		{
			name:     "Filter by duration",
			query:    "?max_duration=30",
			expected: 1, // Should have at least 1 quick task
		},
		{
			name:     "Combined filters",
			query:    "?category=blockchain&difficulty=2",
			expected: 0, // May or may not have results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/available-tasks"+tt.query, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response AvailableTasksResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.NotNil(t, response.Tasks)
			if tt.expected > 0 {
				assert.GreaterOrEqual(t, len(response.Tasks), tt.expected)
			}
		})
	}
}

func TestExecuteTask_Success(t *testing.T) {
	app := setupTaskTestApp()

	executeReq := ExecuteTaskRequest{
		TaskName: "list_farms",
		Params: map[string]interface{}{
			"page":  1,
			"limit": 10,
		},
	}

	body, err := json.Marshal(executeReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/execute-task", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Print response for debugging
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Response status: %d, body: %s", resp.StatusCode, string(body))
		// Reset body for further reading
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response ExecuteTaskResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.TaskID)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Data)
	assert.GreaterOrEqual(t, response.Duration, int64(0))
	assert.NotEmpty(t, response.Timestamp)
}

func TestExecuteTask_ValidationErrors(t *testing.T) {
	app := setupTaskTestApp()

	tests := []struct {
		name        string
		request     ExecuteTaskRequest
		expectedMsg string
	}{
		{
			name: "Missing task name",
			request: ExecuteTaskRequest{
				Params: map[string]interface{}{
					"page": 1,
				},
			},
			expectedMsg: "is not supported",
		},
		{
			name: "Invalid task name",
			request: ExecuteTaskRequest{
				TaskName: "non-existent-task",
				Params: map[string]interface{}{
					"page": 1,
				},
			},
			expectedMsg: "is not supported",
		},
		{
			name: "Invalid parameter type",
			request: ExecuteTaskRequest{
				TaskName: "list_farms",
				Params: map[string]interface{}{
					"page": "invalid", // Should be number
				},
			},
			expectedMsg: "page must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/execute-task", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			var response ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response.Message, tt.expectedMsg)
		})
	}
}

func TestExecuteTask_InvalidJSON(t *testing.T) {
	app := setupTaskTestApp()

	req := httptest.NewRequest("POST", "/execute-task", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response.Error, "Invalid request format")
}

func TestExecuteTask_AllTaskTypes(t *testing.T) {
	app := setupTaskTestApp()

	taskTests := []struct {
		taskName   string
		parameters map[string]interface{}
	}{
		{
			taskName: "list_farms",
			parameters: map[string]interface{}{
				"page":  1,
				"limit": 10,
			},
		},
		{
			taskName: "get_farm",
			parameters: map[string]interface{}{
				"farm_id": 1,
			},
		},
	}

	for _, tt := range taskTests {
		t.Run("Execute_"+tt.taskName, func(t *testing.T) {
			executeReq := ExecuteTaskRequest{
				TaskName: tt.taskName,
				Params:   tt.parameters,
			}

			body, err := json.Marshal(executeReq)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/execute-task", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response ExecuteTaskResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.NotEmpty(t, response.TaskID)
			assert.Equal(t, "success", response.Status)
			assert.NotNil(t, response.Data)
		})
	}
}

func TestExecuteTask_PerformanceMetrics(t *testing.T) {
	app := setupTaskTestApp()

	executeReq := ExecuteTaskRequest{
		TaskName: "list_farms",
		Params: map[string]interface{}{
			"page":  1,
			"limit": 5,
		},
	}

	body, err := json.Marshal(executeReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/execute-task", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response ExecuteTaskResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Verify performance metrics are included
	assert.NotEmpty(t, response.Timestamp)
	assert.GreaterOrEqual(t, response.Duration, int64(0))

	// Verify result contains data
	assert.NotNil(t, response.Data)
	assert.Equal(t, "success", response.Status)
}
