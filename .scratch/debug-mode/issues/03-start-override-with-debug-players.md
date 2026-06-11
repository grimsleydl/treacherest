# Start Override With Debug Players

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Let the host use Start Override to fill missing table seats with stable Debug Players and then start the game through normal role assignment where possible.

Debug Players are synthetic active players. Once created, they persist in the room, are visible in player lists, count for role assignment and game logic, and can later be selected by View As Player. They should be clearly marked as debug seats.

## Acceptance criteria

- [ ] Host can trigger Start Override with Debug Players from the Debug Control Surface in Debug Mode.
- [ ] The action is unavailable or rejected when Debug Mode is disabled.
- [ ] The action is rejected for non-host clients.
- [ ] Missing seats are filled with stable Debug Players using generated names such as `Debug Player 1`.
- [ ] Debug Players persist in the room as active visible players after start.
- [ ] Debug Players are distinguishable from hosts and real players.
- [ ] Role Assignment runs against the real players plus Debug Players.
- [ ] Debug Players count for reveal, elimination, targeting, and win-prompt logic as active seats.
- [ ] Tests cover Debug Player creation, persistence, visibility, host authorization, and role assignment.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
