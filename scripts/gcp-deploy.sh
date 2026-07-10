#!/usr/bin/env sh
set -eu

PROJECT_ID="${PROJECT_ID:-$(gcloud config get-value project 2>/dev/null)}"
REGION="${REGION:-us-central1}"
SERVICE="${SERVICE:-family-tree-api}"
AR_REPOSITORY="${AR_REPOSITORY:-family-tree}"
IMAGE="${IMAGE:-family-tree-api}"
ALLOWED_ORIGINS="${ALLOWED_ORIGINS:-https://family-tree.vercel.app}"
OAUTH_REDIRECT_BASE_URL="${OAUTH_REDIRECT_BASE_URL:-https://family-tree.vercel.app}"
OAUTH_ENABLED_PROVIDERS="${OAUTH_ENABLED_PROVIDERS:-mock}"
ENABLE_MOCK_AUTH="${ENABLE_MOCK_AUTH:-true}"
SEED_TEST_DATA="${SEED_TEST_DATA:-true}"
S3_ENDPOINT="${S3_ENDPOINT:-}"
S3_REGION="${S3_REGION:-us-east-1}"
S3_BUCKET="${S3_BUCKET:-family-tree}"

if [ -z "$PROJECT_ID" ] || [ "$PROJECT_ID" = "(unset)" ]; then
  echo "PROJECT_ID is required or must be configured in gcloud." >&2
  exit 1
fi

gcloud config set project "$PROJECT_ID" >/dev/null

gcloud builds submit . \
  --config cloudbuild.yaml \
  --substitutions="^|^_REGION=${REGION}|_SERVICE=${SERVICE}|_AR_REPOSITORY=${AR_REPOSITORY}|_IMAGE=${IMAGE}|_ALLOWED_ORIGINS=${ALLOWED_ORIGINS}|_OAUTH_REDIRECT_BASE_URL=${OAUTH_REDIRECT_BASE_URL}|_OAUTH_ENABLED_PROVIDERS=${OAUTH_ENABLED_PROVIDERS}|_ENABLE_MOCK_AUTH=${ENABLE_MOCK_AUTH}|_SEED_TEST_DATA=${SEED_TEST_DATA}|_S3_ENDPOINT=${S3_ENDPOINT}|_S3_REGION=${S3_REGION}|_S3_BUCKET=${S3_BUCKET}"
