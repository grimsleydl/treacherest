# Debug Mode Boundary And Host Authorization

Status: done

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Harden the Debug Mode boundary so debug routes, rendered debug UI, and debug-only client scripts are absent unless Debug Mode is enabled. Any debug action that remains callable must still be server-authorized as host-only.

This slice should preserve normal non-debug behavior while making it impossible for production-like runs to reach functional debug behavior by URL, crafted request, or hidden client markup.

## Acceptance criteria

- [x] Debug routes are not registered or reachable when Debug Mode is disabled.
- [x] Debug UI and debug-only client scripts are not rendered when Debug Mode is disabled.
- [x] Debug routes are registered and usable when Debug Mode is enabled.
- [x] Debug actions require the current room player to be a host, even when Debug Mode is enabled.
- [x] Non-host clients cannot use debug endpoints by crafting requests.
- [x] Existing backup-oriented debug behavior still works for hosts when Debug Mode is enabled.
- [x] Tests cover route absence, route availability, host authorization, and non-host rejection.

## Blocked by

None - can start immediately

## Completion notes

- Debug routes are now registered only when Debug Mode is enabled.
- The debug clear action still checks the server-side Debug Mode flag and now also requires the current room player to be a host.
- The base layout no longer renders debug panel markup or debug click-handler script by default.
- Host dashboard full-page renders use the debug-aware layout only when server config enables Debug Mode.

## Verification

- `TMPDIR=/workspace/treacherest/.scratch/go-tmp GOTMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./internal/handlers -run 'TestDebugModeRoutes_' -count=1`
- `TMPDIR=/workspace/treacherest/.scratch/go-tmp GOTMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./internal/views/layouts -run TestBaseLayout -count=1`
- `TMPDIR=/workspace/treacherest/.scratch/go-tmp GOTMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./internal/views/pages -run 'TestHostDashboardLobby_DebugPanelGatedByConfig|TestHostDashboardPlaying' -count=1`
