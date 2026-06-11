# Debug Insights Panel

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## What to build

Add host-only Debug Insights to the Debug Control Surface. Debug Insights should expose normally hidden or derived room facts so the host can verify Role Assignment, private information, Reveal and Elimination state, Rules Mode configuration, Coup state, and Advisory Win Prompt behavior.

The panel must remain absent from non-host views and unavailable outside Debug Mode.

## Acceptance criteria

- [ ] Host can see Debug Insights in Debug Mode.
- [ ] Debug Insights are absent when Debug Mode is disabled.
- [ ] Non-host player views do not render Debug Insights.
- [ ] Debug Insights show each active player's role, Reveal state, Elimination state, and Debug Player marker when present.
- [ ] Debug Insights show private role information attached to each player's role.
- [ ] Debug Insights show the selected Rules Mode and relevant configuration summary.
- [ ] Debug Insights show Coup-specific state when the room uses Coup, including Inquisition state, King Fall, Green Eligibility, and current Advisory Win Prompt when present.
- [ ] Regular player views do not receive hidden role/private information through this panel.
- [ ] Tests cover host visibility, non-host absence, disabled-debug absence, and representative Coup insight content.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
- `.scratch/debug-mode/issues/03-start-override-with-debug-players.md`
- `.scratch/debug-mode/issues/04-start-as-is-best-effort-assignment.md`
