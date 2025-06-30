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
	router.GET("/product/:sku", handleProductDetail)
	router.GET("/product", handleProductDetail) // /product?sku=4522432 or /product?url=...

	log.Printf("Starting server on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET /health - Health check")
	log.Printf("  GET /productos?q=term - Search products")
	log.Printf("  GET /suggestions?term=partial - Get suggestions")
	log.Printf("  GET /promotions?type=promo - Get promotions")
	log.Printf("  GET /categories?id=cat_id - Get category products")
	log.Printf("  GET /product/:sku - Get product detail by SKU")
	log.Printf("  GET /product?sku=sku - Get product detail by SKU parameter")

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
	prods, err := fetchProductsAdvanced(q)
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
	suggestions, err := fetchSuggestionsAdvanced(term)
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
	prods, err := fetchPromotionsAdvanced(promo)
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
	prods, err := fetchCategoryAdvanced(cat)
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

func handleProductDetail(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		sku = c.Query("sku")
	}

	// También permitir URL completa
	if productURL := c.Query("url"); productURL != "" {
		sku = extractSKUFromURL(productURL)
	}

	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "se requiere parámetro 'sku' o 'url'",
			"example": "/product/4522432 o /product?sku=4522432 o /product?url=https://www.lider.cl/supermercado/product/sku/4522432/...",
		})
		return
	}

	detail, err := fetchProductDetailAdvanced(sku)
	if err != nil {
		log.Printf("Error fetching product detail for SKU '%s': %v", sku, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, detail)
}
