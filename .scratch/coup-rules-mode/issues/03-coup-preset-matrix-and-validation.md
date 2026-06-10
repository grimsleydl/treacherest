# Coup Preset Matrix And Validation

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Expand Coup setup from the 5-player happy path to the documented preset matrix for 5-9 players, including the 8-player Wasteland chaos preset. The selected preset should be visible before start, validated against player count, and rejected with clear feedback when it does not fit.

## Acceptance criteria

- [x] Coup supports the documented 5, 6, 7, 8, 8 chaos, and 9 player presets.
- [x] Setup displays a readable summary of the selected Coup preset before start.
- [x] The game cannot start when the selected Coup preset does not match the active player count.
- [x] Wasteland Knight appears only in presets that include Wasteland.
- [x] Tests cover preset distributions and validation failures.

## Completion notes

- Added first-class Coup preset constants, summaries, player-count validation, and an ordered setup option list.
- Expanded Coup assignment to the documented 5-9 player matrix, including the 8-player chaos and 9-player Wasteland presets.
- Added a lobby preset selector with Datastar update handling and readable role summaries.
- Added start-game validation for selected preset/player-count mismatches.
- Added tests for preset distributions, Wasteland placement, selected-preset updates, route wiring, lobby display, and validation failures.

## Blocked by

- `.scratch/coup-rules-mode/issues/02-coup-5-player-happy-path.md`
