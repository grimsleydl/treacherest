# Host Debug Control Surface Shell

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Add the host-only Debug Control Surface as the visible place for Debug Mode tools. It should appear only when Debug Mode is enabled and only in host contexts. It should not appear in ordinary player views.

This slice is the UI shell for later Start Override, Debug Insights, and View As Player tools. It should include clear Debug Mode labeling and preserve or relocate the existing backup-oriented debug actions without mixing them into player-facing UI.

## Acceptance criteria

- [ ] A host can see a Debug Control Surface when Debug Mode is enabled.
- [ ] A host cannot see the Debug Control Surface when Debug Mode is disabled.
- [ ] Non-host player views do not render the Debug Control Surface.
- [ ] Existing backup/debug persistence controls remain available to the host in Debug Mode.
- [ ] The Debug Control Surface has stable targets/containers suitable for future Datastar updates.
- [ ] Tests cover host rendering, non-host rendering, and disabled-debug rendering.

## Blocked by

- `.scratch/debug-mode/issues/01-debug-mode-boundary-and-host-authorization.md`
