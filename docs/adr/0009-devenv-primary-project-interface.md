# Devenv-primary project interface

Status: accepted

Treacherest treats `devenv` as the primary project interface for local development and project tasks, with `just` as the stable command facade over it. Plain Nix flake compatibility may remain where it is useful for packages, images, or external tooling, but the workflow should not be contorted to preserve vanilla `nix develop` behavior when it conflicts with a better devenv-native setup. This aligns Treacherest with the wider project portfolio, where other projects already use `devenv.nix`.
