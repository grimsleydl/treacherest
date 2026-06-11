# Start As-Is Best-Effort Assignment

Status: done

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Let the host use Start Override to start a room with only the current active players, intentionally bypassing normal start validation for underfilled or unusual table states.

Start As-Is should perform best-effort random role assignment for the current active players. For Coup, if at least one active player exists, the assignment must include King and then randomly fill remaining players from the selected Coup preset role pool.

## Acceptance criteria

- [x] Host can trigger Start As-Is from the Debug Control Surface in Debug Mode.
- [x] The action is unavailable or rejected when Debug Mode is disabled.
- [x] The action is rejected for non-host clients.
- [x] Start As-Is can start with fewer active players than normal validation requires.
- [x] Start As-Is assigns roles only to the current active players and does not create Debug Players.
- [x] Coup Start As-Is includes King when at least one active player exists.
- [x] Coup Start As-Is randomly fills remaining active players from the selected Coup preset role pool.
- [x] Underfilled Start As-Is states are visibly/debuggably marked as debug-started or validation-overridden.
- [x] Tests cover underfilled start, Coup King guarantee, non-host rejection, and disabled-debug rejection.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`

## Completion notes

- Added a Debug Mode-only host route: `POST /room/{code}/debug/start-as-is`.
- Start As-Is bypasses normal validation and assigns roles only to current active players.
- Coup Start As-Is uses `AssignCoupRolesBestEffort`, which guarantees one King when at least one active player exists and fills remaining active players from the selected preset pool.
- Added `Room.DebugStartMode` with `as-is` and `with-debug-players` markers for later Debug Insights and validation-override visibility.
- Added a room-aware `Start As-Is` button to the Debug Control Surface.
- Added route-absence, non-host rejection, underfilled Coup, King guarantee, and no-Debug-Player regression coverage.

## Verification

- `gofmt -w internal/handlers/debug_actions.go internal/handlers/debug_mode_test.go internal/game/coup_roles.go internal/game/room.go`
- `templ fmt internal/views/layouts/base.templ internal/views/pages/host_dashboard.templ`
- `build-templ`
- `go test ./internal/handlers -run 'TestDebugModeRoutes_' -count=1`
- `go test ./internal/views/layouts -run TestBaseLayout -count=1`
- `go test ./internal/views/pages -run 'TestHostDashboardLobby_DebugPanelGatedByConfig|TestHostDashboardLobby_DebugControlSurfaceShell|TestHostDashboardLobby_DebugControlSurfaceRequiresHostPlayer|TestHostDashboardLobby_DebugPlayersAreVisiblyMarked|TestGamePage_DoesNotRenderDebugControlSurface|TestHostDashboardPlaying' -count=1`
- `go test ./internal/game -run 'TestAssignCoupRoles' -count=1`
