# High Performance Go Fiber Boilerplate

A high-performance Go API with authentication, products, assets management, and caching. Features include product seeding from JSON, image serving, and comprehensive product management with high-concurrency bulk upload capabilities.

## 🚀 High-Performance Bulk Upload Features

### Timeout Configuration

The application now supports configurable timeouts optimized for bulk operations. You can set these via environment variables:

```bash
# HTTP Server Timeouts - Optimized for Bulk Uploads
READ_TIMEOUT=5m          # 5 minutes for large file reads
WRITE_TIMEOUT=10m        # 10 minutes for bulk operations
IDLE_TIMEOUT=15m         # 15 minutes idle
BODY_LIMIT=104857600     # 100MB for large JSON files

# Database Connection Pool - Optimized for High Concurrency
DB_MAX_OPEN_CONNS=200    # Increased for bulk operations
DB_MAX_IDLE_CONNS=50     # More idle connections
DB_CONN_MAX_LIFETIME=2h  # 2 hours connection lifetime
DB_CONN_MAX_IDLE_TIME=1h # 1 hour idle time

# Redis Timeouts - Optimized for Caching
REDIS_DIAL_TIMEOUT=10s   # 10 seconds dial timeout
REDIS_READ_TIMEOUT=5s    # 5 seconds read timeout
REDIS_WRITE_TIMEOUT=5s   # 5 seconds write timeout
REDIS_POOL_TIMEOUT=10s   # 10 seconds pool timeout
```

### Bulk Upload Performance

- **10 Concurrent Workers**: Processes chunks simultaneously
- **200 Products per Chunk**: Optimized chunk size for reliability
- **COPY Protocol**: Direct database insertion for maximum speed
- **Real-time Progress**: Live progress tracking during upload
- **Graceful Timeout Handling**: Proper error responses and recovery

### Expected Performance

For 10K products:
- **Processing Time**: 30-60 seconds (vs 2-5 minutes previously)
- **Throughput**: 150-300 products/second
- **Success Rate**: >98%
- **Memory Usage**: Optimized with streaming processing

## Features

- 🔐 JWT Authentication
- 📦 Product Management with CRUD operations
- 🖼️ Image serving and management
- 🔍 Full-text search capabilities
- 💾 Redis caching for performance
- 📊 Admin dashboard with bulk operations
- 🚀 High-concurrency bulk upload
- 📈 Real-time progress tracking
- 🛡️ Graceful error handling and recovery

## Quick Start

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your settings
3. Run `go mod tidy` to install dependencies
4. Start the server: `go run main.go`
5. Access the admin dashboard at `http://localhost:3000/admin`

## API Endpoints

### Public Endpoints
- `GET /api/products` - List products with pagination
- `GET /api/products/:id` - Get product by ID
- `GET /api/categories` - List categories
- `POST /api/search` - Search products

### Admin Endpoints
- `GET /admin` - Admin dashboard
- `GET /admin/api/products` - Admin product list
- `POST /admin/api/products` - Create product
- `POST /admin/api/products/bulk` - Bulk upload products
- `DELETE /admin/api/products/bulk-delete` - Delete all products
- `POST /admin/api/cache/clear` - Clear cache

## Bulk Upload Format

Upload a JSON file with the following format:

```json
[
  {
    "Name": "Product Name",
    "Price": 99.99,
    "Category": "Category Name",
    "Stock": 100,
    "Description": "Product description"
  }
]
```

## Performance Optimizations

- **Connection Pooling**: Optimized database and Redis connection pools
- **Caching**: Redis-based caching for frequently accessed data
- **Indexing**: Full-text search indexes and performance indexes
- **Chunked Processing**: Large datasets processed in manageable chunks
- **High Concurrency**: Multiple workers processing simultaneously
- **COPY Protocol**: Direct database insertion bypassing ORM overhead
