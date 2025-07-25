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

## 🚀 Docker Deployment (4 Instances, Redis, PostgreSQL)

This project supports easy deployment using Docker Compose, including 4 app instances for load balancing, Redis for caching, and PostgreSQL for the database.

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) installed

### Steps

1. **Clone the repository:**
   ```bash
   git clone <your-repo-url>
   cd go-fiber-boilerplate
   ```

2. **(Optional) Update environment variables:**
   - Edit `docker-compose.yml` if you want to change default passwords or ports.

3. **Build Docker images:**
   - Standard build (uses cache):
     ```bash
     docker-compose build
     ```
   - Build without cache (recommended if you want a fresh build):
     ```bash
     docker-compose build --no-cache
     ```

4. **Start all services (4 app instances, Redis, PostgreSQL):**
   ```bash
   docker-compose up -d --profile load-balancing
   ```
   This will start:
   - 4 Go Fiber app instances (on ports 3000, 3001, 3002, 3003)
   - Redis (port 6379, password: `your_redis_password`)
   - PostgreSQL (port 5433, user: `user`, password: `password`, db: `testdb`)

5. **Check running containers:**
   ```bash
   docker ps
   ```

6. **View logs:**
   - For all services:
     ```bash
     docker-compose logs -f
     ```
   - For a specific service (e.g., app):
     ```bash
     docker-compose logs -f app
     ```

7. **Restart services:**
   - Restart all services:
     ```bash
     docker-compose restart
     ```
   - Restart a specific service:
     ```bash
     docker-compose restart app
     ```

8. **Access the app:**
   - Main instance: [http://localhost:3000](http://localhost:3000)
   - Other instances: [http://localhost:3001](http://localhost:3001), etc.

9. **Stop all services:**
   ```bash
   docker-compose down
   ```

10. **Remove all containers, networks, and volumes (clean up everything):**
    ```bash
    docker-compose down -v
    ```

### Notes
- The app instances share the same database and Redis cache.
- You can scale the number of app instances by editing `docker-compose.yml`.
- For production, use a reverse proxy (e.g., Nginx, Traefik) for load balancing.
- Data for Redis and PostgreSQL is persisted in Docker volumes (`redis-data`, `postgres-data`).
- If you change dependencies or want to ensure a clean build, use `docker-compose build --no-cache` before starting services.

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
