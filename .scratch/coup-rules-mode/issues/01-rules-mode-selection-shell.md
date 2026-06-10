# Rules Mode Selection Shell

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add a top-level Rules Mode selection path so Treacherest can create a room as Coup or the existing legacy/Treachery mode. The selected Rules Mode should be stored with the game, visible in setup/lobby context, and should not break the existing room creation and join flow.

This slice does not need full Coup behavior yet. Coup can route to a minimal placeholder or disabled setup state as long as the selected Rules Mode is carried end to end.

## Acceptance criteria

- [x] A user can choose a Rules Mode before or during room creation.
- [x] Existing legacy/Treachery room creation still works after the change.
- [x] A Coup selection is stored with the game and visible in the room/lobby.
- [x] Tests cover Rules Mode persistence through create/join/start setup boundaries.

## Completion notes

- Added `treachery` and `coup` Rules Mode values on rooms.
- Room creation defaults to Treachery and accepts Coup via the create-room form.
- Invalid Rules Mode submissions are rejected before a room is created.
- Lobby renders the selected Rules Mode.
- Coup rooms show a disabled setup placeholder and cannot start through the legacy Treachery role assignment path.

## Blocked by

None - can start immediately
