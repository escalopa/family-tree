# Family Tree

A full-stack family tree management system with OAuth authentication, multi-language support, and interactive visualizations.

## Project overview

| Layer | Stack |
|-------|--------|
| Backend | Go, Gin, PostgreSQL, Redis, MinIO |
| Frontend | React, Vite, MUI, D3, i18next |
| Auth | OAuth (Google, etc.), JWT + refresh rotation |

## Repository layout

| Path | Purpose |
|------|---------|
| `be/` | Go API, migrations, seed CLI |
| `fe/` | React SPA (`fe/src/features/tree/` — tree UI module) |
| `_docs/req.md` | Product requirements |
| `.claude/` | Claude Code rules, skills, hooks |
| `scripts/` | `dev-cycle.sh`, `seed-and-test.sh` |

## Development workflow

```bash
make testing-up          # Docker stack (API, FE, Postgres, Redis, MinIO)
make check               # Backend + frontend validation
make check-be            # go build, vet, test ./...
make check-fe            # eslint, vitest, production build
make seed-testdata       # Seed + integration tests (SCALE=medium)
make seed-only SCALE=large CLEAN=1   # Seed only
make test-integration    # BE integration tests (INTEGRATION_TEST=1)
make dev-cycle PHASE=fe  # Focused dev loop script
```

Copy `CLAUDE.local.md.example` → `CLAUDE.local.md` for machine-specific paths. See `CLAUDE.md` for architecture, skills (`/dev-cycle`, `/seed-and-verify`), and MCP setup.

The testing stack enables a local-only Mock SSO provider at
`http://localhost:8090`. Use it from the login page to sign in as one of the
pre-seeded active test users:

| User | Email | Role |
|------|-------|------|
| Mock Super Admin | `superadmin.mock@example.test` | Super Admin |
| Mock Admin | `admin.mock@example.test` | Admin |
| Mock Guest | `guest.mock@example.test` | Guest |

### Frontend development

```bash
cd fe
cp .env.example .env    # if present
npm install
npm run dev             # http://localhost:3000
npm run lint
npm run test            # Vitest (table-driven unit/component tests)
npm run build
```

Environment: `VITE_API_URL` points at the backend (empty = same origin in production).

### Testing

- **Backend unit:** `cd be && go test ./...`
- **Backend integration:** `make test-integration` (requires `INTEGRATION_TEST=1`, running Postgres)
- **Frontend:** `cd fe && npm run test`
- **Full gate:** `make check`

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
