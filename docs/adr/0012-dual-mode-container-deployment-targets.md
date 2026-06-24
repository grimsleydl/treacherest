# Dual-mode container deployment targets

Status: accepted

Treacherest targets OCI images that can run in both single-host Podman deployments and managed container runtimes such as Google Cloud Run. Single-host Podman, preferably managed by Quadlet/systemd, is the self-managed deployment default; managed runtimes remain first-class compatible targets and should not require a separate image shape. Deployment design should therefore avoid Docker Compose assumptions and keep runtime configuration environment-driven.
