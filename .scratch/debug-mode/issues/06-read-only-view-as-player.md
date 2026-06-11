# Read-Only View As Player

Status: done

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Add a host-only View As Player selector to the Debug Control Surface. The host should be able to select a real player or Debug Player and inspect that player's rendered perspective from setup/lobby contexts and during play.

View As Player is read-only for MVP. It should render what the selected player would see, including private role and information visibility, but must not expose or enable player actions through the host impersonation surface.

## Acceptance criteria

- [x] Host can select a real player to View As Player in Debug Mode.
- [x] Host can select a Debug Player to View As Player after Debug Players exist.
- [x] View As Player is unavailable when Debug Mode is disabled.
- [x] View As Player is rejected or absent for non-host clients.
- [x] The rendered perspective includes the selected player's private role and allowed private information.
- [x] The rendered perspective does not expose other players' hidden information beyond what the selected player should see.
- [x] The View As Player surface is read-only and does not expose player action controls such as Inquisition or Royal Guard.
- [x] The selector works in setup/lobby contexts and during play.
- [x] Tests cover host selection, non-host rejection, selected-player privacy, Debug Player selection, and read-only behavior.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
- `.scratch/debug-mode/issues/03-start-override-with-debug-players.md`

## Completion notes

- Added a Debug Mode-only host route: `GET /room/{code}/debug/view-as/{playerID}`.
- Added a `DebugViewAsPlayerSelector` in the Debug Control Surface listing real players and Debug Players.
- Added a read-only `DebugViewAsPlayerPerspective` fragment showing the selected player's own role and rulings without rendering player action controls.
- The read-only perspective shows other players only through role names that are already publicly revealed.
- Added disabled-route, non-host rejection, selected-player privacy, Debug Player selector, and read-only action-control regression coverage.

## Verification

- `gofmt -w internal/handlers/debug_actions.go internal/handlers/debug_mode_test.go`
- `templ fmt internal/views/components/debug_view_as.templ internal/views/layouts/base.templ internal/views/pages/host_dashboard.templ`
- `build-templ`
- `go test ./internal/handlers -run 'TestDebugModeRoutes_' -count=1`
- `go test ./internal/views/layouts -run TestBaseLayout -count=1`
- `go test ./internal/views/pages -run 'TestHostDashboardLobby_DebugPanelGatedByConfig|TestHostDashboardLobby_DebugControlSurfaceShell|TestHostDashboardLobby_DebugControlSurfaceRequiresHostPlayer|TestHostDashboardLobby_DebugPlayersAreVisiblyMarked|TestHostDashboardLobby_DebugInsightsShowRepresentativeCoupState|TestHostDashboardLobby_ViewAsPlayerSelectorIncludesDebugPlayers|TestGamePage_DoesNotRenderDebugControlSurface|TestHostDashboardPlaying' -count=1`
- `go test ./internal/game -run 'TestAssignCoupRoles' -count=1`
