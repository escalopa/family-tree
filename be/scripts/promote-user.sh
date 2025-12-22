#!/bin/bash

# Script to promote a user to super_admin and activate them
# Usage: ./scripts/promote-user.sh user@example.com

set -e

# Check if email is provided
if [ -z "$1" ]; then
    echo "Error: Email address is required"
    echo "Usage: $0 <email>"
    echo "Example: $0 user@example.com"
    exit 1
fi

EMAIL="$1"

# Read database config from environment or use defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-familytree}"
DB_NAME="${DB_NAME:-familytree}"
DB_PASSWORD="${DB_PASSWORD:-secret}"

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

# Run the SQL script
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -f "$(dirname "$0")/promote-user.sql" \
    -v email="$EMAIL"

# Clear password from environment
unset PGPASSWORD
