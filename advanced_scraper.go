package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// AdvancedScraper maneja el scraping con técnicas anti-detección
type AdvancedScraper struct {
	client      *http.Client
	userAgents  []string
	rateLimiter chan time.Time
	retryDelays []time.Duration
	currentUA   int
}

// ScrapingResult contiene el resultado del scraping
type ScrapingResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Source  string      `json:"source"` // "api", "scraping", "cache"
}

// NewAdvancedScraper crea un nuevo scraper avanzado
func NewAdvancedScraper() *AdvancedScraper {
	// Crear jar de cookies
	jar, _ := cookiejar.New(nil)

	// Cliente HTTP con configuración avanzada
	client := &http.Client{
		Jar:     jar,
		Timeout: 45 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        30,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  false,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Detectar redirecciones a queue-it
			if strings.Contains(req.URL.Host, "queue-it.net") {
				return fmt.Errorf("blocked by anti-bot protection (queue-it)")
			}
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// User agents realistas
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	}

	// Rate limiter: máximo 1 request cada 2 segundos
	rateLimiter := make(chan time.Time, 1)
	go func() {
		for {
			rateLimiter <- time.Now()
			time.Sleep(2 * time.Second)
		}
	}()

	return &AdvancedScraper{
		client:      client,
		userAgents:  userAgents,
		rateLimiter: rateLimiter,
		retryDelays: []time.Duration{1 * time.Second, 3 * time.Second, 7 * time.Second, 15 * time.Second},
		currentUA:   0,
	}
}

// makeRequest hace una petición HTTP con todas las técnicas anti-detección
func (s *AdvancedScraper) makeRequest(method, url string, headers map[string]string) (*http.Response, []byte, error) {
	// Rate limiting
	<-s.rateLimiter

	var lastErr error

	for attempt := 0; attempt < len(s.retryDelays)+1; attempt++ {
		if attempt > 0 {
			time.Sleep(s.retryDelays[attempt-1])
		}

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Rotar user agent
		s.currentUA = (s.currentUA + 1) % len(s.userAgents)
		req.Header.Set("User-Agent", s.userAgents[s.currentUA])

		// Headers básicos para parecer un navegador real
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
		req.Header.Set("Accept-Language", "es-CL,es;q=0.9,en;q=0.8")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		req.Header.Set("Sec-Fetch-Site", "none")
		req.Header.Set("Sec-Fetch-User", "?1")
		req.Header.Set("Upgrade-Insecure-Requests", "1")

		// Headers personalizados
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Referrer aleatorio para parecer más natural
		if strings.Contains(url, "lider.cl") {
			req.Header.Set("Referer", "https://www.lider.cl/")
		}

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			continue
		}

		// Verificar si fuimos bloqueados
		if resp.StatusCode == 429 || resp.StatusCode == 503 {
			lastErr = fmt.Errorf("rate limited or service unavailable (status %d)", resp.StatusCode)
			continue
		}

		// Verificar contenido de queue-it en el body
		bodyStr := string(body)
		if strings.Contains(bodyStr, "queue-it.net") || strings.Contains(bodyStr, "Queue-it") {
			lastErr = fmt.Errorf("blocked by anti-bot protection")
			continue
		}

		return resp, body, nil
	}

	return nil, nil, fmt.Errorf("max retries exceeded, last error: %w", lastErr)
}

// FetchProductsAdvanced busca productos con técnicas avanzadas
func (s *AdvancedScraper) FetchProductsAdvanced(query string) *ScrapingResult {
	if query == "" {
		return &ScrapingResult{
			Success: false,
			Error:   "query parameter cannot be empty",
		}
	}

	// Intentar primero con API interna
	apiResult := s.tryAPIEndpoint(fmt.Sprintf("https://apps.lider.cl/supermercado/search?query=%s", url.QueryEscape(query)))
	if apiResult.Success {
		apiResult.Source = "api"
		return apiResult
	}

	// Fallback: scraping de la página de búsqueda
	searchURL := fmt.Sprintf("https://www.lider.cl/supermercado/search?query=%s", url.QueryEscape(query))
	scrapingResult := s.scrapeSearchPage(searchURL)
	if scrapingResult.Success {
		scrapingResult.Source = "scraping"
		return scrapingResult
	}

	return &ScrapingResult{
		Success: false,
		Error:   fmt.Sprintf("API failed: %s, Scraping failed: %s", apiResult.Error, scrapingResult.Error),
		Source:  "none",
	}
}

