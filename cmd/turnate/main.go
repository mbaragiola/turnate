package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	
	"turnate/internal/config"
	"turnate/internal/database"
	"turnate/internal/handlers"
	"turnate/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run auto-migrations
	if err := database.AutoMigrateModels(database.GetDB()); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Set up Gin router
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ContentSecurityMiddleware())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.InputValidationMiddleware())
	r.Use(middleware.ValidateContentType())
	r.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Serve static files
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Create handlers
	authHandler := handlers.NewAuthHandler(cfg)
	userHandler := handlers.NewUserHandler()
	channelHandler := handlers.NewChannelHandler()
	messageHandler := handlers.NewMessageHandler()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "turnate"})
	})

	// Serve the main app
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{"title": "Turnate"})
	})

	// API routes
	api := r.Group("/api/v1")
	api.Use(middleware.APIRateLimitMiddleware())
	{
		// Public auth routes
		auth := api.Group("/auth")
		auth.Use(middleware.AuthRateLimitMiddleware())
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("", userHandler.GetUsers)
				users.GET("/me", authHandler.Profile)
				users.GET("/:id", userHandler.GetUserByID)
				users.PATCH("/:id", userHandler.UpdateUser)
			}

			// Channel routes
			channels := protected.Group("/channels")
			{
				channels.POST("", channelHandler.CreateChannel)
				channels.GET("", channelHandler.GetChannels)
				channels.GET("/:id", channelHandler.GetChannel)
				channels.POST("/:id/join", channelHandler.JoinChannel)
				channels.DELETE("/:id/leave", channelHandler.LeaveChannel)
				channels.GET("/:id/members", channelHandler.GetChannelMembers)
				
				// Message routes (using :id instead of :channelId to avoid conflict)
				channels.POST("/:id/messages", messageHandler.CreateMessage)
				channels.GET("/:id/messages", messageHandler.GetMessages)
				channels.GET("/:id/messages/:threadId/replies", messageHandler.GetThreadMessages)
			}

			// Message routes
			messages := protected.Group("/messages")
			{
				messages.GET("/recent", messageHandler.GetRecentMessages)
			}
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/users", userHandler.GetUsers)
			admin.GET("/channels", channelHandler.GetChannels)
		}
	}

	log.Printf("🚀 Turnate server starting on port %s", cfg.Port)
	log.Printf("🌐 Visit http://localhost:%s to get started!", cfg.Port)
	
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}