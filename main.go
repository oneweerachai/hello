package main

import (
	"log"
	"user-api/config"
	"user-api/handlers"
	"user-api/middleware"
	"user-api/repository"
	"user-api/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize repository
	userRepo := repository.NewInMemoryUserRepository()

	// Initialize service
	userService := services.NewUserService(userRepo)

	// Initialize handler
	userHandler := handlers.NewUserHandler(userService)

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", userHandler.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// User routes
		users := api.Group("/users")
		users.Use(middleware.JSONContentType()) // Apply JSON content type middleware to user routes
		{
			users.POST("", userHandler.CreateUser)   // POST /api/users
			users.GET("", userHandler.GetUsers)      // GET /api/users
			users.GET("/:id", userHandler.GetUser)   // GET /api/users/:id
		}
	}

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("API endpoint: http://localhost:%s/api/users", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
