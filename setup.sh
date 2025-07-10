#!/bin/bash

echo "Setting up Go Fiber Boilerplate..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
# Database Configuration
DATABASE_URL=postgres://postgres:password@localhost:5432/boilerplate?sslmode=disable

# Redis Configuration
REDIS_URL=localhost:6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here-change-this-in-production

# Server Configuration
PORT=3000
EOF
    echo ".env file created. Please update the DATABASE_URL with your actual database credentials."
else
    echo ".env file already exists."
fi

# Install dependencies
echo "Installing Go dependencies..."
go mod tidy

# Build the application
echo "Building the application..."
go build -o main .

echo "Setup completed!"
echo ""
echo "Next steps:"
echo "1. Update the DATABASE_URL in .env with your PostgreSQL credentials"
echo "2. Make sure PostgreSQL and Redis are running"
echo "3. Run the application: ./main"
echo "4. Test the API endpoints"
echo ""
echo "Example test commands:"
echo "# Register a user"
echo "curl -X POST http://localhost:3000/api/auth/register \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"email\":\"test@example.com\",\"password\":\"123456\",\"first_name\":\"John\",\"last_name\":\"Doe\"}'"
echo ""
echo "# Login"
echo "curl -X POST http://localhost:3000/api/auth/login \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"email\":\"test@example.com\",\"password\":\"123456\"}'"
echo ""
echo "# Seed products from JSON"
echo "curl -X POST http://localhost:3000/api/seed/products"
echo ""
echo "# Get products"
echo "curl http://localhost:3000/api/products" 