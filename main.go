package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-api/config"
	"user-api/handlers"
	"user-api/middleware"
	"user-api/repository"
	"user-api/services"
	"user-api/tracing"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize tracing
	tracingShutdown, err := tracing.InitTracing(cfg.Tracing)
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracingShutdown(ctx); err != nil {
			log.Printf("Failed to shutdown tracing: %v", err)
		}
	}()

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

	// Add tracing middleware if enabled
	if cfg.Tracing.Enabled {
		router.Use(middleware.TracingMiddleware(tracing.ServiceName))
		router.Use(middleware.EnhancedTracingMiddleware())
	}

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
	log.Printf("Tracing enabled: %v", cfg.Tracing.Enabled)
	if cfg.Tracing.Enabled {
		log.Printf("Tracing exporter: %s", cfg.Tracing.ExporterType)
		log.Printf("Tracing sampling rate: %.2f", cfg.Tracing.SamplingRate)
	}
	log.Printf("Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("API endpoint: http://localhost:%s/api/users", cfg.Port)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	<-c
	log.Println("Shutting down server...")
}