// FetchProductDetailAdvanced obtiene detalles de producto
func (s *AdvancedScraper) FetchProductDetailAdvanced(sku string) *ScrapingResult {
	if sku == "" {
		return &ScrapingResult{
			Success: false,
			Error:   "SKU parameter cannot be empty",
		}
	}

	// Intentar endpoints de API
	apiEndpoints := []string{
		fmt.Sprintf("https://apps.lider.cl/supermercado/product?sku=%s", sku),
		fmt.Sprintf("https://apps.lider.cl/supermercado/product/%s", sku),
		fmt.Sprintf("https://www.lider.cl/catalogo/api/products/%s", sku),
	}

	for _, endpoint := range apiEndpoints {
		result := s.tryAPIEndpoint(endpoint)
		if result.Success {
			result.Source = "api"
			return result
		}
	}

	// Fallback: scraping de la página del producto
	productURL := fmt.Sprintf("https://www.lider.cl/supermercado/product/sku/%s", sku)
	scrapingResult := s.scrapeProductPage(productURL)
	if scrapingResult.Success {
		scrapingResult.Source = "scraping"
		return scrapingResult
	}

	return &ScrapingResult{
		Success: false,
		Error:   "all methods failed - product may not exist or be blocked",
		Source:  "none",
	}
}

// tryAPIEndpoint intenta hacer una petición a un endpoint de API
func (s *AdvancedScraper) tryAPIEndpoint(endpoint string) *ScrapingResult {
	headers := map[string]string{
		"Accept": "application/json, text/plain, */*",
	}

	resp, body, err := s.makeRequest("GET", endpoint, headers)
	if err != nil {
		return &ScrapingResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ScrapingResult{
			Success: false,
			Error:   fmt.Sprintf("API returned status %d", resp.StatusCode),
		}
	}

	// Intentar parsear como JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return &ScrapingResult{
			Success: false,
			Error:   "failed to parse JSON response",
		}
	}

	return &ScrapingResult{
		Success: true,
		Data:    jsonData,
	}
}

// scrapeSearchPage hace scraping de la página de búsqueda
func (s *AdvancedScraper) scrapeSearchPage(searchURL string) *ScrapingResult {
	resp, body, err := s.makeRequest("GET", searchURL, nil)
	if err != nil {
		return &ScrapingResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ScrapingResult{
			Success: false,
			Error:   fmt.Sprintf("search page returned status %d", resp.StatusCode),
		}
	}

	// Extraer productos del HTML
	products := s.extractProductsFromHTML(string(body))
	if len(products) == 0 {
		return &ScrapingResult{
			Success: false,
			Error:   "no products found in search results",
		}
	}

	return &ScrapingResult{
		Success: true,
		Data:    products,
	}
}

// scrapeProductPage hace scraping de la página de un producto específico
func (s *AdvancedScraper) scrapeProductPage(productURL string) *ScrapingResult {
	resp, body, err := s.makeRequest("GET", productURL, nil)
	if err != nil {
		return &ScrapingResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ScrapingResult{
			Success: false,
			Error:   fmt.Sprintf("product page returned status %d", resp.StatusCode),
		}
	}

	// Extraer detalles del producto del HTML
	product := s.extractProductDetailFromHTML(string(body))
	if product == nil {
		return &ScrapingResult{
			Success: false,
			Error:   "could not extract product details from page",
		}
	}

	return &ScrapingResult{
		Success: true,
		Data:    product,
	}
}

// extractProductsFromHTML extrae productos del HTML de búsqueda
func (s *AdvancedScraper) extractProductsFromHTML(html string) []Product {
	var products []Product

	// Buscar datos JSON incrustados en el HTML
	jsonRegex := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*({.+?});`)
	matches := jsonRegex.FindStringSubmatch(html)

	if len(matches) > 1 {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
			// Extraer productos del estado inicial
			if searchData, ok := data["search"].(map[string]interface{}); ok {
				if results, ok := searchData["results"].([]interface{}); ok {
					for _, item := range results {
						if productMap, ok := item.(map[string]interface{}); ok {
							product := s.mapToProduct(productMap)
							if product.ID != "" {
								products = append(products, product)
							}
						}
					}
				}
			}
		}
	}

	// Fallback: extraer usando patrones HTML
	if len(products) == 0 {
		products = s.extractProductsFromHTMLPatterns(html)
	}

	return products
}

// extractProductDetailFromHTML extrae detalles de producto del HTML
func (s *AdvancedScraper) extractProductDetailFromHTML(html string) *ProductDetail {
	// Buscar datos JSON del producto
	jsonRegex := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*({.+?});`)
	matches := jsonRegex.FindStringSubmatch(html)

	if len(matches) > 1 {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
			if productData, ok := data["product"].(map[string]interface{}); ok {
				return s.mapToProductDetail(productData)
			}
		}
	}

	// Fallback: extraer usando patrones HTML
	return s.extractProductDetailFromHTMLPatterns(html)
}

