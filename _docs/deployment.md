# Deployment Runbook

The target production architecture is now Yandex Cloud first:

- Frontend: Vercel, built from `fe/` by the root `vercel.json`.
- Backend: Yandex Serverless Containers.
- Public backend entrypoint: Yandex API Gateway.
- Database: Yandex Managed Service for YDB, serverless mode, using YQL.
- File uploads: Yandex Object Storage, S3-compatible API.
- Secrets: Yandex Lockbox.
- Infrastructure: Terraform under `infra/yandex`.

## Current Compatibility Status

The infrastructure can be provisioned with Terraform, but the current Go backend is still a PostgreSQL application:

- It uses `pgx` and `pgxpool`.
- It runs goose PostgreSQL migrations.
- Migrations use PostgreSQL features such as `SERIAL`, `JSONB`, arrays, foreign keys, and UUID functions.
- Repository implementations issue PostgreSQL SQL directly.

YDB/YQL is not a PostgreSQL wire-compatible replacement for this code. The backend needs a persistence-layer migration before the serverless container can run fully against YDB. Terraform already exposes YDB runtime values to the container with:

- `DATABASE_BACKEND=ydb`
- `YDB_ENDPOINT`
- `YDB_DATABASE`
- `YDB_AUTH_MODE=metadata`

The next code milestone is to add a YDB repository implementation and YDB/YQL schema bootstrap.

## Terraform Layout

| Path | Purpose |
|------|---------|
| `infra/yandex/versions.tf` | Terraform and Yandex provider setup. |
| `infra/yandex/variables.tf` | Inputs for cloud/folder/image/frontend/secrets. |
| `infra/yandex/main.tf` | YDB, Lockbox, Object Storage, registry, serverless container, API Gateway, IAM. |
| `infra/yandex/outputs.tf` | API URL, image URL, YDB endpoint, bucket outputs. |
| `infra/yandex/templates/api-gateway.yaml.tftpl` | OpenAPI proxy spec for API Gateway. |
| `infra/yandex/terraform.tfvars.example` | Local example variables. |

## One-Time Setup

The target Yandex Cloud deployment is:

- `cloud_id=b1g00m03hogrja9p1rb0`
- `folder_id=b1gkimk9k36atshi4uto`

Before applying Terraform, make sure the local `yc` profile is authenticated to
an account with access to that cloud and folder:

```bash
yc config set cloud-id b1g00m03hogrja9p1rb0
yc config set folder-id b1gkimk9k36atshi4uto
yc resource-manager folder get b1gkimk9k36atshi4uto
```

Create a local tfvars file:

```bash
cd infra/yandex
cp terraform.tfvars.example terraform.tfvars
```

Set at minimum:

```hcl
jwt_secret = "strong-secret"
frontend_origin = "https://YOUR_VERCEL_HOST"
container_image_tag = "latest"
```

Initialize and plan:

```bash
export TF_VAR_yc_token="$(yc iam create-token)"
terraform init
terraform plan
```

Apply:

```bash
terraform apply
```

## Container Image

Terraform creates the Container Registry and outputs the image URL it expects:

```bash
terraform output container_image_url
```

Build and push the backend image before applying a container revision that uses that tag:

```bash
docker build -f be/Dockerfile --target production -t "$(terraform -chdir=infra/yandex output -raw container_image_url)" be
docker push "$(terraform -chdir=infra/yandex output -raw container_image_url)"
```

For a new release, update `container_image_tag`, push that tag, and run `terraform apply`.

## Vercel

Set the frontend environment variable after Terraform outputs the API Gateway URL:

```bash
VITE_API_URL=https://API_GATEWAY_DOMAIN
```

Then update Terraform:

```hcl
frontend_origin = "https://YOUR_VERCEL_HOST"
```

This value is used for CORS and OAuth callback URLs.

## Mock Auth

Until production OAuth credentials are ready, keep:

```hcl
oauth_enabled_providers = "mock"
enable_mock_auth        = true
seed_test_data          = true
```

For real production, switch to provider configuration and disable mock/test seed:

```hcl
oauth_enabled_providers = "google,github"
enable_mock_auth        = false
seed_test_data          = false
```

Provider credentials should go into Lockbox and be mapped into the serverless container as environment variables in Terraform.

## YDB Migration Work

To make the backend fully functional on YDB/YQL:

1. Add a database abstraction boundary that can choose `postgres` or `ydb`.
2. Add YDB schema creation for users, roles, sessions, OAuth states, members, spouses, names, history, scores, and language preferences.
3. Replace PostgreSQL-only types and behaviors:
   - `SERIAL` with generated IDs or YDB sequences/table-based counters.
   - `UUID` storage with strings or bytes.
   - `JSONB` with YDB JSON/string fields.
   - PostgreSQL arrays with child tables or JSON.
   - Postgres-specific transactions/batches with YDB transactions.
4. Implement YDB repositories using the YDB Go SDK and YQL.
5. Add YDB seed logic for the 100-member test family tree.
6. Run integration tests against a real YDB database.

## Manual Steps That Remain

- Push the backend image to Yandex Container Registry.
- Complete the backend YDB persistence migration.
- Configure production OAuth credentials in Lockbox when available.
- Set `VITE_API_URL` in Vercel to the Terraform output `api_gateway_url`.
