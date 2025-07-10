#!/bin/bash

echo "Testing Go Fiber Boilerplate API..."

# Test registration
echo ""
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:3000/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "admin@gmail.com",
    "first_name": "Abdullah",
    "last_name": "Rifat",
    "password": "123456"
  }')

echo "Register Response: $REGISTER_RESPONSE"

# Extract token from registration response if successful
TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "Registration failed or token not found. Trying login..."
    
    # Test login
    echo ""
    echo "2. Testing user login..."
    LOGIN_RESPONSE=$(curl -s -X POST http://localhost:3000/api/auth/login \
      -H 'Content-Type: application/json' \
      -d '{
        "email": "admin@gmail.com",
        "password": "123456"
      }')
    
    echo "Login Response: $LOGIN_RESPONSE"
    
    # Extract token from login response
    TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
fi

if [ ! -z "$TOKEN" ]; then
    echo ""
    echo "3. Testing protected endpoint (profile)..."
    PROFILE_RESPONSE=$(curl -s -X GET http://localhost:3000/api/auth/profile \
      -H "Authorization: Bearer $TOKEN")
    
    echo "Profile Response: $PROFILE_RESPONSE"
    
    echo ""
    echo "4. Testing health check..."
    HEALTH_RESPONSE=$(curl -s -X GET http://localhost:3000/api/health)
    echo "Health Response: $HEALTH_RESPONSE"
else
    echo "Failed to get authentication token. Please check the server logs."
fi

echo ""
echo "API testing completed!" 