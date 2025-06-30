package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// Global advanced scraper instance
var (
	advancedScraper *AdvancedScraper
	scraperOnce     sync.Once
)

// getAdvancedScraper returns the singleton advanced scraper instance
func getAdvancedScraper() *AdvancedScraper {
	scraperOnce.Do(func() {
		advancedScraper = NewAdvancedScraper()
		log.Println("Advanced scraper initialized with anti-bot protection")
	})
	return advancedScraper
}

// fetchProductsAdvanced replaces the original fetchProducts function
func fetchProductsAdvanced(query string) ([]Product, error) {
	if query == "" {
		return nil, fmt.Errorf("query parameter cannot be empty")
	}

	scraper := getAdvancedScraper()
	result := scraper.FetchProductsAdvanced(query)

	if !result.Success {
		return nil, fmt.Errorf("search failed: %s", result.Error)
	}

	// Convert result data to []Product
	products, err := convertToProducts(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert search results: %w", err)
	}

	log.Printf("Successfully fetched %d products for query '%s' using %s", len(products), query, result.Source)
	return products, nil
}

// fetchProductDetailAdvanced replaces the original fetchProductDetail function
func fetchProductDetailAdvanced(sku string) (*ProductDetail, error) {
	if sku == "" {
		return nil, fmt.Errorf("SKU parameter cannot be empty")
	}

	scraper := getAdvancedScraper()
	result := scraper.FetchProductDetailAdvanced(sku)

	if !result.Success {
		return nil, fmt.Errorf("product detail fetch failed: %s", result.Error)
	}

	// Convert result data to ProductDetail
	detail, err := convertToProductDetail(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert product detail: %w", err)
	}

	log.Printf("Successfully fetched product detail for SKU '%s' using %s", sku, result.Source)
	return detail, nil
}

// fetchSuggestionsAdvanced provides suggestions with fallback
func fetchSuggestionsAdvanced(term string) ([]string, error) {
	if term == "" {
		return nil, fmt.Errorf("term parameter cannot be empty")
	}

	// Try original method first (it might work for suggestions)
	suggestions, err := fetchSuggestions(term)
	if err == nil && len(suggestions) > 0 {
		return suggestions, nil
	}

	// Fallback: generate suggestions based on common search terms
	fallbackSuggestions := generateFallbackSuggestions(term)
	if len(fallbackSuggestions) > 0 {
		log.Printf("Using fallback suggestions for term '%s'", term)
		return fallbackSuggestions, nil
	}

	return nil, fmt.Errorf("suggestions failed: %s", err.Error())
}

// fetchPromotionsAdvanced handles promotions with advanced scraping
func fetchPromotionsAdvanced(promoType string) ([]Product, error) {
	if promoType == "" {
		return nil, fmt.Errorf("promoType parameter cannot be empty")
	}

	// Try original method first
	products, err := fetchPromotions(promoType)
	if err == nil && len(products) > 0 {
		return products, nil
	}

	// Fallback: try to scrape promotions page
	scraper := getAdvancedScraper()
	promoURL := fmt.Sprintf("https://www.lider.cl/supermercado/ofertas?type=%s", promoType)
	result := scraper.scrapeSearchPage(promoURL)

	if !result.Success {
		return nil, fmt.Errorf("promotions failed: original API error: %s, scraping error: %s", err.Error(), result.Error)
	}

	products, convErr := convertToProducts(result.Data)
	if convErr != nil {
		return nil, fmt.Errorf("failed to convert promotion results: %w", convErr)
	}

	log.Printf("Successfully fetched %d promotions for type '%s' using scraping", len(products), promoType)
	return products, nil
}

