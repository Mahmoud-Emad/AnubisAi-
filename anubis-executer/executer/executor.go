package executer

import (
	"fmt"
	"log"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
)

// TaskExecutor handles the execution of tasks
type TaskExecutor struct {
	gridClient client.Client
	network    string // dev, test, qa, main
}

// NewTaskExecutor creates a new TaskExecutor instance
func NewTaskExecutor(network string) *TaskExecutor {
	// Default GridProxy endpoints for different networks
	var endpoints []string
	switch network {
	case "dev":
		endpoints = []string{
			"https://gridproxy.dev.grid.tf/",
			"https://gridproxy.02.dev.grid.tf/",
		}
	case "test":
		endpoints = []string{
			"https://gridproxy.test.grid.tf/",
			"https://gridproxy.02.test.grid.tf/",
		}
	case "qa":
		endpoints = []string{
			"https://gridproxy.qa.grid.tf/",
			"https://gridproxy.02.qa.grid.tf/",
		}
	case "main":
		endpoints = []string{
			"https://gridproxy.grid.tf/",
			"https://gridproxy.02.grid.tf/",
		}
	default:
		// Default to main network
		network = "main"
		endpoints = []string{
			"https://gridproxy.grid.tf/",
			"https://gridproxy.02.grid.tf/",
		}
	}

	gridClient := client.NewClient(endpoints...)

	return &TaskExecutor{
		gridClient: gridClient,
		network:    network,
	}
}

// ExecuteTask routes and executes a task based on its name
func (te *TaskExecutor) ExecuteTask(task Task) (interface{}, error) {
	log.Printf("Executing task: %s with params: %v", task.TaskName, task.Params)

	switch task.TaskName {
	case "list_farms":
		return te.listFarms(task.Params)
	case "get_farm":
		return te.getFarm(task.Params)
	default:
		return nil, fmt.Errorf("unknown task: %s", task.TaskName)
	}
}

// ExecuteTaskJSON is a convenience method that takes JSON input and returns JSON output
func (te *TaskExecutor) ExecuteTaskJSON(taskJSON []byte) ([]byte, error) {
	// Parse the task
	task, err := ParseTask(taskJSON)
	if err != nil {
		response := TaskResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse task: %v", err),
		}
		return response.ToJSON()
	}

	// Execute the task
	result, err := te.ExecuteTask(*task)

	// Create response
	var response TaskResponse
	if err != nil {
		response = TaskResponse{
			Success: false,
			Error:   err.Error(),
		}
	} else {
		response = TaskResponse{
			Success: true,
			Data:    result,
		}
	}

	return response.ToJSON()
}

// GetSupportedTasks returns a list of supported task names
func (te *TaskExecutor) GetSupportedTasks() []string {
	return []string{
		"list_farms",
		"get_farm",
	}
}
