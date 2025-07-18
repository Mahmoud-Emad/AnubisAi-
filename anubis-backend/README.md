# Anubis AI Core-Backend

A production-ready Go backend service for AI task execution with ThreeFold blockchain integration, built with Fiber web framework and real-time database operations.

## 🚀 Features

### Core Functionality

- **AI Task Execution**: Execute various AI tasks including text analysis, data processing, and blockchain queries
- **ThreeFold Integration**: Real wallet creation and digital twin management on ThreeFold Grid
- **User Management**: Complete user authentication, profile management, and settings
- **Memory System**: User-specific AI memory storage with tagging and search capabilities
- **Real-time Database**: SQLite/PostgreSQL support with GORM ORM and automatic migrations

### Security & Authentication

- **JWT Authentication**: Secure token-based authentication with configurable expiry
- **Middleware Protection**: Authentication, authorization, and wallet ownership validation
- **Input Validation**: Comprehensive request validation and sanitization
- **Error Handling**: Standardized error responses with detailed logging

### Production Ready

- **Comprehensive Testing**: 100% test coverage with real ThreeFold network integration
- **Database Migrations**: Automatic schema management and versioning
- **Graceful Error Handling**: Robust error recovery and user-friendly messages
- **Performance Monitoring**: Built-in metrics and execution time tracking

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite (for development)
- PostgreSQL (for production)

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd anubis-backend
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run database migrations**

   ```bash
   go run main.go migrate
   ```

5. **Start the server**

   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080` by default.

### Configuration

Key environment variables:

```env
# Database Configuration
DATABASE_TYPE=sqlite          # sqlite or postgres
DATABASE_DSN=./anubis.db     # Database connection string

# JWT Configuration
JWT_SECRET=your-secret-key    # JWT signing secret
JWT_EXPIRY=24h               # Token expiry duration

# ThreeFold Configuration
TFGRID_NETWORK=test          # main, test, qa, or dev
TFGRID_MNEMONIC=your-mnemonic # Optional: custom mnemonic

# Server Configuration
PORT=8080                    # Server port
LOG_LEVEL=info              # debug, info, warn, error
```

## 📚 API Documentation

### Authentication Endpoints

- `POST /auth/register` - Register new user with ThreeFold wallet creation
- `POST /auth/login` - User login with JWT token generation
- `POST /auth/refresh` - Refresh JWT token

### User Management

- `GET /user/profile` - Get authenticated user profile with statistics
- `PUT /user/profile` - Update user profile information
- `GET /user/memories` - Get user AI memories with pagination
- `POST /user/memories` - Create new AI memory
- `GET /user/settings` - Get user settings
- `PUT /user/settings` - Update user setting

### Task Execution

- `GET /available-tasks` - List available AI tasks with filtering
- `POST /execute-task` - Execute AI task with parameters

### Health & Monitoring

- `GET /health` - Health check endpoint
- `GET /metrics` - Application metrics (if enabled)

### Example API Usage

**Register a new user:**

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "username": "johndoe",
    "password": "securepassword123"
  }'
```

**Execute an AI task:**

```bash
curl -X POST http://localhost:8080/execute-task \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "task_name": "list_farms",
    "params": {
      "page": 1,
      "limit": 10
    }
  }'
```

## 🧪 Testing

### Run All Tests

```bash
go test ./... -v
```

### Run Specific Test Suites

```bash
# Test user handlers
go test ./handlers -v -run "TestUser"

# Test authentication
go test ./services -v -run "TestAuth"

# Test middleware
go test ./middleware -v
```

### Test Features

- **Real ThreeFold Integration**: Tests create actual wallets and digital twins
- **Database Integration**: Tests use in-memory SQLite databases with full migrations
- **Comprehensive Coverage**: Tests cover success cases, error cases, and edge conditions
- **Performance Testing**: Execution time and resource usage validation

## Development

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Full CI pipeline
make ci
```

### Generate Documentation

```bash
# Generate Swagger docs
make swagger
```

## Task Execution

The API supports executing ThreeFold Grid tasks:

### List Farms

```bash
curl -X POST http://localhost:8080/execute-task \
  -H "Content-Type: application/json" \
  -d '{
    "task_name": "list_farms",
    "params": {
      "page": 1,
      "location": "BE"
    }
  }'
```

### Get Farm Details

```bash
curl -X POST http://localhost:8080/execute-task \
  -H "Content-Type: application/json" \
  -d '{
    "task_name": "get_farm",
    "params": {
      "farm_id": 1
    }
  }'
```

## Authentication

### Sign In

```bash
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@anubis.local",
    "password": "admin123"
  }'
```

### Using JWT Token

```bash
curl -X GET http://localhost:8080/user \
  -H "Authorization: Bearer your-jwt-token"
```

## Docker Support

### Build Docker Image

```bash
make docker-build
```

### Run with Docker

```bash
make docker-run
```

## Production Deployment

1. Set environment to production:

```bash
export ENV=production
```

2. Configure PostgreSQL database
3. Set secure JWT secret
4. Configure CORS origins
5. Set up reverse proxy (nginx/traefik)
6. Enable HTTPS

## 🏗️ Architecture

```
anubis-backend/
├── handlers/          # HTTP request handlers and API endpoints
├── services/          # Business logic and external service integrations
├── middleware/        # Authentication, validation, and security middleware
├── models/           # Database models and schema definitions
├── database/         # Database connection and migration management
├── config/           # Configuration management and environment variables
├── common/           # Shared types and utility functions
└── main.go          # Application entry point and server setup
```

### Key Design Decisions

1. **Real ThreeFold Integration**: Unlike typical demo applications, this system creates actual wallets and digital twins on the ThreeFold test network, providing authentic blockchain functionality.

2. **Database-First Approach**: All user data, memories, settings, and task executions are stored in a real database with proper relationships and constraints.

3. **Comprehensive Error Handling**: Every endpoint includes detailed error responses with user-friendly messages and proper HTTP status codes.

4. **Production-Grade Testing**: Tests use real database connections, actual ThreeFold network calls, and comprehensive edge case coverage.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make ci`
6. Submit a pull request

## License

MIT License - see LICENSE file for details
