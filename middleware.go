// middleware.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// apiKeyAuthMiddleware fuerza que el header X-API-Key coincida con la variable de entorno API_KEY
// Excluye el endpoint /health de la autenticación
func apiKeyAuthMiddleware() gin.HandlerFunc {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is not set")
	}

	return func(c *gin.Context) {
		// Skip authentication for health check endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		startTime := time.Now()
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		key := c.GetHeader("X-API-Key")
		if key == "" {
			log.Printf("AUTH FAILED: Missing API key - IP: %s, UA: %s, Path: %s",
				clientIP, userAgent, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required",
				"hint":  "Include X-API-Key header with your request",
			})
			return
		}

		if key != apiKey {
			log.Printf("AUTH FAILED: Invalid API key - IP: %s, UA: %s, Path: %s",
				clientIP, userAgent, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid API key",
			})
			return
		}

		// Log successful authentication
		log.Printf("AUTH SUCCESS: IP: %s, Path: %s, Time: %v",
			clientIP, c.Request.URL.Path, time.Since(startTime))

		c.Next()
	}
}
