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

The repository now includes provider-native deployment automation:

| Target | Config |
|--------|--------|
| Vercel frontend | `vercel.json` |
| GCP Cloud Run backend | `cloudbuild.yaml` |
| Deployment env checklist | `.env.deploy.example` |
| Full runbook | `_docs/deployment.md` |

Recommended free stack:

- Vercel for the React SPA.
- GCP Cloud Run for the Go API.
- Supabase free Postgres and Storage for database and uploaded images.
- Redis is optional; when `REDIS_URI` is empty, the API disables Redis-backed rate limiting and still runs.

Cloud Build builds the backend image, runs migrations, optionally loads active mock users and the 100-member test family tree, and deploys the API to Cloud Run.

See `_docs/deployment.md` for the required environment variables and one-time provider setup.

---

## Managing OAuth Providers

OAuth providers can be enabled with environment variables, so adding or removing providers does not require code changes:

```bash
OAUTH_ENABLED_PROVIDERS=google,github
OAUTH_REDIRECT_BASE_URL=https://your-frontend-host
OAUTH_PROVIDER_GOOGLE_CLIENT_ID=...
OAUTH_PROVIDER_GOOGLE_CLIENT_SECRET=...
OAUTH_PROVIDER_GOOGLE_SCOPES=openid,email,profile
OAUTH_PROVIDER_GITHUB_CLIENT_ID=...
OAUTH_PROVIDER_GITHUB_CLIENT_SECRET=...
```

Known providers are `google`, `github`, `gitlab`, `yandex`, and `mock`. Generic OAuth2/OIDC-style providers can also be configured with authorize, token, and user-info URLs; see `_docs/deployment.md`.

Until real OAuth credentials are ready, keep:

```bash
OAUTH_ENABLED_PROVIDERS=mock
ENABLE_MOCK_AUTH=true
SEED_TEST_DATA=true
```

**Note:** Restart or redeploy the backend after changing provider configuration.

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
