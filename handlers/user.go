package handlers

import (
	"log"
	"net/http"

	"disko-backend/middleware"

	"github.com/gin-gonic/gin"
)

// GetUserInfo handles GET /api/user
func GetUserInfo(c *gin.Context) {
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
}

// TestProtected handles GET /api/protected
func TestProtected(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	log.Printf("[API] TestProtected called - UserID: %s, IP: %s, UserAgent: %s", userID, c.ClientIP(), c.GetHeader("User-Agent"))
	c.JSON(http.StatusOK, gin.H{
		"message": "This is a protected endpoint",
		"userID":  userID,
	})
}
