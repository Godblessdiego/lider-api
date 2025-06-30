#!/bin/bash

# Lider API Test Script
# This script tests all available endpoints of the Lider API

# Configuration
BASE_URL="http://localhost:8080"
API_KEY="your-super-secret-api-key-here"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to make API requests
make_request() {
    local method=$1
    local endpoint=$2
    local auth_required=${3:-true}

    echo -e "\n${BLUE}===========================================${NC}"
    echo -e "${BLUE}Testing: $method $endpoint${NC}"
    echo -e "${BLUE}===========================================${NC}"

    if [ "$auth_required" = true ]; then
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
            -H "X-API-Key: $API_KEY" \
            -H "Accept: application/json" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
            -H "Accept: application/json" \
            "$BASE_URL$endpoint")
    fi

    # Extract HTTP code and body
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')

    echo "HTTP Status: $http_code"
    echo "Response:"
    echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        print_success "Request successful"
    else
        print_error "Request failed with status $http_code"
    fi
}

# Check if required tools are available
command -v curl >/dev/null 2>&1 || { print_error "curl is required but not installed. Aborting."; exit 1; }
command -v python3 >/dev/null 2>&1 || print_warning "python3 not found. JSON formatting will be disabled."

# Print script header
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}    Lider API Test Script${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Base URL: $BASE_URL"
echo -e "API Key: ${API_KEY:0:10}..."
echo ""

# Check if server is running
print_status "Checking if server is running..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    print_error "Server is not running at $BASE_URL"
    print_status "Start the server with: go run . or ./lider-api"
    exit 1
fi
print_success "Server is running"

# Test 1: Health Check (no auth required)
make_request "GET" "/health" false

# Test 2: Products Search
make_request "GET" "/productos?q=leche"

# Test 3: Products Search with empty query (should fail)
make_request "GET" "/productos?q="

# Test 4: Suggestions
make_request "GET" "/suggestions?term=lec"

# Test 5: Suggestions with empty term (should fail)
make_request "GET" "/suggestions?term="

# Test 6: Promotions
make_request "GET" "/promotions?type=descuentos"

# Test 7: Promotions with empty type (should fail)
make_request "GET" "/promotions?type="

# Test 8: Categories
make_request "GET" "/categories?id=123"

# Test 9: Categories with empty id (should fail)
make_request "GET" "/categories?id="

# Test 10: Authentication test (invalid API key)
echo -e "\n${BLUE}===========================================${NC}"
echo -e "${BLUE}Testing: Authentication with invalid key${NC}"
echo -e "${BLUE}===========================================${NC}"

response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -H "X-API-Key: invalid-key" \
    -H "Accept: application/json" \
    "$BASE_URL/productos?q=test")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

echo "HTTP Status: $http_code"
echo "Response:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"

if [ "$http_code" = "403" ]; then
    print_success "Authentication properly rejected invalid key"
else
    print_error "Authentication test failed - expected 403, got $http_code"
fi

# Test 11: Authentication test (missing API key)
echo -e "\n${BLUE}===========================================${NC}"
echo -e "${BLUE}Testing: Authentication with missing key${NC}"
echo -e "${BLUE}===========================================${NC}"

response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -H "Accept: application/json" \
    "$BASE_URL/productos?q=test")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

echo "HTTP Status: $http_code"
echo "Response:"
echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"

if [ "$http_code" = "401" ]; then
    print_success "Authentication properly rejected missing key"
else
    print_error "Authentication test failed - expected 401, got $http_code"
fi

# Summary
echo -e "\n${GREEN}===========================================${NC}"
echo -e "${GREEN}    Test Summary${NC}"
echo -e "${GREEN}===========================================${NC}"
print_status "All tests completed!"
print_status "Check the responses above for any issues."
echo ""
print_status "To use the API in your applications:"
echo -e "  curl -H \"X-API-Key: $API_KEY\" \"$BASE_URL/productos?q=search_term\""
echo ""
print_status "Available endpoints:"
echo -e "  GET /health                    - Health check (no auth)"
echo -e "  GET /productos?q=term          - Search products"
echo -e "  GET /suggestions?term=partial  - Get search suggestions"
echo -e "  GET /promotions?type=promo     - Get promotions"
echo -e "  GET /categories?id=cat_id      - Get category products"
