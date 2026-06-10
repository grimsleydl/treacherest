# Manual Reveal And Elimination Tracking

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add Coup-safe manual Reveal and Elimination flows using the decided table-state permission model. A Reveal makes a hidden role public to the table. An Elimination marks a player as out for Coup victory purposes and reveals that player's role.

## Acceptance criteria

- [x] Coup players can be manually Revealed according to the permission model.
- [x] Coup players can be manually Eliminated according to the permission model.
- [x] Eliminated players reveal their role publicly.
- [x] Private role viewing remains distinct from public Reveal.
- [x] SSE updates keep all clients in sync without leaking still-hidden roles.
- [x] Tests cover Reveal, Elimination, and privacy boundaries.

## Completion notes

- Added Coup public Reveal handling for self-reveal and host/spectator record-reveal; Coup reveal is idempotent and does not hide already-public roles.
- Updated elimination so eliminated players become publicly revealed and face up.
- Added player-facing public Reveal controls and host-dashboard record Reveal/Elimination controls.
- Updated host SSE to refresh on public reveal and elimination events.
- Added handler, host SSE, and template tests for permission/state behavior and hidden-role privacy.

## Blocked by

- `.scratch/coup-rules-mode/issues/06-resolve-table-state-control-permissions.md`
