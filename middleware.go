// middleware.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// apiKeyAuthMiddleware fuerza que el header X-API-Key coincida con la variable de entorno API_KEY
func apiKeyAuthMiddleware() gin.HandlerFunc {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is not set")
	}
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			return
		}
		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}
		c.Next()
	}
}
