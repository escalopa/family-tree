#!/bin/bash

# SSL Certificate Initialization Script for Let's Encrypt
# This script sets up SSL certificates for your domain

set -e

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Please copy .env.prod to .env and configure your domain and email"
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

echo "Initializing SSL certificates for domain: $DOMAIN"
echo "Email for notifications: $EMAIL"

# Create required directories
mkdir -p certbot/conf
mkdir -p certbot/www

# Update nginx configuration with actual domain
echo "Updating Nginx configuration with domain: $DOMAIN"
sed "s/your-domain.com/$DOMAIN/g" nginx/nginx-initial.conf > nginx/nginx.conf.tmp
mv nginx/nginx.conf.tmp nginx/nginx.conf

# Start services without SSL first
echo "Starting services without SSL..."
docker compose -f docker-compose.prod.yml --env-file .env up -d postgres redis minio createbuckets migrate backend frontend

# Wait for backend to be healthy
echo "Waiting for backend to be ready..."
sleep 10

# Start nginx with initial configuration
docker compose -f docker-compose.prod.yml --env-file .env up -d nginx

# Wait for nginx to start
echo "Waiting for Nginx to start..."
sleep 5

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

# Update nginx configuration with SSL
echo "Updating Nginx configuration with SSL..."
sed "s/your-domain.com/$DOMAIN/g" nginx/nginx.conf > nginx/nginx.conf.final
mv nginx/nginx.conf.final nginx/nginx.conf

# Restart nginx with SSL configuration
echo "Restarting Nginx with SSL configuration..."
docker compose -f docker-compose.prod.yml --env-file .env restart nginx

# Start certbot renewal service
echo "Starting SSL certificate auto-renewal service..."
docker compose -f docker-compose.prod.yml --env-file .env up -d certbot

echo ""
echo "✅ SSL certificates have been successfully initialized!"
echo ""
echo "Your application is now available at:"
echo "  https://$DOMAIN"
echo ""
echo "SSL certificates will be automatically renewed every 12 hours."
echo ""
