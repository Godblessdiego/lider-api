package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
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

// PriceInfo mapea los precios retornados
type PriceInfo struct {
	BasePriceReference float64 `json:"BasePriceReference"`
	BasePriceSales     float64 `json:"BasePriceSales"`
}

// Images contiene URLs de las imágenes del producto
type Images struct {
	DefaultImage string `json:"defaultImage"`
	MediumImage  string `json:"mediumImage"`
}

// Response mapea la respuesta JSON de /search, /promotions y /category
type Response struct {
	Products []Product `json:"products"`
	NbHits   int       `json:"nbHits"`
	Page     int       `json:"page"`
	NbPages  int       `json:"nbPages"`
}

// SuggestionResponse mapea sugerencias de autocompletar
type SuggestionResponse struct {
	Suggestions []string `json:"suggestions"`
}

// httpClient es un cliente HTTP configurado con timeout
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	},
}

// fetchProducts usa GET al endpoint público de búsqueda
func fetchProducts(query string) ([]Product, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter cannot be empty")
	}

	endpoint := "https://apps.lider.cl/supermercado/search"
	u := fmt.Sprintf("%s?query=%s", endpoint, url.QueryEscape(query))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to make request look more legitimate
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LiderAPI/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "es-CL,es;q=0.9")

	log.Printf("Fetching products for query: %s", query)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	log.Printf("Successfully fetched %d products for query: %s", len(r.Products), query)
	return r.Products, nil
}

// fetchSuggestions usa GET al endpoint de sugerencias
func fetchSuggestions(term string) ([]string, error) {
	if term == "" {
		return nil, fmt.Errorf("term parameter cannot be empty")
	}

	endpoint := "https://apps.lider.cl/supermercado/suggestions"
	u := fmt.Sprintf("%s?term=%s", endpoint, url.QueryEscape(term))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LiderAPI/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "es-CL,es;q=0.9")

	log.Printf("Fetching suggestions for term: %s", term)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("suggestions failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sr SuggestionResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	log.Printf("Successfully fetched %d suggestions for term: %s", len(sr.Suggestions), term)
	return sr.Suggestions, nil
}

// fetchPromotions usa GET al endpoint de promociones
func fetchPromotions(promoType string) ([]Product, error) {
	if promoType == "" {
		return nil, fmt.Errorf("promoType parameter cannot be empty")
	}

	endpoint := "https://apps.lider.cl/supermercado/promotions"
	u := fmt.Sprintf("%s?type=%s", endpoint, url.QueryEscape(promoType))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LiderAPI/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "es-CL,es;q=0.9")

	log.Printf("Fetching promotions for type: %s", promoType)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("promotions failed with status %d: %s", resp.StatusCode, string(body))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	log.Printf("Successfully fetched %d promotions for type: %s", len(r.Products), promoType)
	return r.Products, nil
}

// fetchCategory usa GET al endpoint de categoría
func fetchCategory(categoryID string) ([]Product, error) {
	if categoryID == "" {
		return nil, fmt.Errorf("categoryID parameter cannot be empty")
	}

	endpoint := "https://apps.lider.cl/supermercado/category"
	u := fmt.Sprintf("%s?id=%s", endpoint, url.QueryEscape(categoryID))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LiderAPI/1.0)")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "es-CL,es;q=0.9")

	log.Printf("Fetching category for ID: %s", categoryID)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("category failed with status %d: %s", resp.StatusCode, string(body))
	}

	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	log.Printf("Successfully fetched %d products for category: %s", len(r.Products), categoryID)
	return r.Products, nil
}
