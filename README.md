# High Performance Go Fiber Boilerplate

A high-performance Go API with authentication, products management, and caching. Optimized for bulk uploads with 10K+ products support.

## 🚀 Features

- 🔐 JWT Authentication
- 📦 Product Management with CRUD operations
- 🖼️ Image serving and management
- 🔍 Full-text search capabilities
- 💾 Redis caching for performance
- 📊 Admin dashboard with bulk operations
- 🚀 High-concurrency bulk upload (10K+ products)
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

- **Connection Pooling**: Optimized database and Redis connection pools
- **Caching**: Redis-based caching for frequently accessed data
- **Indexing**: Full-text search indexes and performance indexes
- **Chunked Processing**: Large datasets processed in manageable chunks
- **High Concurrency**: Multiple workers processing simultaneously
- **COPY Protocol**: Direct database insertion bypassing ORM overhead

## 🐳 Docker Deployment

### Resource Allocation (2 vCPU, 4GB RAM)

| Service | CPU | Memory | Purpose |
|---------|-----|--------|---------|
| **Main App (app)** | 0.6 CPU | 768MB RAM | Primary bulk upload handler |
| **Secondary (app-1)** | 0.2 CPU | 256MB RAM | Load balancing |
| **PostgreSQL** | 0.7 CPU | 768MB RAM | Database |
| **Redis** | 0.3 CPU | 512MB RAM | Caching |

**Total**: 1.8 CPU cores (90%), 3.1GB RAM (78%)

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
- **Large Chunks**: 500 products per chunk (increased from 100)
- **6 Concurrent Workers**: Optimized for 2 vCPU environment
- **Pre-category Insertion**: All categories created before bulk upload
- **Memory Optimization**: Streaming file reads for large files
- **Conflict Handling**: ON CONFLICT DO NOTHING for safe concurrent uploads

**Performance Metrics:**
- **Throughput**: 500-1000+ products/second
- **Memory Usage**: Optimized for 2 vCPU, 4GB RAM instance
- **File Size Support**: Up to 500MB JSON files
- **Timeout**: Dynamic calculation based on file size (10-30 minutes)

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
