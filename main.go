package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"disko-backend/handlers"
	"disko-backend/middleware"
	"disko-backend/models"
	"disko-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Simple in-memory rate limiting (for production, use Redis)
var rateLimitStore = make(map[string]time.Time)

func isRateLimited(key string, duration time.Duration) bool {
	if lastRequest, exists := rateLimitStore[key]; exists {
		if time.Since(lastRequest) < duration {
			return true
		}
	}
	return false
}

func setRateLimit(key string, duration time.Duration) {
	rateLimitStore[key] = time.Now()

	// Clean up old entries (simple cleanup)
	go func() {
		time.Sleep(duration * 2)
		delete(rateLimitStore, key)
	}()
}

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// getAppVersion reads the version from the .version file
func getAppVersion() string {
	versionBytes, err := os.ReadFile("static/.version")
	if err != nil {
		log.Printf("[Version] Error reading version file: %v", err)
		return "0.0.0"
	}
	version := string(versionBytes)
	version = strings.TrimSpace(version)
	log.Printf("[Version] App version: %s", version)
	return version
}

// getPublicStats returns public statistics for the landing page
func getPublicStats() gin.H {
	// Get database connection
	if models.DB == nil {
		log.Printf("[Stats] Database connection failed")
		return gin.H{"boards": 0, "ideas": 0, "feedback": 0}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count all boards
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardsCount, err := boardsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("[Stats] Error counting boards: %v", err)
		boardsCount = 0
	}

	// Count all ideas
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ideasCount, err := ideasCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("[Stats] Error counting ideas: %v", err)
		ideasCount = 0
	}

	// Count feedback (thumbs up and emoji reactions)
	feedbackCount := 0
	cursor, err := ideasCollection.Find(ctx, bson.M{})
	if err == nil {
		defer cursor.Close(ctx)
		var ideas []bson.M
		if err := cursor.All(ctx, &ideas); err == nil {
			for _, idea := range ideas {
				// Count thumbs up
				if thumbsUp, exists := idea["thumbsUp"]; exists {
					if thumbsUpInt, ok := thumbsUp.(int32); ok {
						feedbackCount += int(thumbsUpInt)
					} else if thumbsUpInt, ok := thumbsUp.(int64); ok {
						feedbackCount += int(thumbsUpInt)
					} else if thumbsUpInt, ok := thumbsUp.(int); ok {
						feedbackCount += thumbsUpInt
					}
				}

				// Count emoji reactions
				if emojiReactions, exists := idea["emojiReactions"]; exists {
					if reactionsArray, ok := emojiReactions.([]interface{}); ok {
						feedbackCount += len(reactionsArray)
					}
				}
			}
		}
	}

	log.Printf("[Stats] Landing page stats - Boards: %d, Ideas: %d, Feedback: %d", boardsCount, ideasCount, feedbackCount)
	return gin.H{"boards": boardsCount, "ideas": ideasCount, "feedback": feedbackCount}
}

