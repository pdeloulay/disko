package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck handles GET /health
func HealthCheck(c *gin.Context) {
	log.Printf("[Health] Health check from IP: %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// Ping handles GET /api/ping
func Ping(c *gin.Context) {
	log.Printf("[API] Health check from IP: %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
