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

Before applying Terraform, make sure the local `yc` profile points to the public
Yandex Cloud endpoint and is authenticated to an account with access to that
cloud and folder. If your workstation has internal or alternate Yandex Cloud
profiles, create a dedicated public profile:

```bash
yc config profile create public || yc config profile activate public
yc config set endpoint api.cloud.yandex.net:443
yc config set cloud-id b1g00m03hogrja9p1rb0
yc config set folder-id b1gkimk9k36atshi4uto
yc init --federation-endpoint auth.yandex.cloud --username=<yandex-account-email>
yc resource-manager folder get b1gkimk9k36atshi4uto
```

## Vercel Project And Domain

The Vercel project is `family-tree` under the `escalopa` account. The production
domain is attached in Vercel:

- `family.escalopa.com`

Vercel currently reports that DNS is not configured. Add this record at the DNS
provider for `escalopa.com`:

```text
Type:  A
Name:  family
Value: 76.76.21.21
```

Vercel also accepts the generated CNAME target it recommended:

```text
Type:  CNAME
Name:  family
Value: 154e7ade97ff9174.vercel-dns-017.com.
```

Use the `A` record unless the DNS provider requires a CNAME for this subdomain.
After changing DNS, verify with:

```bash
vercel domains verify family.escalopa.com
```

## Terraform State Bootstrap

GitHub Actions deploys from `main` only and uses an Object Storage bucket for
Terraform state. That bucket and the CI service account are also Terraform
resources, managed by `infra/yandex/bootstrap`.

Run the bootstrap once from an authenticated local `yc` profile:

```bash
export TF_VAR_cloud_id=b1g00m03hogrja9p1rb0
export TF_VAR_folder_id=b1gkimk9k36atshi4uto
export TF_VAR_yc_token="$(yc iam create-token)"
terraform -chdir=infra/yandex/bootstrap init
terraform -chdir=infra/yandex/bootstrap apply
```

Then copy the generated CI credentials to GitHub secrets without printing them:

```bash
terraform -chdir=infra/yandex/bootstrap output -raw yc_service_account_key_json | gh secret set YC_SERVICE_ACCOUNT_KEY_JSON
terraform -chdir=infra/yandex/bootstrap output -raw tf_state_access_key_id | gh secret set TF_STATE_ACCESS_KEY_ID
terraform -chdir=infra/yandex/bootstrap output -raw tf_state_secret_access_key | gh secret set TF_STATE_SECRET_ACCESS_KEY
```

The remaining production secrets must be supplied by you:

```bash
printf '%s' '<vercel-token>' | gh secret set VERCEL_TOKEN
printf '%s' 'team_nKYA059W7ZnhVw72wPcXyky8' | gh secret set VERCEL_ORG_ID
printf '%s' 'prj_baUersa3B9BrZ7v3Nakpko6iyL35' | gh secret set VERCEL_PROJECT_ID
printf '%s' '<strong-jwt-secret>' | gh secret set JWT_SECRET
printf '%s' '<google-oauth-client-id>' | gh secret set OAUTH_GOOGLE_CLIENT_ID
printf '%s' '<google-oauth-client-secret>' | gh secret set OAUTH_GOOGLE_CLIENT_SECRET
```

Do not add these values to `.env`, tfvars, or committed files.

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

## Production Auth

Production deploys use Google OAuth and do not enable mock auth or seed data:

```hcl
oauth_enabled_providers = "google"
enable_mock_auth        = false
seed_test_data          = false
```

Provider credentials are written to Yandex Lockbox by Terraform and exposed to
the serverless container as environment variables.

## CI/CD Policy

- Pull requests run backend tests/build, frontend lint/build, and Terraform validation.
- Pushes to `main` run the same checks, then deploy production.
- No preview/testing deployment is created for non-main branches.
- The production frontend is deployed through Vercel CLI.
- Yandex Cloud resources are created or updated through Terraform only.

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
