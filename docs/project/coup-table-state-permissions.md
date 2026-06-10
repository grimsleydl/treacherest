# Coup Table-State Control Permissions

## Decision

Coup table-state actions use a trust-the-table permission model with role-safe defaults:

- Public Reveal: the target player may reveal their own hidden role. A host/spectator may record a reveal only after the table has publicly seen or agreed to that reveal. Other players should not reveal someone else's hidden role from their own device.
- Elimination: a player may mark themself eliminated. A host/spectator may mark any player eliminated to keep table state accurate. Other players should not eliminate someone else from their own device.
- Advisory Win Prompt rejection: any active non-host player may reject a public advisory prompt after table discussion. A host/spectator may also record the table's rejection.
- Confirmed Win: any active non-host player may confirm a public advisory prompt after the table agrees. A host/spectator may record the table's confirmed decision, but does not count as a player vote.

## Rationale

Coup is for in-person Commander tables. The app should help record table state without becoming the hidden-information authority or requiring a hard moderator. The safest default is self-service for actions that expose a player's own hidden information, plus host/spectator override for public table state.

Advisory Win Prompts are explicitly non-final. Confirmation and rejection record the table's public decision; they do not reveal private role information by themselves.

## Host And Spectator Behavior

Host/spectator controls are moderator controls:

- They can record public reveals, eliminations, rejected prompts, and confirmed wins after table agreement.
- They should not receive private role information unless they are also a player with an assigned role.
- They should not be treated as a living player for victory checks, Inquisition confirmation, or win-prompt quorum.

## Implementation Guidance

Future implementation issues should enforce these defaults:

- Reveal endpoints should allow self-reveal and host/spectator record-reveal.
- Elimination endpoints should allow self-elimination and host/spectator record-elimination.
- Confirm/reject endpoints for Advisory Win Prompts should allow any active non-host player or host/spectator.
- UI copy should make clear that the app records public table decisions and does not adjudicate hidden information automatically.
