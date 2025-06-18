package handlers

import (
	"net/http"
	"strings"
	"user-api/models"
	"user-api/services"
	"user-api/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest

	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Trim whitespace from string fields
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)
	req.DateOfBirth = strings.TrimSpace(req.DateOfBirth)

	// Create user through service
	user, err := h.userService.CreateUser(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.ConflictResponse(c, "User creation failed", err)
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			utils.ValidationErrorResponse(c, err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create user", err)
		return
	}

	// Return success response
	utils.CreatedResponse(c, "User created successfully", user.ToResponse())
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.NotFoundResponse(c, "User not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	utils.OKResponse(c, "User retrieved successfully", user.ToResponse())
}

// GetUsers handles GET /api/users
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get users", err)
		return
	}

	// Convert users to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	utils.OKResponse(c, "Users retrieved successfully", userResponses)
}

// HealthCheck handles GET /health
func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Server is running",
		"timestamp": gin.H{"now": "2024-01-01T00:00:00Z"}, // You can use time.Now() here
	})
}
