package services

import (
	"anubis-backend/config"
	"fmt"
	"log"
)

// TaskExecutor interface for executing tasks
type TaskExecutor interface {
	ExecuteTask(taskName string, params map[string]interface{}) (interface{}, error)
	GetSupportedTasks() []string
}

var executor TaskExecutor

// InitTaskService initializes the task service with the executer
func InitTaskService(cfg *config.Config) error {
	// Import the executer package from anubis-executer
	// For now, we'll create a simple implementation
	// In production, this would use the actual executer package
	
	executor = &SimpleTaskExecutor{
		network: cfg.TFGrid.Network,
	}
	
	log.Println("Task service initialized successfully")
	return nil
}

// ExecuteTask executes a task using the configured executor
func ExecuteTask(taskName string, params map[string]interface{}) (interface{}, error) {
	if executor == nil {
		return nil, fmt.Errorf("task executor not initialized")
	}
	
	return executor.ExecuteTask(taskName, params)
}

// GetSupportedTasks returns the list of supported tasks
func GetSupportedTasks() []string {
	if executor == nil {
		return []string{}
	}
	
	return executor.GetSupportedTasks()
}

// SimpleTaskExecutor is a simple implementation for demonstration
// In production, this would be replaced with the actual executer from anubis-executer
type SimpleTaskExecutor struct {
	network string
}

// ExecuteTask implements the TaskExecutor interface
func (e *SimpleTaskExecutor) ExecuteTask(taskName string, params map[string]interface{}) (interface{}, error) {
	log.Printf("Executing task: %s with params: %v", taskName, params)
	
	switch taskName {
	case "list_farms":
		return e.listFarms(params)
	case "get_farm":
		return e.getFarm(params)
	default:
		return nil, fmt.Errorf("unsupported task: %s", taskName)
	}
}

// GetSupportedTasks implements the TaskExecutor interface
func (e *SimpleTaskExecutor) GetSupportedTasks() []string {
	return []string{"list_farms", "get_farm"}
}

// listFarms simulates the list_farms task
func (e *SimpleTaskExecutor) listFarms(params map[string]interface{}) (interface{}, error) {
	// This is a mock implementation
	// In production, this would use the actual anubis-executer
	
	response := map[string]interface{}{
		"farms": []map[string]interface{}{
			{
				"farmId": 1,
				"name":   "Freefarm",
				"certificationType": "NotCertified",
				"dedicated": false,
				"pricingPolicyId": 1,
				"stellarAddress": "GCIHPMKWFMP7OLU3ICJZN5AWLWVAKZNZIFPC6XKFMFDX5BLBA5KNVULR",
				"twinId": 2,
			},
			{
				"farmId": 2,
				"name":   "MixNMatch",
				"certificationType": "NotCertified",
				"dedicated": false,
				"pricingPolicyId": 1,
				"stellarAddress": "GCZL3MUFKCHUH3PQPWAERMBYBGQXFYQXBU5ONFGJRTARFYGTLGFGOPCH",
				"twinId": 8,
			},
		},
		"total_count": 2,
		"page":       1,
		"page_size":  5,
		"network":    e.network,
	}
	
	return response, nil
}

// getFarm simulates the get_farm task
func (e *SimpleTaskExecutor) getFarm(params map[string]interface{}) (interface{}, error) {
	farmIDParam, exists := params["farm_id"]
	if !exists {
		return nil, fmt.Errorf("farm_id parameter is required")
	}
	
	// Convert farm_id to int
	var farmID int
	switch v := farmIDParam.(type) {
	case float64:
		farmID = int(v)
	case int:
		farmID = v
	default:
		return nil, fmt.Errorf("invalid farm_id type")
	}
	
	// Mock response based on farm ID
	if farmID == 1 {
		return map[string]interface{}{
			"farmId": 1,
			"name":   "Freefarm",
			"certificationType": "NotCertified",
			"dedicated": false,
			"pricingPolicyId": 1,
			"stellarAddress": "GCIHPMKWFMP7OLU3ICJZN5AWLWVAKZNZIFPC6XKFMFDX5BLBA5KNVULR",
			"twinId": 2,
			"publicIps": []map[string]interface{}{
				{
					"contract_id": 1230264,
					"gateway":     "185.69.167.1",
					"id":          "0012817744-000068-6c055",
					"ip":          "185.69.167.209/24",
				},
			},
		}, nil
	}
	
	return nil, fmt.Errorf("farm with ID %d not found", farmID)
}
