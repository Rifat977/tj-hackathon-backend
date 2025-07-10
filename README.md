# High Performance Go API

A high-performance Go API built with Fiber, featuring authentication, product management, caching, and full-text search capabilities.

## Features

### Performance Optimizations
- **Connection Pooling**: Uses pgxpool for efficient database connections
- **Redis Caching**: Session storage and response caching
- **Full-text Search**: PostgreSQL GIN indexes for fast product search
- **Gzip Compression**: Automatic response compression
- **Pagination**: Cursor-based pagination for large datasets

### Authentication
- JWT-based authentication
- Session management with Redis
- Password hashing with bcrypt
- Role-based access control

### Database
- PostgreSQL with optimized indexes
- Full-text search capabilities
- Proper foreign key relationships
- Connection pooling

### Caching Strategy
- Redis for session storage
- Product listings and categories caching
- Search results caching with TTL
- User profiles caching

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login (JWT tokens)
- `POST /api/auth/logout` - User logout
- `GET /api/auth/profile` - Get user profile
- `PUT /api/auth/profile` - Update user profile

### Products
- `GET /api/products` - Product listing with pagination, filters
- `GET /api/products/:id` - Single product details
- `GET /api/products/search` - Product search with query params
- `GET /api/categories` - Product categories list
- `GET /api/categories/:id/products` - Products by category

### Seed Data
- `POST /api/seed/products` - Seed products from JSON file
- `DELETE /api/seed/clear` - Clear all products and categories

### Health Check
- `GET /api/health` - Health check endpoint

## Setup Instructions

### Prerequisites
- Go 1.23+
- PostgreSQL 12+
- Redis 6+

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-fiber-boilerplate
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Set up the database:
```sql
CREATE DATABASE boilerplate;
```

5. Run the application:
```bash
go run main.go
```

### Environment Variables

```env
# Database Configuration
DATABASE_URL=postgres://username:password@localhost:5432/boilerplate?sslmode=disable

# Redis Configuration
REDIS_URL=localhost:6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Server Configuration
PORT=3000
```

## API Usage Examples

### Register a new user
```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Seed products from JSON (1000 products)
```bash
curl -X POST http://localhost:3000/api/seed/products
```

### Clear all products and categories
```bash
curl -X DELETE http://localhost:3000/api/seed/clear
```

### Get products with pagination
```bash
curl "http://localhost:3000/api/products?page=1&limit=10"
```

### Search products
```bash
curl "http://localhost:3000/api/products/search?q=laptop&min_price=500&max_price=2000&sort_by=price&sort_order=ASC"
```

### Get user profile (authenticated)
```bash
curl -X GET http://localhost:3000/api/auth/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Database Seeding

### Product Seeding Features
- **JSON-based seeding**: Seeds all 1000 products from `assets/products.json`
- **Automatic categories**: Categories are created automatically from product data
- **Idempotent operation**: Running seed multiple times won't create duplicates
- **Comprehensive data**: All product fields including EAN, SKU, images, etc.
- **Performance optimized**: Efficient bulk operations with proper indexing
- **GORM-only implementation**: Uses only GORM methods with transaction support
- **Progress tracking**: Real-time progress logging for large datasets
- **Error handling**: Robust error handling with detailed logging

### Clear Functionality Features
- **Safe deletion**: Handles foreign key constraints properly with GORM
- **Transaction support**: Uses database transactions for data integrity
- **Complete cleanup**: Removes all products and categories permanently
- **Sequence reset**: Resets auto-increment counters
- **Verification**: Built-in verification to ensure complete data removal
- **GORM-only approach**: No raw SQL fallbacks needed

### API Response Format
Both seeding and clearing endpoints now return detailed information:

**Seed Response:**
```json
{
  "success": true,
  "message": "Products seeded successfully from JSON file",
  "data": {
    "products_before": 0,
    "products_after": 1000,
    "products_added": 1000,
    "categories_before": 0,
    "categories_after": 34,
    "categories_added": 34
  }
}
```

**Clear Response:**
```json
{
  "success": true,
  "message": "Products and categories cleared successfully",
  "data": {
    "products_before": 1000,
    "products_after": 0,
    "products_removed": 1000,
    "categories_before": 34,
    "categories_after": 0,
    "categories_removed": 34
  }
}
```

## Performance Features

### Database Optimizations
- Full-text search indexes for product search
- Proper foreign key relationships
- Connection pooling with pgxpool
- Query optimization for joins

### Caching Strategy
- Redis for session storage
- Cache product listings and categories
- Cache search results (with TTL)
- Cache user profiles

### API Performance
- Pagination with cursor-based approach
- Efficient search with database indexes
- Gzip compression
- Response time monitoring

## Development

### Running with hot reload
```bash
air
```

### Running tests
```bash
go test ./...
```

### Code formatting
```bash
gofmt -s -w .
```

## License

This project is licensed under the MIT License.
