package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates Clerk JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract the token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Verify the JWT token with Clerk
		claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired token",
					"details": err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("userID", claims.Subject)
		c.Set("sessionID", claims.SessionID)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware validates Clerk JWT tokens but doesn't require them
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without setting user context
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without setting user context
			c.Next()
			return
		}

		token := tokenParts[1]

		// Try to verify the JWT token
		claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			// Invalid token, continue without setting user context
			c.Next()
			return
		}

		// Store user information in context if valid
		c.Set("userID", claims.Subject)
		c.Set("sessionID", claims.SessionID)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", fmt.Errorf("user ID not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user ID is not a string")
	}

	return userIDStr, nil
}

// GetSessionID extracts the session ID from the Gin context
func GetSessionID(c *gin.Context) (string, error) {
	sessionID, exists := c.Get("sessionID")
	if !exists {
		return "", fmt.Errorf("session ID not found in context")
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		return "", fmt.Errorf("session ID is not a string")
	}

	return sessionIDStr, nil
}

// InitializeClerk initializes the Clerk client with the secret key
func InitializeClerk() error {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		return fmt.Errorf("CLERK_SECRET_KEY environment variable is required")
	}

	clerk.SetKey(secretKey)
	return nil
}

// RequireAuth is a helper function to check if user is authenticated
func RequireAuth(c *gin.Context) bool {
	_, err := GetUserID(c)
	return err == nil
}
