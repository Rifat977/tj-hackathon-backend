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

## 🚀 Docker Deployment (2 Instances, Redis, PostgreSQL) - Optimized for 2 vCPU, 4GB RAM

This project supports easy deployment using Docker Compose, including 2 app instances optimized for least_connection load balancing, Redis for caching, and PostgreSQL for the database.

### Prerequisites
- Docker and Docker Compose installed
- 2 vCPU, 4GB RAM minimum (Google Cloud Compute e2-medium or similar)
- At least 10GB free disk space

### Resource Allocation (Optimized for 2 vCPU, 4GB RAM)

| Service | CPU | Memory | Purpose |
|---------|-----|--------|---------|
| **Main App (app)** | 0.6 CPU | 768MB RAM | Primary bulk upload handler |
| **Secondary (app-1)** | 0.2 CPU | 256MB RAM | Load balancing (least_connection) |
| **PostgreSQL** | 0.7 CPU | 768MB RAM | Database with bulk upload optimization |
| **Redis** | 0.3 CPU | 512MB RAM | Caching and session storage |
| **System Overhead** | 0.2 CPU | 0.7GB RAM | OS and Docker overhead |

**Total Usage**: 1.8 CPU cores (90%), 3.1GB RAM (78%) - ✅ **Safe for 2 vCPU, 4GB RAM**

### Quick Start

1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd go-fiber-boilerplate
   ```

2. **Build and start services**:
   ```bash
   # Build all services
   docker-compose build --no-cache
   
   # Start with load balancing (2 instances)
   docker-compose --profile load-balancing up -d
   
   # Or start single instance for development
   docker-compose up -d
   ```

3. **Access your application**:
   - Main app: http://localhost:3000
   - Secondary app: http://localhost:3001 (load balancing)
   - Admin dashboard: http://localhost:3000/admin
   - Database: localhost:5433
   - Redis: localhost:6379

### Load Balancer Configuration

For **least_connection** load balancing, use this nginx configuration:

```nginx
upstream go_fiber_backend {
    least_conn;
    server localhost:3000 max_fails=3 fail_timeout=30s;
    server localhost:3001 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://go_fiber_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for bulk uploads
        proxy_connect_timeout 60s;
        proxy_send_timeout 900s;
        proxy_read_timeout 900s;
    }
}
```

### Bulk Upload Performance

**Optimized for 10K+ products**:
- **File size limit**: 500MB
- **Processing time**: 45-60 seconds for 10K products
- **Memory usage**: ~2.5GB peak during bulk upload
- **Workers**: 4 concurrent workers (optimized for 2 vCPU)
- **Chunk size**: 100 products per chunk
- **Success rate**: >99% with deadlock prevention

### Environment Variables

```bash
# HTTP Server Timeouts - Optimized for Bulk Uploads
READ_TIMEOUT=10m          # 10 minutes for large file reads
WRITE_TIMEOUT=15m         # 15 minutes for bulk operations
IDLE_TIMEOUT=20m          # 20 minutes idle
BODY_LIMIT=524288000      # 500MB for large JSON files

# Database Connection Pool - Optimized for High Concurrency
DB_MAX_OPEN_CONNS=150     # Primary instance
DB_MAX_IDLE_CONNS=30      # Primary instance
DB_CONN_MAX_LIFETIME=1h   # 1 hour connection lifetime
DB_CONN_MAX_IDLE_TIME=30m # 30 minutes idle time

# Redis Timeouts - Optimized for Caching
REDIS_POOL_SIZE=30        # Primary instance
REDIS_MIN_IDLE_CONNS=8    # Primary instance
REDIS_DIAL_TIMEOUT=10s    # 10 seconds dial timeout
REDIS_READ_TIMEOUT=5s     # 5 seconds read timeout
REDIS_WRITE_TIMEOUT=5s    # 5 seconds write timeout
REDIS_POOL_TIMEOUT=10s    # 10 seconds pool timeout
```

### Management Commands

```bash
# View logs
docker-compose logs -f app

