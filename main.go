package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware for better API compatibility
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Apply API key authentication
	router.Use(apiKeyAuthMiddleware())

	// Health check endpoint (without auth)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "lider-api"})
	})

	// API routes
	router.GET("/productos", handleSearch)
	router.GET("/suggestions", handleSuggestions)
	router.GET("/promotions", handlePromotions)
	router.GET("/categories", handleCategories)

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func handleSearch(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "se requiere parámetro 'q'",
			"example": "/productos?q=leche",
		})
		return
	}
	prods, err := fetchProducts(q)
	if err != nil {
		log.Printf("Error fetching products for query '%s': %v", q, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"query":    q,
		"count":    len(prods),
		"products": prods,
	})
}

func handleSuggestions(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "se requiere parámetro 'term'",
			"example": "/suggestions?term=lec",
		})
		return
	}
	suggestions, err := fetchSuggestions(term)
	if err != nil {
		log.Printf("Error fetching suggestions for term '%s': %v", term, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"term":        term,
		"count":       len(suggestions),
		"suggestions": suggestions,
	})
}

func handlePromotions(c *gin.Context) {
	promo := c.Query("type")
	if promo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "se requiere parámetro 'type'",
			"example": "/promotions?type=descuentos",
		})
		return
	}
	prods, err := fetchPromotions(promo)
	if err != nil {
		log.Printf("Error fetching promotions for type '%s': %v", promo, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"type":     promo,
		"count":    len(prods),
		"products": prods,
	})
}

func handleCategories(c *gin.Context) {
	cat := c.Query("id")
	if cat == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "se requiere parámetro 'id'",
			"example": "/categories?id=123",
		})
		return
	}
	prods, err := fetchCategory(cat)
	if err != nil {
		log.Printf("Error fetching category for id '%s': %v", cat, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"category_id": cat,
		"count":       len(prods),
		"products":    prods,
	})
}
