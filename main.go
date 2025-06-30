package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// Carga la llave de entorno
	// Asegúrate de exportar API_KEY antes de correr:
	// export API_KEY=tu_llave_secreta

	router := gin.Default()

	// Aplica seguridad: autenticación por API key
	router.Use(apiKeyAuthMiddleware())

	// Rutas separadas en bloques
	router.GET("/productos", handleSearch)
	router.GET("/suggestions", handleSuggestions)
	router.GET("/promotions", handlePromotions)
	router.GET("/categories", handleCategories)

	// Ejecuta en puerto 8080
	router.Run(":8080")
}
