#!/bin/bash

# Test script for assets functionality
BASE_URL="http://localhost:3000"

echo "🚀 Testing Assets Functionality"
echo "================================"

# Test 1: Seed products from JSON
echo ""
echo "1. Seeding products from JSON..."
curl -X POST "$BASE_URL/api/seed/products" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

# Wait a moment for seeding to complete
sleep 2

# Test 2: Get products (first page)
echo ""
echo "2. Getting products (first page)..."
curl "$BASE_URL/api/products?page=1&limit=5" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

# Test 3: Get categories
echo ""
echo "3. Getting categories..."
curl "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

# Test 4: Get banners
echo ""
echo "4. Getting banners..."
curl "$BASE_URL/api/banners" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "✅ Testing completed!" 