// fetchCategoryAdvanced handles category products with advanced scraping
func fetchCategoryAdvanced(categoryID string) ([]Product, error) {
	if categoryID == "" {
		return nil, fmt.Errorf("categoryID parameter cannot be empty")
	}

	// Try original method first
	products, err := fetchCategory(categoryID)
	if err == nil && len(products) > 0 {
		return products, nil
	}

	// Fallback: try to scrape category page
	scraper := getAdvancedScraper()
	categoryURL := fmt.Sprintf("https://www.lider.cl/supermercado/category/%s", categoryID)
	result := scraper.scrapeSearchPage(categoryURL)

	if !result.Success {
		return nil, fmt.Errorf("category failed: original API error: %s, scraping error: %s", err.Error(), result.Error)
	}

	products, convErr := convertToProducts(result.Data)
	if convErr != nil {
		return nil, fmt.Errorf("failed to convert category results: %w", convErr)
	}

	log.Printf("Successfully fetched %d products for category '%s' using scraping", len(products), categoryID)
	return products, nil
}

// convertToProducts converts interface{} to []Product
func convertToProducts(data interface{}) ([]Product, error) {
	var products []Product

	switch v := data.(type) {
	case []Product:
		return v, nil
	case []interface{}:
		for _, item := range v {
			if productMap, ok := item.(map[string]interface{}); ok {
				product := mapInterfaceToProduct(productMap)
				if product.ID != "" || product.DisplayName != "" {
					products = append(products, product)
				}
			}
		}
	case map[string]interface{}:
		// Check if it's a response wrapper
		if productsData, ok := v["products"].([]interface{}); ok {
			for _, item := range productsData {
				if productMap, ok := item.(map[string]interface{}); ok {
					product := mapInterfaceToProduct(productMap)
					if product.ID != "" || product.DisplayName != "" {
						products = append(products, product)
					}
				}
			}
		}
	default:
		// Try to marshal/unmarshal through JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		// Try as array first
		var productArray []Product
		if err := json.Unmarshal(jsonData, &productArray); err == nil {
			return productArray, nil
		}

		// Try as response wrapper
		var response Response
		if err := json.Unmarshal(jsonData, &response); err == nil {
			return response.Products, nil
		}

		return nil, fmt.Errorf("unsupported data format: %T", data)
	}

	return products, nil
}

// convertToProductDetail converts interface{} to *ProductDetail
func convertToProductDetail(data interface{}) (*ProductDetail, error) {
	switch v := data.(type) {
	case *ProductDetail:
		return v, nil
	case ProductDetail:
		return &v, nil
	case map[string]interface{}:
		return mapInterfaceToProductDetail(v), nil
	default:
		// Try JSON marshal/unmarshal
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		var detail ProductDetail
		if err := json.Unmarshal(jsonData, &detail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to ProductDetail: %w", err)
		}

		return &detail, nil
	}
}

// mapInterfaceToProduct converts map[string]interface{} to Product
func mapInterfaceToProduct(data map[string]interface{}) Product {
	product := Product{}

	if id, ok := data["id"].(string); ok {
		product.ID = id
	} else if id, ok := data["ID"].(string); ok {
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
	} else if name, ok := data["name"].(string); ok {
		product.DisplayName = name
	}

	// Handle price data
	if priceData, ok := data["price"].(map[string]interface{}); ok {
		if baseRef, ok := priceData["BasePriceReference"].(float64); ok {
			product.Price.BasePriceReference = baseRef
		} else if baseRef, ok := priceData["original"].(float64); ok {
			product.Price.BasePriceReference = baseRef
		}

		if baseSales, ok := priceData["BasePriceSales"].(float64); ok {
			product.Price.BasePriceSales = baseSales
		} else if baseSales, ok := priceData["current"].(float64); ok {
			product.Price.BasePriceSales = baseSales
		}
	}

	// Handle images
	if imageData, ok := data["images"].(map[string]interface{}); ok {
		if defaultImg, ok := imageData["defaultImage"].(string); ok {
			product.Images.DefaultImage = defaultImg
		}
		if mediumImg, ok := imageData["mediumImage"].(string); ok {
			product.Images.MediumImage = mediumImg
		}
	} else if images, ok := data["images"].([]interface{}); ok && len(images) > 0 {
		if img1, ok := images[0].(string); ok {
			product.Images.DefaultImage = img1
		}
		if len(images) > 1 {
			if img2, ok := images[1].(string); ok {
				product.Images.MediumImage = img2
			}
		}
	}

	return product
}

