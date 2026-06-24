#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root/nix/app"

export CGO_ENABLED=0

templ generate
npm run build:css

go test ./internal/config ./internal/views/... ./internal/game/ability -count=1
npm run test:theme-lab
