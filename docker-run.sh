#!/bin/bash

# Docker Run Script for Go Fiber Application

echo "🐳 Go Fiber Application Docker Runner"
echo "======================================"

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTION]"
    echo ""
    echo "Options:"
    echo "  single     Run single instance (port 3000)"
    echo "  load       Run multiple instances for load balancing (ports 3000-3003)"
    echo "  dev        Run development mode with hot reload (port 3000)"
    echo "  stop       Stop all containers"
    echo "  logs       Show logs from all containers"
    echo "  clean      Stop and remove all containers and images"
    echo ""
    echo "Examples:"
    echo "  $0 single"
    echo "  $0 load"
    echo "  $0 dev"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo "❌ Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Function to build if image doesn't exist
build_if_needed() {
    if ! docker image inspect go-fiber-boilerplate-app:latest > /dev/null 2>&1; then
        echo "📦 Building Docker image..."
        docker compose build
    fi
}

case "$1" in
    "single")
        check_docker
        build_if_needed
        echo "🚀 Starting single instance on port 3000..."
        docker compose up -d app
        echo "✅ Application is running on http://localhost:3000"
        ;;
    "load")
        check_docker
        build_if_needed
        echo "🔄 Starting multiple instances for load balancing..."
        echo "   Instance 1: http://localhost:3000"
        echo "   Instance 2: http://localhost:3001"
        echo "   Instance 3: http://localhost:3002"
        echo "   Instance 4: http://localhost:3003"
        docker compose --profile load-balancing up -d
        echo "✅ Load balancing setup complete!"
        echo "💡 You can now configure your load balancer to distribute traffic across these ports"
        ;;
    "dev")
        check_docker
        build_if_needed
        echo "🛠️  Starting development mode with hot reload..."
        docker compose --profile development up -d
        echo "✅ Development server is running on http://localhost:3000"
        echo "🔄 Hot reload is enabled - changes will be reflected automatically"
        ;;
    "stop")
        echo "🛑 Stopping all containers..."
        docker compose down
        echo "✅ All containers stopped"
        ;;
    "logs")
        echo "📋 Showing logs from all containers..."
        docker compose logs -f
        ;;
    "clean")
        echo "🧹 Cleaning up Docker resources..."
        docker compose down --rmi all --volumes --remove-orphans
        docker system prune -f
        echo "✅ Cleanup complete"
        ;;
    *)
        show_usage
        exit 1
        ;;
esac