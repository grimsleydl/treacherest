# Public Inquisition Flow

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Implement the default Public Inquisition Result flow. A Blue Knight calls Inquisition in-app, names a suspected Red Knight, is revealed, and waits for one living non-Blue witness to confirm the Inquisition Notice. After confirmation, a correct guess reveals Red publicly; an incorrect guess keeps the target hidden and tells Blue to lose half their current life total, rounded up.

## Acceptance criteria

- [x] A Blue Knight can call Inquisition once per Blue per game.
- [x] Calling Inquisition reveals Blue publicly.
- [x] The app requires one living non-Blue witness confirmation before showing the result.
- [x] A correct public Inquisition reveals Red to all players.
- [x] An incorrect Inquisition keeps the target role hidden and displays the half-current-life-rounded-up penalty.
- [x] The app records Inquisition success for later Green Eligibility and Advisory Win Prompt logic.
- [x] Tests cover success, failure, witness confirmation, and no-result-before-confirmation behavior.

## Completion notes

- Added Coup Inquisition state with per-Blue attempt tracking, pending notice, last result, and success flag.
- Added call and witness-confirm endpoints; calling reveals Blue, confirmation resolves success/failure, and correct public results reveal Red.
- Enforced one call per Blue and one living non-Blue witness before resolution.
- Added in-game Inquisition UI for Blue call, pending notice confirmation, and failed-result life-loss guidance.
- Added tests for pending/no-result behavior, success, failure, route wiring, once-per-Blue, and witness restrictions.

## Blocked by

- `.scratch/coup-rules-mode/issues/07-manual-reveal-and-elimination-tracking.md`
