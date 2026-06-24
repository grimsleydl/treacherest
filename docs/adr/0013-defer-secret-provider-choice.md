# Defer secret provider choice

Status: accepted

Treacherest will not choose a full secrets-management provider as part of the dev environment standardization work. The binding rule for now is that secrets must not be committed to git and must not be baked into container images; provider choices for local development, CI, Podman/Quadlet, and managed runtimes are deferred until deployment work needs them.
