# `cmd/server` Test Signature Drift

## Summary

`go test ./cmd/server` currently fails before running tests because `cmd/server/main_test.go` still calls `handlers.New` with an old argument list.

This is pre-existing test drift and is unrelated to the Coup role-image ingestion work. The production server package still builds after the Coup image loader wiring.

## Reproduction

From `nix/app`:

```sh
TMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOTMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOCACHE=/workspace/treacherest/.scratch/go-cache \
CGO_ENABLED=0 \
go test ./cmd/server
```

Current failure:

```text
cmd/server/main_test.go:56:44: not enough arguments in call to handlers.New
	have (*store.MemoryStore, *game.CardService, *config.ServerConfig)
	want (*store.MemoryStore, *game.CardService, *config.ServerConfig, *game.BackupService)
cmd/server/main_test.go:383:57: not enough arguments in call to handlers.New
	have (*store.MemoryStore, *game.CardService, *config.ServerConfig)
	want (*store.MemoryStore, *game.CardService, *config.ServerConfig, *game.BackupService)
```

## Verified Around It

For the Coup image-ingestion change, these commands passed:

```sh
TMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOTMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOCACHE=/workspace/treacherest/.scratch/go-cache \
CGO_ENABLED=0 \
go test ./internal/game -run 'TestLoadCoupRoleImages|TestAssignCoupRoles|TestCoup'

TMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOTMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOCACHE=/workspace/treacherest/.scratch/go-cache \
CGO_ENABLED=0 \
go test ./scripts/coup-images

TMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOTMPDIR=/workspace/treacherest/.scratch/go-tmp \
GOCACHE=/workspace/treacherest/.scratch/go-cache \
CGO_ENABLED=0 \
go build -o /workspace/treacherest/.scratch/go-tmp/server-check ./cmd/server
```

## Likely Fix

Update the `handlers.New` calls in `nix/app/cmd/server/main_test.go` to provide a `*game.BackupService`, matching the current production constructor usage in `nix/app/cmd/server/main.go`.

Keep that fix separate from Coup role-image tooling so it can be reviewed as test maintenance.
