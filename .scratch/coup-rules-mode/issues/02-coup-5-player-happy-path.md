# Coup 5-Player Happy Path

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Implement the thinnest demoable Coup path: a 5-player Coup game assigns King, Blue Knight, Black Knight, Red Knight, and Green Knight; King is public after assignment; each non-King role remains hidden; each player can privately view their own role.

This should establish Coup as its own user-facing role taxonomy, not aliases for Leader, Guardian, Assassin, or Traitor.

## Acceptance criteria

- [ ] A Coup room with five active players can start successfully.
- [ ] The assigned roles are exactly King, Blue Knight, Black Knight, Red Knight, and Green Knight.
- [ ] King is revealed publicly after role assignment.
- [ ] Each player can privately view their own Coup role without seeing other hidden roles.
- [ ] Existing legacy/Treachery role assignment behavior is not changed.

## Blocked by

- `.scratch/coup-rules-mode/issues/01-rules-mode-selection-shell.md`
