module lider-api

go 1.24.4
require (
    github.com/gin-gonic/gin v1.9.0
)


// File: .env.example
# Define aquí tu llave secreta para la API
API_KEY=tu_llave_secreta_aqui


// File: middleware.go
package main

import (
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
)

// apiKeyAuthMiddleware valida la llave en el header X-API-Key
func apiKeyAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.GetHeader("X-API-Key")
        if key == "" || key != os.Getenv("API_KEY") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key inválida"})
            return
        }
        c.Next()
    }
}
