#!/bin/bash

# Test script for clearing functionality
BASE_URL="http://localhost:3000"

echo "🧹 Testing Clear Functionality"
echo "=============================="

echo ""
echo "1. Testing clear endpoint..."
response=$(curl -X DELETE "$BASE_URL/api/seed/clear" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

# Extract HTTP status and response body
http_status=$(echo "$response" | tail -n1 | cut -d: -f2)
response_body=$(echo "$response" | sed '$d')

echo "Response: $response_body"
echo "HTTP Status: $http_status"

# Parse and display the data if successful
if [ "$http_status" = "200" ]; then
    echo ""
    echo "Clear operation details:"
    echo "$response_body" | jq -r '.data | "Products removed: \(.products_removed), Categories removed: \(.categories_removed)"' 2>/dev/null || echo "Could not parse response data"
fi

echo ""
echo "2. Checking if products are cleared..."
products_response=$(curl "$BASE_URL/api/products?page=1&limit=5" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

products_status=$(echo "$products_response" | tail -n1 | cut -d: -f2)
products_body=$(echo "$products_response" | sed '$d')

echo "HTTP Status: $products_status"
total_products=$(echo "$products_body" | jq -r '.pagination.total // 0' 2>/dev/null)
echo "Total products: $total_products"

echo ""
echo "3. Checking if categories are cleared..."
categories_response=$(curl "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

categories_status=$(echo "$categories_response" | tail -n1 | cut -d: -f2)
categories_body=$(echo "$categories_response" | sed '$d')

echo "HTTP Status: $categories_status"
total_categories=$(echo "$categories_body" | jq 'length // 0' 2>/dev/null)
echo "Total categories: $total_categories"

echo ""
echo "✅ Clear functionality test completed!"
echo ""
echo "Summary:"
if [ "$total_products" = "0" ] && [ "$total_categories" = "0" ]; then
    echo "✅ SUCCESS: All data cleared successfully"
else
    echo "❌ FAILURE: Data not fully cleared (Products: $total_products, Categories: $total_categories)"
fi 