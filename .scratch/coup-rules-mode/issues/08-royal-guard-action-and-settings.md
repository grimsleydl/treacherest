# Royal Guard Action And Settings

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add Blue Knight Royal Guard support. A Blue Knight must be revealed to use Royal Guard. The default rule allows a revealed Blue Knight, once each combat, to have any number of untapped creatures they control block creatures attacking the King as though those creatures were attacking the Blue Knight. Normal blocking restrictions apply.

Support a blocker-limit setting so the default can be changed to one blocker or another configured limit.

## Acceptance criteria

- [x] Blue Knight players can access Royal Guard guidance/action UI.
- [x] Using Royal Guard reveals Blue if Blue was still hidden.
- [x] The default displayed Royal Guard rule allows any number of untapped Blue-controlled blockers.
- [x] The blocker-limit setting can change the displayed/configured limit.
- [x] The rule text makes clear that Royal Guard protects only the King player by default.
- [x] Tests cover Blue reveal-on-use and blocker-limit configuration.

## Completion notes

- Added `CoupRoyalGuardBlockerLimit` room state with default unlimited blockers.
- Added Royal Guard rule text generation for unlimited, one-blocker, and numeric blocker-limit variants.
- Added a Blue-only Royal Guard action endpoint; using Royal Guard reveals Blue publicly and refreshes clients through the existing reveal SSE path.
- Added lobby configuration for Royal Guard blocker limit and in-game Blue Knight guidance/action UI.
- Added tests for rule text, reveal-on-use, route wiring, settings persistence, and configured UI text.

## Blocked by

- `.scratch/coup-rules-mode/issues/07-manual-reveal-and-elimination-tracking.md`
