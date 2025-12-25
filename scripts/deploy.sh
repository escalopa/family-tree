#!/bin/bash

# Production Deployment Script
# This script deploys the application to production

set -e

echo "🚀 Starting production deployment..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please copy .env.prod to .env and configure your settings"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

# Validate required variables
if [ -z "$DOMAIN" ] || [ "$DOMAIN" = "your-domain.com" ]; then
    echo "Error: Please set DOMAIN in .env file"
    exit 1
fi

echo "Deploying to domain: $DOMAIN"

# Pull latest code (if using git)
if [ -d .git ]; then
    echo "📥 Pulling latest code..."
    git pull
fi

# Update backend configuration with production values
echo "⚙️  Updating backend configuration..."
if [ -f be/config.yaml ]; then
    # Update allowed origins
    sed -i.bak "s|http://localhost:3000|https://$DOMAIN|g" be/config.yaml
    sed -i.bak "s|http://localhost:5173|https://$DOMAIN|g" be/config.yaml

    # Update redirect base URL
    sed -i.bak "s|redirect_base_url:.*|redirect_base_url: \"https://$DOMAIN\"|g" be/config.yaml

    # Update cookie settings for production
    sed -i.bak "s|secure: false|secure: true|g" be/config.yaml
    sed -i.bak "s|enable_hsts: false|enable_hsts: true|g" be/config.yaml

    # Update server mode to release
    sed -i.bak "s|mode: \"debug\"|mode: \"release\"|g" be/config.yaml

    # Update database connection
    sed -i.bak "s|host=localhost|host=postgres|g" be/config.yaml

    # Update redis connection
    sed -i.bak "s|redis://localhost:6379|redis://${REDIS_PASSWORD:+:$REDIS_PASSWORD@}redis:6379|g" be/config.yaml

    rm -f be/config.yaml.bak
fi

# Build and start services
echo "🏗️  Building Docker images..."
docker-compose -f docker-compose.prod.yml build --no-cache

echo "🔄 Starting services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
sleep 15

# Check service health
echo "🔍 Checking service health..."
docker-compose -f docker-compose.prod.yml ps

# Show logs
echo ""
echo "📋 Recent logs:"
docker-compose -f docker-compose.prod.yml logs --tail=50

echo ""
echo "✅ Deployment completed successfully!"
echo ""
echo "Your application is available at:"
echo "  https://$DOMAIN"
echo ""
echo "To view logs, run:"
echo "  docker-compose -f docker-compose.prod.yml logs -f [service-name]"
echo ""
echo "To stop services, run:"
echo "  docker-compose -f docker-compose.prod.yml down"
echo ""
