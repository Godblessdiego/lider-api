# Lider API

Una API REST para consultar productos, promociones y sugerencias del supermercado Lider de Chile.

## 🚀 Características

- **Búsqueda de productos** por término de búsqueda
- **Sugerencias de autocompletado** para términos de búsqueda
- **Consulta de promociones** por tipo
- **Productos por categoría** usando ID de categoría
- **Autenticación por API Key** para seguridad
- **Logging completo** de solicitudes y errores
- **CORS habilitado** para uso desde frontend
- **Timeout configurado** para evitar colgadas

## 📋 Requisitos

- Go 1.24.4 o superior
- Variable de entorno `API_KEY` configurada

## 🛠️ Instalación

1. Clona el repositorio:
```bash
git clone <repository-url>
cd lider-api
```

2. Instala las dependencias:
```bash
go mod tidy
```

3. Configura las variables de entorno:
```bash
cp .env.example .env
# Edita .env con tu API key
```

4. Ejecuta la aplicación:
```bash
go run .
```

O compila y ejecuta:
```bash
go build -o lider-api
./lider-api
```

## ⚙️ Configuración

### Variables de Entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
API_KEY=tu-clave-api-super-secreta
PORT=8080
GIN_MODE=release
```

- `API_KEY`: Clave requerida para autenticación (obligatoria)
- `PORT`: Puerto donde correrá el servidor (default: 8080)
- `GIN_MODE`: Modo de Gin (release/debug)

## 🔑 Autenticación

Todas las solicitudes (excepto `/health`) requieren el header `X-API-Key`:

```bash
curl -H "X-API-Key: tu-clave-api" http://localhost:8080/productos?q=leche
```

## 📚 Endpoints de la API

### Health Check

```http
GET /health
```

**Respuesta:**
```json
{
  "status": "ok",
  "service": "lider-api"
}
```

### Búsqueda de Productos

```http
GET /productos?q={término_búsqueda}
```

**Parámetros:**
- `q` (requerido): Término de búsqueda

**Ejemplo:**
```bash
curl -H "X-API-Key: tu-clave" "http://localhost:8080/productos?q=leche"
```

**Respuesta:**
```json
{
  "query": "leche",
  "count": 25,
  "products": [
    {
      "ID": "12345",
      "brand": "Soprole",
      "description": "Leche entera 1L",
      "displayName": "Leche Soprole Entera 1L",
      "price": {
        "BasePriceReference": 1200,
        "BasePriceSales": 1000
      },
      "images": {
        "defaultImage": "https://...",
        "mediumImage": "https://..."
      }
    }
  ]
}
```

### Sugerencias de Autocompletado

```http
GET /suggestions?term={término_parcial}
```

**Parámetros:**
- `term` (requerido): Término parcial para autocompletar

**Ejemplo:**
```bash
curl -H "X-API-Key: tu-clave" "http://localhost:8080/suggestions?term=lec"
```

**Respuesta:**
```json
{
  "term": "lec",
  "count": 5,
  "suggestions": [
    "leche",
    "leche descremada",
    "leche condensada",
    "leche evaporada",
    "lechuga"
  ]
}
```

### Promociones

```http
GET /promotions?type={tipo_promoción}
```

**Parámetros:**
- `type` (requerido): Tipo de promoción (ej: "descuentos", "ofertas")

**Ejemplo:**
```bash
curl -H "X-API-Key: tu-clave" "http://localhost:8080/promotions?type=descuentos"
```

**Respuesta:**
```json
{
  "type": "descuentos",
  "count": 15,
  "products": [
    {
      "ID": "67890",
      "brand": "Coca Cola",
      "description": "Bebida Cola 1.5L",
      "displayName": "Coca Cola 1.5L",
      "price": {
        "BasePriceReference": 2000,
        "BasePriceSales": 1500
      },
      "images": {
        "defaultImage": "https://...",
        "mediumImage": "https://..."
      }
    }
  ]
}
```

### Productos por Categoría

```http
GET /categories?id={id_categoría}
```

**Parámetros:**
- `id` (requerido): ID de la categoría

**Ejemplo:**
```bash
curl -H "X-API-Key: tu-clave" "http://localhost:8080/categories?id=123"
```

**Respuesta:**
```json
{
  "category_id": "123",
  "count": 30,
  "products": [...]
}
```

## 🚨 Manejo de Errores

### Códigos de Estado

- `200`: Éxito
- `400`: Solicitud incorrecta (parámetros faltantes)
- `401`: No autorizado (API key faltante)
- `403`: Prohibido (API key inválida)
- `500`: Error interno del servidor

### Formato de Errores

```json
{
  "error": "Descripción del error",
  "message": "Detalles técnicos del error",
  "example": "/productos?q=leche"
}
```

## 🔧 Desarrollo

### Estructura del Proyecto

```
lider-api/
├── main.go           # Servidor principal y handlers
├── middleware.go     # Middleware de autenticación
├── scraper.go        # Funciones para consultar APIs de Lider
├── go.mod           # Dependencias de Go
├── go.sum           # Checksums de dependencias
├── .env             # Variables de entorno (no en git)
├── .gitignore       # Archivos ignorados por git
└── README.md        # Esta documentación
```

### Ejecutar en Modo Debug

```bash
export GIN_MODE=debug
go run .
```

### Compilar para Producción

```bash
go build -ldflags="-s -w" -o lider-api
```

## 🔒 Seguridad

- **API Key obligatoria**: Todas las solicitudes requieren autenticación
- **Logging de seguridad**: Se registran intentos de acceso fallidos
- **CORS configurado**: Permite acceso desde dominios específicos
- **Timeouts**: Evita ataques de denegación de servicio
- **Rate limiting**: Considera implementar límites de solicitudes

## 📊 Logging

La aplicación registra:
- Solicitudes exitosas con IP y tiempo de respuesta
- Errores de autenticación con IP y User-Agent
- Errores de consulta a APIs externas
- Inicio y parada del servidor

## 🚀 Deployment

### Docker (Recomendado)

Crear `Dockerfile`:
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o lider-api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/lider-api .
CMD ["./lider-api"]
```

### Systemd Service

Crear `/etc/systemd/system/lider-api.service`:
```ini
[Unit]
Description=Lider API Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/lider-api
ExecStart=/opt/lider-api/lider-api
Restart=always
RestartSec=10
Environment=API_KEY=tu-clave-api
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

## 🤝 Contribuciones

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/nueva-caracteristica`)
3. Commit tus cambios (`git commit -am 'Agrega nueva característica'`)
4. Push a la rama (`git push origin feature/nueva-caracteristica`)
5. Abre un Pull Request

## 📝 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 🆘 Soporte

Si tienes problemas o preguntas:

1. Revisa los logs de la aplicación
2. Verifica que tu API key esté configurada correctamente
3. Asegúrate de que los endpoints de Lider estén disponibles
4. Abre un issue en GitHub con detalles del problema

## 📈 Métricas y Monitoreo

Para producción, considera agregar:
- Métricas de Prometheus
- Health checks avanzados
- Alertas por errores
- Dashboard de monitoreo