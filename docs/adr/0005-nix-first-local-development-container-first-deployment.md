# Nix-first local development and container-first deployment

Status: superseded by ADR-0009

Treacherest initially planned to standardize on Nix as the local development and build contract, and OCI containers as the deployment and runtime contract. This was refined by ADR-0009: `devenv` is the primary project interface for local development and project tasks, while Nix remains the underlying reproducibility layer where it is useful for packages, images, or external tooling. The container-first deployment side of this decision remains in force.
