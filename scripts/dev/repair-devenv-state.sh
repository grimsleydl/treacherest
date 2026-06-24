#!/usr/bin/env bash
set -eu

root="${1:-.}"
dotfile="$root/.devenv"
repaired=0

fail_unwritable_dir() {
  local path="$1"

  echo "Generated devenv state is not writable: $path" >&2
  echo "Run this from the repo root, then reload direnv." >&2
  echo "For a host/aibox shared workspace, prefer writable generated state:" >&2
  echo "  sudo chmod -R a+rwX .devenv .direnv" >&2
  echo "If ownership is badly mixed, regenerate the ignored state instead:" >&2
  echo "  rm -rf .devenv .direnv" >&2
  echo "  direnv allow" >&2
  echo "  direnv reload" >&2
  exit 1
}

chmod_shared() {
  local path

  for path in "$@"; do
    [ -e "$path" ] && chmod a+rwX "$path" 2>/dev/null || true
  done
}

share_generated_state() {
  [ -e "$dotfile" ] || return

  chmod_shared "$dotfile" "$dotfile/state" "$dotfile/load-exports"
  chmod_shared "$dotfile"/nix-eval-cache.db*

  if [ -d "$dotfile/state" ]; then
    chmod_shared "$dotfile/state"/tasks.db*
  fi
}

repair_sqlite_family() {
  local db="$1"
  local needs_repair=0
  local path

  for path in "$db" "$db-shm" "$db-wal"; do
    if [ -e "$path" ] && [ ! -w "$path" ]; then
      needs_repair=1
    fi
  done

  if [ "$needs_repair" -eq 0 ]; then
    return
  fi

  chmod_shared "$db" "$db-shm" "$db-wal"

  needs_repair=0
  for path in "$db" "$db-shm" "$db-wal"; do
    if [ -e "$path" ] && [ ! -w "$path" ]; then
      needs_repair=1
    fi
  done

  if [ "$needs_repair" -eq 0 ]; then
    return
  fi

  if ! rm -f "$db" "$db-shm" "$db-wal"; then
    echo "Could not remove stale devenv cache files for $db" >&2
    echo "Run: sudo chmod -R a+rwX .devenv .direnv" >&2
    exit 1
  fi

  repaired=1
}

if [ ! -e "$dotfile" ]; then
  exit 0
fi

share_generated_state

if [ ! -d "$dotfile" ]; then
  fail_unwritable_dir "$dotfile"
fi

if [ ! -w "$dotfile" ]; then
  fail_unwritable_dir "$dotfile"
fi

if [ -d "$dotfile/state" ] && [ ! -w "$dotfile/state" ]; then
  fail_unwritable_dir "$dotfile/state"
fi

repair_sqlite_family "$dotfile/nix-eval-cache.db"

if [ -d "$dotfile/state" ]; then
  repair_sqlite_family "$dotfile/state/tasks.db"
fi

if [ "$repaired" -eq 1 ]; then
  echo "Repaired generated devenv cache files under .devenv"
fi
