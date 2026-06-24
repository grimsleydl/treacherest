#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
registry="${1:-ghcr.io/grimsleydl/treacherest}"
host_port="${2:-8080}"
tag="sha-$(git -C "$repo_root" rev-parse --short HEAD)"
image="${registry}:${tag}"
container_name="treacherest-smoke-${tag}-$$"

if (echo >"/dev/tcp/127.0.0.1/${host_port}") >/dev/null 2>&1; then
  echo "Host port ${host_port} is already in use; image-run uses strict explicit ports." >&2
  exit 1
fi

cleanup() {
  podman rm -f "$container_name" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cd "$repo_root"
nix run .#containers.x86_64-linux.default.copyTo -- "containers-storage:${image}"

podman run --rm --detach --name "$container_name" -p "127.0.0.1:${host_port}:8080" "$image" >/dev/null

ready_url="http://127.0.0.1:${host_port}/health/ready"
for _ in $(seq 1 60); do
  if curl -fsS "$ready_url" >/dev/null 2>&1; then
    echo "Image smoke passed: ${ready_url}"
    exit 0
  fi
  sleep 1
done

echo "Image did not become ready at ${ready_url}" >&2
podman logs "$container_name" >&2 || true
exit 1
