#!/bin/bash

# Maintenance Script for Family Tree Application
# Provides common maintenance tasks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to show menu
show_menu() {
    echo ""
    echo "========================================="
    echo "   Family Tree Maintenance Menu"
    echo "========================================="
    echo "1. View service status"
    echo "2. View logs (all services)"
    echo "3. View logs (specific service)"
    echo "4. Restart all services"
    echo "5. Restart specific service"
    echo "6. Update application"
    echo "7. Backup database and files"
    echo "8. Clean up Docker resources"
    echo "9. Check SSL certificate status"
    echo "10. Renew SSL certificates"
    echo "11. View disk usage"
    echo "12. Database shell access"
    echo "13. Redis shell access"
    echo "0. Exit"
    echo "========================================="
    echo -n "Select an option: "
}

# Function to view service status
view_status() {
    print_info "Service Status:"
    docker compose -f docker-compose.prod.yml --env-file .env ps
}

# Function to view all logs
view_all_logs() {
    print_info "Viewing all logs (Press Ctrl+C to exit)..."
    docker compose -f docker-compose.prod.yml --env-file .env logs -f --tail=100
}

# Function to view specific service logs
view_service_logs() {
    echo "Available services: backend, frontend, nginx, postgres, redis, minio"
    read -p "Enter service name: " service
    print_info "Viewing logs for $service (Press Ctrl+C to exit)..."
    docker compose -f docker-compose.prod.yml --env-file .env logs -f --tail=100 $service
}

# Function to restart all services
restart_all() {
    print_warning "Restarting all services..."
    docker compose -f docker-compose.prod.yml --env-file .env restart
    print_info "All services restarted"
}

# Function to restart specific service
restart_service() {
    echo "Available services: backend, frontend, nginx, postgres, redis, minio"
    read -p "Enter service name: " service
    print_info "Restarting $service..."
    docker compose -f docker-compose.prod.yml --env-file .env restart $service
    print_info "$service restarted"
}

# Function to update application
update_app() {
    print_warning "Updating application..."
    if [ -f scripts/deploy.sh ]; then
        ./scripts/deploy.sh
    else
        print_error "deploy.sh not found"
    fi
}

# Function to backup
backup() {
    print_info "Starting backup..."
    if [ -f scripts/backup.sh ]; then
        ./scripts/backup.sh
    else
        print_error "backup.sh not found"
    fi
}

# Function to clean up Docker
cleanup_docker() {
    print_warning "Cleaning up Docker resources..."
    echo "This will remove:"
    echo "  - Stopped containers"
    echo "  - Unused networks"
    echo "  - Dangling images"
    echo "  - Build cache"
    read -p "Continue? (y/N): " confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        docker system prune -f
        print_info "Cleanup completed"
    else
        print_info "Cleanup cancelled"
    fi
}

# Function to check SSL status
check_ssl() {
    print_info "Checking SSL certificate status..."
    if [ -d certbot/conf/live ]; then
        for cert in certbot/conf/live/*/cert.pem; do
            domain=$(basename $(dirname $cert))
            expiry=$(openssl x509 -enddate -noout -in $cert | cut -d= -f2)
            print_info "Domain: $domain"
            print_info "Expires: $expiry"
        done
    else
        print_warning "No SSL certificates found"
    fi
}

# Function to renew SSL
renew_ssl() {
    print_info "Renewing SSL certificates..."
    docker compose -f docker-compose.prod.yml --env-file .env run --rm certbot renew
    docker compose -f docker-compose.prod.yml --env-file .env restart nginx
    print_info "SSL certificates renewed and Nginx restarted"
}

# Function to check disk usage
check_disk() {
    print_info "Disk Usage:"
    df -h /
    echo ""
    print_info "Docker Disk Usage:"
    docker system df
}

# Function to access database shell
db_shell() {
    print_info "Connecting to PostgreSQL..."
    docker compose -f docker-compose.prod.yml --env-file .env exec postgres psql -U familytree -d familytree
}

# Function to access Redis shell
redis_shell() {
    print_info "Connecting to Redis..."
    docker compose -f docker-compose.prod.yml --env-file .env exec redis redis-cli
}

# Main loop
while true; do
    show_menu
    read choice

    case $choice in
        1) view_status ;;
        2) view_all_logs ;;
        3) view_service_logs ;;
        4) restart_all ;;
        5) restart_service ;;
        6) update_app ;;
        7) backup ;;
        8) cleanup_docker ;;
        9) check_ssl ;;
        10) renew_ssl ;;
        11) check_disk ;;
        12) db_shell ;;
        13) redis_shell ;;
        0)
            print_info "Goodbye!"
            exit 0
            ;;
        *)
            print_error "Invalid option. Please try again."
            ;;
    esac

    echo ""
    read -p "Press Enter to continue..."
done
