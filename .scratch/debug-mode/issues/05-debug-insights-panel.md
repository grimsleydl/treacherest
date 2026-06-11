# Debug Insights Panel

Status: done

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Add host-only Debug Insights to the Debug Control Surface. Debug Insights should expose normally hidden or derived room facts so the host can verify Role Assignment, private information, Reveal and Elimination state, Rules Mode configuration, Coup state, and Advisory Win Prompt behavior.

The panel must remain absent from non-host views and unavailable outside Debug Mode.

## Acceptance criteria

- [x] Host can see Debug Insights in Debug Mode.
- [x] Debug Insights are absent when Debug Mode is disabled.
- [x] Non-host player views do not render Debug Insights.
- [x] Debug Insights show each active player's role, Reveal state, Elimination state, and Debug Player marker when present.
- [x] Debug Insights show private role information attached to each player's role.
- [x] Debug Insights show the selected Rules Mode and relevant configuration summary.
- [x] Debug Insights show Coup-specific state when the room uses Coup, including Inquisition state, King Fall, Green Eligibility, and current Advisory Win Prompt when present.
- [x] Regular player views do not receive hidden role/private information through this panel.
- [x] Tests cover host visibility, non-host absence, disabled-debug absence, and representative Coup insight content.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
- `.scratch/debug-mode/issues/03-start-override-with-debug-players.md`
- `.scratch/debug-mode/issues/04-start-as-is-best-effort-assignment.md`

## Completion notes

- Debug Control Surface now receives the room object when rendered from host dashboard pages.
- Added `DebugInsights` rendering for host-only Debug Mode pages.
- Insights include Rules Mode, Coup preset, Debug Start mode, active-player role/reveal/elimination/debug-seat state, private role rulings, Coup King Fall, Green Eligibility, Inquisition status, and advisory/confirmed win outcome.
- Existing disabled-debug, non-host, and player-page tests now explicitly assert that Debug Insights are absent outside the host Debug Mode surface.

## Verification

- `templ fmt internal/views/layouts/base.templ internal/views/pages/host_dashboard.templ`
- `build-templ`
- `go test ./internal/views/layouts -run TestBaseLayout -count=1`
- `go test ./internal/views/pages -run 'TestHostDashboardLobby_DebugPanelGatedByConfig|TestHostDashboardLobby_DebugControlSurfaceShell|TestHostDashboardLobby_DebugControlSurfaceRequiresHostPlayer|TestHostDashboardLobby_DebugPlayersAreVisiblyMarked|TestHostDashboardLobby_DebugInsightsShowRepresentativeCoupState|TestGamePage_DoesNotRenderDebugControlSurface|TestHostDashboardPlaying' -count=1`
