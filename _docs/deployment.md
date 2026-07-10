# Deployment Runbook

This project is configured for a low-cost deployment path:

- Frontend: Vercel, built from `fe/` by the root `vercel.json`.
- Backend: Google Cloud Run, built and deployed by `cloudbuild.yaml`.
- Database: Supabase Postgres.
- Object storage: Supabase Storage S3-compatible endpoint.
- Redis: optional. If `REDIS_URI` is empty, rate limiting is disabled and the app still runs.

Cloud Run has a monthly free tier, but Google Cloud still requires a billing-enabled project. Keep `--min-instances=0` so idle backend compute can scale to zero.

## Automated Flow

1. Push to the connected repository branch.
2. Vercel builds and deploys the frontend.
3. Cloud Build builds the backend container image.
4. Cloud Build runs migrations and optionally loads test data.
5. Cloud Build deploys the image to Cloud Run.

Provider-side Git integration is the deployment trigger. No deployment command needs to run from a developer laptop after Vercel and Cloud Build are connected to the repository.

## One-Time Setup

### Supabase

1. Create or choose a Supabase project.
2. Copy a Postgres connection string into `DATABASE_DSN`.
   - Prefer the session pooler connection string for the long-running Cloud Run backend.
   - Keep `sslmode=require` in the URL.
3. Create a private Storage bucket, for example `family-tree`.
4. Create Supabase Storage S3 access keys.
5. Set:
   - `S3_ENDPOINT=https://PROJECT_REF.storage.supabase.co/storage/v1/s3`
   - `S3_REGION=us-east-1`
   - `S3_BUCKET=family-tree`
   - `S3_ACCESS_KEY`
   - `S3_SECRET_KEY`

### Google Cloud

Use `scripts/gcp-bootstrap.sh` once per GCP project. It enables required APIs, creates the Artifact Registry Docker repository, grants Cloud Build and Cloud Run access to Secret Manager, and writes required secrets.

```bash
export PROJECT_ID=your-gcp-project-id
export REGION=us-central1
export DATABASE_DSN='postgresql://postgres.PROJECT_REF:PASSWORD@aws-0-region.pooler.supabase.com:5432/postgres?sslmode=require'
export S3_ACCESS_KEY='...'
export S3_SECRET_KEY='...'
export JWT_SECRET='...' # optional; generated if omitted

sh scripts/gcp-bootstrap.sh
```

For a manual deploy:

```bash
export PROJECT_ID=your-gcp-project-id
export REGION=us-central1
export ALLOWED_ORIGINS=https://your-vercel-app.vercel.app
export OAUTH_REDIRECT_BASE_URL=https://your-vercel-app.vercel.app
export S3_ENDPOINT=https://PROJECT_REF.storage.supabase.co/storage/v1/s3
export S3_BUCKET=family-tree

sh scripts/gcp-deploy.sh
```

For automatic backend deployments:

1. In Google Cloud, connect the GitHub repository to Cloud Build.
2. Create a Cloud Build trigger for the deploy branch.
3. Use `cloudbuild.yaml` as the build configuration.
4. Set trigger substitutions if your values differ from the defaults:
   - `_REGION`
   - `_SERVICE`
   - `_AR_REPOSITORY`
   - `_IMAGE`
   - `_ALLOWED_ORIGINS`
   - `_OAUTH_REDIRECT_BASE_URL`
   - `_OAUTH_ENABLED_PROVIDERS`
   - `_ENABLE_MOCK_AUTH`
   - `_SEED_TEST_DATA`
   - `_S3_ENDPOINT`
   - `_S3_REGION`
   - `_S3_BUCKET`

The default deployment stays testable before production OAuth credentials exist:

- `_OAUTH_ENABLED_PROVIDERS=mock`
- `_ENABLE_MOCK_AUTH=true`
- `_SEED_TEST_DATA=true`

With this setup, the login button signs in as `superadmin.mock@example.test`, and the seed step loads the 100-member test family tree plus mock users.

For real production, change these values:

- `_OAUTH_ENABLED_PROVIDERS=google,github` or whichever providers are active.
- `_ENABLE_MOCK_AUTH=false`
- `_SEED_TEST_DATA=false`
- Add the matching `OAUTH_PROVIDER_<PROVIDER>_*` Cloud Run environment variables or Secret Manager mappings.

### Vercel

1. Import the repository as a Vercel project.
2. Keep the root directory as the repository root. `vercel.json` handles the frontend build.
3. Add:
   - `VITE_API_URL=https://YOUR_CLOUD_RUN_SERVICE_URL`
4. After Vercel gives you the frontend URL, update Cloud Build trigger substitutions:
   - `_ALLOWED_ORIGINS=https://YOUR_VERCEL_HOST`
   - `_OAUTH_REDIRECT_BASE_URL=https://YOUR_VERCEL_HOST`

## Environment Variables And Secrets

Copy `.env.deploy.example` as a checklist.

### GCP Bootstrap Inputs

| Variable | Purpose |
|----------|---------|
| `PROJECT_ID` | GCP project ID. |
| `REGION` | GCP region, defaults to `us-central1`. |
| `DATABASE_DSN` | Supabase Postgres connection string; stored as a GCP secret. |
| `JWT_SECRET` | Token signing secret; generated if omitted. |
| `S3_ACCESS_KEY` | Supabase Storage S3 access key; stored as a GCP secret. |
| `S3_SECRET_KEY` | Supabase Storage S3 secret key; stored as a GCP secret. |

### Cloud Build Substitutions

| Substitution | Purpose |
|--------------|---------|
| `_REGION` | Cloud Run and Artifact Registry region. |
| `_SERVICE` | Cloud Run service name. |
| `_AR_REPOSITORY` | Artifact Registry Docker repository name. |
| `_IMAGE` | Backend image name. |
| `_ALLOWED_ORIGINS` | Comma-separated frontend URLs allowed by CORS. |
| `_OAUTH_REDIRECT_BASE_URL` | Public frontend URL used for OAuth callbacks. |
| `_OAUTH_ENABLED_PROVIDERS` | Comma-separated provider list, for example `mock` or `google,github`. |
| `_ENABLE_MOCK_AUTH` | Enables the fallback mock provider when real OAuth credentials are not ready. |
| `_SEED_TEST_DATA` | Loads mock users and 100 test family members during deploy when `true`. |
| `_S3_ENDPOINT` | S3-compatible endpoint. |
| `_S3_REGION` | S3 region value required by the provider. |
| `_S3_BUCKET` | Bucket name. |

### GCP Secret Manager

| Secret | Runtime env var |
|--------|------------------|
| `DATABASE_DSN` | `DATABASE_DSN` |
| `JWT_SECRET` | `JWT_SECRET` |
| `S3_ACCESS_KEY` | `S3_ACCESS_KEY` |
| `S3_SECRET_KEY` | `S3_SECRET_KEY` |

### Required for Frontend

| Variable | Purpose |
|----------|---------|
| `VITE_API_URL` | Public Cloud Run API URL. |

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

- Choosing or creating the Supabase project.
- Supplying Supabase database/storage secrets.
- Connecting Vercel and Cloud Build to the GitHub repository.
- Updating OAuth provider dashboards with the final callback URL:
  `https://YOUR_FRONTEND_HOST/auth/<provider>/callback`.

Everything else is represented in repository config.
