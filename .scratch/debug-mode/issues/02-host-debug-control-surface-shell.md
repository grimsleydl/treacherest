# Host Debug Control Surface Shell

Status: done

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Add the host-only Debug Control Surface as the visible place for Debug Mode tools. It should appear only when Debug Mode is enabled and only in host contexts. It should not appear in ordinary player views.

This slice is the UI shell for later Start Override, Debug Insights, and View As Player tools. It should include clear Debug Mode labeling and preserve or relocate the existing backup-oriented debug actions without mixing them into player-facing UI.

## Acceptance criteria

- [x] A host can see a Debug Control Surface when Debug Mode is enabled.
- [x] A host cannot see the Debug Control Surface when Debug Mode is disabled.
- [x] Non-host player views do not render the Debug Control Surface.
- [x] Existing backup/debug persistence controls remain available to the host in Debug Mode.
- [x] The Debug Control Surface has stable targets/containers suitable for future Datastar updates.
- [x] Tests cover host rendering, non-host rendering, and disabled-debug rendering.

## Blocked by

- `.scratch/debug-mode/issues/01-debug-mode-boundary-and-host-authorization.md`

## Completion notes

- Reworked the debug-only host panel into a named `debug-control-surface` shell.
- Preserved the existing backup persistence controls and client handler under `debug-persistence-controls`.
- Added stable empty containers for future Start Override, Debug Insights, and View As Player updates:
  - `debug-start-override-controls`
  - `debug-insights-container`
  - `debug-view-as-player-container`
- Qualified host dashboard debug rendering on both server Debug Mode and the current room player being host.
- Added player-page regression coverage so ordinary player views do not render the Debug Control Surface or debug-only script.

## Verification

- `templ fmt internal/views/layouts/base.templ internal/views/pages/host_dashboard.templ`
- `build-templ`
- `go test ./internal/views/layouts -run TestBaseLayout -count=1`
- `go test ./internal/views/pages -run 'TestHostDashboardLobby_DebugPanelGatedByConfig|TestHostDashboardLobby_DebugControlSurfaceShell|TestHostDashboardLobby_DebugControlSurfaceRequiresHostPlayer|TestGamePage_DoesNotRenderDebugControlSurface|TestHostDashboardPlaying' -count=1`
