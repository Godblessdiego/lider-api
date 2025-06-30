// main.go
package main

import (
	"encoding/json"
	"errors"
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

type PriceInfo struct {
	BasePriceReference float64 `json:"BasePriceReference"`
	BasePriceSales     float64 `json:"BasePriceSales"`
}

type Images struct {
	DefaultImage string `json:"defaultImage"`
	MediumImage  string `json:"mediumImage"`
}

// Response mapea la estructura de la búsqueda pública
type Response struct {
	Products []Product `json:"products"`
	NbHits   int       `json:"nbHits"`
	Page     int       `json:"page"`
	NbPages  int       `json:"nbPages"`
}

type SuggestionResponse struct {
	Suggestions []string `json:"suggestions"`
}

func fetchProducts(query string) ([]Product, error) {
	base := "https://apps.lider.cl/supermercado/search"
	escape := url.QueryEscape(query)

	// primera página
	u := fmt.Sprintf("%s?query=%s&page=1", base, escape)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("search failed: %s — %s", resp.Status, string(body)))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	all := r.Products
	for p := 2; p <= r.NbPages; p++ {
		u := fmt.Sprintf("%s?query=%s&page=%d", base, escape, p)
		resp, err := http.Get(u)
		if err != nil {
			continue
		}
		chunk, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
		var rp Response
		if err := json.Unmarshal(chunk, &rp); err != nil {
			continue
		}
		all = append(all, rp.Products...)
	}
	return all, nil
}

func fetchSuggestions(term string) ([]string, error) {
	// placeholder: implementar cuando el endpoint esté disponible
	return []string{}, nil
}

func fetchPromotions(promoType string) ([]Product, error) {
	// placeholder: implementar cuando el endpoint esté disponible
	return []Product{}, nil
}

func fetchCategory(categoryID string) ([]Product, error) {
	// placeholder: implementar cuando el endpoint esté disponible
	return []Product{}, nil
}

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
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), apiKeyAuthMiddleware())

	router.GET("/productos", handleSearch)
	router.GET("/suggestions", handleSuggestions)
	router.GET("/promotions", handlePromotions)
	router.GET("/categories", handleCategories)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
