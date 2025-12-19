# Treacherest Container Build & Deploy
# Usage: just <recipe>

# Default recipe - show help
default:
    @just --list

# Container registry settings
registry := "ghcr.io/grimsleydl/treacherest"

# ============================================================
# Container Build Recipes
# ============================================================

# Build production container
build-container:
    nix build .#containers.x86_64-linux.default
    @echo "Built production container. Use 'just load-container' to load into Docker."

# Build development container (includes debugging tools)
build-container-dev:
    nix build .#containers.x86_64-linux.dev
    @echo "Built dev container. Use 'just load-container-dev' to load into Docker."

# Build minimal container (just the binary)
build-container-minimal:
    nix build .#containers.x86_64-linux.minimal
    @echo "Built minimal container. Use 'just load-container-minimal' to load into Docker."

# ============================================================
# Container Load Recipes (into Docker daemon)
# ============================================================

# Load production container into Docker
load-container: build-container
    ./result/copyTo docker-daemon:{{registry}}:latest
    @echo "Loaded {{registry}}:latest into Docker"

# Load dev container into Docker
load-container-dev: build-container-dev
    ./result/copyTo docker-daemon:{{registry}}:dev
    @echo "Loaded {{registry}}:dev into Docker"

# Load minimal container into Docker
load-container-minimal: build-container-minimal
    ./result/copyTo docker-daemon:{{registry}}:minimal
    @echo "Loaded {{registry}}:minimal into Docker"

# ============================================================
# Container Push Recipes (to GHCR)
# ============================================================

# Push production container to GHCR
push-container: load-container
    docker push {{registry}}:latest
    @echo "Pushed {{registry}}:latest to GHCR"

# Push dev container to GHCR
push-container-dev: load-container-dev
    docker push {{registry}}:dev
    @echo "Pushed {{registry}}:dev to GHCR"

# Push minimal container to GHCR
push-container-minimal: load-container-minimal
    docker push {{registry}}:minimal
    @echo "Pushed {{registry}}:minimal to GHCR"

# Push all container variants to GHCR
push-all: push-container push-container-dev push-container-minimal
    @echo "All containers pushed to GHCR"

# ============================================================
# Direct Push (no Docker daemon required)
# ============================================================

# Push production container directly to GHCR (requires skopeo auth)
push-direct: build-container
    ./result/copyTo docker://{{registry}}:latest
    @echo "Pushed {{registry}}:latest directly to GHCR"

# Push dev container directly to GHCR
push-direct-dev: build-container-dev
    ./result/copyTo docker://{{registry}}:dev
    @echo "Pushed {{registry}}:dev directly to GHCR"

# Push minimal container directly to GHCR
push-direct-minimal: build-container-minimal
    ./result/copyTo docker://{{registry}}:minimal
    @echo "Pushed {{registry}}:minimal directly to GHCR"

# ============================================================
# Container Run Recipes
# ============================================================

# Run production container locally
run-container: load-container
    docker run --rm -p 8080:8080 {{registry}}:latest

# Run dev container locally with shell
run-container-dev: load-container-dev
    docker run --rm -it -p 8080:8080 {{registry}}:dev

# Run minimal container locally
run-container-minimal: load-container-minimal
    docker run --rm -p 8080:8080 {{registry}}:minimal

# ============================================================
# Auth & Utility
# ============================================================

# Login to GitHub Container Registry
ghcr-login:
    @echo "Logging into GHCR..."
    @echo "Make sure GITHUB_TOKEN is set with packages:write scope"
    echo $GITHUB_TOKEN | docker login ghcr.io -u grimsleydl --password-stdin

# Show container image details
inspect-container: load-container
    docker inspect {{registry}}:latest | jq '.[0].Config'

# List all local treacherest images
list-images:
    docker images | grep treacherest || echo "No treacherest images found"

# Remove all local treacherest images
clean-images:
    docker images --format '{{{{.Repository}}}}:{{{{.Tag}}}}' | grep treacherest | xargs -r docker rmi
    @echo "Cleaned up local treacherest images"

# ============================================================
# Tagged Releases
# ============================================================

# Build and push a tagged release (e.g., just release v1.0.0)
release tag: build-container
    ./result/copyTo docker-daemon:{{registry}}:{{tag}}
    docker push {{registry}}:{{tag}}
    @echo "Released {{registry}}:{{tag}}"
