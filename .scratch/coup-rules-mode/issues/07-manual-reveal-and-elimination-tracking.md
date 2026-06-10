# Manual Reveal And Elimination Tracking

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add Coup-safe manual Reveal and Elimination flows using the decided table-state permission model. A Reveal makes a hidden role public to the table. An Elimination marks a player as out for Coup victory purposes and reveals that player's role.

## Acceptance criteria

- [ ] Coup players can be manually Revealed according to the permission model.
- [ ] Coup players can be manually Eliminated according to the permission model.
- [ ] Eliminated players reveal their role publicly.
- [ ] Private role viewing remains distinct from public Reveal.
- [ ] SSE updates keep all clients in sync without leaking still-hidden roles.
- [ ] Tests cover Reveal, Elimination, and privacy boundaries.

## Blocked by

- `.scratch/coup-rules-mode/issues/06-resolve-table-state-control-permissions.md`
