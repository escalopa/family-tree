# Family Tree

A full-stack family tree management system with OAuth authentication, multi-language support, and interactive visualizations.

## Quick Start (Local Development)

### Prerequisites

- Docker & Docker Compose

### 1. Configuration

**Backend:** Copy and edit `be/config.example.yaml` → `be/config.yaml`

```yaml
# Key settings to change:
jwt:
  secret: "your-secret-key"  # Change this!

oauth:
  redirect_base_url: "http://localhost:3000"
  providers:
    google:  # Enable providers you need
      client_id: "your-google-client-id"
      client_secret: "your-google-client-secret"

database:
  dsn: "host=postgres port=5432 user=familytree password=secret dbname=familytree sslmode=disable"

redis:
  uri: "redis://redis:6379/0"

s3:
  endpoint: "http://minio:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
```

### 2. Start Application

```bash
# Start all services (backend, frontend, database, redis, minio)
make testing-up

# Wait ~30 seconds for services to initialize, then access at:
# http://localhost:3000
```

### 3. Create First Admin User

Login via OAuth, then promote your user:

```bash
make promote-user EMAIL=your-email@example.com
```

**Stop services:**

```bash
make testing-down
```

---

## Production Deployment

### 1. Environment Setup

Copy `env.prod.example` → `.env` and configure:

```bash
DOMAIN=yourdomain.com
EMAIL=your-email@example.com

POSTGRES_PASSWORD=strong-password-here
REDIS_PASSWORD=redis-password-here
MINIO_ROOT_PASSWORD=minio-password-here
```

### 2. Backend Configuration

Create `be/config.yaml` for production:

```yaml
server:
  mode: "release"
  allowed_origins:
    - "https://yourdomain.com"
  enable_hsts: true
  cookie:
    secure: true  # Enable for HTTPS

jwt:
  secret: "production-secret-key"  # Use strong random key

oauth:
  providers:
    google:
      client_id: "prod-google-client-id"
      client_secret: "prod-google-client-secret"
  redirect_base_url: "https://yourdomain.com"

database:
  dsn: "host=postgres port=5432 user=familytree password=${POSTGRES_PASSWORD} dbname=familytree sslmode=require"

redis:
  uri: "redis://:${REDIS_PASSWORD}@redis:6379/0"

s3:
  endpoint: "http://minio:9000"
  access_key: "${MINIO_ROOT_USER}"
  secret_key: "${MINIO_ROOT_PASSWORD}"
```

### 3. Deploy

```bash
# Initialize SSL certificates
make prod-init

# Start all services
make prod-up

# Run migrations
make migrate-up DB_HOST=localhost DB_PASSWORD=your-db-password

# Create admin user
make promote-user EMAIL=admin@yourdomain.com
```

---

## Managing OAuth Providers

Edit `be/config.yaml` to enable/disable providers:

```yaml
oauth:
  providers:
    # Google OAuth
    google:
      order: 1  # Display order on login page
      client_id: "your-client-id"
      client_secret: "your-client-secret"
      # Get credentials: https://console.cloud.google.com/apis/credentials

    # GitHub OAuth (uncomment to enable)
    # github:
    #   order: 2
    #   client_id: "github-client-id"
    #   client_secret: "github-client-secret"
    #   # Get credentials: https://github.com/settings/developers

    # GitLab OAuth (uncomment to enable)
    # gitlab:
    #   order: 3
    #   client_id: "gitlab-client-id"
    #   client_secret: "gitlab-client-secret"
    #   # Get credentials: https://gitlab.com/-/profile/applications

    # Yandex OAuth (uncomment to enable)
    # yandex:
    #   order: 4
    #   client_id: "yandex-client-id"
    #   client_secret: "yandex-client-secret"
    #   # Get credentials: https://oauth.yandex.com/

  redirect_base_url: "https://yourdomain.com"  # Your app URL
```

**Note:** Restart backend after changing providers.

---

## Useful Commands

```bash
# Local Development
make testing-up           # Start all services (http://localhost:3000)
make testing-down         # Stop all services
make testing-logs         # View logs
make promote-user EMAIL=user@example.com

# Production
make prod-up              # Start production
make prod-down            # Stop production
make prod-logs            # View logs
make prod-backup          # Backup data

# Database
make migrate-up           # Run migrations
make migrate-status       # Check migration status
make db-recreate          # Reset database (⚠️ deletes data)
```

---

## Configuration Reference

| File | Purpose |
|------|---------|
| `be/config.yaml` | Backend configuration (server, database, OAuth, etc.) |
| `fe/.env` | Frontend API endpoint |
| `.env` | Production environment variables (Docker Compose) |

See `be/config.example.yaml` for all available backend options.
