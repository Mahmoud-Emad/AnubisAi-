// Package handlers provides HTTP request handlers for task execution and management.
// This file contains handlers for ThreeFold Grid task operations including
// listing available tasks and executing them with proper validation and logging.
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
)

// TaskInfo represents comprehensive information about a supported ThreeFold Grid task.
// This structure provides detailed documentation for each available task including
// parameter specifications and usage examples for API consumers.
type TaskInfo struct {
	Name        string                 `json:"name" example:"list_farms"`                    // Task identifier
	Description string                 `json:"description" example:"List ThreeFold farms"`   // Human-readable description
	Parameters  map[string]interface{} `json:"parameters"`                                   // Parameter specifications
	Example     map[string]interface{} `json:"example"`                                      // Usage example
	Category    string                 `json:"category" example:"farms"`                     // Task category for organization
	Version     string                 `json:"version" example:"1.0"`                        // Task version for compatibility
}

// AvailableTasksResponse represents the response for available tasks endpoint.
// This provides a comprehensive list of all supported ThreeFold Grid operations.
type AvailableTasksResponse struct {
	Tasks     []TaskInfo `json:"tasks"`                                           // Available tasks with full documentation
	Count     int        `json:"count" example:"2"`                               // Total number of available tasks
	Timestamp time.Time  `json:"timestamp" example:"2024-01-01T12:00:00Z"`       // Response timestamp
	RequestID string     `json:"request_id,omitempty" example:"req_123456789"`   // Request identifier
}

// AvailableTasks godoc
// @Summary Get available tasks
// @Description Returns a comprehensive list of all supported ThreeFold Grid tasks with detailed documentation
// @Description Each task includes parameter specifications, examples, and usage information for API consumers
// @Tags tasks
// @Produce json
// @Success 200 {object} AvailableTasksResponse "List of available tasks retrieved successfully"
// @Router /available-tasks [get]
func AvailableTasks(c *fiber.Ctx) error {
	// Define all supported tasks with comprehensive documentation
	tasks := []TaskInfo{
		{
			Name:        "list_farms",
			Description: "List ThreeFold farms with optional filtering and pagination support",
			Category:    "farms",
			Version:     "1.0",
			Parameters: map[string]interface{}{
				"page":     "integer (optional) - Page number for pagination (default: 1, max: 1000)",
				"location": "string (optional) - Filter by country code (e.g., 'BE', 'US') or location name",
				"name":     "string (optional) - Filter by farm name using case-insensitive contains search",
				"farm_id":  "integer (optional) - Filter by specific farm ID for exact match",
			},
			Example: map[string]interface{}{
				"task_name": "list_farms",
				"params": map[string]interface{}{
					"page":     1,
					"location": "BE",
					"name":     "freefarm",
				},
			},
		},
		{
			Name:        "get_farm",
			Description: "Get comprehensive information about a specific ThreeFold farm including resources and public IPs",
			Category:    "farms",
			Version:     "1.0",
			Parameters: map[string]interface{}{
				"farm_id": "integer (required) - The unique ID of the farm to retrieve (must be > 0)",
			},
			Example: map[string]interface{}{
				"task_name": "get_farm",
				"params": map[string]interface{}{
					"farm_id": 1,
				},
			},
		},
	}

	response := AvailableTasksResponse{
		Tasks:     tasks,
		Count:     len(tasks),
		Timestamp: time.Now(),
		RequestID: c.Get("X-Request-ID", ""),
	}

	return c.JSON(response)
}

// ExecuteTaskRequest represents the request for task execution with comprehensive validation.
// This structure defines the expected format for task execution requests.
type ExecuteTaskRequest struct {
	TaskName string                 `json:"task_name" validate:"required" example:"list_farms"`  // Task identifier (required)
	Params   map[string]interface{} `json:"params" example:"{\"page\": 1}"`                      // Task parameters (optional)
}

// ExecuteTaskResponse represents the response for task execution with comprehensive result information.
// This structure provides detailed execution results including performance metrics and error details.
type ExecuteTaskResponse struct {
	TaskID    uuid.UUID   `json:"task_id" example:"123e4567-e89b-12d3-a456-426614174000"`     // Unique task execution ID
	Status    string      `json:"status" example:"success"`                                    // Execution status (success/failed)
	Data      interface{} `json:"data,omitempty"`                                             // Task result data (on success)
	Error     string      `json:"error,omitempty" example:"farm_id parameter is required"`   // Error message (on failure)
	Duration  int64       `json:"duration_ms" example:"150"`                                  // Execution time in milliseconds
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T12:00:00Z"`                  // Execution timestamp
	RequestID string      `json:"request_id,omitempty" example:"req_123456789"`              // Request identifier for tracing
}

