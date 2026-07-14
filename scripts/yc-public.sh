#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="${YC_PUBLIC_IMAGE:-family-tree-yc-public:latest}"
CONFIG_DIR="${YC_PUBLIC_CONFIG_DIR:-$ROOT_DIR/.yc-public}"

mkdir -p "$CONFIG_DIR"
docker build -q -t "$IMAGE_NAME" "$ROOT_DIR/tools/yc-public" >/dev/null

docker_args=(
  --rm
  -it
  -v "$CONFIG_DIR:/root/.config/yandex-cloud"
  -v "$ROOT_DIR:/workspace"
  -w /workspace
)

if [[ "${YC_PUBLIC_HOST_NETWORK:-0}" == "1" ]]; then
  docker_args+=(--network host)
fi

exec docker run "${docker_args[@]}" "$IMAGE_NAME" "$@"
