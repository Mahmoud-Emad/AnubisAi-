package executer

import (
	"encoding/json"
	"testing"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

func TestNewTaskExecutor(t *testing.T) {
	tests := []struct {
		name            string
		network         string
		expectedNetwork string
	}{
		{"main network", "main", "main"},
		{"dev network", "dev", "dev"},
		{"test network", "test", "test"},
		{"qa network", "qa", "qa"},
		{"invalid network defaults to main", "invalid", "main"},
		{"empty network defaults to main", "", "main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewTaskExecutor(tt.network)

			if executor == nil {
				t.Errorf("expected non-nil executor")
				return
			}

			if executor.network != tt.expectedNetwork {
				t.Errorf("expected network %q, got %q", tt.expectedNetwork, executor.network)
			}

			if executor.gridClient == nil {
				t.Errorf("expected non-nil gridClient")
			}
		})
	}
}

func TestGetSupportedTasks(t *testing.T) {
	executor := NewTaskExecutor("test")
	tasks := executor.GetSupportedTasks()

	expectedTasks := []string{"list_farms", "get_farm"}

	if len(tasks) != len(expectedTasks) {
		t.Errorf("expected %d tasks, got %d", len(expectedTasks), len(tasks))
	}

	for i, expected := range expectedTasks {
		if i >= len(tasks) || tasks[i] != expected {
			t.Errorf("expected task %q at index %d, got %q", expected, i, tasks[i])
		}
	}
}

func TestExecuteTask(t *testing.T) {
	// Create mock farms for testing
	mockFarms := []types.Farm{
		{FarmID: 1, Name: "TestFarm1"},
		{FarmID: 2, Name: "TestFarm2"},
	}

	tests := []struct {
		name          string
		task          Task
		expectedError bool
		errorMessage  string
	}{
		{
			name: "valid list_farms task",
			task: Task{
				TaskName: "list_farms",
				Params:   map[string]interface{}{},
			},
			expectedError: false,
		},
		{
			name: "valid get_farm task",
			task: Task{
				TaskName: "get_farm",
				Params:   map[string]interface{}{"farm_id": float64(1)},
			},
			expectedError: false,
		},
		{
			name: "invalid task name",
			task: Task{
				TaskName: "invalid_task",
				Params:   map[string]interface{}{},
			},
			expectedError: true,
			errorMessage:  "unknown task: invalid_task",
		},
		{
			name: "get_farm missing farm_id",
			task: Task{
				TaskName: "get_farm",
				Params:   map[string]interface{}{},
			},
			expectedError: true,
			errorMessage:  "farm_id parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor with mock client
			executor := &TaskExecutor{
				gridClient: &MockGridClient{
					farms: mockFarms,
				},
				network: "test",
			}

			result, err := executor.ExecuteTask(tt.task)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("expected error message %q, got %q", tt.errorMessage, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected non-nil result")
			}
		})
	}
}

func TestExecuteTaskJSON(t *testing.T) {
	// Create mock farms for testing
	mockFarms := []types.Farm{
		{FarmID: 1, Name: "TestFarm1"},
	}

	tests := []struct {
		name          string
		taskJSON      string
		expectedError bool
		checkSuccess  bool
	}{
		{
			name:         "valid JSON task",
			taskJSON:     `{"task_name": "list_farms", "params": {}}`,
			checkSuccess: true,
		},
		{
			name:         "valid get_farm task",
			taskJSON:     `{"task_name": "get_farm", "params": {"farm_id": 1}}`,
			checkSuccess: true,
		},
		{
			name:     "invalid JSON",
			taskJSON: `{"task_name": "list_farms", "params":}`,
		},
		{
			name:     "unknown task",
			taskJSON: `{"task_name": "unknown_task", "params": {}}`,
		},
		{
			name:     "missing farm_id",
			taskJSON: `{"task_name": "get_farm", "params": {}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor with mock client
			executor := &TaskExecutor{
				gridClient: &MockGridClient{
					farms: mockFarms,
				},
				network: "test",
			}

			responseJSON, err := executor.ExecuteTaskJSON([]byte(tt.taskJSON))

			if err != nil {
				t.Errorf("ExecuteTaskJSON should not return error, got: %v", err)
				return
			}

			// Parse response
			var response TaskResponse
			if err := json.Unmarshal(responseJSON, &response); err != nil {
				t.Errorf("failed to parse response JSON: %v", err)
				return
			}

			if tt.checkSuccess {
				if !response.Success {
					t.Errorf("expected success=true, got success=false with error: %s", response.Error)
				}
				if response.Data == nil {
					t.Errorf("expected non-nil data for successful response")
				}
			} else {
				if response.Success {
					t.Errorf("expected success=false, got success=true")
				}
				if response.Error == "" {
					t.Errorf("expected non-empty error message for failed response")
				}
			}
		})
	}
}

func TestExecuteTaskJSONStructure(t *testing.T) {
	// Test the structure of successful responses
	mockFarms := []types.Farm{
		{FarmID: 1, Name: "TestFarm1"},
		{FarmID: 2, Name: "TestFarm2"},
	}

	executor := &TaskExecutor{
		gridClient: &MockGridClient{
			farms: mockFarms,
		},
		network: "test",
	}

	// Test list_farms response structure
	taskJSON := `{"task_name": "list_farms", "params": {}}`
	responseJSON, err := executor.ExecuteTaskJSON([]byte(taskJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(responseJSON, &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check top-level structure
	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("expected success=true, got %v", response["success"])
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Errorf("expected data to be map[string]interface{}, got %T", response["data"])
		return
	}

	// Check data structure for list_farms
	expectedFields := []string{"farms", "total_count", "page", "page_size", "network"}
	for _, field := range expectedFields {
		if _, exists := data[field]; !exists {
			t.Errorf("expected field %q in data, but not found", field)
		}
	}

	// Check network field
	if network, ok := data["network"].(string); !ok || network != "test" {
		t.Errorf("expected network=test, got %v", data["network"])
	}
}