// ExecuteTask godoc
// @Summary Execute a task
// @Description Execute a ThreeFold Grid task with comprehensive validation and error handling
// @Description This endpoint processes task execution requests, validates parameters, logs execution history, and returns detailed results
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body ExecuteTaskRequest true "Task execution request with task name and parameters"
// @Success 200 {object} ExecuteTaskResponse "Task executed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or parameters"
// @Failure 500 {object} ErrorResponse "Internal server error during task execution"
// @Router /execute-task [post]
func ExecuteTask(c *fiber.Ctx) error {
	var req ExecuteTaskRequest

	// Parse and validate JSON request body
	if err := c.BodyParser(&req); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid request format",
			"Failed to parse JSON request body: "+err.Error())
	}

	// Validate task name against supported tasks
	supportedTasks := []string{"list_farms", "get_farm"}
	isSupported := false
	for _, task := range supportedTasks {
		if task == req.TaskName {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Unsupported task",
			fmt.Sprintf("Task '%s' is not supported. Available tasks: %v", req.TaskName, supportedTasks))
	}

	// Validate required parameters based on task type
	if err := validateTaskParameters(req.TaskName, req.Params); err != nil {
		return NewErrorResponse(c, fiber.StatusBadRequest,
			"Invalid parameters",
			err.Error())
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if userIDValue := c.Locals("user_id"); userIDValue != nil {
		if uid, ok := userIDValue.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Create task execution record for audit and monitoring
	taskExecution := &models.TaskExecution{
		TaskName:   req.TaskName,
		Status:     "running",
		Parameters: mustMarshalJSON(req.Params),
	}

	if userID != nil {
		taskExecution.UserID = *userID
	}

	// Save execution record to database for audit trail
	db := database.GetDB()
	if err := db.Create(taskExecution).Error; err != nil {
		return NewErrorResponse(c, fiber.StatusInternalServerError,
			"Failed to create task execution record",
			"Database error: "+err.Error())
	}

	// Execute the task with performance monitoring
	startTime := time.Now()
	result, err := services.ExecuteTask(req.TaskName, req.Params)
	duration := time.Since(startTime).Milliseconds()

	// Update task execution record with results
	taskExecution.Duration = duration
	if err != nil {
		taskExecution.Status = "failed"
		taskExecution.ErrorMsg = err.Error()
	} else {
		taskExecution.Status = "success"
		taskExecution.Response = mustMarshalJSON(result)
	}

	// Save updated execution record (non-blocking for performance)
	if updateErr := db.Save(taskExecution).Error; updateErr != nil {
		// Log the error but don't fail the request
		// In production, this should use structured logging
		c.Set("X-Warning", "Failed to update task execution record: "+updateErr.Error())
	}

	// Return comprehensive response with execution details
	response := ExecuteTaskResponse{
		TaskID:    taskExecution.ID,
		Duration:  duration,
		Timestamp: time.Now(),
		RequestID: c.Get("X-Request-ID", ""),
	}

	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response.Status = "success"
	response.Data = result
	return c.JSON(response)
}

// validateTaskParameters validates parameters for specific tasks with comprehensive checks.
// This function ensures that all required parameters are present and have valid types.
func validateTaskParameters(taskName string, params map[string]interface{}) error {
	switch taskName {
	case "get_farm":
		// Validate required farm_id parameter
		farmIDParam, exists := params["farm_id"]
		if !exists {
			return fmt.Errorf("farm_id parameter is required for get_farm task")
		}

		// Validate farm_id type and value
		switch v := farmIDParam.(type) {
		case float64:
			if v <= 0 {
				return fmt.Errorf("farm_id must be a positive integer, got: %v", v)
			}
		case int:
			if v <= 0 {
				return fmt.Errorf("farm_id must be a positive integer, got: %v", v)
			}
		default:
			return fmt.Errorf("farm_id must be an integer, got: %T", farmIDParam)
		}

	case "list_farms":
		// Validate optional parameters for list_farms
		if pageParam, exists := params["page"]; exists {
			switch v := pageParam.(type) {
			case float64:
				if v <= 0 || v > 1000 {
					return fmt.Errorf("page must be between 1 and 1000, got: %v", v)
				}
			case int:
				if v <= 0 || v > 1000 {
					return fmt.Errorf("page must be between 1 and 1000, got: %v", v)
				}
			default:
				return fmt.Errorf("page must be an integer, got: %T", pageParam)
			}
		}

		// Validate location parameter if provided
		if locationParam, exists := params["location"]; exists {
			if _, ok := locationParam.(string); !ok {
				return fmt.Errorf("location must be a string, got: %T", locationParam)
			}
		}

		// Validate name parameter if provided
		if nameParam, exists := params["name"]; exists {
			if _, ok := nameParam.(string); !ok {
				return fmt.Errorf("name must be a string, got: %T", nameParam)
			}
		}
	}
	return nil
}

// mustMarshalJSON marshals data to JSON with safe error handling.
// This helper function ensures consistent JSON serialization across the application.
// Returns empty JSON object on error to maintain data integrity.
func mustMarshalJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	data, err := json.Marshal(v)
	if err != nil {
		// Return empty JSON object instead of empty string for consistency
		return "{}"
	}
	return string(data)
}
