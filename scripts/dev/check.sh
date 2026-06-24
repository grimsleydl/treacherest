#!/usr/bin/env bash
set -u

repo_root="$(git rev-parse --show-toplevel)"
status=0

run_check() {
  name="$1"
  shift

  printf '\n==> %s\n' "$name"
  if "$@"; then
    printf 'PASS: %s\n' "$name"
  else
    code=$?
    printf 'FAIL: %s exited with %s\n' "$name" "$code" >&2
    status=1
  fi
}

run_check "go test ./..." bash -lc "cd '$repo_root/nix/app' && CGO_ENABLED=0 go test ./..."
run_check "theme readability tests" bash -lc "cd '$repo_root/nix/app' && npm run test:theme-lab"
run_check "package build" bash -lc "cd '$repo_root' && nix build .#packages.x86_64-linux.default --no-link"

exit "$status"
