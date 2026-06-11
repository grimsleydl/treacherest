# Debug Mode Boundary And Host Authorization

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Harden the Debug Mode boundary so debug routes, rendered debug UI, and debug-only client scripts are absent unless Debug Mode is enabled. Any debug action that remains callable must still be server-authorized as host-only.

This slice should preserve normal non-debug behavior while making it impossible for production-like runs to reach functional debug behavior by URL, crafted request, or hidden client markup.

## Acceptance criteria

- [ ] Debug routes are not registered or reachable when Debug Mode is disabled.
- [ ] Debug UI and debug-only client scripts are not rendered when Debug Mode is disabled.
- [ ] Debug routes are registered and usable when Debug Mode is enabled.
- [ ] Debug actions require the current room player to be a host, even when Debug Mode is enabled.
- [ ] Non-host clients cannot use debug endpoints by crafting requests.
- [ ] Existing backup-oriented debug behavior still works for hosts when Debug Mode is enabled.
- [ ] Tests cover route absence, route availability, host authorization, and non-host rejection.

## Blocked by

None - can start immediately
