# Coup Privacy Regression Coverage

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add focused regression coverage for Coup privacy boundaries. The tests should prove hidden roles, Information Policy outputs, Conspiracy Knowledge, Network variants, Public Inquisition Result, and Private Inquisition Result only render to the correct clients.

## Acceptance criteria

- [x] Tests prove hidden Coup roles are not rendered to other players before Reveal.
- [x] Tests prove King-to-Blue information is visible only to King.
- [x] Tests prove Red-to-Black information is visible only to Red.
- [x] Tests prove Black-to-Red and Network variants only reveal information to intended players.
- [x] Tests prove Private Inquisition Result is visible only to the Blue inquisitor.
- [x] Tests include a multi-client or equivalent integration path for SSE/privacy behavior.

## Completion notes

- Added page-level regression coverage for generated Coup information policy output.
- Covered King-to-Blue, Red-to-Black, Black-to-Red, and Black Network privacy boundaries.
- Added public Inquisition result visibility coverage and reused the private Inquisition regression.
- Added a handler-level multi-client equivalent test using the same `renderToString(pages.GameContent(...))` path used by game SSE updates.

## Verification

- `go test ./internal/views/pages -run 'TestGameBody_CoupInformationPolicyPrivacyBoundaries|TestGameBody_CoupPublicInquisitionResultVisibleToAllClients|TestGameBody_CoupPrivateInquisitionResultOnlyInformsInquisitor|TestGameBody_CoupPrivacy|TestGameBody_CoupPrivateInformationScopedToRecipient'`
- `go test ./internal/handlers -run 'TestRenderGameContent_CoupPrivacyIsScopedPerClientLikeSSE'`

## Blocked by

- `.scratch/coup-rules-mode/issues/04-coup-information-policies.md`
- `.scratch/coup-rules-mode/issues/09-public-inquisition-flow.md`
- `.scratch/coup-rules-mode/issues/10-private-inquisition-result-variant.md`
