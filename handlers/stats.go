package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"disko-backend/middleware"
	"disko-backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// GetStats returns statistics for the authenticated user
func GetStats(c *gin.Context) {
	startTime := time.Now()

	// Get authenticated user ID
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.Printf("[Stats] Failed to get user ID: %v - IP: %s", err, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	log.Printf("[Stats] Starting stats collection for user: %s - IP: %s", userID, c.ClientIP())

	// Get database connection
	if models.DB == nil {
		log.Printf("[Stats] Database connection failed - IP: %s", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Database connection failed",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize stats
	stats := gin.H{
		"boards":   0,
		"ideas":    0,
		"feedback": 0,
	}

	// Count boards for this user
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardsCount, err := boardsCollection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		log.Printf("[Stats] Error counting boards for user %s: %v - IP: %s", userID, err, c.ClientIP())
	} else {
		stats["boards"] = boardsCount
		log.Printf("[Stats] Boards count for user %s: %d - IP: %s", userID, boardsCount, c.ClientIP())
	}

	// Count ideas for this user's boards
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ideasCount, err := ideasCollection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		log.Printf("[Stats] Error counting ideas for user %s: %v - IP: %s", userID, err, c.ClientIP())
	} else {
		stats["ideas"] = ideasCount
		log.Printf("[Stats] Ideas count for user %s: %d - IP: %s", userID, ideasCount, c.ClientIP())
	}

	// Count feedback (thumbs up and emoji reactions) for this user's ideas
	feedbackCount := 0

	// Get all ideas for this user and count reactions manually
	cursor, err := ideasCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		log.Printf("[Stats] Error finding ideas for feedback count for user %s: %v - IP: %s", userID, err, c.ClientIP())
	} else {
		defer cursor.Close(ctx)

		var ideas []bson.M
		if err := cursor.All(ctx, &ideas); err != nil {
			log.Printf("[Stats] Error reading ideas for feedback count for user %s: %v - IP: %s", userID, err, c.ClientIP())
		} else {
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

	stats["feedback"] = feedbackCount
	log.Printf("[Stats] Feedback count for user %s: %d - IP: %s", userID, feedbackCount, c.ClientIP())

	duration := time.Since(startTime)
	log.Printf("[Stats] Stats collected successfully for user %s - Duration: %v, IP: %s", userID, duration, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"timestamp": time.Now().UTC(),
		"userID":    userID,
	})
}
