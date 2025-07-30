package middleware

import (
	"context"
	"fmt"
	"log"
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
		log.Printf("[Auth] AuthMiddleware called - Path: %s, Method: %s, IP: %s, UserAgent: %s", c.Request.URL.Path, c.Request.Method, c.ClientIP(), c.GetHeader("User-Agent"))

		if authHeader == "" {
			log.Printf("[Auth] AuthMiddleware failed - No authorization header, IP: %s", c.ClientIP())
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
			log.Printf("[Auth] AuthMiddleware failed - Invalid token format, IP: %s", c.ClientIP())
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
		log.Printf("[Auth] AuthMiddleware - Token received, length: %d, IP: %s", len(token), c.ClientIP())

		// Verify the JWT token with Clerk
		claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
			Token: token,
		})
		if err != nil {
			log.Printf("[Auth] AuthMiddleware failed - Token verification error: %v, IP: %s", err, c.ClientIP())
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

		log.Printf("[Auth] AuthMiddleware success - UserID: %s, SessionID: %s, IP: %s", claims.Subject, claims.SessionID, c.ClientIP())

		c.Next()
	}
}

// OptionalAuthMiddleware validates Clerk JWT tokens but doesn't require them
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		log.Printf("[Auth] OptionalAuthMiddleware called - Path: %s, Method: %s, IP: %s, UserAgent: %s", c.Request.URL.Path, c.Request.Method, c.ClientIP(), c.GetHeader("User-Agent"))

		if authHeader == "" {
			log.Printf("[Auth] OptionalAuthMiddleware - No auth header, continuing without auth, IP: %s", c.ClientIP())
			// No auth header, continue without setting user context
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Printf("[Auth] OptionalAuthMiddleware - Invalid token format, continuing without auth, IP: %s", c.ClientIP())
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
			log.Printf("[Auth] OptionalAuthMiddleware - Token verification failed: %v, continuing without auth, IP: %s", err, c.ClientIP())
			// Invalid token, continue without setting user context
			c.Next()
			return
		}

		// Store user information in context if valid
		c.Set("userID", claims.Subject)
		c.Set("sessionID", claims.SessionID)
		c.Set("claims", claims)

		log.Printf("[Auth] OptionalAuthMiddleware success - UserID: %s, SessionID: %s, IP: %s", claims.Subject, claims.SessionID, c.ClientIP())

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("[Auth] GetUserID failed - UserID not found in context, IP: %s", c.ClientIP())
		return "", fmt.Errorf("user ID not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("[Auth] GetUserID failed - UserID is not a string, IP: %s", c.ClientIP())
		return "", fmt.Errorf("user ID is not a string")
	}

	log.Printf("[Auth] GetUserID success - UserID: %s, IP: %s", userIDStr, c.ClientIP())
	return userIDStr, nil
}

// GetSessionID extracts the session ID from the Gin context
func GetSessionID(c *gin.Context) (string, error) {
	sessionID, exists := c.Get("sessionID")
	if !exists {
		log.Printf("[Auth] GetSessionID failed - SessionID not found in context, IP: %s", c.ClientIP())
		return "", fmt.Errorf("session ID not found in context")
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		log.Printf("[Auth] GetSessionID failed - SessionID is not a string, IP: %s", c.ClientIP())
		return "", fmt.Errorf("session ID is not a string")
	}

	log.Printf("[Auth] GetSessionID success - SessionID: %s, IP: %s", sessionIDStr, c.ClientIP())
	return sessionIDStr, nil
}

// InitializeClerk initializes the Clerk client with the secret key
func InitializeClerk() error {
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		log.Printf("[Auth] InitializeClerk failed - CLERK_SECRET_KEY not set")
		return fmt.Errorf("CLERK_SECRET_KEY environment variable is required")
	}

	log.Printf("[Auth] InitializeClerk success - Clerk client initialized")
	clerk.SetKey(secretKey)
	return nil
}

// RequireAuth is a helper function to check if user is authenticated
func RequireAuth(c *gin.Context) bool {
	_, err := GetUserID(c)
	isAuthenticated := err == nil
	log.Printf("[Auth] RequireAuth check - IsAuthenticated: %t, IP: %s", isAuthenticated, c.ClientIP())
	return isAuthenticated
}
