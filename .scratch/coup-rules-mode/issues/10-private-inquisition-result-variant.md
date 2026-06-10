# Private Inquisition Result Variant

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add the Private Inquisition Result variant. In this variant, a correct Inquisition tells only the Blue inquisitor that the named player is Red. Red remains hidden to the table, while the app still records that Inquisition succeeded for Green Eligibility and Advisory Win Prompt logic.

## Acceptance criteria

- [ ] Coup setup can select Public Inquisition Result or Private Inquisition Result.
- [ ] Under Private Inquisition Result, a correct guess informs only the Blue inquisitor.
- [ ] Under Private Inquisition Result, Red is not publicly Revealed by the app.
- [ ] Inquisition success is still recorded for downstream rules.
- [ ] Tests prove private result information does not leak to other clients.

## Blocked by

- `.scratch/coup-rules-mode/issues/09-public-inquisition-flow.md`
