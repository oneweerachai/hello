package handlers

import (
	"net/http"
	"strings"
	"user-api/models"
	"user-api/services"
	"user-api/tracing"
	"user-api/utils"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *services.UserService
	tracer      trace.Tracer
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		tracer:      tracing.GetTracer("user-api/handlers"),
	}
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx, span := tracing.StartSpan(c.Request.Context(), h.tracer, "CreateUser")
	defer span.End()

	// Update context in gin
	c.Request = c.Request.WithContext(ctx)

	var req models.CreateUserRequest

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("validation_error"))
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Add request attributes to span
	tracing.AddSpanAttributes(span,
		tracing.AttrUserEmail.String(req.Email),
		attribute.String("user.first_name", req.FirstName),
		attribute.String("user.last_name", req.LastName),
	)

	// Trim whitespace from string fields
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)
	req.DateOfBirth = strings.TrimSpace(req.DateOfBirth)

	// Create user through service
	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		tracing.RecordError(span, err)

		if strings.Contains(err.Error(), "already exists") {
			tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("conflict_error"))
			utils.ConflictResponse(c, "User creation failed", err)
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("validation_error"))
			utils.ValidationErrorResponse(c, err)
			return
		}
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("internal_error"))
		utils.InternalServerErrorResponse(c, "Failed to create user", err)
		return
	}

	// Add success attributes
	tracing.AddSpanAttributes(span,
		tracing.AttrUserID.String(user.ID),
		attribute.String("operation.result", "success"),
	)

	tracing.AddSpanEvent(span, "user.created",
		tracing.AttrUserID.String(user.ID),
		tracing.AttrUserEmail.String(user.Email),
	)

	// Return success response
	utils.CreatedResponse(c, "User created successfully", user.ToResponse())
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	ctx, span := tracing.StartSpan(c.Request.Context(), h.tracer, "GetUser")
	defer span.End()

	// Update context in gin
	c.Request = c.Request.WithContext(ctx)

	id := c.Param("id")

	// Add request attributes
	tracing.AddSpanAttributes(span, tracing.AttrUserID.String(id))

	user, err := h.userService.GetUserByID(ctx, id)
	if err != nil {
		tracing.RecordError(span, err)

		if strings.Contains(err.Error(), "not found") {
			tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("not_found"))
			utils.NotFoundResponse(c, "User not found")
			return
		}
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("internal_error"))
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	// Add success attributes
	tracing.AddSpanAttributes(span,
		tracing.AttrUserEmail.String(user.Email),
		attribute.String("operation.result", "success"),
	)

	utils.OKResponse(c, "User retrieved successfully", user.ToResponse())
}

// GetUsers handles GET /api/users
func (h *UserHandler) GetUsers(c *gin.Context) {
	ctx, span := tracing.StartSpan(c.Request.Context(), h.tracer, "GetUsers")
	defer span.End()

	// Update context in gin
	c.Request = c.Request.WithContext(ctx)

	users, err := h.userService.GetAllUsers(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		tracing.AddSpanAttributes(span, tracing.AttrErrorType.String("internal_error"))
		utils.InternalServerErrorResponse(c, "Failed to get users", err)
		return
	}

	// Convert users to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	// Add success attributes
	tracing.AddSpanAttributes(span,
		attribute.Int("users.count", len(users)),
		attribute.String("operation.result", "success"),
	)

	utils.OKResponse(c, "Users retrieved successfully", userResponses)
}

// HealthCheck handles GET /health
func (h *UserHandler) HealthCheck(c *gin.Context) {
	ctx, span := tracing.StartSpan(c.Request.Context(), h.tracer, "HealthCheck")
	defer span.End()

	// Update context in gin
	c.Request = c.Request.WithContext(ctx)

	traceID := tracing.GetTraceID(ctx)

	response := gin.H{
		"status":    "success",
		"message":   "Server is running",
		"timestamp": gin.H{"now": "2024-01-01T00:00:00Z"}, // You can use time.Now() here
	}

	if traceID != "" {
		response["trace_id"] = traceID
	}

	tracing.AddSpanAttributes(span, attribute.String("operation.result", "success"))

	c.JSON(http.StatusOK, response)
}
