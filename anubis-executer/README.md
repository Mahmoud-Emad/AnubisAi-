# Anubis Task Executor

A Go backend module for the Anubis AI platform that executes AI-generated task instructions to interact with the ThreeFold Grid using the official `tfgrid-sdk-go`.

## Features

- **Real GridProxy Integration**: Direct connection to ThreeFold Grid APIs
- **Farm Operations**: List and retrieve farm information with filtering
- **Pagination Support**: Configurable page size (max 5 per page)
- **Multi-Network Support**: dev, test, qa, main networks
- **JSON API**: Clean JSON input/output for AI integration
- **Error Handling**: Comprehensive validation and error responses

## Supported Tasks

- `list_farms` - List ThreeFold farms with optional filtering
- `get_farm` - Get specific farm details by ID

## Installation

```bash
git clone <repository>
cd anubis-executer
go mod tidy
```

## Usage

### CLI Tool

```bash
# Show help and supported tasks
go run main.go

# Run demo with test cases
go run main.go demo
```

### As a Library

```go
import "anubis-executer/executer"

// Create executor for main network
executor := executer.NewTaskExecutor("main")

// Execute a task
task := executer.Task{
    TaskName: "list_farms",
    Params: map[string]interface{}{
        "name": "freefarm",
        "page": 1,
    },
}

result, err := executor.ExecuteTask(task)
if err != nil {
    log.Fatal(err)
}

// Or use JSON interface
taskJSON := `{"task_name": "get_farm", "params": {"farm_id": 1}}`
responseJSON, err := executor.ExecuteTaskJSON([]byte(taskJSON))
```

## Task Examples

### List Farms
```json
{
  "task_name": "list_farms",
  "params": {
    "page": 1,
    "location": "BE",
    "name": "freefarm"
  }
}
```

### Get Farm
```json
{
  "task_name": "get_farm",
  "params": {
    "farm_id": 1
  }
}
```

## Response Format

All responses follow this structure:

```json
{
  "success": true,
  "data": {
    "farms": [...],
    "total_count": 4134,
    "page": 1,
    "page_size": 5,
    "network": "main"
  }
}
```

Error responses:
```json
{
  "success": false,
  "error": "farm_id parameter is required"
}
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./executer -run TestListFarms
```

## Networks

- `main` - Production ThreeFold Grid
- `test` - Test network
- `qa` - QA network  
- `dev` - Development network

## Architecture

```
anubis-executer/
├── executer/
│   ├── executor.go      # Main task execution logic
│   ├── handlers.go      # Task-specific handlers
│   ├── task_types.go    # Data structures
│   └── *_test.go        # Unit tests
├── main.go              # CLI interface
└── README.md
```

## Contributing

1. Add new task handlers in `handlers.go`
2. Update the task router in `executor.go`
3. Add corresponding tests
4. Update documentation
