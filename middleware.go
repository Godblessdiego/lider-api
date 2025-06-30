package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// apiKeyAuthMiddleware fuerza que haya un header X-API-Key igual a tu API_KEY
func apiKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") != os.Getenv("API_KEY") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key inválida"})
			return
		}
		c.Next()
	}
}
