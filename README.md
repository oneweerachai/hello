# User REST API

A REST API for user management built with Go and Gin framework.

## Features

- Create new users with validation
- Retrieve users by ID
- List all users
- In-memory storage (for demonstration)
- Input validation with proper error messages
- Structured JSON responses
- CORS support
- Request logging
- Health check endpoint

## API Endpoints

### Health Check
- **GET** `/health` - Check if the server is running

### User Management
- **POST** `/api/users` - Create a new user
- **GET** `/api/users` - Get all users
- **GET** `/api/users/:id` - Get user by ID

## User Model

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "1234567890",
  "date_of_birth": "1990-01-15",
  "address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "USA"
  }
}
```

## Required Fields
- `first_name` (2-50 characters)
- `last_name` (2-50 characters)
- `email` (valid email format)

## Optional Fields
- `phone` (10-15 characters)
- `date_of_birth` (YYYY-MM-DD format)
- `address` (object with street, city, state, postal_code, country)

## Getting Started

### Prerequisites
- Go 1.21 or higher

### Installation

1. Clone the repository
```bash
git clone <repository-url>
cd user-api
```

2. Install dependencies
```bash
go mod tidy
```

3. Run the application
```bash
go run main.go
```

The server will start on port 8080 by default.

### Environment Variables
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment mode (default: development)

## Usage Examples

### Create a User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "1234567890",
    "date_of_birth": "1990-01-15",
    "address": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "USA"
    }
  }'
```

### Get All Users
```bash
curl http://localhost:8080/api/users
```

### Get User by ID
```bash
curl http://localhost:8080/api/users/{user-id}
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Response Format

All API responses follow this structure:

### Success Response
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": "uuid",
    "first_name": "John",
    "last_name": "Doe",
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "phone": "1234567890",
    "date_of_birth": "1990-01-15",
    "address": {...},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Error Response
```json
{
  "status": "error",
  "message": "Validation failed",
  "error": "first_name is required; email must be a valid email address"
}
```

## Project Structure

```
user-api/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── config/
│   └── config.go          # Configuration management
├── models/
│   └── user.go            # User model and validation
├── repository/
│   └── user_repository.go # Data access layer
├── services/
│   └── user_service.go    # Business logic
├── handlers/
│   └── user_handler.go    # HTTP handlers
├── middleware/
│   └── middleware.go      # HTTP middleware
└── utils/
    └── response.go        # Response utilities
```
