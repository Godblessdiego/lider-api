package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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

// ProductDetail representa detalles completos de un producto individual
type ProductDetail struct {
	SKU            string      `json:"sku"`
	Name           string      `json:"name"`
	Brand          string      `json:"brand"`
	Description    string      `json:"description"`
	Price          DetailPrice `json:"price"`
	Images         []string    `json:"images"`
	Specifications []Spec      `json:"specifications"`
	Availability   bool        `json:"availability"`
	Stock          int         `json:"stock"`
	Rating         float64     `json:"rating"`
	ReviewCount    int         `json:"reviewCount"`
	Category       string      `json:"category"`
	URL            string      `json:"url"`
}

// DetailPrice contiene información detallada de precios
type DetailPrice struct {
	Current  float64 `json:"current"`
	Original float64 `json:"original"`
	Discount float64 `json:"discount"`
	Currency string  `json:"currency"`
	PerUnit  string  `json:"perUnit"`
}

// Spec representa una especificación del producto
type Spec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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

// fetchProductDetail obtiene detalles completos de un producto por SKU
func fetchProductDetail(sku string) (*ProductDetail, error) {
	if sku == "" {
		return nil, fmt.Errorf("SKU parameter cannot be empty")
	}

	// Primero intentamos obtener el producto via API interna
	detail, err := fetchProductDetailViaAPI(sku)
	if err == nil {
		return detail, nil
	}

	log.Printf("API method failed for SKU %s, trying web scraping: %v", sku, err)

	// Si falla la API, intentamos web scraping
	return fetchProductDetailViaScraping(sku)
}

// fetchProductDetailViaAPI intenta obtener datos via API interna
func fetchProductDetailViaAPI(sku string) (*ProductDetail, error) {
	// Intentar diferentes endpoints de API que podrían existir
	endpoints := []string{
		fmt.Sprintf("https://apps.lider.cl/supermercado/product/%s", sku),
		fmt.Sprintf("https://www.lider.cl/catalogo/api/products/%s", sku),
		fmt.Sprintf("https://api.lider.cl/v1/products/%s", sku),
	}

	for _, endpoint := range endpoints {
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LiderAPI/1.0)")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Language", "es-CL,es;q=0.9")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			var detail ProductDetail
			if err := json.Unmarshal(body, &detail); err == nil {
				log.Printf("Successfully fetched product detail via API for SKU: %s", sku)
				return &detail, nil
			}
		}
		resp.Body.Close()
	}

	return nil, fmt.Errorf("no working API endpoint found for SKU: %s", sku)
}

// fetchProductDetailViaScraping obtiene datos mediante web scraping
func fetchProductDetailViaScraping(sku string) (*ProductDetail, error) {
	// Construir URL del producto
	productURL := fmt.Sprintf("https://www.lider.cl/supermercado/product/sku/%s", sku)

	req, err := http.NewRequest("GET", productURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers para simular un navegador real
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-CL,es;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	log.Printf("Scraping product detail for SKU: %s", sku)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scraping failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	html := string(body)

	// Extraer datos usando expresiones regulares
	detail := &ProductDetail{
		SKU: sku,
		URL: productURL,
		Price: DetailPrice{
			Currency: "CLP",
		},
	}

	// Extraer nombre del producto
	if name := extractWithRegex(html, `<h1[^>]*class="[^"]*product-title[^"]*"[^>]*>([^<]+)</h1>`); name != "" {
		detail.Name = strings.TrimSpace(name)
	}

	// Extraer marca
	if brand := extractWithRegex(html, `<span[^>]*class="[^"]*brand[^"]*"[^>]*>([^<]+)</span>`); brand != "" {
		detail.Brand = strings.TrimSpace(brand)
	}

	// Extraer precio actual
	if priceStr := extractWithRegex(html, `<span[^>]*class="[^"]*price[^"]*"[^>]*>\$?([0-9,.]+)</span>`); priceStr != "" {
		if price := parsePrice(priceStr); price > 0 {
			detail.Price.Current = price
		}
	}

	// Extraer precio original
	if originalPriceStr := extractWithRegex(html, `<span[^>]*class="[^"]*original-price[^"]*"[^>]*>\$?([0-9,.]+)</span>`); originalPriceStr != "" {
		if price := parsePrice(originalPriceStr); price > 0 {
			detail.Price.Original = price
		}
	}

	// Calcular descuento
	if detail.Price.Original > 0 && detail.Price.Current > 0 {
		detail.Price.Discount = ((detail.Price.Original - detail.Price.Current) / detail.Price.Original) * 100
	}

	// Extraer disponibilidad
	if strings.Contains(html, "en-stock") || strings.Contains(html, "available") {
		detail.Availability = true
	}

	// Extraer rating
	if ratingStr := extractWithRegex(html, `"rating"\s*:\s*([0-9.]+)`); ratingStr != "" {
		fmt.Sscanf(ratingStr, "%f", &detail.Rating)
	}

	// Extraer imágenes
	imageRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"[^>]*(?:class="[^"]*product-image[^"]*"|alt="[^"]*product[^"]*")`)
	matches := imageRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			detail.Images = append(detail.Images, match[1])
		}
	}

	// Si no encontramos datos básicos, la página podría haber cambiado
	if detail.Name == "" && detail.Price.Current == 0 {
		return nil, fmt.Errorf("could not extract product data - page structure may have changed")
	}

	log.Printf("Successfully scraped product detail for SKU: %s", sku)
	return detail, nil
}

// extractWithRegex extrae texto usando expresión regular
func extractWithRegex(html, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parsePrice convierte string de precio a float64
func parsePrice(priceStr string) float64 {
	// Remover puntos de miles y reemplazar coma decimal
	clean := strings.ReplaceAll(priceStr, ".", "")
	clean = strings.ReplaceAll(clean, ",", ".")
	clean = strings.ReplaceAll(clean, "$", "")
	clean = strings.TrimSpace(clean)

	var price float64
	fmt.Sscanf(clean, "%f", &price)
	return price
}

// extractSKUFromURL extrae SKU de una URL de producto de Lider
func extractSKUFromURL(productURL string) string {
	re := regexp.MustCompile(`/sku/(\d+)/`)
	matches := re.FindStringSubmatch(productURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
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

// fetchProductDetailViaAPI obtiene el detalle completo de un producto por SKU
