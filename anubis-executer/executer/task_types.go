package executer

import (
	"encoding/json"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// Task represents a JSON task instruction from the AI
type Task struct {
	TaskName string                 `json:"task_name"`
	Params   map[string]interface{} `json:"params"`
}

// TaskResponse represents the response from executing a task
type TaskResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Farm represents a ThreeFold farm (using the real GridProxy types.Farm)
type Farm = types.Farm

// ParseTask parses JSON bytes into a Task struct
func ParseTask(data []byte) (*Task, error) {
	var task Task
	err := json.Unmarshal(data, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ToJSON converts a TaskResponse to JSON bytes
func (tr *TaskResponse) ToJSON() ([]byte, error) {
	return json.Marshal(tr)
}
