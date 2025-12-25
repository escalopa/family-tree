#!/bin/bash

# Backup Script for Production Data
# This script backs up the database, uploaded files, and configurations

set -e

BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_PATH="$BACKUP_DIR/backup_$TIMESTAMP"

echo "🗄️  Starting backup..."

# Create backup directory
mkdir -p "$BACKUP_PATH"

# Load environment variables if available
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Backup PostgreSQL database
echo "📦 Backing up PostgreSQL database..."
docker exec postgres pg_dump -U ${POSTGRES_USER:-familytree} ${POSTGRES_DB:-familytree} | gzip > "$BACKUP_PATH/postgres_backup.sql.gz"

# Backup MinIO data (if container is running)
if [ "$(docker ps -q -f name=minio)" ]; then
    echo "📦 Backing up MinIO data..."
    docker exec minio tar czf - /data | cat > "$BACKUP_PATH/minio_backup.tar.gz"
fi

# Backup configuration files
echo "📦 Backing up configuration files..."
cp be/config.yaml "$BACKUP_PATH/config.yaml" 2>/dev/null || echo "Warning: config.yaml not found"
cp .env "$BACKUP_PATH/.env" 2>/dev/null || echo "Warning: .env not found"

# Backup SSL certificates
if [ -d certbot/conf ]; then
    echo "📦 Backing up SSL certificates..."
    tar czf "$BACKUP_PATH/ssl_certs.tar.gz" certbot/conf
fi

# Create a backup manifest
echo "Creating backup manifest..."
cat > "$BACKUP_PATH/manifest.txt" << EOF
Backup created: $(date)
Domain: ${DOMAIN:-unknown}
Database: ${POSTGRES_DB:-familytree}
EOF

# Compress the entire backup
echo "🗜️  Compressing backup..."
tar czf "$BACKUP_DIR/backup_$TIMESTAMP.tar.gz" -C "$BACKUP_DIR" "backup_$TIMESTAMP"
rm -rf "$BACKUP_PATH"

# Keep only last 7 backups
echo "🧹 Cleaning up old backups..."
cd "$BACKUP_DIR"
ls -t backup_*.tar.gz | tail -n +8 | xargs -r rm

echo ""
echo "✅ Backup completed successfully!"
echo "Backup file: $BACKUP_DIR/backup_$TIMESTAMP.tar.gz"
echo ""
