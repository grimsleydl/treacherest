# Treacherest command facade

default:
    @just --list

registry := "ghcr.io/grimsleydl/treacherest"
coup_role_image_source := ".scratch/coup-role-images/final"
go_cache := ".scratch/go-cache"
go_tmp := ".scratch/go-tmp"
image_port := env_var_or_default("TREACHEREST_IMAGE_PORT", "8080")

# Start the devenv-managed local development process graph
dev:
    devenv up

# Run the normal Go test suite
test:
    devenv shell -- bash -lc 'cd nix/app && CGO_ENABLED=0 go test ./...'

# Run the desired full verification gate; this is expected to fail while known-red tests exist
check:
    devenv shell -- bash scripts/dev/check.sh

# Run the temporary known-green verification subset
check-known-green:
    devenv shell -- bash scripts/dev/check-known-green.sh

# Build the Nix package artifact
build:
    nix build .#packages.x86_64-linux.default

# Build the production OCI image
image:
    nix build .#containers.x86_64-linux.default

# Run the production OCI image locally and smoke-test /health/ready with a strict host port
image-run port=image_port:
    devenv shell -- bash scripts/dev/image-run-smoke.sh "{{registry}}" "{{port}}"

# Push the production OCI image; defaults to sha-<shortsha> when no tag is provided
image-push tag="":
    devenv shell -- bash scripts/dev/image-push.sh "{{registry}}" "{{tag}}"

# Push a release tag for the production OCI image
release tag:
    devenv shell -- bash scripts/dev/image-push.sh "{{registry}}" "{{tag}}"

# Create the local staging directory for generated Coup role images
prepare-coup-role-images:
    mkdir -p "{{coup_role_image_source}}"
    @echo "Save generated Coup role images in {{coup_role_image_source}}"

# Import generated Coup role images into canonical app assets
import-coup-role-images source=coup_role_image_source:
    mkdir -p "{{source}}" "{{go_cache}}" "{{go_tmp}}"
    src="$(realpath "{{source}}")"; cache="$(realpath "{{go_cache}}")"; tmp="$(realpath "{{go_tmp}}")"; cd nix/app; TMPDIR="$tmp" GOTMPDIR="$tmp" GOCACHE="$cache" CGO_ENABLED=0 go run ./scripts/coup-images -source "$src"

# Verify Coup role image import and runtime loading tests
test-coup-role-images:
    mkdir -p "{{go_cache}}" "{{go_tmp}}"
    cache="$(realpath "{{go_cache}}")"; tmp="$(realpath "{{go_tmp}}")"; cd nix/app; TMPDIR="$tmp" GOTMPDIR="$tmp" GOCACHE="$cache" CGO_ENABLED=0 go test ./scripts/coup-images ./internal/game -run 'TestImportCoupRoleImages|TestLoadCoupRoleImages' -count=1

_serve-test port="8080":
    devenv shell -- bash scripts/dev/serve-test.sh "{{port}}"
