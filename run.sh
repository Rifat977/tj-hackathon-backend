#!/bin/bash

# High Performance Go API Runner

echo "🚀 Starting High Performance Go API..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from example..."
    cp .env.example .env
    echo "📝 Please edit .env file with your configuration"
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo "📦 Installing air for hot reloading..."
    go install github.com/cosmtrek/air@latest
fi

# Run the application with air for hot reloading
echo "🎯 Starting server with hot reload on :3000..."
air 