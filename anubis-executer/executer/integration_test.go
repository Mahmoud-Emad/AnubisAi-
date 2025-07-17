package executer

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// Integration tests that run against the real GridProxy API
// These tests are skipped by default and can be enabled with the INTEGRATION_TESTS environment variable

func TestIntegrationListFarms(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	executor := NewTaskExecutor("main")

	// Test basic list farms
	task := Task{
		TaskName: "list_farms",
		Params:   map[string]interface{}{},
	}

	result, err := executor.ExecuteTask(task)
	if err != nil {
		t.Fatalf("failed to list farms: %v", err)
	}

	response, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map response, got %T", result)
	}

	// Check response structure
	farms, ok := response["farms"]
	if !ok {
		t.Errorf("expected farms field in response")
	}

	totalCount, ok := response["total_count"].(int)
	if !ok {
		t.Errorf("expected total_count to be int, got %T", response["total_count"])
	}

	if totalCount <= 0 {
		t.Errorf("expected positive total_count, got %d", totalCount)
	}

	t.Logf("Successfully retrieved %d farms", totalCount)
	t.Logf("Farms type: %T", farms)
}

func TestIntegrationGetFarm(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	executor := NewTaskExecutor("main")

	// Test get specific farm (farm ID 1 should exist)
	task := Task{
		TaskName: "get_farm",
		Params:   map[string]interface{}{"farm_id": float64(1)},
	}

	result, err := executor.ExecuteTask(task)
	if err != nil {
		t.Fatalf("failed to get farm: %v", err)
	}

	farm, ok := result.(Farm)
	if !ok {
		t.Fatalf("expected Farm type, got %T", result)
	}

	if farm.FarmID != 1 {
		t.Errorf("expected farm ID 1, got %d", farm.FarmID)
	}

	if farm.Name == "" {
		t.Errorf("expected non-empty farm name")
	}

	t.Logf("Successfully retrieved farm: ID=%d, Name=%s", farm.FarmID, farm.Name)
}

func TestIntegrationJSONInterface(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	executor := NewTaskExecutor("main")

	// Test JSON interface
	taskJSON := `{"task_name": "list_farms", "params": {"page": 1}}`

	responseJSON, err := executor.ExecuteTaskJSON([]byte(taskJSON))
	if err != nil {
		t.Fatalf("failed to execute task JSON: %v", err)
	}

	var response TaskResponse
	if err := json.Unmarshal(responseJSON, &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success=true, got error: %s", response.Error)
	}

	if response.Data == nil {
		t.Errorf("expected non-nil data")
	}

	t.Logf("JSON interface test successful")
}

func TestIntegrationErrorHandling(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	executor := NewTaskExecutor("main")

	// Test error case - missing farm_id
	taskJSON := `{"task_name": "get_farm", "params": {}}`

	responseJSON, err := executor.ExecuteTaskJSON([]byte(taskJSON))
	if err != nil {
		t.Fatalf("ExecuteTaskJSON should not return error: %v", err)
	}

	var response TaskResponse
	if err := json.Unmarshal(responseJSON, &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Success {
		t.Errorf("expected success=false for missing farm_id")
	}

	if response.Error == "" {
		t.Errorf("expected non-empty error message")
	}

	expectedError := "farm_id parameter is required"
	if response.Error != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, response.Error)
	}

	t.Logf("Error handling test successful: %s", response.Error)
}

func TestIntegrationPagination(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	executor := NewTaskExecutor("main")

	// Test pagination
	tests := []struct {
		page     int
		expected int
	}{
		{1, 5}, // First page should have 5 farms (our limit)
		{2, 5}, // Second page should also have 5 farms
	}

	for _, tt := range tests {
		task := Task{
			TaskName: "list_farms",
			Params:   map[string]interface{}{"page": float64(tt.page)},
		}

		result, err := executor.ExecuteTask(task)
		if err != nil {
			t.Fatalf("failed to list farms page %d: %v", tt.page, err)
		}

		response, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map response, got %T", result)
		}

		farms, ok := response["farms"].([]Farm)
		if !ok {
			t.Fatalf("expected farms to be []Farm, got %T", response["farms"])
		}

		if len(farms) != tt.expected {
			t.Errorf("page %d: expected %d farms, got %d", tt.page, tt.expected, len(farms))
		}

		page, ok := response["page"].(uint64)
		if !ok {
			t.Errorf("expected page to be uint64, got %T", response["page"])
		} else if int(page) != tt.page {
			t.Errorf("expected page %d, got %d", tt.page, int(page))
		}

		t.Logf("Page %d: retrieved %d farms", tt.page, len(farms))
	}
}

func TestIntegrationNetworkConfiguration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	// Test different networks (main should work, others might not have data)
	networks := []string{"main", "test"}

	for _, network := range networks {
		t.Run(network, func(t *testing.T) {
			executor := NewTaskExecutor(network)

			if executor.network != network {
				t.Errorf("expected network %s, got %s", network, executor.network)
			}

			// Try to ping the network (this tests connectivity)
			err := executor.gridClient.Ping()
			if err != nil {
				t.Logf("Network %s ping failed (expected for some networks): %v", network, err)
				return
			}

			t.Logf("Network %s is accessible", network)
		})
	}
}

func TestIntegrationPerformance(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	executor := NewTaskExecutor("main")

	// Test response time
	start := time.Now()

	task := Task{
		TaskName: "list_farms",
		Params:   map[string]interface{}{},
	}

	_, err := executor.ExecuteTask(task)
	if err != nil {
		t.Fatalf("failed to list farms: %v", err)
	}

	duration := time.Since(start)

	// API should respond within reasonable time (adjust as needed)
	maxDuration := 10 * time.Second
	if duration > maxDuration {
		t.Errorf("API response took too long: %v (max: %v)", duration, maxDuration)
	}

	t.Logf("API response time: %v", duration)
}
