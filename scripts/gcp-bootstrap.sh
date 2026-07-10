#!/usr/bin/env sh
set -eu

PROJECT_ID="${PROJECT_ID:-$(gcloud config get-value project 2>/dev/null)}"
REGION="${REGION:-us-central1}"
AR_REPOSITORY="${AR_REPOSITORY:-family-tree}"

if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "(unset)" ]; then
  echo "PROJECT_ID is required or must be configured in gcloud." >&2
  exit 1
fi

if [ -z "${DATABASE_DSN:-}" ]; then
  echo "DATABASE_DSN is required. Use the Supabase Postgres pooler or direct URL with sslmode=require." >&2
  exit 1
fi

if [ -z "${S3_ACCESS_KEY:-}" ] || [ -z "${S3_SECRET_KEY:-}" ]; then
  echo "S3_ACCESS_KEY and S3_SECRET_KEY are required for Supabase Storage uploads." >&2
  exit 1
fi

if [ -z "${JWT_SECRET:-}" ]; then
  if command -v openssl >/dev/null 2>&1; then
    JWT_SECRET="$(openssl rand -base64 48)"
  else
    JWT_SECRET="generated-$(date +%s)-$(whoami)"
  fi
  export JWT_SECRET
  echo "Generated JWT_SECRET for Secret Manager."
fi

echo "Using project: $PROJECT_ID"
gcloud config set project "$PROJECT_ID" >/dev/null

gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com

if ! gcloud artifacts repositories describe "$AR_REPOSITORY" --location="$REGION" >/dev/null 2>&1; then
  gcloud artifacts repositories create "$AR_REPOSITORY" \
    --repository-format=docker \
    --location="$REGION" \
    --description="Family Tree Cloud Run images"
fi

PROJECT_NUMBER="$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')"
CLOUD_BUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

for role in \
  roles/run.admin \
  roles/iam.serviceAccountUser \
  roles/artifactregistry.writer \
  roles/secretmanager.secretAccessor
do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:${CLOUD_BUILD_SA}" \
    --role="$role" \
    --quiet >/dev/null
done

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${COMPUTE_SA}" \
  --role="roles/secretmanager.secretAccessor" \
  --quiet >/dev/null

upsert_secret() {
  name="$1"
  value="$2"
  if ! gcloud secrets describe "$name" >/dev/null 2>&1; then
    printf '%s' "$value" | gcloud secrets create "$name" --data-file=- >/dev/null
  else
    printf '%s' "$value" | gcloud secrets versions add "$name" --data-file=- >/dev/null
  fi
}

upsert_secret DATABASE_DSN "$DATABASE_DSN"
upsert_secret JWT_SECRET "$JWT_SECRET"
upsert_secret S3_ACCESS_KEY "$S3_ACCESS_KEY"
upsert_secret S3_SECRET_KEY "$S3_SECRET_KEY"

echo "GCP bootstrap complete."
