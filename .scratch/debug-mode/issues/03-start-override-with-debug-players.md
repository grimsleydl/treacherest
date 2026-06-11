# Start Override With Debug Players

Status: done

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Let the host use Start Override to fill missing table seats with stable Debug Players and then start the game through normal role assignment where possible.

Debug Players are synthetic active players. Once created, they persist in the room, are visible in player lists, count for role assignment and game logic, and can later be selected by View As Player. They should be clearly marked as debug seats.

## Acceptance criteria

- [x] Host can trigger Start Override with Debug Players from the Debug Control Surface in Debug Mode.
- [x] The action is unavailable or rejected when Debug Mode is disabled.
- [x] The action is rejected for non-host clients.
- [x] Missing seats are filled with stable Debug Players using generated names such as `Debug Player 1`.
- [x] Debug Players persist in the room as active visible players after start.
- [x] Debug Players are distinguishable from hosts and real players.
- [x] Role Assignment runs against the real players plus Debug Players.
- [x] Debug Players count for reveal, elimination, targeting, and win-prompt logic as active seats.
- [x] Tests cover Debug Player creation, persistence, visibility, host authorization, and role assignment.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`

## Completion notes

- Added a Debug Mode-only host route: `POST /room/{code}/debug/start-with-debug-players`.
- The route fills missing seats to the active target count, then delegates to the existing normal start path.
- Coup rooms use the selected Coup preset player count; non-Coup rooms use the room role configuration target count when available.
- Generated seats use stable names and IDs such as `Debug Player 1` and `debug-{ROOM}-1`.
- Added `Player.IsDebug` so synthetic seats are distinguishable from real players and hosts.
- Host dashboard player lists render a visible `Debug` badge for debug seats.
- The Debug Control Surface now includes a room-aware `Start with Debug Players` Datastar POST control.
- Added an elimination regression to verify Debug Players persist as ordinary active seats for existing game actions.

## Verification

- `gofmt -w internal/handlers/debug_actions.go internal/handlers/debug_mode_test.go internal/game/player.go`
- `templ fmt internal/views/layouts/base.templ internal/views/pages/host_dashboard.templ`
- `build-templ`
- `go test ./internal/handlers -run 'TestDebugModeRoutes_' -count=1`
- `go test ./internal/views/layouts -run TestBaseLayout -count=1`
- `go test ./internal/views/pages -run 'TestHostDashboardLobby_DebugPanelGatedByConfig|TestHostDashboardLobby_DebugControlSurfaceShell|TestHostDashboardLobby_DebugControlSurfaceRequiresHostPlayer|TestHostDashboardLobby_DebugPlayersAreVisiblyMarked|TestGamePage_DoesNotRenderDebugControlSurface|TestHostDashboardPlaying' -count=1`
