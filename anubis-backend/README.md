# Anubis AI Core-Backend API

A comprehensive REST API serving as the core backend for the Anubis AI platform, providing endpoints for user management, task execution, and AI memory management with ThreeFold Grid integration.

## Features

- **Task Execution**: Execute ThreeFold Grid tasks with real-time results
- **User Management**: Complete user authentication and profile management
- **AI Memories**: Store and retrieve AI conversation memories
- **User Settings**: Customizable user preferences
- **Real-time Monitoring**: Health checks and performance metrics
- **Comprehensive Documentation**: Auto-generated Swagger/OpenAPI docs
- **Database Support**: SQLite (development) and PostgreSQL (production)
- **Security**: JWT authentication, rate limiting, CORS protection

## Quick Start

### Prerequisites

- Go 1.21 or higher
- SQLite (for development)
- PostgreSQL (for production)

### Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd anubis-backend
```

2. Install dependencies:

```bash
make deps
```

3. Copy environment configuration:

```bash
cp .env.example .env
```

4. Run the application:

```bash
make run
```

The API will be available at `http://localhost:8080`

### Development Mode

For development with hot reload:

```bash
make dev
```

## API Documentation

Once the server is running, visit:

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **API Home**: `http://localhost:8080/home`
- **Health Check**: `http://localhost:8080/health-check`

## Available Endpoints

### Public Endpoints

- `GET /health-check` - Health status
- `GET /home` - API information
- `GET /available-tasks` - List supported tasks
- `POST /execute-task` - Execute ThreeFold Grid tasks
- `POST /auth/signin` - User authentication
- `POST /auth/signup` - User registration
- `POST /auth/refresh` - Refresh JWT token
- `POST /reset-password` - Password reset request

### Protected Endpoints (Require Authentication)

- `GET /user` - Get user profile
- `PUT /user` - Update user profile
- `GET /user/memories` - Get user memories
- `POST /user/memories` - Create user memory
- `GET /user/settings` - Get user settings
- `PUT /user/settings` - Update user settings

## Configuration

The application uses environment variables for configuration. See `.env.example` for all available options.

### Key Configuration Options

- `ENV`: Environment (development/production)
- `PORT`: Server port (default: 8080)
- `DB_TYPE`: Database type (sqlite/postgres)
- `JWT_SECRET`: JWT signing secret
- `TFGRID_NETWORK`: ThreeFold Grid network (main/test/qa/dev)

## Database

### Development (SQLite)

```bash
# Database is automatically created at ./data/anubis.db
make run
```

### Production (PostgreSQL)

```bash
# Set environment variables
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=anubis
export DB_USER=anubis
export DB_PASSWORD=your-password

make run
```

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration
```

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

## Architecture

```
anubis-backend/
├── config/          # Configuration management
├── database/        # Database connection and migrations
├── docs/           # Swagger documentation
├── handlers/       # HTTP request handlers
├── middleware/     # HTTP middleware
├── models/         # Database models
├── routes/         # Route definitions
├── services/       # Business logic
├── main.go         # Application entry point
└── README.md
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make ci`
6. Submit a pull request

## License

MIT License - see LICENSE file for details
