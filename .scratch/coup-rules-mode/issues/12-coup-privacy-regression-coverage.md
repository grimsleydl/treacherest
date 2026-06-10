# Coup Privacy Regression Coverage

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add focused regression coverage for Coup privacy boundaries. The tests should prove hidden roles, Information Policy outputs, Conspiracy Knowledge, Network variants, Public Inquisition Result, and Private Inquisition Result only render to the correct clients.

## Acceptance criteria

- [ ] Tests prove hidden Coup roles are not rendered to other players before Reveal.
- [ ] Tests prove King-to-Blue information is visible only to King.
- [ ] Tests prove Red-to-Black information is visible only to Red.
- [ ] Tests prove Black-to-Red and Network variants only reveal information to intended players.
- [ ] Tests prove Private Inquisition Result is visible only to the Blue inquisitor.
- [ ] Tests include a multi-client or equivalent integration path for SSE/privacy behavior.

## Blocked by

- `.scratch/coup-rules-mode/issues/04-coup-information-policies.md`
- `.scratch/coup-rules-mode/issues/09-public-inquisition-flow.md`
- `.scratch/coup-rules-mode/issues/10-private-inquisition-result-variant.md`
