# Agent Dev Workflow

Use the same command contract as humans.

## Primary Commands

- Enter the project shell with `devenv shell`.
- Start local development with `just dev`.
- Use `just check-known-green` for a passing transition gate.
- Use `just check` when you need the full desired gate; it is currently allowed
  to fail on documented known-red application tests.
- Use `just build`, `just image`, and `just image-run` for package and image
  verification.

Do not use Docker Compose or Podman Compose for local development.

## Ports

- Interactive dev starts at `8888` and may auto-increment through `8899`.
- Runtime/container app port is fixed at `8080`.
- Container smoke tests use explicit host ports and should fail on collision.
- Browser tests should use `BASE_URL` when they need a specific server URL.

## Secrets

Do not read, stage, commit, or bake secrets into images. Secrets-provider choice
is intentionally deferred.

## Known Red

Before claiming a workflow regression, check
`docs/project/dev-environment-workflow.md` for the current known-red test
baseline. Use `just check-known-green` when a passing gate is required during
the transition.
