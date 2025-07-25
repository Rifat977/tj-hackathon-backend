# High Performance Go Fiber Boilerplate

A high-performance Go API with authentication, products management, and caching. Optimized for bulk uploads with 10K+ products support.

## 🚀 Features

- 🔐 JWT Authentication
- 📦 Product Management with CRUD operations
- 🖼️ Image serving and management
- 🔍 Full-text search capabilities
- 💾 Redis caching for performance
- 📊 Admin dashboard with bulk operations
- ⚡ Lightning-fast bulk upload (10K+ products in seconds)
- 📈 Real-time progress tracking
- 🛡️ Graceful error handling and recovery

## 📋 API Endpoints

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

## 📤 Bulk Upload Format

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

## ⚡ Performance Optimizations

- **Connection Pooling**: Optimized database and Redis connection pools for lightning-fast operations
- **Caching**: Redis-based caching for frequently accessed data
- **Indexing**: Full-text search indexes and lightning-fast performance indexes
- **Chunked Processing**: Large datasets processed in lightning-fast 1000-product chunks
- **High Concurrency**: 8 workers processing simultaneously for maximum throughput
- **COPY Protocol**: Direct database insertion bypassing ORM overhead
- **Streaming JSON**: Memory-efficient parsing for large files
- **Parallel Processing**: Categories processed in parallel batches

## 🐳 Docker Deployment

### Resource Allocation (2 vCPU, 4GB RAM)

| Service | CPU | Memory | Purpose |
|---------|-----|--------|---------|
| **Main App (app)** | 0.5 CPU | 768MB RAM | Primary bulk upload handler |
| **Secondary (app-1)** | 0.2 CPU | 256MB RAM | Load balancing |
| **PostgreSQL** | 0.8 CPU | 1GB RAM | Database |
| **Redis** | 0.2 CPU | 512MB RAM | Caching |

**Total**: 1.7 CPU cores (85%), 2.5GB RAM (63%) - Leaves room for frontend app

### Quick Start

```bash
# Build and start with load balancing
docker-compose build --no-cache
docker-compose --profile load-balancing up -d

# Access
# Main app: http://localhost:3000
# Secondary: http://localhost:3001
# Admin: http://localhost:3000/admin
```

### Load Balancer (Nginx)

```nginx
upstream go_fiber_backend {
    least_conn;
    server localhost:3000 max_fails=3 fail_timeout=30s;
    server localhost:3001 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    location / {
        proxy_pass http://go_fiber_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 60s;
        proxy_send_timeout 900s;
        proxy_read_timeout 900s;
    }
}
```

## 🚀 Bulk Upload Performance

**Ultra-High-Performance Optimizations:**
- **COPY Protocol**: Direct PostgreSQL COPY for maximum speed
- **Large Chunks**: 1000 products per chunk (increased from 500)
- **8 Concurrent Workers**: Optimized for 2 vCPU environment
- **Parallel Category Processing**: All categories created before bulk upload
- **Memory Optimization**: Streaming file reads for large files
- **Conflict Handling**: ON CONFLICT DO NOTHING for safe concurrent uploads
- **Streaming JSON**: Memory-efficient parsing for files >10MB

**Performance Metrics:**
- **Throughput**: 1000-2000+ products/second (lightning-fast)
- **Memory Usage**: Optimized for 2 vCPU, 4GB RAM instance
- **File Size Support**: Up to 500MB JSON files
- **Timeout**: Dynamic calculation based on file size (10-30 minutes)
- **Processing Time**: 50-80% faster than previous version

## 🔧 Management Commands

```bash
# View logs
docker-compose logs -f app

# Restart services
docker-compose restart

# Stop all services
docker-compose --profile load-balancing down

# Clean up
docker-compose --profile load-balancing down -v
```

## 📈 Health Monitoring

```bash
# Check service health
docker-compose ps

# Monitor resource usage
docker stats

# Check logs for errors
docker-compose logs --tail=100 app
```

## 🛡️ Production Resilience & Recovery

### Key Features
- **Graceful Shutdown**: Proper signal handling with 30-second timeout
- **Database Retry Logic**: Exponential backoff retry (5 attempts)
- **Health Checks**: Comprehensive monitoring for all services
- **Auto-restart**: `unless-stopped` policy for containers
- **Resource Limits**: CPU and memory constraints

### Monitoring Scripts
```bash
# Health monitoring
./scripts/health-monitor.sh --daemon

# Production deployment
./scripts/production-deploy.sh deploy rolling load-balancing

# Manual rollback
./scripts/production-deploy.sh rollback
```

### Failure Recovery
- **Database Failure**: Automatic retry with exponential backoff
- **Redis Failure**: Graceful degradation without caching
- **App Instance Failure**: Automatic container restart
- **System Failure**: Full restart with data persistence

### Environment Variables
```bash
# HTTP Timeouts
READ_TIMEOUT=10m
WRITE_TIMEOUT=15m
IDLE_TIMEOUT=20m
BODY_LIMIT=524288000

# Database Pool
DB_MAX_OPEN_CONNS=150
DB_MAX_IDLE_CONNS=30

# Redis Pool
REDIS_POOL_SIZE=30
REDIS_MIN_IDLE_CONNS=8
```

## 🚨 Troubleshooting

**High Memory Usage**: `docker stats` → `docker-compose restart`
**Bulk Upload Timeout**: Check file size (max 500MB), verify logs
**Database Issues**: `docker-compose logs postgres` → `docker-compose restart postgres`
**Emergency Restart**: `docker-compose down && docker-compose --profile load-balancing up -d`

## 🗄️ Database Performance Optimizations

### 📊 PostgreSQL Tuning
- **Shared Buffers**: 256MB (25% of 1GB RAM)
- **Work Memory**: 16MB for better sort/join performance
- **Effective Cache**: 1GB (75% of available RAM)
- **Slow Query Logging**: Enabled for queries >100ms
- **Bulk Operations**: Optimized for high-throughput uploads

### 🗂️ Database Indexes
**Product Table Indexes:**
- `idx_products_id` - Primary key optimization
- `idx_products_active` - Active products filter
- `idx_products_category_id` - Category filtering
- `idx_products_price` - Price-based queries
- `idx_products_created_at` - Pagination optimization
- `idx_products_combined_fts` - Full-text search

**Composite Indexes:**
- `idx_products_active_category` - Active products by category
- `idx_products_active_created` - Active products by creation date
- `idx_products_pagination` - Optimized pagination

### 📈 Performance Monitoring
```bash
# Monitor database performance
./scripts/db-monitor.sh

# Check slow queries
docker exec -it go-fiber-postgres psql -U user -d testdb -c "SELECT * FROM slow_queries LIMIT 10;"

# View index usage
docker exec -it go-fiber-postgres psql -U user -d testdb -c "SELECT * FROM index_usage_stats;"

# Monitor table statistics
docker exec -it go-fiber-postgres psql -U user -d testdb -c "SELECT * FROM table_stats;"
```

### 🔧 Query Optimization Tips
- **Use Indexed Columns**: Filter by `active`, `category_id`, `created_at`
- **Limit Results**: Always use pagination for large datasets
- **Avoid SELECT ***: Use specific column selection
- **Monitor Slow Queries**: Check logs for queries >100ms
- **Regular Maintenance**: Run VACUUM and ANALYZE periodically
