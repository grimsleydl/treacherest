# MTG Treacherest

A real-time multiplayer game of deception and hidden roles, built with Go,
Templ, Datastar, and DaisyUI.

## Development

Treacherest uses `devenv` as the primary project environment and `just` as the
stable command facade.

`nix develop` remains a compatibility tool shell, not the primary workflow
surface. Use `devenv shell` and the `just` commands below for normal work.

### Prerequisites

- Nix
- `devenv` 2.1 or newer
- Podman only for image smoke tests and local container runs

Docker Compose and Podman Compose are not part of the local development
workflow.

### Getting Started

1. Enter the project shell:

   ```bash
   devenv shell
   ```

2. Start the local development process graph:

   ```bash
   just dev
   ```

3. Open the `http://localhost:<port>` URL printed by the startup logs.

Interactive development starts from port `8888` and auto-increments within the
Treacherest dev range when the base port is occupied. Runtime and container
paths stay fixed on app port `8080`.

### Standard Commands

- `just dev`: start the `devenv up` process graph.
- `just test`: run the normal Go test suite.
- `just check`: run the desired full verification gate. This command is honest
  about current known-red tests and may fail until those application defects are
  fixed.
- `just check-known-green`: run the temporary trusted passing subset.
- `just build`: build the Nix package artifact.
- `just image`: build the production OCI image.
- `just image-load [tag]`: load the Nix-built image into local Podman storage.
- `just image-smoke [port] [tag]`: run the loaded image with an explicit strict
  host port and verify `/health/ready`.
- `just image-run [port] [tag]`: load and smoke-test the image locally.
- `just image-push [tag]`: push the image manually until CI/CD exists. With no
  tag it uses the short git SHA for a clean tree and `latest` for a dirty tree.
- `just release <tag>`: push a release tag.

Project-specific Coup role image helpers remain available through `just --list`.

### Testing

`just check-known-green` is the day-to-day transition gate. It currently covers:

- Templ generation
- CSS build
- selected Go packages that pass today
- theme readability tests

`just check` is the full target gate and is expected to stay red while known
application failures remain. Known-red failures are tracked in
`docs/project/dev-environment-workflow.md`.

### Dependencies

Go dependencies are managed through gomod2nix. To add or update dependencies:

```bash
go get package@version
gomod2nix generate
```

Do not commit secrets or bake them into images. Secrets-provider selection is
intentionally deferred.
