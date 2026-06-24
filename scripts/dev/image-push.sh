#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
registry="${1:-ghcr.io/grimsleydl/treacherest}"
tag="${2:-}"

if [ -z "$tag" ]; then
  tag="sha-$(git -C "$repo_root" rev-parse --short HEAD)"
fi

case "$tag" in
  latest)
    echo "Pushing latest as a convenience tag only. Do not use latest as production deployment identity." >&2
    ;;
esac

cd "$repo_root"
echo "Pushing nix2container image: ${registry}:${tag}"
nix run .#containers.x86_64-linux.default.copyTo -- "docker://${registry}:${tag}"
echo "Pushed ${registry}:${tag}"
