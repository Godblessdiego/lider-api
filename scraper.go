package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

// Product representa un producto extraído de Lider
type Product struct {
	ID          string    `json:"ID"`
	Brand       string    `json:"brand"`
	Description string    `json:"description"`
	DisplayName string    `json:"displayName"`
	Price       PriceInfo `json:"price"`
	Images      Images    `json:"images"`
}

// PriceInfo mapea precios de la respuesta
type PriceInfo struct {
	BasePriceReference float64 `json:"BasePriceReference"`
	BasePriceSales     float64 `json:"BasePriceSales"`
}

// Images contiene URLs de imágenes
type Images struct {
	DefaultImage string `json:"defaultImage"`
	MediumImage  string `json:"mediumImage"`
}

// Response mapea la estructura JSON de /search
type Response struct {
	Products []Product `json:"products"`
	NbHits   int       `json:"nbHits"`
	Page     int       `json:"page"`
	NbPages  int       `json:"nbPages"`
}

// SuggestionResponse mapea /suggestions
type SuggestionResponse struct {
	Suggestions []string `json:"suggestions"`
}

// fetchProducts usa GET al endpoint público de búsqueda con paginación
func fetchProducts(query string) ([]Product, error) {
	endpoint := "https://apps.lider.cl/supermercado/search"
	escape := url.QueryEscape(query)

	// primera página
	u := fmt.Sprintf("%s?query=%s&page=1", endpoint, escape)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s – %s", resp.Status, string(body))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	all := r.Products
	// páginas adicionales
	for p := 2; p <= r.NbPages; p++ {
		u := fmt.Sprintf("%s?query=%s&page=%d", endpoint, escape, p)
		resp, err := http.Get(u)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
		var rp Response
		if err := json.Unmarshal(body, &rp); err != nil {
			continue
		}
		all = append(all, rp.Products...)
	}
	return all, nil
}

// fetchSuggestions usa GET al endpoint de sugerencias
func fetchSuggestions(term string) ([]string, error) {
	endpoint := "https://apps.lider.cl/supermercado/suggestions"
	u := fmt.Sprintf("%s?term=%s", endpoint, url.QueryEscape(term))

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("suggest failed: %s – %s", resp.Status, string(body))
	}

	var sr SuggestionResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, err
	}
	return sr.Suggestions, nil
}

// fetchPromotions usa GET al endpoint de promociones
func fetchPromotions(promoType string) ([]Product, error) {
	endpoint := "https://apps.lider.cl/supermercado/promotions"
	u := fmt.Sprintf("%s?type=%s", endpoint, url.QueryEscape(promoType))

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("promotions failed: %s – %s", resp.Status, string(body))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return r.Products, nil
}

// fetchCategory usa GET al endpoint de categoría
func fetchCategory(categoryID string) ([]Product, error) {
	endpoint := "https://apps.lider.cl/supermercado/category"
	u := fmt.Sprintf("%s?id=%s", endpoint, url.QueryEscape(categoryID))

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("category failed: %s – %s", resp.Status, string(body))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return r.Products, nil
}

// apiKeyAuthMiddleware fuerza que el header X-API-Key coincida con la variable de entorno API_KEY
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

// Handlers
func handleSearch(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere q"})
		return
	}
	prods, err := fetchProducts(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prods)
}

func handleSuggestions(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere term"})
		return
	}
	suggestions, err := fetchSuggestions(term)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, suggestions)
}

func handlePromotions(c *gin.Context) {
	p := c.Query("type")
	if p == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere type"})
		return
	}
	prods, err := fetchPromotions(p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prods)
}

func handleCategories(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere id"})
		return
	}
	prods, err := fetchCategory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prods)
}

func main() {
	if os.Getenv("API_KEY") == "" {
		fmt.Println("ERROR: debes definir API_KEY en el entorno")
		return
	}

	router := gin.Default()
	router.Use(apiKeyAuthMiddleware())

	router.GET("/productos", handleSearch)
	router.GET("/suggestions", handleSuggestions)
	router.GET("/promotions", handlePromotions)
	router.GET("/categories", handleCategories)

	router.Run(":8080")
}
