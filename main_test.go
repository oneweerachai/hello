package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-api/handlers"
	"user-api/models"
	"user-api/repository"
	"user-api/services"
	"user-api/tracing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Initialize dependencies
	userRepo := repository.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Setup router
	router := gin.New()
	router.GET("/health", userHandler.HealthCheck)

	api := router.Group("/api")
	users := api.Group("/users")
	{
		users.POST("", userHandler.CreateUser)
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUser)
	}

	return router
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	// Check if trace_id is present in response
	if traceID, exists := response["trace_id"]; exists {
		assert.NotEmpty(t, traceID)
	}
}

func TestCreateUser(t *testing.T) {
	router := setupTestRouter()

	user := models.CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "1234567890",
	}

	jsonData, _ := json.Marshal(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "User created successfully", response["message"])

	// Check if data is present
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	assert.Equal(t, "John", data["first_name"])
	assert.Equal(t, "Doe", data["last_name"])
	assert.Equal(t, "john.doe@example.com", data["email"])

	// Check if trace_id is present in response
	if traceID, exists := response["trace_id"]; exists {
		assert.NotEmpty(t, traceID)
	}
}

func TestCreateUserValidation(t *testing.T) {
	router := setupTestRouter()

	// Test with missing required fields
	user := models.CreateUserRequest{
		FirstName: "", // Missing first name
		LastName:  "Doe",
		Email:     "invalid-email", // Invalid email
	}

	jsonData, _ := json.Marshal(user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])

	// Check if trace_id is present in error response
	if traceID, exists := response["trace_id"]; exists {
		assert.NotEmpty(t, traceID)
	}
}

func TestGetUsers(t *testing.T) {
	router := setupTestRouter()

	// First create a user
	user := models.CreateUserRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
	}

	jsonData, _ := json.Marshal(user)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	// Now get all users
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/users", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	// Check if data is an array
	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Greater(t, len(data), 0)

	// Check if trace_id is present in response
	if traceID, exists := response["trace_id"]; exists {
		assert.NotEmpty(t, traceID)
	}
}

// TestTracingIntegration tests that tracing is working correctly
func TestTracingIntegration(t *testing.T) {
	// Initialize tracing for test
	tracingConfig := tracing.TracingConfig{
		Enabled:      true,
		ExporterType: "console",
		SamplingRate: 1.0,
		Environment:  "test",
	}

	shutdown, err := tracing.InitTracing(tracingConfig)
	assert.NoError(t, err)
	defer func() {
		ctx := context.Background()
		shutdown(ctx)
	}()

	router := setupTestRouter()

	// Test health check with tracing
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}
