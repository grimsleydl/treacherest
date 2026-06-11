# Private Inquisition Result Variant

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add the Private Inquisition Result variant. In this variant, a correct Inquisition tells only the Blue inquisitor that the named player is Red. Red remains hidden to the table, while the app still records that Inquisition succeeded for Green Eligibility and Advisory Win Prompt logic.

## Acceptance criteria

- [x] Coup setup can select Public Inquisition Result or Private Inquisition Result.
- [x] Under Private Inquisition Result, a correct guess informs only the Blue inquisitor.
- [x] Under Private Inquisition Result, Red is not publicly Revealed by the app.
- [x] Inquisition success is still recorded for downstream rules.
- [x] Tests prove private result information does not leak to other clients.

## Completion notes

- Added a Coup Inquisition result policy with Public as the default and Private as an opt-in setup variant.
- Private successful Inquisition records success without publicly revealing Red.
- Private successful Inquisition result text identifies Red only to the Blue inquisitor; other clients see a generic success notice.
- Added handler, setup UI, routing, and rendering tests for the private result path.

## Verification

- `go test ./internal/views/pages -run 'TestGameBody_CoupPrivateInquisitionResultOnlyInformsInquisitor|TestGameBody_CoupInquisition|TestGameBody_CoupPrivacy|TestGameBody_CoupRoyalGuardBlockerLimit|TestLobbyPage/shows_coup'`
- `go test ./internal/handlers -run 'TestConfirmCoupInquisition_PrivateResultRecordsSuccessWithoutRevealingRed|TestUpdateCoupInquisitionSettings|TestSetupRouter_RoutesCoupInquisitionSettings|Test.*CoupInquisition|TestHandler_StartGame_Coup|TestToggleReveal_Coup|TestEliminatePlayer_Coup|TestUseCoupRoyalGuard|TestSetupRouter_RoutesCoup|TestUpdateCoup'`
- `go test ./internal/game -run 'TestAssignCoupRoles|TestCoup|TestPlayer'`

## Blocked by

- `.scratch/coup-rules-mode/issues/09-public-inquisition-flow.md`
