# Dev Environment Baseline

Last verified: 2026-06-24.

This document records the current Treacherest development, build, test,
container, port, and deployment workflow before the dev environment
standardization work changes behavior.

## Scope

- Baseline only. No runtime behavior is changed in this slice.
- `.scratch/dev-env-standardization/` remains the local planning tracker and is
  not staged or committed.
- The repo is on `main`, with local `main` ahead of `forgejo/main` by 208
  commits at the time of this inventory.

## Dirty Worktree Constraints

- `flake.lock` is intentionally parked. It contains broad input updates and was
  not committed because `nix build .#packages.x86_64-linux.default --no-link`
  failed while building `mtg-treachery-card-images` with a fixed-output hash
  mismatch:
  - specified: `sha256-4YBITZeMVZmEJQXwUg1RG/9BFI3LUsxHh7FaczBcypg=`
  - got: `sha256-oHi3P3pVFRjVJs8V6D2NX6+D29c5RFXY5zqeuoqaNVU=`
- Do not stage, overwrite, revert, or "fix up" the parked `flake.lock` change
  unless the lockfile update is explicitly in scope.
- `aibox.yaml` is local Aibox workspace/session configuration. It is ignored so
  it does not get committed accidentally.
- Preserve unrelated user and other-agent changes. Do not use destructive git
  cleanup commands without explicit approval.

## Local Development Commands

Current shell entrypoint:

- `nix develop`

Current devshell commands from `nix/local/shells.nix`:

- `dev`: starts templ watch mode with proxy `http://localhost:8888`, then starts
  Air from `nix/app`.
- `dev-css`: stops existing PostCSS and templ watchers, starts `npm run
  watch:css`, starts templ watch mode with proxy `http://localhost:8080`, then
  starts Air.
- `run`: builds templates and runs `go run cmd/server/main.go` with
  `HOST=${HOST:-localhost}`, `PORT=${PORT:-8080}`, and
  `CONFIG_PATH=${CONFIG_PATH:-../../configs/server-production.yaml}`.
- `build`: runs `templ generate`, builds `cmd/server/main.go`, and writes
  `nix/app/bin/server`.
- `build-templ`: runs `templ generate`.
- `fmt`: runs `go fmt ./...` and `templ fmt .`.
- `update-deps`: runs `go mod tidy` and `gomod2nix generate`.
- `import-deps`: runs `gomod2nix import`.
- `setup-css`: runs `npm install`.
- `build-css`: runs `npm run build:css`.
- `build-css-prod`: runs `npm run build:css:prod`.
- `backup-templates`: copies `nix/app/internal/views` to a timestamped backup.
- `download-cards`, `download-cards-sample`, and `download-cards-info`: run the
  card image helper scripts.

The current Air config is `nix/app/.air.toml`. Its build command runs
`templ generate`, `npm run build:css`, and `go build -o ./tmp/main
./cmd/server`. Its runtime command is:

```sh
HOST=localhost PORT=8888 CONFIG_PATH=../../configs/server-development.yaml ./tmp/main
```

## Test Commands

Current devshell test commands:

- `test-all`: runs `go test ./...` from `nix/app`.
- `test-coverage`: runs `go test -v -coverprofile=build/coverage/coverage.out
  ./...`, then writes HTML and function coverage reports under
  `nix/app/build/coverage/`.
- `test-playwright`: runs `playwright test` from `nix/app`.
- `test-playwright-ui`: runs `playwright test --ui`.
- `test-playwright-debug`: runs `playwright test --debug`.

Current local caveat: in this environment, direct `go test` fails before app
tests run because `gcc` is not on `PATH`. The same is true inside the current
`nix develop` shell unless `CGO_ENABLED=0` is set explicitly.

Known pre-existing red checks:

- `nix develop --command bash -lc 'cd nix/app; CGO_ENABLED=0 go test ./internal/game -run TestHiddenDistribution -count=1'`
  currently fails in `TestHiddenDistribution`:
  - `hidden_distribution_assigns_roles_from_random_preset` leaves players
    without roles and no leader assigned.
  - `fully_random_distribution_assigns_random_roles` can assign a duplicate
    card.
- The handoff notes existing full-suite `internal/handlers` failures around
  live server, SSE, and handler logic. They were not reverified in this slice.

## Port Assumptions

Current app config defaults:

- `configs/server-development.yaml`: `server.host: localhost`,
  `server.port: "8888"`, `metricsPort: "9090"`, metrics disabled, debug mode
  enabled.
- `configs/server-production.yaml`: `server.host: "0.0.0.0"`,
  `server.port: "8080"`, `metricsPort: "9090"`, metrics disabled.
- `HOST` and `PORT` environment variables override config values through the
  app config loader.

Current command-level assumptions:

- `dev` uses Air with `HOST=localhost`, `PORT=8888`, and the development config.
- `dev-css` starts templ with proxy `http://localhost:8080`, while Air still
  uses the `nix/app/.air.toml` runtime command on port `8888`. This is a
  pre-existing mismatch.
- `run` defaults to `localhost:8080` and the production config.
- `nix/app/README.md` still tells developers to visit `http://localhost:8080`
  after `dev`, which is stale relative to the current Air `PORT=8888`.

Current Playwright assumptions:

- `nix/app/playwright.config.ts` uses `baseURL: 'http://localhost:8080'` and
  expects the web server on port `8080`.
- `nix/app/tests/playwright/sse-countdown-sync.spec.ts` hardcodes
  `http://localhost:8080`.
- This is another pre-existing mismatch with the current `dev` command.

Current container assumptions:

- All three Nix container variants set `HOST=0.0.0.0` and `PORT=8080`.
- All three variants expose `8080/tcp`.
- Local run recipes map host port `8080` to container port `8080`.

## Container, Image, And Release Commands

Current `justfile` container recipes:

- Build only: `just build-container`, `just build-container-dev`,
  `just build-container-minimal`.
- Load into local Podman storage: `just load-container`,
  `just load-container-dev`, `just load-container-minimal`.
- Push to GHCR via local storage: `just push-container`,
  `just push-container-dev`, `just push-container-minimal`, `just push-all`.
- Push directly without local storage: `just push-direct`,
  `just push-direct-dev`, `just push-direct-minimal`.
- Run locally with Podman: `just run-container`, `just run-container-dev`,
  `just run-container-minimal`.
- Registry/auth/utilities: `just ghcr-login`, `just inspect-container`,
  `just list-images`, `just clean-images`.
- Tagged releases: `just release <tag>`, `just release-direct <tag>`.

Current image registry default:

- `ghcr.io/grimsleydl/treacherest`

The project does not yet have a committed CI/CD pipeline for these commands.

## Health And Smoke-Test Seam

The current HTTP readiness endpoints are:

- `GET /health/live`: returns `200 OK` with body `OK`.
- `GET /health/ready`: returns `200 OK` with body `OK`.

These endpoints are the intended process and container smoke-test seam for
later slices.

## Verification Performed

- Inspected `git status --short --branch --untracked-files=all`.
- Inspected current command definitions in `nix/local/shells.nix`,
  `nix/app/.air.toml`, and `justfile`.
- Inspected current port defaults in `configs/server-development.yaml`,
  `configs/server-production.yaml`, `nix/containers/containers.nix`,
  `nix/app/playwright.config.ts`, Playwright specs, and `nix/app/README.md`.
- Inspected `/health/live` and `/health/ready` route definitions in
  `nix/app/internal/handlers/router.go`.
- Reverified the focused `internal/game` hidden distribution failure with
  `CGO_ENABLED=0` inside `nix develop`.
