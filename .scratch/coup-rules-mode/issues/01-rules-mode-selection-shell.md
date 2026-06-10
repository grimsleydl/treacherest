# Rules Mode Selection Shell

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add a top-level Rules Mode selection path so Treacherest can create a room as Coup or the existing legacy/Treachery mode. The selected Rules Mode should be stored with the game, visible in setup/lobby context, and should not break the existing room creation and join flow.

This slice does not need full Coup behavior yet. Coup can route to a minimal placeholder or disabled setup state as long as the selected Rules Mode is carried end to end.

## Acceptance criteria

- [ ] A user can choose a Rules Mode before or during room creation.
- [ ] Existing legacy/Treachery room creation still works after the change.
- [ ] A Coup selection is stored with the game and visible in the room/lobby.
- [ ] Tests cover Rules Mode persistence through create/join/start setup boundaries.

## Blocked by

None - can start immediately