// mapToProduct convierte un map a Product
func (s *AdvancedScraper) mapToProduct(data map[string]interface{}) Product {
	product := Product{}

	if id, ok := data["id"].(string); ok {
		product.ID = id
	}
	if brand, ok := data["brand"].(string); ok {
		product.Brand = brand
	}
	if desc, ok := data["description"].(string); ok {
		product.Description = desc
	}
	if name, ok := data["displayName"].(string); ok {
		product.DisplayName = name
	}

	// Extraer precios
	if priceData, ok := data["price"].(map[string]interface{}); ok {
		if baseRef, ok := priceData["BasePriceReference"].(float64); ok {
			product.Price.BasePriceReference = baseRef
		}
		if baseSales, ok := priceData["BasePriceSales"].(float64); ok {
			product.Price.BasePriceSales = baseSales
		}
	}

	// Extraer imágenes
	if imageData, ok := data["images"].(map[string]interface{}); ok {
		if defaultImg, ok := imageData["defaultImage"].(string); ok {
			product.Images.DefaultImage = defaultImg
		}
		if mediumImg, ok := imageData["mediumImage"].(string); ok {
			product.Images.MediumImage = mediumImg
		}
	}

	return product
}

// mapToProductDetail convierte un map a ProductDetail
func (s *AdvancedScraper) mapToProductDetail(data map[string]interface{}) *ProductDetail {
	detail := &ProductDetail{}

	if sku, ok := data["sku"].(string); ok {
		detail.SKU = sku
	}
	if name, ok := data["name"].(string); ok {
		detail.Name = name
	}
	if brand, ok := data["brand"].(string); ok {
		detail.Brand = brand
	}
	if desc, ok := data["description"].(string); ok {
		detail.Description = desc
	}

	// Extraer precios
	if priceData, ok := data["price"].(map[string]interface{}); ok {
		if current, ok := priceData["current"].(float64); ok {
			detail.Price.Current = current
		}
		if original, ok := priceData["original"].(float64); ok {
			detail.Price.Original = original
		}
		detail.Price.Currency = "CLP"
	}

	// Extraer imágenes
	if images, ok := data["images"].([]interface{}); ok {
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				detail.Images = append(detail.Images, imgStr)
			}
		}
	}

	if avail, ok := data["availability"].(bool); ok {
		detail.Availability = avail
	}
	if stock, ok := data["stock"].(float64); ok {
		detail.Stock = int(stock)
	}
	if rating, ok := data["rating"].(float64); ok {
		detail.Rating = rating
	}
	if category, ok := data["category"].(string); ok {
		detail.Category = category
	}

	detail.URL = fmt.Sprintf("https://www.lider.cl/supermercado/product/sku/%s", detail.SKU)

	return detail
}

// extractProductsFromHTMLPatterns extrae productos usando patrones HTML
func (s *AdvancedScraper) extractProductsFromHTMLPatterns(html string) []Product {
	var products []Product

	// Patrones regex para extraer información básica
	productRegex := regexp.MustCompile(`data-testid="product-item"[^>]*>(.*?)</div>`)
	matches := productRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			productHTML := match[1]
			product := Product{}

			// Extraer ID del producto
			if idMatch := regexp.MustCompile(`data-product-id="([^"]+)"`).FindStringSubmatch(productHTML); len(idMatch) > 1 {
				product.ID = idMatch[1]
			}

			// Extraer nombre
			if nameMatch := regexp.MustCompile(`data-testid="product-title"[^>]*>([^<]+)`).FindStringSubmatch(productHTML); len(nameMatch) > 1 {
				product.DisplayName = strings.TrimSpace(nameMatch[1])
			}

			// Extraer precio
			if priceMatch := regexp.MustCompile(`data-testid="product-price"[^>]*>.*?\$([0-9,.]+)`).FindStringSubmatch(productHTML); len(priceMatch) > 1 {
				priceStr := strings.ReplaceAll(priceMatch[1], ".", "")
				priceStr = strings.ReplaceAll(priceStr, ",", ".")
				if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
					product.Price.BasePriceSales = price
					product.Price.BasePriceReference = price
				}
			}

			if product.ID != "" && product.DisplayName != "" {
				products = append(products, product)
			}
		}
	}

	return products
}

// extractProductDetailFromHTMLPatterns extrae detalles usando patrones HTML
func (s *AdvancedScraper) extractProductDetailFromHTMLPatterns(html string) *ProductDetail {
	detail := &ProductDetail{}

	// Extraer SKU de la URL o metadatos
	if skuMatch := regexp.MustCompile(`"sku":"([^"]+)"`).FindStringSubmatch(html); len(skuMatch) > 1 {
		detail.SKU = skuMatch[1]
	}

	// Extraer nombre del producto
	if nameMatch := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`).FindStringSubmatch(html); len(nameMatch) > 1 {
		detail.Name = strings.TrimSpace(nameMatch[1])
	}

	// Extraer precio
	if priceMatch := regexp.MustCompile(`"price":\s*([0-9,.]+)`).FindStringSubmatch(html); len(priceMatch) > 1 {
		priceStr := strings.ReplaceAll(priceMatch[1], ".", "")
		priceStr = strings.ReplaceAll(priceStr, ",", ".")
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			detail.Price.Current = price
			detail.Price.Original = price
			detail.Price.Currency = "CLP"
		}
	}

	// Si no pudimos extraer información básica, retornar nil
	if detail.SKU == "" && detail.Name == "" {
		return nil
	}

	detail.Availability = true
	return detail
}