func main() {
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

	// Initialize notification service
	utils.InitNotificationService()

	// Initialize WebSocket manager
	utils.InitWebSocketManager()

	// Initialize Gin router
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	// Add custom request logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request is processed
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[GIN] %s | %3d | %13v | %15s | %-7s %s | %s",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
			userAgent,
		)
	})

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// Test modal endpoint
	router.GET("/test-modal", func(c *gin.Context) {
		log.Printf("[Test] Modal test page accessed - IP: %s", c.ClientIP())
		c.File("test_modal.html")
	})

	// Web routes
	router.GET("/", func(c *gin.Context) {
		log.Printf("[Template] Rendering index.html for IP: %s", c.ClientIP())

		// Get public stats for the landing page
		stats := getPublicStats()

		// Get app version
		version := getAppVersion()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":               "Disko",
			"clerkPublishableKey": os.Getenv("CLERK_PUBLISHABLE_KEY"),
			"clerkFrontendApiUrl": os.Getenv("CLERK_FRONTEND_API_URL"),
			"stats":               stats,
			"version":             version,
		})
	})

	router.GET("/dashboard", func(c *gin.Context) {
		startTime := time.Now()
		userAgent := c.GetHeader("User-Agent")
		referer := c.GetHeader("Referer")
		acceptLanguage := c.GetHeader("Accept-Language")

		log.Printf("[Template] Dashboard route accessed - IP: %s, UserAgent: %s, Referer: %s, AcceptLanguage: %s",
			c.ClientIP(), userAgent, referer, acceptLanguage)

		// Log environment variables for debugging
		clerkKey := os.Getenv("CLERK_PUBLISHABLE_KEY")
		clerkApiUrl := os.Getenv("CLERK_FRONTEND_API_URL")
		log.Printf("[Template] Dashboard environment - ClerkKey: %s, ClerkApiUrl: %s",
			clerkKey != "", clerkApiUrl != "")

		// Get app version
		version := getAppVersion()

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title":               "Dashboard - Disko",
			"clerkPublishableKey": clerkKey,
			"clerkFrontendApiUrl": clerkApiUrl,
			"version":             version,
		})

		duration := time.Since(startTime)
		log.Printf("[Template] Dashboard rendered successfully - Duration: %v, IP: %s", duration, c.ClientIP())
	})

	// Board route - authentication handled by frontend
	router.GET("/board/:id", func(c *gin.Context) {
		startTime := time.Now()
		boardID := c.Param("id")
		userAgent := c.GetHeader("User-Agent")
		referer := c.GetHeader("Referer")
		acceptLanguage := c.GetHeader("Accept-Language")

		log.Printf("[Template] Board route accessed - BoardID: %s, IP: %s, UserAgent: %s, Referer: %s, AcceptLanguage: %s",
			boardID, c.ClientIP(), userAgent, referer, acceptLanguage)

		// Log environment variables for debugging
		clerkKey := os.Getenv("CLERK_PUBLISHABLE_KEY")
		clerkApiUrl := os.Getenv("CLERK_FRONTEND_API_URL")
		log.Printf("[Template] Board environment - ClerkKey: %s, ClerkApiUrl: %s",
			clerkKey != "", clerkApiUrl != "")

		// Get app version
		version := getAppVersion()

		c.HTML(http.StatusOK, "board.html", gin.H{
			"title":               "Board - Disko",
			"boardID":             boardID,
			"clerkPublishableKey": clerkKey,
			"clerkFrontendApiUrl": clerkApiUrl,
			"version":             version,
		})

		duration := time.Since(startTime)
		log.Printf("[Template] Board rendered successfully - BoardID: %s, Duration: %v, IP: %s",
			boardID, duration, c.ClientIP())
	})

	// Public board route with rate limiting (for public access)
	router.GET("/public/:publicLink", func(c *gin.Context) {
		startTime := time.Now()
		publicLink := c.Param("publicLink")
		userAgent := c.GetHeader("User-Agent")
		referer := c.GetHeader("Referer")
		acceptLanguage := c.GetHeader("Accept-Language")
		clientIP := c.ClientIP()

		log.Printf("[Template] Public Board route accessed - PublicLink: %s, IP: %s, UserAgent: %s, Referer: %s, AcceptLanguage: %s",
			publicLink, clientIP, userAgent, referer, acceptLanguage)

		// Rate limiting for public board access
		rateLimitKey := "public_board_" + publicLink + "_" + clientIP
		if isRateLimited(rateLimitKey, 10*time.Second) {
			log.Printf("[Template] Public Board route - Rate limited: %s, IP: %s", publicLink, clientIP)
			c.HTML(http.StatusTooManyRequests, "error.html", gin.H{
				"title":   "Rate Limited - Disko",
				"message": "Too many requests. Please try again in a few seconds.",
			})
			return
		}
		setRateLimit(rateLimitKey, 10*time.Second)

		// Log environment variables for debugging
		clerkKey := os.Getenv("CLERK_PUBLISHABLE_KEY")
		clerkApiUrl := os.Getenv("CLERK_FRONTEND_API_URL")
		log.Printf("[Template] Public Board environment - ClerkKey: %s, ClerkApiUrl: %s",
			clerkKey != "", clerkApiUrl != "")

		// Check if board exists and is public
		collection := models.GetCollection(models.BoardsCollection)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"public_link": publicLink, "is_public": true}
		var board models.Board
		if err := collection.FindOne(ctx, filter).Decode(&board); err != nil {
			log.Printf("[Template] Public Board route - Board not found or not public: %s", publicLink)
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"title":   "Board Not Found - Disko",
				"message": "This board does not exist or is not publicly accessible.",
			})
			return
		}

		log.Printf("[Template] Public Board route - Board is public: %s", publicLink)

		// Get app version
		version := getAppVersion()

		c.HTML(http.StatusOK, "board.html", gin.H{
			"title":               "Board - Disko",
			"publicLink":          publicLink,
			"isPublic":            true, // Always true for public route
			"boardID":             "",   // No board ID for public view
			"clerkPublishableKey": clerkKey,
			"clerkFrontendApiUrl": clerkApiUrl,
			"version":             version,
		})

		duration := time.Since(startTime)
		log.Printf("[Template] Public Board rendered successfully - PublicLink: %s, Duration: %v, IP: %s",
			publicLink, duration, clientIP)
	})

	// API routes group
	api := router.Group("/api")
	{
		// Public endpoints
		api.GET("/ping", handlers.Ping)

		// Public board access endpoint
		api.GET("/boards/:id/public", handlers.GetPublicBoard)
		api.GET("/boards/:id/ideas/public", handlers.GetPublicBoardIdeas)
		api.GET("/boards/:id/release/public", handlers.GetPublicReleasedIdeas)

		// Public feedback endpoints
		api.POST("/ideas/:id/thumbsup", handlers.AddThumbsUp)
		api.POST("/ideas/:id/emoji", handlers.AddEmojiReaction)

		// WebSocket endpoint for real-time updates
		api.GET("/ws/boards/:boardId", utils.HandleWebSocket)

		// Protected endpoints (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User info endpoint
			protected.GET("/user", handlers.GetUserInfo)

			// Test protected endpoint
			protected.GET("/protected", handlers.TestProtected)

			// Board management endpoints
			protected.POST("/boards", handlers.CreateBoard)
			protected.GET("/boards", handlers.GetBoards)
			protected.GET("/boards/:id", handlers.GetBoard)
			protected.PUT("/boards/:id", handlers.UpdateBoard)

			protected.DELETE("/boards/:id", handlers.DeleteBoard)

			// Idea management endpoints
			protected.POST("/boards/:id/ideas", handlers.CreateIdea)
			protected.GET("/boards/:id/ideas", handlers.GetBoardIdeas)
			protected.GET("/boards/:id/search", handlers.SearchBoardIdeas)
			protected.GET("/boards/:id/release", handlers.GetReleasedIdeas)
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
