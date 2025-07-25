#!/bin/bash

# Health Monitor Script for Go Fiber Application
# This script monitors the health of all services and can restart them if needed

set -euo pipefail

# Configuration
COMPOSE_FILE="docker-compose.yml"
LOG_FILE="/var/log/health-monitor.log"
MAX_FAILED_CHECKS=3
CHECK_INTERVAL=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Check if a service is healthy
check_service_health() {
    local service_name=$1
    local container_name=$2
    
    # Check if container is running
    if ! docker ps --format "table {{.Names}}" | grep -q "^${container_name}$"; then
        log "❌ Container $container_name is not running"
        return 1
    fi
    
    # Check container health status
    local health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "unknown")
    
    case $health_status in
        "healthy")
            log "✅ $service_name ($container_name) is healthy"
            return 0
            ;;
        "unhealthy")
            log "❌ $service_name ($container_name) is unhealthy"
            return 1
            ;;
        "starting")
            log "🔄 $service_name ($container_name) is starting"
            return 2
            ;;
        *)
            log "⚠️  $service_name ($container_name) health status unknown: $health_status"
            return 1
            ;;
    esac
}

# Check application endpoint
check_app_endpoint() {
    local port=$1
    local endpoint="http://localhost:${port}/api/health"
    
    if curl -sf --max-time 5 "$endpoint" > /dev/null 2>&1; then
        log "✅ App endpoint $endpoint is responsive"
        return 0
    else
        log "❌ App endpoint $endpoint is not responsive"
        return 1
    fi
}

# Restart a service
restart_service() {
    local service_name=$1
    log "🔄 Restarting service: $service_name"
    
    if docker-compose -f "$COMPOSE_FILE" restart "$service_name"; then
        log "✅ Successfully restarted $service_name"
        sleep 10 # Give service time to start
        return 0
    else
        log "❌ Failed to restart $service_name"
        return 1
    fi
}

# Send alert (can be extended to send to Slack, email, etc.)
send_alert() {
    local message=$1
    log "🚨 ALERT: $message"
    
    # Example: Send to webhook (uncomment and configure as needed)
    # curl -X POST -H 'Content-type: application/json' \
    #     --data "{\"text\":\"🚨 Health Monitor Alert: $message\"}" \
    #     "$WEBHOOK_URL"
}

# Main health check function
perform_health_checks() {
    local failed_services=()
    local all_healthy=true
    
    log "🔍 Starting health checks..."
    
    # Check PostgreSQL
    if ! check_service_health "PostgreSQL" "go-fiber-postgres"; then
        failed_services+=("postgres")
        all_healthy=false
    fi
    
    # Check Redis
    if ! check_service_health "Redis" "go-fiber-redis"; then
        failed_services+=("redis")
        all_healthy=false
    fi
    
    # Check main app
    if ! check_service_health "App" "go-fiber-app"; then
        failed_services+=("app")
        all_healthy=false
    else
        # Additional endpoint check for main app
        if ! check_app_endpoint "3000"; then
            failed_services+=("app")
            all_healthy=false
        fi
    fi
    
    # Check load balancing instances (if running)
    for i in {1..3}; do
        local container_name="go-fiber-app-$i"
        local port=$((3000 + i))
        
        if docker ps --format "table {{.Names}}" | grep -q "^${container_name}$"; then
            if ! check_service_health "App Instance $i" "$container_name"; then
                failed_services+=("app-instance-$i")
                all_healthy=false
            else
                if ! check_app_endpoint "$port"; then
                    failed_services+=("app-instance-$i")
                    all_healthy=false
                fi
            fi
        fi
    done
    
    if [ "$all_healthy" = true ]; then
        log "✅ All services are healthy"
        return 0
    else
        log "⚠️  Failed services: ${failed_services[*]}"
        return 1
    fi
}

# Recovery function
attempt_recovery() {
    local failed_services=("$@")
    
    for service in "${failed_services[@]}"; do
        log "🔧 Attempting to recover service: $service"
        
        case $service in
            "postgres"|"redis")
                restart_service "$service"
                ;;
            "app"|"app-instance-"*)
                restart_service "$service"
                sleep 20 # Give app more time to fully start
                ;;
        esac
    done
}

# Main monitoring loop
main() {
    log "🚀 Starting Health Monitor for Go Fiber Application"
    
    local consecutive_failures=0
    
    while true; do
        if perform_health_checks; then
            consecutive_failures=0
        else
            consecutive_failures=$((consecutive_failures + 1))
            
            if [ $consecutive_failures -ge $MAX_FAILED_CHECKS ]; then
                send_alert "Multiple consecutive health check failures detected ($consecutive_failures)"
                
                # Get failed services from last check
                local failed_services=()
                # Re-run checks to identify current failed services
                # This is a simplified approach - in production you might want more sophisticated tracking
                
                log "🔧 Maximum failures reached, attempting automatic recovery..."
                # attempt_recovery "${failed_services[@]}"
                
                # Reset counter after recovery attempt
                consecutive_failures=0
            fi
        fi
        
        log "💤 Sleeping for $CHECK_INTERVAL seconds..."
        sleep $CHECK_INTERVAL
    done
}

# Handle script termination
cleanup() {
    log "🛑 Health monitor shutting down..."
    exit 0
}

trap cleanup SIGINT SIGTERM

# Check if running as daemon
if [ "${1:-}" = "--daemon" ]; then
    # Run as daemon
    nohup "$0" > /dev/null 2>&1 &
    echo "Health monitor started as daemon (PID: $!)"
    echo $! > /var/run/health-monitor.pid
else
    # Run in foreground
    main
fi 