# Restart services
docker-compose restart

# Scale services (if needed)
docker-compose up -d --scale app=1 --scale app-instance-1=1

# Stop all services
docker-compose --profile load-balancing down

# Clean up (removes volumes)
docker-compose --profile load-balancing down -v
```

### Health Monitoring

```bash
# Check service health
docker-compose ps

# Monitor resource usage
docker stats

# Check logs for errors
docker-compose logs --tail=100 app
```

### Troubleshooting

**High Memory Usage**:
- Monitor with `docker stats`
- Restart services if needed: `docker-compose restart`

**Bulk Upload Timeout**:
- Check file size (max 500MB)
- Verify network connectivity
- Monitor logs: `docker-compose logs -f app`

**Database Connection Issues**:
- Check PostgreSQL health: `docker-compose logs postgres`
- Verify connection pool settings
- Restart database: `docker-compose restart postgres`

## 🛡️ Production Resilience & Recovery

This application is designed with comprehensive production resilience features to ensure high availability, graceful failure handling, and automatic recovery.

### 🚀 Resilience Features

#### **Process Management**
- **Graceful Shutdown**: Proper signal handling (SIGTERM, SIGINT, SIGQUIT) with 30-second timeout
- **Database Connection Cleanup**: Automatic closure of database pools and connections
- **Panic Recovery**: Fiber middleware to catch and handle panics without crashing
- **Init System**: Tini process manager in Docker for proper signal handling and zombie reaping

#### **Connection Resilience**
- **Database Retry Logic**: Exponential backoff retry (5 attempts, 2-32s intervals)
- **Connection Pooling**: Optimized connection pools for PostgreSQL and Redis
- **Health Checks**: Comprehensive health monitoring for all services
- **Dependency Management**: Proper service startup order with health validation

#### **Resource Management**
- **Memory Limits**: Container resource limits and reservations
- **CPU Limits**: Proper CPU allocation to prevent resource starvation
- **Storage Persistence**: Docker volumes for data persistence across restarts

### 📊 Monitoring & Health Checks

#### **Service Health Checks**
- **PostgreSQL**: Custom health check with connection validation
- **Redis**: Redis-cli ping with authentication
- **Application**: HTTP endpoint monitoring with timeout handling
- **Container Health**: Docker health check with configurable intervals

#### **Automated Monitoring**
Run the health monitor daemon:
```bash
./scripts/health-monitor.sh --daemon
```

Features:
- Continuous health monitoring (30-second intervals)
- Automatic service restart on failure
- Alert system (configurable for Slack/email)
- Detailed logging with timestamps

#### **Manual Health Check**
```bash
# Check all services
./scripts/production-deploy.sh health-check

# Check specific service
docker-compose ps
docker inspect --format='{{.State.Health.Status}}' go-fiber-app
```

### 🔄 Zero-Downtime Deployment

#### **Rolling Deployment**
Deploy without downtime using the production deployment script:
```bash
# Rolling deployment with load balancing
./scripts/production-deploy.sh deploy rolling load-balancing

# Single instance deployment
./scripts/production-deploy.sh deploy rolling single
```

#### **Deployment Features**
- **Pre-deployment Validation**: Disk space, Docker setup, compose file validation
- **Automatic Backup**: Creates backup before deployment
- **Health Validation**: Waits for services to be healthy before proceeding
- **Endpoint Testing**: Validates all application endpoints
- **Automatic Rollback**: Rolls back on failure (configurable)

#### **Manual Rollback**
```bash
./scripts/production-deploy.sh rollback
```

### 🗂️ Backup & Recovery

#### **Automatic Backups**
Backups are created automatically during deployments and include:
- Docker Compose configuration
- Docker images export
- Deployment metadata (Git commit, timestamps)

#### **Manual Backup**
```bash
./scripts/production-deploy.sh backup
```

#### **Data Persistence**
All critical data is persisted in Docker volumes:
- `postgres-data`: Database data
- `redis-data`: Cache data with persistence
- `app-logs-*`: Application logs for debugging

### ⚙️ Production Configuration

#### **Environment Variables for Resilience**
```bash
# Server Timeouts
READ_TIMEOUT=5m
WRITE_TIMEOUT=10m
IDLE_TIMEOUT=15m

