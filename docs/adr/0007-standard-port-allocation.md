# Standard port allocation

Status: accepted

Treacherest uses a fixed runtime/container port of `8080`, while interactive local development may choose an available host port from a small project-owned range and must print the actual localhost URL at startup. Automatic port fallback is appropriate for `just dev`; CI, release, deployment, and container smoke-test commands should use explicit ports and fail fast on collision. This keeps production and container behavior predictable while avoiding unnecessary friction when multiple local projects are running.
