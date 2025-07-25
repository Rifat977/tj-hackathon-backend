#!/bin/bash

# Production Deployment Script for Go Fiber Application
# Supports zero-downtime deployment with health checks and rollback

set -euo pipefail

# Configuration
COMPOSE_FILE="docker-compose.yml"
BACKUP_DIR="./backups"
LOG_FILE="./logs/deployment.log"
HEALTH_CHECK_TIMEOUT=120
ROLLBACK_ON_FAILURE=true

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging function
log() {
    local level=${2:-INFO}
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} ${level}: $1" | tee -a "$LOG_FILE"
}

error() {
    log "${RED}$1${NC}" "ERROR"
}

success() {
    log "${GREEN}$1${NC}" "SUCCESS"
}

warning() {
    log "${YELLOW}$1${NC}" "WARNING"
}

# Create necessary directories
setup_directories() {
    mkdir -p "$BACKUP_DIR" "$(dirname "$LOG_FILE")" scripts logs
}

# Backup current deployment
backup_current_deployment() {
    local backup_name="backup-$(date +%Y%m%d-%H%M%S)"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    log "Creating backup: $backup_name"
    mkdir -p "$backup_path"
    
    # Backup docker-compose file
    cp "$COMPOSE_FILE" "$backup_path/"
    
    # Export current images
    log "Exporting current Docker images..."
    docker images --format "table {{.Repository}}:{{.Tag}}" | grep "go-fiber" | while read image; do
        if [ "$image" != "REPOSITORY:TAG" ]; then
            local image_file=$(echo "$image" | tr '/:' '_')
            docker save "$image" > "$backup_path/${image_file}.tar"
        fi
    done
    
    # Create backup metadata
    cat > "$backup_path/metadata.txt" << EOF
Backup Created: $(date)
Git Commit: $(git rev-parse HEAD 2>/dev/null || echo "unknown")
Docker Compose Version: $(docker-compose version --short)
Running Containers: $(docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}" | grep go-fiber || echo "none")
EOF
    
    success "Backup created: $backup_path"
    echo "$backup_path" > "$BACKUP_DIR/latest-backup.txt"
}

# Health check function
wait_for_service_health() {
    local service_name=$1
    local container_name=$2
    local timeout=${3:-$HEALTH_CHECK_TIMEOUT}
    local interval=5
    local elapsed=0
    
    log "Waiting for $service_name to become healthy (timeout: ${timeout}s)"
    
    while [ $elapsed -lt $timeout ]; do
        local health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "unknown")
        
        case $health_status in
            "healthy")
                success "$service_name is healthy"
                return 0
                ;;
            "unhealthy")
                error "$service_name is unhealthy"
                return 1
                ;;
            "starting")
                log "$service_name is starting... (${elapsed}s elapsed)"
                ;;
            *)
                warning "$service_name health status unknown: $health_status"
                ;;
        esac
        
        sleep $interval
        elapsed=$((elapsed + interval))
    done
    
    error "Timeout waiting for $service_name to become healthy"
    return 1
}

# Validate application endpoints
validate_endpoints() {
    local endpoints=(
        "http://localhost:3000/api/health"
        "http://localhost:3001/api/health"
        "http://localhost:3002/api/health"
        "http://localhost:3003/api/health"
    )
    
    log "Validating application endpoints..."
    
    for endpoint in "${endpoints[@]}"; do
        local port=$(echo "$endpoint" | cut -d: -f3 | cut -d/ -f1)
        local container_name="go-fiber-app"
        
        if [ "$port" != "3000" ]; then
            local instance_num=$((port - 3000))
            container_name="go-fiber-app-$instance_num"
        fi
        
        # Check if container exists and is running
        if docker ps --format "{{.Names}}" | grep -q "^${container_name}$"; then
            if curl -sf --max-time 10 "$endpoint" > /dev/null 2>&1; then
                success "Endpoint $endpoint is responsive"
            else
                error "Endpoint $endpoint is not responsive"
                return 1
            fi
        else
            log "Container $container_name not running, skipping endpoint $endpoint"
        fi
    done
    
    success "All active endpoints are responsive"
}

# Rolling deployment for zero downtime
rolling_deployment() {
    local profile=${1:-"load-balancing"}
    
    log "Starting rolling deployment with profile: $profile"
    
    # Get list of app containers
    local containers=()
    if [ "$profile" = "load-balancing" ]; then
        containers=("app" "app-instance-1" "app-instance-2" "app-instance-3")
    else
        containers=("app")
    fi
    
    # Update containers one by one
    for container in "${containers[@]}"; do
        log "Updating container: $container"
        
        # Stop the specific container
        docker-compose -f "$COMPOSE_FILE" --profile "$profile" stop "$container" || true
        
        # Remove the container
        docker-compose -f "$COMPOSE_FILE" --profile "$profile" rm -f "$container" || true
        
        # Start the updated container
        docker-compose -f "$COMPOSE_FILE" --profile "$profile" up -d "$container"
        
        # Wait for health check
        local container_name="go-fiber-$container"
        if [ "$container" = "app" ]; then
            container_name="go-fiber-app"
        fi
        
        if ! wait_for_service_health "$container" "$container_name" 60; then
            error "Failed to start $container"
            return 1
        fi
        
        # Brief pause between container updates
        sleep 5
    done
    
    success "Rolling deployment completed"
}

