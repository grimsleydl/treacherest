# Coup Preset Matrix And Validation

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Expand Coup setup from the 5-player happy path to the documented preset matrix for 5-9 players, including the 8-player Wasteland chaos preset. The selected preset should be visible before start, validated against player count, and rejected with clear feedback when it does not fit.

## Acceptance criteria

- [ ] Coup supports the documented 5, 6, 7, 8, 8 chaos, and 9 player presets.
- [ ] Setup displays a readable summary of the selected Coup preset before start.
- [ ] The game cannot start when the selected Coup preset does not match the active player count.
- [ ] Wasteland Knight appears only in presets that include Wasteland.
- [ ] Tests cover preset distributions and validation failures.

## Blocked by

- `.scratch/coup-rules-mode/issues/02-coup-5-player-happy-path.md`
