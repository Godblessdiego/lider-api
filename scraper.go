package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Product representa un producto extraído de Lider
// Extiende según tu JSON real

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

// Response mapea una respuesta genérica con productos
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

// handleSearch maneja GET /productos?q=
func handleSearch(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(400, gin.H{"error": "parámetro 'q' es requerido"})
		return
	}
	products, err := fetchProducts(query)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, products)
}

// handleSuggestions maneja GET /suggestions?term=
func handleSuggestions(ctx *gin.Context) {
	term := ctx.Query("term")
	if term == "" {
		ctx.JSON(400, gin.H{"error": "parámetro 'term' es requerido"})
		return
	}
	suggestions, err := fetchSuggestions(term)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, suggestions)
}

// handlePromotions maneja GET /promotions?type=
func handlePromotions(ctx *gin.Context) {
	promo := ctx.Query("type")
	if promo == "" {
		ctx.JSON(400, gin.H{"error": "parámetro 'type' es requerido"})
		return
	}
	products, err := fetchPromotions(promo)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, products)
}

// handleCategories maneja GET /categories?id=
func handleCategories(ctx *gin.Context) {
	cat := ctx.Query("id")
	if cat == "" {
		ctx.JSON(400, gin.H{"error": "parámetro 'id' es requerido"})
		return
	}
	products, err := fetchCategory(cat)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, products)
}

// fetchProducts hace POST a /search y pagina resultados
func fetchProducts(query string) ([]Product, error) {
	url := "https://apps.lider.cl/search"
	// Payload base
	base := map[string]interface{}{"query": query, "size": 20}
	// Primera llamada para leer nbPages
	data, err := doPost(url, base)
	if err != nil {
		return nil, err
	}
	var res Response
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	all := res.Products
	// Paginación concurrente
	for p := 2; p <= res.NbPages; p++ {
		base["page"] = p
		chunk, _ := postAndParse(url, base)
		all = append(all, chunk...)
	}
	return all, nil
}

// fetchSuggestions, fetchPromotions, fetchCategory y helpers…

func fetchSuggestions(term string) ([]string, error) {
	url := "https://apps.lider.cl/suggestions"
	payload := map[string]interface{}{"term": term}
	data, err := doPost(url, payload)
	if err != nil {
		return nil, err
	}
	var sr SuggestionResponse
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, err
	}
	return sr.Suggestions, nil
}

func fetchPromotions(p string) ([]Product, error) {
	return postAndParse("https://apps.lider.cl/promotions", map[string]interface{}{"promoType": p})
}
func fetchCategory(c string) ([]Product, error) {
	return postAndParse("https://apps.lider.cl/category", map[string]interface{}{"categoryId": c, "size": 20})
}

func postAndParse(url string, payload map[string]interface{}) ([]Product, error) {
	data, err := doPost(url, payload)
	if err != nil {
		return nil, err
	}
	var r Response
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return r.Products, nil
}

func doPost(url string, payload map[string]interface{}) ([]byte, error) {
	bodyReq, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bodyReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("petición fallida: " + resp.Status)
	}
	return io.ReadAll(resp.Body)
}
