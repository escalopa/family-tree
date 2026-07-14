#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_DSN:-}" ]; then
  echo "DATABASE_DSN is required" >&2
  exit 1
fi

go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$DATABASE_DSN" up

if [ "${SEED_TEST_DATA:-false}" = "true" ]; then
  go run ./cmd/seed-testing
fi
