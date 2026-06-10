# Coup 5-Player Happy Path

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Implement the thinnest demoable Coup path: a 5-player Coup game assigns King, Blue Knight, Black Knight, Red Knight, and Green Knight; King is public after assignment; each non-King role remains hidden; each player can privately view their own role.

This should establish Coup as its own user-facing role taxonomy, not aliases for Leader, Guardian, Assassin, or Traitor.

## Acceptance criteria

- [x] A Coup room with five active players can start successfully.
- [x] The assigned roles are exactly King, Blue Knight, Black Knight, Red Knight, and Green Knight.
- [x] King is revealed publicly after role assignment.
- [x] Each player can privately view their own Coup role without seeing other hidden roles.
- [x] Existing legacy/Treachery role assignment behavior is not changed.

## Completion notes

- Added the initial five-player Coup role assignment path.
- Added first-class Coup role cards for King, Blue Knight, Black Knight, Red Knight, and Green Knight.
- Valid five-player Coup rooms start through the existing Datastar redirect/countdown path.
- Unsupported Coup player counts are rejected without assigning roles.
- The game page keeps non-King Coup roles private while still showing the current player's own role and the public King.

## Blocked by

- `.scratch/coup-rules-mode/issues/01-rules-mode-selection-shell.md`
