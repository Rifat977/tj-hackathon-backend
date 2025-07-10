#!/bin/bash

# High Performance Go API Production Build

echo "🚀 Building High Performance Go API for production..."

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

# Build the application
echo "🔨 Building application..."
go build -o main .

# Run the application
echo "🎯 Starting server on :3000..."
./main 