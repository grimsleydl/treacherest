# Start As-Is Best-Effort Assignment

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Let the host use Start Override to start a room with only the current active players, intentionally bypassing normal start validation for underfilled or unusual table states.

Start As-Is should perform best-effort random role assignment for the current active players. For Coup, if at least one active player exists, the assignment must include King and then randomly fill remaining players from the selected Coup preset role pool.

## Acceptance criteria

- [ ] Host can trigger Start As-Is from the Debug Control Surface in Debug Mode.
- [ ] The action is unavailable or rejected when Debug Mode is disabled.
- [ ] The action is rejected for non-host clients.
- [ ] Start As-Is can start with fewer active players than normal validation requires.
- [ ] Start As-Is assigns roles only to the current active players and does not create Debug Players.
- [ ] Coup Start As-Is includes King when at least one active player exists.
- [ ] Coup Start As-Is randomly fills remaining active players from the selected Coup preset role pool.
- [ ] Underfilled Start As-Is states are visibly/debuggably marked as debug-started or validation-overridden.
- [ ] Tests cover underfilled start, Coup King guarantee, non-host rejection, and disabled-debug rejection.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
