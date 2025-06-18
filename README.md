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
- Request logging with trace correlation
- Health check endpoint
- **Distributed tracing with OpenTelemetry**
- **Trace context propagation across all layers**
- **Comprehensive span instrumentation**
- **Configurable trace exporters (console/OTLP)**

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

#### Server Configuration
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment mode (default: development)

#### Tracing Configuration
- `TRACING_ENABLED` - Enable/disable tracing (default: true in development, false in production)
- `TRACING_EXPORTER` - Trace exporter type: "console" or "otlp" (default: console in dev, otlp in prod)
- `TRACING_OTLP_ENDPOINT` - OTLP endpoint URL (default: http://localhost:4318/v1/traces)
- `TRACING_SAMPLING_RATE` - Sampling rate 0.0-1.0 (default: 1.0 in dev, 0.1 in prod)

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

## Distributed Tracing

This API includes comprehensive distributed tracing using OpenTelemetry, providing full observability across all layers.

### Tracing Features

- **Complete request lifecycle tracing** for all API endpoints
- **Multi-layer span instrumentation**: HTTP handlers, service layer, repository layer
- **Rich span attributes**: HTTP details, user information, operation results
- **Error tracking** with detailed error attributes and stack traces
- **Trace context propagation** between all service layers
- **Trace ID correlation** in logs and API responses
- **Configurable exporters**: Console (development) or OTLP (production)
- **Sampling control** for performance optimization

### Tracing Configuration

#### Development Setup (Console Exporter)
```bash
export TRACING_ENABLED=true
export TRACING_EXPORTER=console
export TRACING_SAMPLING_RATE=1.0
go run main.go
```

#### Production Setup (OTLP Exporter)
```bash
export TRACING_ENABLED=true
export TRACING_EXPORTER=otlp
export TRACING_OTLP_ENDPOINT=http://jaeger:4318/v1/traces
export TRACING_SAMPLING_RATE=0.1
go run main.go
```

#### Disable Tracing
```bash
export TRACING_ENABLED=false
go run main.go
```

### Trace Attributes

The API automatically captures the following span attributes:

#### HTTP Layer
- `http.method` - HTTP method (GET, POST, etc.)
- `http.url` - Request URL
- `http.status_code` - Response status code
- `http.user_agent` - User agent string
- `http.client_ip` - Client IP address
- `http.request.size` - Request payload size
- `http.response.size` - Response payload size

#### User Operations
- `user.id` - User ID for user-specific operations
- `user.email` - User email address
- `user.first_name` - User first name
- `user.last_name` - User last name

#### Database Operations
- `db.operation` - Database operation (create, get_by_id, etc.)
- `db.table` - Table/collection name
- `users.count` - Number of users returned

#### Error Tracking
- `error.type` - Error category (validation_error, not_found, etc.)
- `error.message` - Detailed error message

### Trace Context Propagation

Trace context is automatically propagated through:
1. **HTTP requests** via standard trace headers
2. **Service layer calls** via Go context
3. **Repository operations** via Go context
4. **Log messages** with trace ID correlation

### Observability Integration

#### Jaeger Setup
```bash
# Run Jaeger with Docker
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

# Configure API to use Jaeger
export TRACING_OTLP_ENDPOINT=http://localhost:4318/v1/traces
```

#### Viewing Traces
1. Open Jaeger UI: http://localhost:16686
2. Select service: `user-api`
3. Search for traces by operation or trace ID

### API Response with Trace ID

All API responses include trace IDs for correlation:

```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {...},
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

### Log Correlation

All log messages include trace and span IDs:

```
[2024-01-01T12:00:00Z] POST /api/users 201 45.2ms 127.0.0.1 trace_id=4bf92f3577b34da6a3ce929d0e0e4736 span_id=00f067aa0ba902b7
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
├── tracing/
│   └── tracing.go         # OpenTelemetry tracing setup
└── utils/
    └── response.go        # Response utilities
```