// mapInterfaceToProductDetail converts map[string]interface{} to *ProductDetail
func mapInterfaceToProductDetail(data map[string]interface{}) *ProductDetail {
	detail := &ProductDetail{}

	if sku, ok := data["sku"].(string); ok {
		detail.SKU = sku
	} else if sku, ok := data["SKU"].(string); ok {
		detail.SKU = sku
	}

	if name, ok := data["name"].(string); ok {
		detail.Name = name
	} else if name, ok := data["displayName"].(string); ok {
		detail.Name = name
	}

	if brand, ok := data["brand"].(string); ok {
		detail.Brand = brand
	}

	if desc, ok := data["description"].(string); ok {
		detail.Description = desc
	}

	// Handle price data
	if priceData, ok := data["price"].(map[string]interface{}); ok {
		if current, ok := priceData["current"].(float64); ok {
			detail.Price.Current = current
		} else if current, ok := priceData["BasePriceSales"].(float64); ok {
			detail.Price.Current = current
		}

		if original, ok := priceData["original"].(float64); ok {
			detail.Price.Original = original
		} else if original, ok := priceData["BasePriceReference"].(float64); ok {
			detail.Price.Original = original
		}

		if currency, ok := priceData["currency"].(string); ok {
			detail.Price.Currency = currency
		} else {
			detail.Price.Currency = "CLP"
		}
	}

	// Handle images
	if images, ok := data["images"].([]interface{}); ok {
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				detail.Images = append(detail.Images, imgStr)
			}
		}
	} else if imageData, ok := data["images"].(map[string]interface{}); ok {
		if defaultImg, ok := imageData["defaultImage"].(string); ok {
			detail.Images = append(detail.Images, defaultImg)
		}
		if mediumImg, ok := imageData["mediumImage"].(string); ok {
			detail.Images = append(detail.Images, mediumImg)
		}
	}

	if avail, ok := data["availability"].(bool); ok {
		detail.Availability = avail
	} else if avail, ok := data["available"].(bool); ok {
		detail.Availability = avail
	} else {
		detail.Availability = true // Default to available
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

	// Generate URL if SKU is available
	if detail.SKU != "" && detail.URL == "" {
		detail.URL = fmt.Sprintf("https://www.lider.cl/supermercado/product/sku/%s", detail.SKU)
	}

	return detail
}

// generateFallbackSuggestions provides basic suggestions when API fails
func generateFallbackSuggestions(term string) []string {
	commonSuggestions := map[string][]string{
		"lec": {"leche", "leche descremada", "leche entera", "leche condensada", "lechuga"},
		"pan": {"pan", "pan integral", "pan molde", "pan hallulla", "panceta"},
		"arr": {"arroz", "arroz grado 1", "arroz integral", "arrollado"},
		"car": {"carne", "carne molida", "carne vacuno", "carnitas", "carbón"},
		"pol": {"pollo", "pollo entero", "pollo trozado", "pollo pechuga", "polenta"},
		"que": {"queso", "queso gauda", "queso mantecoso", "queso fresco", "queque"},
		"hue": {"huevos", "huevos blancos", "huevos color", "huevos codorniz"},
		"yog": {"yogurt", "yogurt natural", "yogurt griego", "yogurt light"},
		"man": {"mantequilla", "manzana", "manjar", "mandarina", "mango"},
		"cer": {"cereal", "cerveza", "cernir", "cerdo"},
	}

	// Find matches for the beginning of the term
	for prefix, suggestions := range commonSuggestions {
		if len(term) >= 3 && term[:3] == prefix {
			return suggestions
		}
	}

	// Fallback to generic suggestions
	return []string{
		term + " natural",
		term + " light",
		term + " premium",
		term + " casero",
		term + " integral",
	}
}
