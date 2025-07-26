package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"disko-backend/handlers"
	"disko-backend/middleware"
	"disko-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize MongoDB connection
	if err := models.ConnectDatabase(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := models.DisconnectDatabase(); err != nil {
			log.Println("Error disconnecting from MongoDB:", err)
		}
	}()

	// Initialize Clerk authentication
	if err := middleware.InitializeClerk(); err != nil {
		log.Fatal("Failed to initialize Clerk:", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	// Web routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":               "Disko",
			"clerkPublishableKey": os.Getenv("CLERK_PUBLISHABLE_KEY"),
		})
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":               "Dashboard - Disko",
			"clerkPublishableKey": os.Getenv("CLERK_PUBLISHABLE_KEY"),
		})
	})

	router.GET("/board/:publicLink", func(c *gin.Context) {
		publicLink := c.Param("publicLink")
		c.HTML(http.StatusOK, "board.html", gin.H{
			"title":               "Board - Disko",
			"publicLink":          publicLink,
			"clerkPublishableKey": os.Getenv("CLERK_PUBLISHABLE_KEY"),
		})
	})

	// API routes group
	api := router.Group("/api")
	{
		// Public endpoints
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		// Protected endpoints (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User info endpoint
			protected.GET("/user", func(c *gin.Context) {
				userID, err := middleware.GetUserID(c)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": gin.H{
							"code":    "INTERNAL_ERROR",
							"message": "Failed to get user ID",
						},
					})
					return
				}

				sessionID, _ := middleware.GetSessionID(c)

				c.JSON(http.StatusOK, gin.H{
					"userID":    userID,
					"sessionID": sessionID,
				})
			})

			// Test protected endpoint
			protected.GET("/protected", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				c.JSON(http.StatusOK, gin.H{
					"message": "This is a protected endpoint",
					"userID":  userID,
				})
			})

			// Board management endpoints
			protected.POST("/boards", handlers.CreateBoard)
			protected.GET("/boards", handlers.GetBoards)
			protected.PUT("/boards/:id", handlers.UpdateBoard)
			protected.DELETE("/boards/:id", handlers.DeleteBoard)

			// Idea management endpoints
			protected.POST("/boards/:id/ideas", handlers.CreateIdea)
			protected.GET("/boards/:id/ideas", handlers.GetBoardIdeas)
			protected.PUT("/ideas/:id", handlers.UpdateIdea)
			protected.DELETE("/ideas/:id", handlers.DeleteIdea)
			protected.PUT("/ideas/:id/position", handlers.UpdateIdeaPosition)
			protected.PUT("/ideas/:id/status", handlers.UpdateIdeaStatus)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
