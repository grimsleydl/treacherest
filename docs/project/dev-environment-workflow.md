# Dev Environment Workflow

Last updated: 2026-06-24.

Treacherest uses `devenv` as the primary project interface and `just` as the
human command facade. The accepted local workflow is native development through
`devenv`; containers are runtime and deployment artifacts.

Plain `nix develop` remains a compatibility tool shell for callers that need
flake-provided packages, but it is not the command surface. Use `devenv shell`
for the project environment and `just --list` for supported workflow commands.

## Primary Local Workflow

```bash
devenv shell
just dev
```

`just dev` delegates to `devenv up`. The `devenv` process graph starts:

- CSS watch/build through `npm run watch:css`.
- Templ watch/proxy through `templ generate --watch`.
- The Go server watcher through Air.

The Go server process receives `HOST`, `PORT`, and `CONFIG_PATH` from the
`devenv` process definition. Air no longer hardcodes the development port.

Startup logs print the effective URL:

```text
Treacherest dev server: http://localhost:<port>
```

## Ports

- Runtime/container app port: fixed `8080`.
- Interactive development host port: starts at `8888`.
- Interactive development range: `8888-8899`.
- `devenv` auto-allocates from the base port; the process exits if allocation
  moves outside the Treacherest range.
- Container smoke tests use explicit host-port mapping and fail fast if the
  selected host port is already in use.
- Production, release, and deployment-like paths do not use automatic fallback
  ports.

Browser tests use `BASE_URL` when supplied. Without `BASE_URL`, Playwright uses
`http://localhost:8080` and starts a deterministic test server through the
private `just _serve-test` helper.

## Standard Command Contract

- `just dev`: start local development through `devenv up`.
- `just test`: run the normal Go test suite.
- `just check`: run the desired full verification gate.
- `just check-known-green`: run the temporary trusted passing subset.
- `just build`: build the Nix package artifact.
- `just image`: build the production OCI image.
- `just image-load [tag]`: load the Nix-built production image into local
  Podman storage. Without a tag this uses the short git SHA for a clean tree
  and `latest` for a dirty tree.
- `just image-smoke [port] [tag]`: run the already-loaded production image and
  smoke-test `/health/ready`.
- `just image-run [port] [tag]`: load the production image locally and
  smoke-test `/health/ready`.
- `just image-push [tag]`: manually push the image until CI/CD exists. Without a
  tag this uses the short git SHA for a clean tree and `latest` for a dirty
  tree.
- `just release <tag>`: push a release tag.

Project-specific commands may exist, but they should name the local concern
explicitly and not replace the standard contract.

## Checks

`just check` is intentionally the full desired gate. It does not hide known
application failures.

`just check-known-green` is transitional and should shrink away once `just
check` is clean. It currently verifies:

- `templ generate`
- `npm run build:css`
- `go test ./internal/config ./internal/views/... ./internal/game/ability -count=1`
- `npm run test:theme-lab`

Known-red baseline failures observed during this migration:

- `cmd/server`: config startup test path fails with `PORT environment variable
  must be set`.
- `internal/game`: `TestHiddenDistribution` role assignment failures.
- `internal/handlers`: live server, SSE, redirect, and handler status
  expectations fail in the full suite.
- `internal/store`: `TestDefaultGameSize` role-count mismatch.

## Images And Deployment

OCI containers are the runtime artifact. `just image` builds the production
image through the Nix container output. `just image-load` copies that
nix2container output into local Podman storage. `just image-smoke` runs the
loaded image with an explicit host port and verifies:

```text
GET /health/ready
```

`just image-run` is the convenience form for `image-load` followed by
`image-smoke`.

The load and push steps use the nix2container `copyTo` app. That copy path may
build or run `skopeo` locally because `skopeo` performs OCI image copies between
Nix outputs, local container storage, and registries.

Image tags should use immutable identities:

- Short git SHAs for source revision tags.
- `latest` for dirty local working-tree images and other convenience-only
  cases.
- Semantic release tags through `just release <tag>`.
- `latest` is convenience only and must not be used as the production deployment
  identity.

OCI labels include image source, revision, title, version, and license where
available.

The same production image shape is intended to work for single-host
Podman/Quadlet deployments and managed container runtimes such as Google Cloud
Run. Quadlet is deployment tooling, not the local development loop.

Manual `just image-push` remains available because the repository does not have
CI/CD yet. CI image publishing remains the target release model.

## Secrets

Secrets-provider selection is deferred. The current rule is:

- Do not commit secrets.
- Do not bake secrets into images.
- Keep committed runtime config limited to non-secret defaults.

## Compose

Docker Compose and Podman Compose are not part of the Treacherest local
development standard.
