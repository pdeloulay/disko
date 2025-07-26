package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// Set up test environment
	os.Setenv("CLERK_SECRET_KEY", "test_secret_key")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth middleware to a test route
	router.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		userID, _ := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	t.Run("Invalid Authorization Header Format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})
}

func TestOptionalAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/optional", OptionalAuthMiddleware(), func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"authenticated": false})
		} else {
			c.JSON(http.StatusOK, gin.H{"authenticated": true, "userID": userID})
		}
	})

	t.Run("No Authorization Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/optional", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"authenticated":false`)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"authenticated":false`)
	})
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("User ID Exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", "test_user_123")

		userID, err := GetUserID(c)

		assert.NoError(t, err)
		assert.Equal(t, "test_user_123", userID)
	})

	t.Run("User ID Missing", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID, err := GetUserID(c)

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "user ID not found in context")
	})

	t.Run("User ID Wrong Type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", 123) // Set as int instead of string

		userID, err := GetUserID(c)

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "user ID is not a string")
	})
}

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Authenticated User", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userID", "test_user_123")

		result := RequireAuth(c)

		assert.True(t, result)
	})

	t.Run("Unauthenticated User", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		result := RequireAuth(c)

		assert.False(t, result)
	})
}
