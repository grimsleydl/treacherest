# Non-Host Debug Control Surface Discovery

Status: ready-for-agent

## Parent

`.scratch/debug-mode/PRD.md`

## Type

AFK

## What to build

Make Debug Mode controls discoverable and reachable from non-host mode when the server is running with Debug Mode enabled. A developer who creates or joins a room through the normal player flow should be able to find the Debug Control Surface without needing to know that it currently appears only on the host dashboard.

The implementation should not weaken the Debug Mode safety boundary. Debug UI, debug routes, and debug-only behavior must still be absent when Debug Mode is disabled. Any privileged debug action that exposes hidden state or mutates room state should remain server-authorized. If the selected UX exposes host-only controls from a player page, the server must still reject unauthorized action attempts.

Prefer a narrow, reviewable MVP: expose an obvious Debug Mode entry point or panel on non-host/player room pages, and use copy or disabled states to make clear which controls are informational, which require host authority, and which are unavailable from the current player context.

## Acceptance criteria

- [ ] In Debug Mode, a non-host/player room view includes an obvious way to access or view Debug Mode tooling.
- [ ] The non-host/player debug entry point is absent when Debug Mode is disabled.
- [ ] The non-host/player debug surface does not leak hidden role/private information unless the current request is explicitly authorized to see it.
- [ ] Host-only debug actions still require host authorization on the server.
- [ ] Non-host attempts to invoke privileged debug endpoints are rejected and covered by tests.
- [ ] The UX makes the current authority level clear, for example by showing view-only controls, disabled host-only controls, or a link to switch to the host dashboard where appropriate.
- [ ] Tests cover Debug Mode enabled/disabled rendering on non-host pages and server-side authorization for any exposed debug actions.

## Blocked by

- `.scratch/debug-mode/issues/02-host-debug-control-surface-shell.md`
- `.scratch/debug-mode/issues/05-debug-insights-panel.md`
- `.scratch/debug-mode/issues/06-read-only-view-as-player.md`

## Notes

- This issue exists because starting with `dev` correctly enables Debug Mode, but the shipped MVP renders the full Debug Control Surface only for host-mode pages.
- Do not implement production admin features here; this is local Debug Mode discoverability and access only.
