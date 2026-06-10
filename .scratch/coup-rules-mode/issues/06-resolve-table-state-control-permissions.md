# Resolve Table-State Control Permissions

Status: ready-for-human

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Decide who can perform table-state actions in Coup: public Reveal, Elimination, rejected Advisory Win Prompt recording, and Confirmed Win. The decision should choose a permission model that works for in-person Commander tables and does not leak hidden information.

This is a HITL decision issue. Do not implement the flows until the permission model is decided.

## Acceptance criteria

- [ ] The project has a documented decision for who may trigger public Reveal.
- [ ] The project has a documented decision for who may mark Elimination.
- [ ] The project has a documented decision for who may confirm or reject Advisory Win Prompts.
- [ ] The decision covers host/spectator behavior.
- [ ] Follow-up AFK implementation issues can reference the decision without asking the human again.

## Blocked by

- `.scratch/coup-rules-mode/issues/02-coup-5-player-happy-path.md`
