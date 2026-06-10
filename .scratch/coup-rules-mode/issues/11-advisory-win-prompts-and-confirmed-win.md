# Advisory Win Prompts And Confirmed Win

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add Advisory Win Prompts for Coup. The app may say a faction or player looks like they might have won based on tracked Reveal, Elimination, Inquisition, King Fall, and Strict Green Eligibility state. The app must require a human/table confirmation before producing a Confirmed Win and ending the game.

## Acceptance criteria

- [ ] The app can detect and display advisory prompts for King-side, Black, Red, Green-sharing, and Wasteland outcomes from tracked state.
- [ ] Advisory Win Prompts explain the tracked facts that triggered them.
- [ ] A prompt does not end the game until confirmed according to the permission model.
- [ ] A Confirmed Win ends the Coup game and shows the winning outcome.
- [ ] Rejected prompts do not incorrectly end or hide the game.
- [ ] Tests cover at least King-side, Black, Red, and Green Eligibility edge cases.

## Blocked by

- `.scratch/coup-rules-mode/issues/07-manual-reveal-and-elimination-tracking.md`
- `.scratch/coup-rules-mode/issues/09-public-inquisition-flow.md`
- `.scratch/coup-rules-mode/issues/10-private-inquisition-result-variant.md`
