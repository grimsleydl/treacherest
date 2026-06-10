# Public Inquisition Flow

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Implement the default Public Inquisition Result flow. A Blue Knight calls Inquisition in-app, names a suspected Red Knight, is revealed, and waits for one living non-Blue witness to confirm the Inquisition Notice. After confirmation, a correct guess reveals Red publicly; an incorrect guess keeps the target hidden and tells Blue to lose half their current life total, rounded up.

## Acceptance criteria

- [ ] A Blue Knight can call Inquisition once per Blue per game.
- [ ] Calling Inquisition reveals Blue publicly.
- [ ] The app requires one living non-Blue witness confirmation before showing the result.
- [ ] A correct public Inquisition reveals Red to all players.
- [ ] An incorrect Inquisition keeps the target role hidden and displays the half-current-life-rounded-up penalty.
- [ ] The app records Inquisition success for later Green Eligibility and Advisory Win Prompt logic.
- [ ] Tests cover success, failure, witness confirmation, and no-result-before-confirmation behavior.

## Blocked by

- `.scratch/coup-rules-mode/issues/07-manual-reveal-and-elimination-tracking.md`
