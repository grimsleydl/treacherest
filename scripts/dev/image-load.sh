#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
registry="${1:-ghcr.io/grimsleydl/treacherest}"
tag="${2:-}"

if [ -z "$tag" ]; then
  if [ -z "$(git -C "$repo_root" status --short --untracked-files=normal)" ]; then
    tag="$(git -C "$repo_root" rev-parse --short HEAD)"
  else
    tag="latest"
  fi
fi

image="${registry}:${tag}"

cd "$repo_root"
echo "Loading nix2container image into Podman storage: ${image}"
nix run .#containers.x86_64-linux.default.copyTo -- "containers-storage:${image}"
echo "Loaded ${image}"