# Database Connection Pool
DB_MAX_OPEN_CONNS=200
DB_MAX_IDLE_CONNS=50
DB_CONN_MAX_LIFETIME=2h
DB_CONN_MAX_IDLE_TIME=1h

# Redis Timeouts
REDIS_DIAL_TIMEOUT=10s
REDIS_READ_TIMEOUT=5s
REDIS_WRITE_TIMEOUT=5s
REDIS_POOL_TIMEOUT=10s
```

#### **Docker Compose Resilience**
- **Restart Policy**: `unless-stopped` for automatic restart
- **Health Checks**: 20s intervals with 5s timeout
- **Service Dependencies**: Proper startup order with health conditions
- **Resource Limits**: CPU and memory constraints
- **Network Isolation**: Custom bridge network for security

### 🚨 Failure Scenarios & Recovery

#### **Database Connection Failure**
- **Detection**: Health checks fail, connection errors logged
- **Recovery**: Automatic retry with exponential backoff
- **Fallback**: Application continues serving cached data where possible

#### **Redis Connection Failure**
- **Detection**: Redis ping failures, connection timeouts
- **Recovery**: Automatic reconnection with configured timeouts
- **Fallback**: Application degrades gracefully without caching

#### **Application Instance Failure**
- **Detection**: Health check failures, unresponsive endpoints
- **Recovery**: Automatic container restart via Docker
- **Load Balancing**: Traffic routed to healthy instances

#### **Complete System Failure**
- **Recovery**: Full system restart with data persistence
- **Backup Restoration**: Automated rollback to last known good state
- **Data Integrity**: PostgreSQL ensures ACID compliance

### 📈 Performance & Scaling

#### **Horizontal Scaling**
Scale application instances:
```bash
# Start with load balancing (4 instances)
docker-compose --profile load-balancing up -d

# Scale specific service
docker-compose up -d --scale app=6
```

#### **Resource Monitoring**
```bash
# Monitor resource usage
docker stats

# Check container logs
docker-compose logs -f --tail=100

# Monitor database performance
docker exec go-fiber-postgres psql -U user -d testdb -c "SELECT * FROM public.health_check();"
```

### 🔧 Troubleshooting

#### **Common Issues**
1. **Port Conflicts**: Ensure ports 3000-3003, 5433, 6379 are available
2. **Memory Issues**: Check Docker resource limits and system memory
3. **Permission Issues**: Ensure proper file permissions for scripts
4. **Network Issues**: Verify Docker network configuration

#### **Debug Commands**
```bash
# Check service health
docker-compose ps
docker inspect go-fiber-app | grep -A 10 Health

# View detailed logs
docker-compose logs --tail=50 app
docker-compose logs --tail=50 postgres
docker-compose logs --tail=50 redis

# Test database connection
docker exec go-fiber-postgres pg_isready -U user -d testdb

# Test Redis connection
docker exec go-fiber-redis redis-cli -a your_redis_password ping

# Check application endpoints
curl -i http://localhost:3000/api/health
```

#### **Emergency Procedures**
```bash
# Force restart all services
docker-compose down && docker-compose --profile load-balancing up -d

# Clean restart (removes containers but keeps data)
docker-compose down -v && docker-compose --profile load-balancing up -d

# Complete reset (WARNING: removes all data)
docker-compose down -v --remove-orphans
docker system prune -f
```

### 🔐 Security Considerations

- **Non-root User**: Applications run as non-root user in containers
- **Network Isolation**: Services communicate through isolated Docker network
- **Secret Management**: Use environment variables for sensitive data
- **Resource Limits**: Prevents resource exhaustion attacks
- **Health Monitoring**: Detects and responds to anomalous behavior
