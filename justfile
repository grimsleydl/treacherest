# Treacherest Container Build & Deploy
# Usage: just <recipe>

# Default recipe - show help
default:
    @just --list

# Container registry settings
registry := "ghcr.io/grimsleydl/treacherest"
coup_role_image_source := ".scratch/coup-role-images/final"

# ============================================================
# Coup Role Image Recipes
# ============================================================

# Create the local staging directory for generated Coup role images
prepare-coup-role-images:
    mkdir -p "{{coup_role_image_source}}"
    @echo "Save generated Coup role images in {{coup_role_image_source}}"

# Import generated Coup role images into canonical app assets
import-coup-role-images source=coup_role_image_source:
    mkdir -p "{{source}}"
    src="$$(realpath "{{source}}")"; cd nix/app; go run ./scripts/coup-images -source "$$src"

# Verify Coup role image import and runtime loading tests
test-coup-role-images:
    cd nix/app && go test ./scripts/coup-images ./internal/game -run 'TestImportCoupRoleImages|TestLoadCoupRoleImages' -count=1

# ============================================================
# Container Build Recipes (nix build only)
# ============================================================

# Build production container
build-container:
    nix build .#containers.x86_64-linux.default
    @echo "Built production container"

# Build dev container
build-container-dev:
    nix build .#containers.x86_64-linux.dev
    @echo "Built dev container"

# Build minimal container
build-container-minimal:
    nix build .#containers.x86_64-linux.minimal
    @echo "Built minimal container"

# ============================================================
# Container Load Recipes (build + load into Podman)
# ============================================================

# Build and load production container into Podman
load-container:
    nix run .#containers.x86_64-linux.default.copyTo -- containers-storage:{{registry}}:latest
    @echo "Loaded {{registry}}:latest into Podman"

# Build and load dev container into Podman
load-container-dev:
    nix run .#containers.x86_64-linux.dev.copyTo -- containers-storage:{{registry}}:dev
    @echo "Loaded {{registry}}:dev into Podman"

# Build and load minimal container into Podman
load-container-minimal:
    nix run .#containers.x86_64-linux.minimal.copyTo -- containers-storage:{{registry}}:minimal
    @echo "Loaded {{registry}}:minimal into Podman"

# ============================================================
# Container Push Recipes (to GHCR)
# ============================================================

# Push production container to GHCR
push-container: load-container
    podman push {{registry}}:latest
    @echo "Pushed {{registry}}:latest to GHCR"

# Push dev container to GHCR
push-container-dev: load-container-dev
    podman push {{registry}}:dev
    @echo "Pushed {{registry}}:dev to GHCR"

# Push minimal container to GHCR
push-container-minimal: load-container-minimal
    podman push {{registry}}:minimal
    @echo "Pushed {{registry}}:minimal to GHCR"

# Push all container variants to GHCR
push-all: push-container push-container-dev push-container-minimal
    @echo "All containers pushed to GHCR"

# ============================================================
# Direct Push (no local storage required)
# ============================================================

# Push production container directly to GHCR
push-direct:
    nix run .#containers.x86_64-linux.default.copyTo -- docker://{{registry}}:latest
    @echo "Pushed {{registry}}:latest directly to GHCR"

# Push dev container directly to GHCR
push-direct-dev:
    nix run .#containers.x86_64-linux.dev.copyTo -- docker://{{registry}}:dev
    @echo "Pushed {{registry}}:dev directly to GHCR"

# Push minimal container directly to GHCR
push-direct-minimal:
    nix run .#containers.x86_64-linux.minimal.copyTo -- docker://{{registry}}:minimal
    @echo "Pushed {{registry}}:minimal directly to GHCR"

# ============================================================
# Container Run Recipes
# ============================================================

# Run production container locally
run-container: load-container
    podman run --rm -p 8080:8080 {{registry}}:latest

# Run dev container locally with shell
run-container-dev: load-container-dev
    podman run --rm -it -p 8080:8080 {{registry}}:dev

# Run minimal container locally
run-container-minimal: load-container-minimal
    podman run --rm -p 8080:8080 {{registry}}:minimal

# ============================================================
# Auth & Utility
# ============================================================

# Login to GitHub Container Registry
ghcr-login:
    @echo "Logging into GHCR..."
    @echo "Make sure GITHUB_TOKEN is set with packages:write scope"
    echo $GITHUB_TOKEN | podman login ghcr.io -u grimsleydl --password-stdin

# Show container image details
inspect-container: load-container
    podman inspect {{registry}}:latest | jq '.[0].Config'

# List all local treacherest images
list-images:
    podman images | grep treacherest || echo "No treacherest images found"

# Remove all local treacherest images
clean-images:
    podman images --format '{{{{.Repository}}}}:{{{{.Tag}}}}' | grep treacherest | xargs -r podman rmi
    @echo "Cleaned up local treacherest images"

# ============================================================
# Tagged Releases
# ============================================================

# Build and push a tagged release (e.g., just release v1.0.0)
release tag:
    nix run .#containers.x86_64-linux.default.copyTo -- containers-storage:{{registry}}:{{tag}}
    podman push {{registry}}:{{tag}}
    @echo "Released {{registry}}:{{tag}}"

# Direct push a tagged release (e.g., just release-direct v1.0.0)
release-direct tag:
    nix run .#containers.x86_64-linux.default.copyTo -- docker://{{registry}}:{{tag}}
    @echo "Released {{registry}}:{{tag}} directly to GHCR"
