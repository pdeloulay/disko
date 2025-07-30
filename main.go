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
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		log.Printf("[Health] Health check from IP: %s", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

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

	// Private board route with JWT enforcement (for board owners only)
	router.GET("/board/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		startTime := time.Now()
		boardID := c.Param("id")
		userAgent := c.GetHeader("User-Agent")
		referer := c.GetHeader("Referer")
		acceptLanguage := c.GetHeader("Accept-Language")

		log.Printf("[Template] Private Board route accessed - BoardID: %s, IP: %s, UserAgent: %s, Referer: %s, AcceptLanguage: %s",
			boardID, c.ClientIP(), userAgent, referer, acceptLanguage)

		// Get authenticated user ID (required by AuthMiddleware)
		userID, err := middleware.GetUserID(c)
		if err != nil {
			log.Printf("[Template] Private Board route - Auth error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			return
		}

		log.Printf("[Template] Private Board route - User authenticated: %s", userID)

		// Log environment variables for debugging
		clerkKey := os.Getenv("CLERK_PUBLISHABLE_KEY")
		clerkApiUrl := os.Getenv("CLERK_FRONTEND_API_URL")
		log.Printf("[Template] Private Board environment - ClerkKey: %s, ClerkApiUrl: %s",
			clerkKey != "", clerkApiUrl != "")

		// Check if user owns this board
		collection := models.GetCollection(models.BoardsCollection)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": boardID, "user_id": userID}
		var board models.Board
		if err := collection.FindOne(ctx, filter).Decode(&board); err != nil {
			log.Printf("[Template] Private Board route - User does not own board: %s, BoardID: %s, Error: %v", userID, boardID, err)
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"title":   "Board Not Found - Disko",
				"message": "This board does not exist or you don't have permission to access it.",
			})
			return
		}

		log.Printf("[Template] Private Board route - User owns board: %s, BoardID: %s, PublicLink: %s", userID, boardID, board.PublicLink)

		// Get app version
		version := getAppVersion()

		c.HTML(http.StatusOK, "board.html", gin.H{
			"title":               "Board - Disko",
			"publicLink":          board.PublicLink,
			"isPublic":            false, // Always false for private route
			"boardID":             boardID,
			"isOwner":             true, // User is always owner in authenticated route
			"clerkPublishableKey": clerkKey,
			"clerkFrontendApiUrl": clerkApiUrl,
			"version":             version,
		})

		duration := time.Since(startTime)
		log.Printf("[Template] Private Board rendered successfully - BoardID: %s, UserID: %s, Duration: %v, IP: %s",
			boardID, userID, duration, c.ClientIP())
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
		api.GET("/ping", func(c *gin.Context) {
			log.Printf("[API] Health check from IP: %s", c.ClientIP())
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		// Public board access endpoint
		api.GET("/boards/:id/public", func(c *gin.Context) {
			boardID := c.Param("id")
			log.Printf("[API] GetPublicBoard called - BoardID: %s, IP: %s, UserAgent: %s", boardID, c.ClientIP(), c.GetHeader("User-Agent"))
			handlers.GetPublicBoard(c)
		})
		api.GET("/boards/:id/ideas/public", func(c *gin.Context) {
			boardID := c.Param("id")
			log.Printf("[API] GetPublicBoardIdeas called - BoardID: %s, IP: %s, UserAgent: %s", boardID, c.ClientIP(), c.GetHeader("User-Agent"))
			handlers.GetPublicBoardIdeas(c)
		})
		api.GET("/boards/:id/release/public", func(c *gin.Context) {
			boardID := c.Param("id")
			log.Printf("[API] GetReleasedIdeas (public) called - BoardID: %s, IP: %s, UserAgent: %s", boardID, c.ClientIP(), c.GetHeader("User-Agent"))
			c.Header("X-Public-Access", "true")
			handlers.GetReleasedIdeas(c)
		})

		// Public feedback endpoints
		api.POST("/ideas/:id/thumbsup", func(c *gin.Context) {
			ideaID := c.Param("id")
			log.Printf("[API] AddThumbsUp called - IdeaID: %s, IP: %s, UserAgent: %s", ideaID, c.ClientIP(), c.GetHeader("User-Agent"))
			handlers.AddThumbsUp(c)
		})
		api.POST("/ideas/:id/emoji", func(c *gin.Context) {
			ideaID := c.Param("id")
			log.Printf("[API] AddEmojiReaction called - IdeaID: %s, IP: %s, UserAgent: %s", ideaID, c.ClientIP(), c.GetHeader("User-Agent"))
			handlers.AddEmojiReaction(c)
		})

		// WebSocket endpoint for real-time updates
		api.GET("/ws/boards/:boardId", func(c *gin.Context) {
			boardID := c.Param("boardId")
			log.Printf("[WebSocket] HandleWebSocket called - BoardID: %s, IP: %s, UserAgent: %s", boardID, c.ClientIP(), c.GetHeader("User-Agent"))
			utils.HandleWebSocket(c)
		})

		// Protected endpoints (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User info endpoint
			protected.GET("/user", func(c *gin.Context) {
				userID, err := middleware.GetUserID(c)
				log.Printf("[API] GetUserInfo called - IP: %s, UserAgent: %s", c.ClientIP(), c.GetHeader("User-Agent"))
				if err != nil {
					log.Printf("[API] GetUserInfo failed - Error: %v, IP: %s", err, c.ClientIP())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": gin.H{
							"code":    "INTERNAL_ERROR",
							"message": "Failed to get user ID",
						},
					})
					return
				}

				sessionID, _ := middleware.GetSessionID(c)
				log.Printf("[API] GetUserInfo success - UserID: %s, SessionID: %s, IP: %s", userID, sessionID, c.ClientIP())

				c.JSON(http.StatusOK, gin.H{
					"userID":    userID,
					"sessionID": sessionID,
				})
			})

			// Test protected endpoint
			protected.GET("/protected", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] TestProtected called - UserID: %s, IP: %s, UserAgent: %s", userID, c.ClientIP(), c.GetHeader("User-Agent"))
				c.JSON(http.StatusOK, gin.H{
					"message": "This is a protected endpoint",
					"userID":  userID,
				})
			})

			// Board management endpoints
			protected.POST("/boards", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] CreateBoard called - UserID: %s, IP: %s, UserAgent: %s", userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.CreateBoard(c)
			})
			protected.GET("/boards", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] GetBoards called - UserID: %s, IP: %s, UserAgent: %s", userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.GetBoards(c)
			})
			protected.GET("/boards/:id", func(c *gin.Context) {
				boardID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] GetBoard called - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.GetBoard(c)
			})
			protected.PUT("/boards/:id", func(c *gin.Context) {
				boardID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] UpdateBoard called - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.UpdateBoard(c)
			})
			protected.DELETE("/boards/:id", func(c *gin.Context) {
				boardID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] DeleteBoard called - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.DeleteBoard(c)
			})

			// Idea management endpoints
			protected.POST("/boards/:id/ideas", func(c *gin.Context) {
				boardID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] CreateIdea called - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.CreateIdea(c)
			})
			protected.GET("/boards/:id/ideas", func(c *gin.Context) {
				boardID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] GetBoardIdeas called - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.GetBoardIdeas(c)
			})
			protected.GET("/boards/:id/search", func(c *gin.Context) {
				boardID := c.Param("id")
				query := c.Query("q")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] SearchBoardIdeas called - BoardID: %s, Query: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, query, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.SearchBoardIdeas(c)
			})
			protected.GET("/boards/:id/release", func(c *gin.Context) {
				boardID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] GetReleasedIdeas (protected) called - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s", boardID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.GetReleasedIdeas(c)
			})
			protected.PUT("/ideas/:id", func(c *gin.Context) {
				ideaID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] UpdateIdea called - IdeaID: %s, UserID: %s, IP: %s, UserAgent: %s", ideaID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.UpdateIdea(c)
			})
			protected.DELETE("/ideas/:id", func(c *gin.Context) {
				ideaID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] DeleteIdea called - IdeaID: %s, UserID: %s, IP: %s, UserAgent: %s", ideaID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.DeleteIdea(c)
			})
			protected.PUT("/ideas/:id/position", func(c *gin.Context) {
				ideaID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] UpdateIdeaPosition called - IdeaID: %s, UserID: %s, IP: %s, UserAgent: %s", ideaID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.UpdateIdeaPosition(c)
			})
			protected.PUT("/ideas/:id/status", func(c *gin.Context) {
				ideaID := c.Param("id")
				userID, _ := middleware.GetUserID(c)
				log.Printf("[API] UpdateIdeaStatus called - IdeaID: %s, UserID: %s, IP: %s, UserAgent: %s", ideaID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
				handlers.UpdateIdeaStatus(c)
			})
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
