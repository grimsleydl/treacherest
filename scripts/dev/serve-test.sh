#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
port="${1:-8080}"

cd "$repo_root/nix/app"

export CGO_ENABLED=0
export HOST=localhost
export PORT="$port"
export CONFIG_PATH=../../configs/server-development.yaml
export SHUTDOWN_TIMEOUT=250ms

templ generate
npm run build:css

exec go run ./cmd/server
