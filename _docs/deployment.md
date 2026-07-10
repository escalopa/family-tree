# Deployment Runbook

This project is configured for a low-cost deployment path:

- Frontend: Vercel, built from `fe/` by the root `vercel.json`.
- Backend: Render Web Service, built from `be/` by `render.yaml`.
- Database: Supabase Postgres.
- Object storage: Supabase Storage S3-compatible endpoint.
- Redis: optional. If `REDIS_URI` is empty, rate limiting is disabled and the app still runs.

## Automated Flow

1. Push to the connected repository branch.
2. Vercel builds and deploys the frontend.
3. Render builds the backend, runs migrations, optionally loads test data, and deploys the API.
Provider-side Git integration is the deployment trigger. No deployment command needs to run from a developer laptop after the projects are connected.

## One-Time Setup

### Supabase

1. Create a Supabase project.
2. Copy a Postgres connection string into `DATABASE_DSN`.
   - Prefer the session pooler connection string for the long-running Render backend.
   - Keep `sslmode=require` in the URL.
3. Create a private Storage bucket, for example `family-tree`.
4. Create Supabase Storage S3 access keys.
5. Set:
   - `S3_ENDPOINT=https://PROJECT_REF.storage.supabase.co/storage/v1/s3`
   - `S3_REGION=us-east-1`
   - `S3_BUCKET=family-tree`
   - `S3_ACCESS_KEY`
   - `S3_SECRET_KEY`

### Render

1. Create a new Blueprint from this repository.
2. Render reads `render.yaml` and creates the `family-tree-api` web service.
3. Add the missing secret environment variables:
   - `DATABASE_DSN`
   - `ALLOWED_ORIGINS`
   - `OAUTH_REDIRECT_BASE_URL`
   - `S3_ENDPOINT`
   - `S3_REGION`
   - `S3_BUCKET`
   - `S3_ACCESS_KEY`
   - `S3_SECRET_KEY`
4. Keep the generated `JWT_SECRET`; do not replace it with a committed value.

The default Blueprint keeps the deployment testable before production OAuth credentials exist:

- `OAUTH_ENABLED_PROVIDERS=mock`
- `ENABLE_MOCK_AUTH=true`
- `SEED_TEST_DATA=true`

With this setup, the login button signs in as `superadmin.mock@example.test`, and the seed step loads the 100-member test family tree plus mock users.

For real production, change these values:

- `OAUTH_ENABLED_PROVIDERS=google,github` or whichever providers are active.
- `ENABLE_MOCK_AUTH=false`
- `SEED_TEST_DATA=false`
- Add the matching `OAUTH_PROVIDER_<PROVIDER>_*` secrets.

### Vercel

1. Import the repository as a Vercel project.
2. Keep the root directory as the repository root. `vercel.json` handles the frontend build.
3. Add:
   - `VITE_API_URL=https://YOUR_RENDER_API_HOST`
4. After Vercel gives you the frontend URL, update Render:
   - `ALLOWED_ORIGINS=https://YOUR_VERCEL_HOST`
   - `OAUTH_REDIRECT_BASE_URL=https://YOUR_VERCEL_HOST`

## Environment Variables

Copy `.env.deploy.example` as a checklist for provider dashboards.

### Required for Backend

| Variable | Purpose |
|----------|---------|
| `DATABASE_DSN` | Postgres connection string. |
| `JWT_SECRET` | Signing secret for access and refresh tokens. Render can generate it. |
| `ALLOWED_ORIGINS` | Comma-separated frontend URLs allowed by CORS. |
| `OAUTH_REDIRECT_BASE_URL` | Public frontend URL used for OAuth callbacks. |
| `OAUTH_ENABLED_PROVIDERS` | Comma-separated provider list, for example `mock` or `google,github`. |
| `S3_ENDPOINT` | S3-compatible endpoint. |
| `S3_REGION` | S3 region value required by the provider. |
| `S3_BUCKET` | Bucket name. |
| `S3_ACCESS_KEY` | S3 access key. |
| `S3_SECRET_KEY` | S3 secret key. |

### Optional for Backend

| Variable | Purpose |
|----------|---------|
| `REDIS_URI` | Enables Redis-backed rate limiting. |
| `SEED_TEST_DATA` | Loads mock users and 100 test family members during deploy when `true`. |
| `ENABLE_MOCK_AUTH` | Enables the mock provider when no real OAuth credentials are available. |
| `COOKIE_DOMAIN` | Set only when cookies must span subdomains. |
| `RATE_LIMIT_AUTH_ENABLED` | Enables auth endpoint rate limiting when Redis exists. |
| `RATE_LIMIT_API_ENABLED` | Enables API endpoint rate limiting when Redis exists. |
| `RATE_LIMIT_UPLOAD_ENABLED` | Enables upload endpoint rate limiting when Redis exists. |

### Required for Frontend

| Variable | Purpose |
|----------|---------|
| `VITE_API_URL` | Public backend API URL. |

## Configurable OAuth Providers

Known providers still work with their existing adapters: `google`, `github`, `gitlab`, `yandex`, and `mock`.

Unknown providers can be configured without code changes when they expose OAuth2-style authorize, token, and user-info endpoints:

```text
OAUTH_ENABLED_PROVIDERS=custom
OAUTH_PROVIDER_CUSTOM_CLIENT_ID=...
OAUTH_PROVIDER_CUSTOM_CLIENT_SECRET=...
OAUTH_PROVIDER_CUSTOM_AUTH_URL=https://provider.example/oauth/authorize
OAUTH_PROVIDER_CUSTOM_TOKEN_URL=https://provider.example/oauth/token
OAUTH_PROVIDER_CUSTOM_USER_INFO_URL=https://provider.example/userinfo
OAUTH_PROVIDER_CUSTOM_SCOPES=openid,email,profile
OAUTH_PROVIDER_CUSTOM_ID_FIELD=sub
OAUTH_PROVIDER_CUSTOM_EMAIL_FIELD=email
OAUTH_PROVIDER_CUSTOM_NAME_FIELD=name,display_name
OAUTH_PROVIDER_CUSTOM_PICTURE_FIELD=picture,avatar_url
```

Field variables accept comma-separated fallbacks.

## Manual Steps That Remain

- Creating the Vercel, Render, and Supabase projects.
- Pasting secrets into provider dashboards.
- Updating OAuth provider dashboards with the final callback URL:
  `https://YOUR_FRONTEND_HOST/auth/<provider>/callback`.

Everything else is represented in repository config.
