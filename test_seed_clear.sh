#!/bin/bash

# Comprehensive test script for seed and clear functionality
BASE_URL="http://localhost:3000"

echo "🌱 Testing Seed and Clear Functionality"
echo "======================================="

# Helper function to extract data from JSON response
extract_count() {
    local response="$1"
    local field="$2"
    echo "$response" | jq -r ".data.$field // 0" 2>/dev/null
}

# Test 1: Clear any existing data
echo ""
echo "1. Clearing any existing data..."
clear_response=$(curl -X DELETE "$BASE_URL/api/seed/clear" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

clear_status=$(echo "$clear_response" | tail -n1 | cut -d: -f2)
clear_body=$(echo "$clear_response" | sed '$d')
echo "HTTP Status: $clear_status"

if [ "$clear_status" = "200" ]; then
    products_removed=$(extract_count "$clear_body" "products_removed")
    categories_removed=$(extract_count "$clear_body" "categories_removed")
    echo "Cleared: $products_removed products, $categories_removed categories"
fi

sleep 1

# Test 2: Verify data is cleared
echo ""
echo "2. Verifying data is cleared..."
products_check=$(curl "$BASE_URL/api/products?page=1&limit=1" \
  -H "Content-Type: application/json" \
  -s | jq '.pagination.total // 0' 2>/dev/null)

categories_check=$(curl "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -s | jq 'length // 0' 2>/dev/null)

echo "Products count: $products_check"
echo "Categories count: $categories_check"

# Test 3: Seed products from JSON
echo ""
echo "3. Seeding products from JSON..."
seed_response=$(curl -X POST "$BASE_URL/api/seed/products" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

seed_status=$(echo "$seed_response" | tail -n1 | cut -d: -f2)
seed_body=$(echo "$seed_response" | sed '$d')
echo "HTTP Status: $seed_status"

if [ "$seed_status" = "200" ]; then
    products_added=$(extract_count "$seed_body" "products_added")
    categories_added=$(extract_count "$seed_body" "categories_added")
    products_after=$(extract_count "$seed_body" "products_after")
    categories_after=$(extract_count "$seed_body" "categories_after")
    echo "Added: $products_added products, $categories_added categories"
    echo "Total after seeding: $products_after products, $categories_after categories"
fi

sleep 3

# Test 4: Verify seeding worked
echo ""
echo "4. Verifying seeding worked..."
products_verify=$(curl "$BASE_URL/api/products?page=1&limit=1" \
  -H "Content-Type: application/json" \
  -s | jq '.pagination.total // 0' 2>/dev/null)

categories_verify=$(curl "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -s | jq 'length // 0' 2>/dev/null)

echo "Products count: $products_verify"
echo "Categories count: $categories_verify"

# Test 5: Test clear again
echo ""
echo "5. Testing clear functionality again..."
clear2_response=$(curl -X DELETE "$BASE_URL/api/seed/clear" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS:%{http_code}" \
  -s)

clear2_status=$(echo "$clear2_response" | tail -n1 | cut -d: -f2)
clear2_body=$(echo "$clear2_response" | sed '$d')
echo "HTTP Status: $clear2_status"

if [ "$clear2_status" = "200" ]; then
    products_removed2=$(extract_count "$clear2_body" "products_removed")
    categories_removed2=$(extract_count "$clear2_body" "categories_removed")
    echo "Cleared: $products_removed2 products, $categories_removed2 categories"
fi

sleep 1

# Test 6: Final verification
echo ""
echo "6. Final verification - should be empty again..."
products_final=$(curl "$BASE_URL/api/products?page=1&limit=1" \
  -H "Content-Type: application/json" \
  -s | jq '.pagination.total // 0' 2>/dev/null)

categories_final=$(curl "$BASE_URL/api/categories" \
  -H "Content-Type: application/json" \
  -s | jq 'length // 0' 2>/dev/null)

echo "Products count: $products_final"
echo "Categories count: $categories_final"

echo ""
echo "✅ Comprehensive seed and clear test completed!"
echo ""
echo "Summary:"
echo "=========="
echo "Initial clear: $products_removed products, $categories_removed categories removed"
echo "Seeding result: $products_added products, $categories_added categories added"
echo "Second clear: $products_removed2 products, $categories_removed2 categories removed"
echo "Final state: $products_final products, $categories_final categories"
echo ""

# Overall test result
if [ "${products_final:-0}" = "0" ] && [ "${categories_final:-0}" = "0" ] && [ "${products_added:-0}" -gt "0" ]; then
    echo "✅ SUCCESS: All tests passed!"
    echo "   - Seeding worked (added ${products_added:-0} products)"
    echo "   - Clear function works properly"
    echo "   - Final state is clean"
else
    echo "❌ FAILURE: Some tests failed"
    if [ "${products_added:-0}" = "0" ]; then
        echo "   - Seeding failed (no products added)"
    fi
    if [ "${products_final:-0}" != "0" ] || [ "${categories_final:-0}" != "0" ]; then
        echo "   - Clear function failed (data remaining)"
    fi
fi 