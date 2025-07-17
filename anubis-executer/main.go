package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"anubis-executer/executer"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		runDemo()
		return
	}

	// Default: run as a simple CLI tool
	fmt.Println("Anubis Task Executor")
	fmt.Println("Usage:")
	fmt.Println("  go run main.go demo          - Run demo with test cases")
	fmt.Println("  go run main.go               - Show this help")
	fmt.Println("")
	fmt.Println("Supported tasks:")
	executor := executer.NewTaskExecutor("main")
	for _, task := range executor.GetSupportedTasks() {
		fmt.Printf("  - %s\n", task)
	}
}

func runDemo() {
	log.Println("Starting Anubis Task Executor Demo")

	// Create a new task executor (using main network)
	executor := executer.NewTaskExecutor("main")

	// Test cases with various JSON task examples
	testCases := []string{
		// List all farms (first page, max 5)
		`{"task_name": "list_farms", "params": {}}`,

		// List farms with name filter
		`{"task_name": "list_farms", "params": {"name": "freefarm"}}`,

		// Get specific farm details
		`{"task_name": "get_farm", "params": {"farm_id": 1}}`,

		// List farms page 2
		`{"task_name": "list_farms", "params": {"page": 2}}`,

		// Test error case - unknown task
		`{"task_name": "unknown_task", "params": {}}`,

		// Test error case - missing required parameter
		`{"task_name": "get_farm", "params": {}}`,
	}

	// Execute each test case
	for i, testCase := range testCases {
		fmt.Printf("\n=== Test Case %d ===\n", i+1)
		fmt.Printf("Input: %s\n", testCase)

		// Execute the task
		responseJSON, err := executor.ExecuteTaskJSON([]byte(testCase))
		if err != nil {
			log.Printf("Error executing task: %v", err)
			continue
		}

		// Pretty print the response
		var prettyResponse map[string]interface{}
		if err := json.Unmarshal(responseJSON, &prettyResponse); err == nil {
			prettyJSON, _ := json.MarshalIndent(prettyResponse, "", "  ")
			fmt.Printf("Response: %s\n", string(prettyJSON))
		} else {
			fmt.Printf("Response: %s\n", string(responseJSON))
		}
	}

	fmt.Println("\nAnubis Task Executor Demo completed!")
}