# Rollback to previous deployment
rollback_deployment() {
    local backup_path=$(cat "$BACKUP_DIR/latest-backup.txt" 2>/dev/null || echo "")
    
    if [ -z "$backup_path" ] || [ ! -d "$backup_path" ]; then
        error "No backup found for rollback"
        return 1
    fi
    
    warning "Rolling back to backup: $(basename "$backup_path")"
    
    # Stop current services
    docker-compose -f "$COMPOSE_FILE" down
    
    # Restore docker-compose file
    cp "$backup_path/$COMPOSE_FILE" "./"
    
    # Load backup images
    for image_file in "$backup_path"/*.tar; do
        if [ -f "$image_file" ]; then
            log "Loading image: $(basename "$image_file")"
            docker load < "$image_file"
        fi
    done
    
    # Start services
    docker-compose -f "$COMPOSE_FILE" --profile load-balancing up -d
    
    success "Rollback completed"
}

# Pre-deployment checks
pre_deployment_checks() {
    log "Running pre-deployment checks..."
    
    # Check Docker and Docker Compose
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed or not in PATH"
        return 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed or not in PATH"
        return 1
    fi
    
    # Check if compose file exists
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "Docker Compose file not found: $COMPOSE_FILE"
        return 1
    fi
    
    # Check available disk space (at least 2GB)
    local available_space=$(df . | tail -1 | awk '{print $4}')
    if [ "$available_space" -lt 2097152 ]; then
        error "Insufficient disk space (need at least 2GB)"
        return 1
    fi
    
    # Validate compose file
    if ! docker-compose -f "$COMPOSE_FILE" config > /dev/null 2>&1; then
        error "Invalid Docker Compose file"
        return 1
    fi
    
    success "Pre-deployment checks passed"
}

# Post-deployment validation
post_deployment_validation() {
    log "Running post-deployment validation..."
    
    # Wait for all services to be healthy
    if ! wait_for_service_health "PostgreSQL" "go-fiber-postgres" 60; then
        return 1
    fi
    
    if ! wait_for_service_health "Redis" "go-fiber-redis" 30; then
        return 1
    fi
    
    if ! wait_for_service_health "Main App" "go-fiber-app" 60; then
        return 1
    fi
    
    # Validate endpoints
    if ! validate_endpoints; then
        return 1
    fi
    
    success "Post-deployment validation passed"
}

# Main deployment function
deploy() {
    local deployment_type=${1:-"rolling"}
    local profile=${2:-"load-balancing"}
    
    log "Starting deployment (type: $deployment_type, profile: $profile)"
    
    # Setup
    setup_directories
    
    # Pre-deployment checks
    if ! pre_deployment_checks; then
        error "Pre-deployment checks failed"
        exit 1
    fi
    
    # Create backup
    backup_current_deployment
    
    # Build new images
    log "Building new Docker images..."
    if ! docker-compose -f "$COMPOSE_FILE" build --no-cache; then
        error "Failed to build Docker images"
        exit 1
    fi
    
    # Deploy based on type
    case $deployment_type in
        "rolling")
            if ! rolling_deployment "$profile"; then
                if [ "$ROLLBACK_ON_FAILURE" = true ]; then
                    warning "Deployment failed, initiating rollback..."
                    rollback_deployment
                fi
                exit 1
            fi
            ;;
        "blue-green")
            error "Blue-green deployment not yet implemented"
            exit 1
            ;;
        "recreate")
            log "Recreating all services..."
            docker-compose -f "$COMPOSE_FILE" --profile "$profile" up -d --force-recreate
            ;;
        *)
            error "Unknown deployment type: $deployment_type"
            exit 1
            ;;
    esac
    
    # Post-deployment validation
    if ! post_deployment_validation; then
        if [ "$ROLLBACK_ON_FAILURE" = true ]; then
            warning "Post-deployment validation failed, initiating rollback..."
            rollback_deployment
        fi
        exit 1
    fi
    
    success "Deployment completed successfully!"
    log "Services status:"
    docker-compose -f "$COMPOSE_FILE" ps
}

# Script usage
usage() {
    cat << EOF
Production Deployment Script for Go Fiber Application

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    deploy [rolling|recreate] [single|load-balancing]  Deploy the application
    rollback                                           Rollback to previous deployment
    health-check                                       Check health of all services
    backup                                             Create a backup of current deployment

Examples:
    $0 deploy rolling load-balancing    # Rolling deployment with load balancing
    $0 deploy recreate single           # Recreate deployment with single instance
    $0 rollback                         # Rollback to previous deployment
    $0 health-check                     # Check service health

EOF
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        deploy "${2:-rolling}" "${3:-load-balancing}"
        ;;
    "rollback")
        rollback_deployment
        ;;
    "health-check")
        if post_deployment_validation; then
            success "All services are healthy"
        else
            error "Some services are unhealthy"
            exit 1
        fi
        ;;
    "backup")
        setup_directories
        backup_current_deployment
        ;;
    "help"|"--help"|"-h")
        usage
        ;;
    *)
        error "Unknown command: $1"
        usage
        exit 1
        ;;
esac 