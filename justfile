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

# Repair generated devenv cache files when direnv reports failed enter tasks
repair-devenv:
    bash scripts/dev/repair-devenv-state.sh

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

# Load the nix2container production image into local Podman storage
image-load tag="":
    devenv shell -- bash scripts/dev/image-load.sh "{{registry}}" "{{tag}}"

# Smoke-test the loaded production image on a strict explicit host port
image-smoke port=image_port tag="":
    devenv shell -- bash scripts/dev/image-smoke.sh "{{registry}}" "{{port}}" "{{tag}}"

# Load and smoke-test the production image locally
image-run port=image_port tag="":
    devenv shell -- bash scripts/dev/image-load.sh "{{registry}}" "{{tag}}"
    devenv shell -- bash scripts/dev/image-smoke.sh "{{registry}}" "{{port}}" "{{tag}}"

# Push the production OCI image; defaults to short git SHA when clean, latest when dirty
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
