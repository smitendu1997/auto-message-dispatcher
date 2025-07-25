#!/bin/bash

# Auto-Message-Dispatcher Docker Build Script
set -e

echo "Building Auto-Message-Dispatcher Docker Services..."
cp ./config/.env .env
# Check if .env file exists
if [ ! -f ./config/.env ]; then
    echo "Error: .env file not found. Please copy .env.example to .env and configure it."
    exit 1
fi

# Source environment variables
set -a
source .env
set +a

# Build the application
echo "Building Docker images..."
docker-compose build

echo "Starting services..."
docker-compose up -d

echo "Waiting for services to be healthy..."
sleep 30

echo "Checking service status..."
docker-compose ps

rm .env
echo "Importing database schema..."
if docker exec -i mysql-server mysql -u "$MYSQL_DB_USERNAME" -p"$MYSQL_DB_PASSWORD" "$MYSQL_DB_SCHEMA" < ./models/dbConf/schema.sql; then
    echo "✅ Schema imported successfully!"
else
    echo "❌ Error importing schema!"
fi

echo ""
echo "✅ Auto-Message-Dispatcher services are running!"
echo ""
echo "Services:"
echo "- MySQL: localhost:3306"
echo "- Redis: localhost:6379"
echo "- Auto-Message-Dispatcher Messaging API: localhost:8080"
echo ""
echo "To view logs: docker-compose logs -f"
echo "To stop services: docker-compose down"
echo "To stop and remove volumes: docker-compose down -v"