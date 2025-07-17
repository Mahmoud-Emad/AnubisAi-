package executer

import (
	"encoding/json"
	"testing"
)

func TestParseTask(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Task
		expectError bool
	}{
		{
			name:  "valid list_farms task",
			input: `{"task_name": "list_farms", "params": {"page": 1}}`,
			expected: Task{
				TaskName: "list_farms",
				Params: map[string]interface{}{
					"page": float64(1), // JSON numbers are float64
				},
			},
			expectError: false,
		},
		{
			name:  "valid get_farm task",
			input: `{"task_name": "get_farm", "params": {"farm_id": 123}}`,
			expected: Task{
				TaskName: "get_farm",
				Params: map[string]interface{}{
					"farm_id": float64(123),
				},
			},
			expectError: false,
		},
		{
			name:  "empty params",
			input: `{"task_name": "list_farms", "params": {}}`,
			expected: Task{
				TaskName: "list_farms",
				Params:   map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       `{"task_name": "list_farms", "params":}`,
			expected:    Task{},
			expectError: true,
		},
		{
			name:        "missing task_name",
			input:       `{"params": {}}`,
			expected:    Task{TaskName: "", Params: map[string]interface{}{}},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := ParseTask([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if task.TaskName != tt.expected.TaskName {
				t.Errorf("expected task_name %q, got %q", tt.expected.TaskName, task.TaskName)
			}

			// Compare params
			if len(task.Params) != len(tt.expected.Params) {
				t.Errorf("expected %d params, got %d", len(tt.expected.Params), len(task.Params))
			}

			for key, expectedValue := range tt.expected.Params {
				if actualValue, exists := task.Params[key]; !exists {
					t.Errorf("expected param %q not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("expected param %q to be %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestTaskResponseToJSON(t *testing.T) {
	tests := []struct {
		name     string
		response TaskResponse
		expected string
	}{
		{
			name: "success response",
			response: TaskResponse{
				Success: true,
				Data:    map[string]interface{}{"farms": []string{"farm1", "farm2"}},
			},
			expected: `{"success":true,"data":{"farms":["farm1","farm2"]}}`,
		},
		{
			name: "error response",
			response: TaskResponse{
				Success: false,
				Error:   "farm_id parameter is required",
			},
			expected: `{"success":false,"error":"farm_id parameter is required"}`,
		},
		{
			name: "success response with nil data",
			response: TaskResponse{
				Success: true,
				Data:    nil,
			},
			expected: `{"success":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := tt.response.ToJSON()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Parse both to compare structure (order-independent)
			var expected, actual map[string]interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("failed to parse expected JSON: %v", err)
			}
			if err := json.Unmarshal(jsonBytes, &actual); err != nil {
				t.Fatalf("failed to parse actual JSON: %v", err)
			}

			// Compare success field
			if expected["success"] != actual["success"] {
				t.Errorf("expected success %v, got %v", expected["success"], actual["success"])
			}

			// Compare error field if present
			if expectedError, exists := expected["error"]; exists {
				if actualError, exists := actual["error"]; !exists {
					t.Errorf("expected error field but got none")
				} else if expectedError != actualError {
					t.Errorf("expected error %q, got %q", expectedError, actualError)
				}
			}

			// Compare data field if present
			if expectedData, exists := expected["data"]; exists {
				if actualData, exists := actual["data"]; !exists {
					t.Errorf("expected data field but got none")
				} else {
					// For complex data comparison, convert back to JSON strings
					expectedDataJSON, _ := json.Marshal(expectedData)
					actualDataJSON, _ := json.Marshal(actualData)
					if string(expectedDataJSON) != string(actualDataJSON) {
						t.Errorf("expected data %s, got %s", expectedDataJSON, actualDataJSON)
					}
				}
			}
		})
	}
}
