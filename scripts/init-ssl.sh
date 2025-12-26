#!/bin/bash

# SSL Certificate Initialization Script for Let's Encrypt
# This script sets up SSL certificates for your domain
# After running this script, start the application with:
#   docker compose -f docker-compose.prod.yml --env-file .env up -d

set -e

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please copy env.prod.example to .env and configure your domain and email"
    exit 1
fi

# Load environment variables
set -a
source .env
set +a

# Validate required variables
if [ -z "$DOMAIN" ] || [ "$DOMAIN" = "your-domain.com" ]; then
    echo "Error: Please set DOMAIN in .env file"
    exit 1
fi

if [ -z "$EMAIL" ] || [ "$EMAIL" = "your-email@example.com" ]; then
    echo "Error: Please set EMAIL in .env file"
    exit 1
fi

echo "=========================================="
echo "SSL Certificate Initialization"
echo "=========================================="
echo "Domain: $DOMAIN"
echo "Email: $EMAIL"
echo ""

# Create required directories
echo "Creating certbot directories..."
mkdir -p certbot/conf
mkdir -p certbot/www

# Update nginx configuration with actual domain
echo "Updating Nginx configuration with domain: $DOMAIN"
sed "s/your-domain.com/$DOMAIN/g" nginx/nginx.conf > nginx/nginx.conf.tmp
mv nginx/nginx.conf.tmp nginx/nginx.conf

# Start all services to obtain certificate
echo "Starting services to obtain SSL certificate..."
docker compose -f docker-compose.prod.yml --env-file .env up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 15

# Obtain SSL certificate
echo "Obtaining SSL certificate from Let's Encrypt..."
docker compose -f docker-compose.prod.yml --env-file .env run --rm \
    --entrypoint certbot \
    certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN

# Reload nginx to apply SSL certificate
echo "Reloading Nginx with SSL configuration..."
docker compose -f docker-compose.prod.yml --env-file .env exec nginx nginx -s reload

echo ""
echo "=========================================="
echo "✅ SSL Setup Complete!"
echo "=========================================="
echo ""
echo "Your application is now available at:"
echo "  https://$DOMAIN"
echo ""
echo "SSL certificates will be automatically renewed every 12 hours."
echo ""
echo "To stop the application:"
echo "  docker compose -f docker-compose.prod.yml --env-file .env down"
echo ""
echo "To view logs:"
echo "  docker compose -f docker-compose.prod.yml --env-file .env logs -f"
echo ""
