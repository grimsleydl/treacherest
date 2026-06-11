# Advisory Win Prompts And Confirmed Win

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add Advisory Win Prompts for Coup. The app may say a faction or player looks like they might have won based on tracked Reveal, Elimination, Inquisition, King Fall, and Strict Green Eligibility state. The app must require a human/table confirmation before producing a Confirmed Win and ending the game.

## Acceptance criteria

- [x] The app can detect and display advisory prompts for King-side, Black, Red, Green-sharing, and Wasteland outcomes from tracked state.
- [x] Advisory Win Prompts explain the tracked facts that triggered them.
- [x] A prompt does not end the game until confirmed according to the permission model.
- [x] A Confirmed Win ends the Coup game and shows the winning outcome.
- [x] Rejected prompts do not incorrectly end or hide the game.
- [x] Tests cover at least King-side, Black, Red, and Green Eligibility edge cases.

## Completion notes

- Added a small Coup advisory win detector for King-side, Black, Red, and Wasteland outcomes.
- Added Green-sharing facts for King-side and Red outcomes, including the strict Red-sharing lock before King fall.
- Added player/host prompt UI with Confirm Win and Reject Prompt actions.
- Added confirmed Coup win display for ended games.
- Added confirm/reject endpoints using the documented permission model: active non-host players or host/spectator can record the table decision.
- King elimination now records King fall and locks strict Green eligibility before the King is marked eliminated.

## Verification

- `go test ./internal/game -run 'TestCurrentCoupAdvisoryWin|TestAssignCoupRoles|TestCoup|TestPlayer'`
- `go test ./internal/handlers -run 'Test.*CoupWinPrompt|TestSetupRouter_RoutesCoupWinPromptDecisions|TestEliminatePlayer_CoupKingFallLocksGreenEligibilityBeforeBlueDies|Test.*CoupInquisition|TestHandler_StartGame_Coup|TestToggleReveal_Coup|TestEliminatePlayer_Coup|TestUseCoupRoyalGuard|TestSetupRouter_RoutesCoup|TestUpdateCoup'`
- `go test ./internal/views/pages -run 'TestGameBody_Coup|TestHostDashboardPlaying_Coup|TestLobbyPage/shows_coup'`

## Blocked by

- `.scratch/coup-rules-mode/issues/07-manual-reveal-and-elimination-tracking.md`
- `.scratch/coup-rules-mode/issues/09-public-inquisition-flow.md`
- `.scratch/coup-rules-mode/issues/10-private-inquisition-result-variant.md`
