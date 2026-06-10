# Royal Guard Action And Settings

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add Blue Knight Royal Guard support. A Blue Knight must be revealed to use Royal Guard. The default rule allows a revealed Blue Knight, once each combat, to have any number of untapped creatures they control block creatures attacking the King as though those creatures were attacking the Blue Knight. Normal blocking restrictions apply.

Support a blocker-limit setting so the default can be changed to one blocker or another configured limit.

## Acceptance criteria

- [ ] Blue Knight players can access Royal Guard guidance/action UI.
- [ ] Using Royal Guard reveals Blue if Blue was still hidden.
- [ ] The default displayed Royal Guard rule allows any number of untapped Blue-controlled blockers.
- [ ] The blocker-limit setting can change the displayed/configured limit.
- [ ] The rule text makes clear that Royal Guard protects only the King player by default.
- [ ] Tests cover Blue reveal-on-use and blocker-limit configuration.

## Blocked by

- `.scratch/coup-rules-mode/issues/07-manual-reveal-and-elimination-tracking.md`
