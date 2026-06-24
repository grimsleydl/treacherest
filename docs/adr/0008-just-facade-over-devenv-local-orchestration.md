# Just facade over devenv local orchestration

Status: accepted

Treacherest uses `just` as the stable command facade and `devenv` as the local development environment and process orchestrator. `just dev` should delegate to `devenv up`, while `devenv.nix` owns local process definitions, readiness checks, development tasks, and automatic dev port allocation. Docker Compose and Podman Compose are intentionally outside the local development standard; Quadlet may be considered for deployment, not for the interactive dev loop.
