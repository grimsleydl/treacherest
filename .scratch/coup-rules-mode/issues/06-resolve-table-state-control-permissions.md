# Resolve Table-State Control Permissions

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Decide who can perform table-state actions in Coup: public Reveal, Elimination, rejected Advisory Win Prompt recording, and Confirmed Win. The decision should choose a permission model that works for in-person Commander tables and does not leak hidden information.

This is a HITL decision issue. Do not implement the flows until the permission model is decided.

## Acceptance criteria

- [x] The project has a documented decision for who may trigger public Reveal.
- [x] The project has a documented decision for who may mark Elimination.
- [x] The project has a documented decision for who may confirm or reject Advisory Win Prompts.
- [x] The decision covers host/spectator behavior.
- [x] Follow-up AFK implementation issues can reference the decision without asking the human again.

## Completion notes

- Documented the decision in `docs/project/coup-table-state-permissions.md`.
- Public Reveal is self-reveal by default, with host/spectator record-reveal after public table agreement.
- Elimination is self-elimination by default, with host/spectator record-elimination.
- Advisory Win Prompt confirm/reject can be recorded by any active non-host player or by host/spectator after table agreement.
- Host/spectator controls are moderator controls and should not grant private role information or count as player votes.

## Blocked by

- `.scratch/coup-rules-mode/issues/02-coup-5-player-happy-path.md`
