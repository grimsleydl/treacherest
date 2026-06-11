# Debug Mode PRD

Status: ready-for-agent

## Problem Statement

Treacherest hidden-role flows are hard to test locally because they normally require enough real browser sessions to satisfy the selected Rules Mode. Coup makes this more obvious: the host often needs to verify private information, role assignment, Inquisition visibility, Green Eligibility state, and advisory win prompts across many player perspectives.

Today Debug Mode exists only as a small backup-oriented panel. It does not give the host a clear, safe, server-gated way to start underfilled games, add synthetic seats, inspect hidden state, or view the game as a player. The result is slow local testing and increased risk that agents or developers confuse debug-only behavior with production player behavior.

## Solution

Expand Debug Mode into a host-only Debug Control Surface that is available only when Debug Mode is enabled. The Debug Control Surface lets the host start games through explicit Start Override options, fill missing seats with stable Debug Players, start an underfilled room as-is with best-effort role assignment, inspect hidden/derived room facts through Debug Insights, and use a read-only View As Player control to inspect player perspectives.

Outside Debug Mode, debug routes, debug UI, and debug-only client scripts should not be registered, rendered, served, or functional. In Debug Mode, controls remain host-only and server-authorized; Datastar can keep the UI reactive, but server-side checks are the boundary.

## User Stories

1. As a local developer, I want Debug Mode capabilities to be available when I launch the app for development, so that I can test hidden-role flows without manual setup friction.
2. As a local developer, I want debug routes to be absent outside Debug Mode, so that production-like runs cannot accidentally expose privileged behavior.
3. As a host, I want debug controls to appear only on the host surface, so that regular player views stay representative of real play.
4. As a host, I want debug actions to be rejected unless my current room player is the host, so that a non-host client cannot craft debug requests.
5. As a host, I want a clear Debug Control Surface, so that debug tools are discoverable without mixing them into normal player controls.
6. As a host, I want Start Override options, so that I can test a game even when normal start validation fails.
7. As a host, I want to start with Debug Players, so that I can fill missing table seats without opening extra browser sessions.
8. As a host, I want Debug Players to persist as room players once created, so that the table state remains stable across inspection and backup/restore.
9. As a host, I want Debug Players to have stable generated names like `Debug Player 1`, so that logs and View As Player selections remain understandable.
10. As a host, I want Debug Players to appear in normal player lists, so that the table state is visible and realistic.
11. As a real player in a debug-started game, I want Debug Players to be clearly labeled, so that I know which seats are synthetic.
12. As a host, I want Debug Players to count as active seats, so that role assignment, targeting, reveal, elimination, and win prompt logic can be exercised realistically.
13. As a host, I want to start a room as-is, so that I can intentionally test incomplete or unusual table states.
14. As a host, I want Start As-Is to assign roles to the current active players automatically, so that I do not need to map every role manually before testing.
15. As a host testing Coup Start As-Is, I want the King to be included whenever at least one active player exists, so that underfilled Coup games still have their central role.
16. As a host testing Coup Start As-Is, I want remaining roles to be selected randomly from the selected preset pool, so that the app can quickly create varied debug states.
17. As a host, I want Debug Insights to show every player's assigned role, so that I can verify Role Assignment without peeking at every client.
18. As a host, I want Debug Insights to show public Reveal and Elimination state, so that I can confirm Game State Tracking after actions.
19. As a host, I want Debug Insights to show private role information, so that I can verify Information Policy outputs.
20. As a host testing Coup, I want Debug Insights to show Coup-specific state, so that I can verify Inquisition, King Fall, Green Eligibility, and advisory win behavior.
21. As a host, I want Debug Insights to show the selected Rules Mode and relevant configuration, so that I can understand which rules produced the current state.
22. As a host, I want Debug Insights to show the current Advisory Win Prompt, so that I can understand why the app thinks a win may have happened.
23. As a host, I want a View As Player selector in setup/lobby contexts, so that I can inspect player context before and after debug setup actions.
24. As a host, I want a View As Player selector during play, so that I can inspect each player perspective from one browser.
25. As a host, I want View As Player to be read-only, so that inspecting a player perspective does not accidentally perform player actions.
26. As a host, I want View As Player to render the selected player's private information exactly as that player would see it, so that privacy bugs are easier to spot.
27. As a host, I want View As Player to work for Debug Players, so that synthetic seats can be inspected without a browser session.
28. As a regular player, I do not want to see debug controls, so that my view remains representative of the actual game experience.
29. As a regular player, I do not want debug-only secrets to leak through my rendered page, so that hidden-role privacy remains testable.
30. As an agent implementing this feature, I want the Debug Mode safety boundary documented and enforced, so that future changes do not accidentally expose debug behavior.

## Implementation Decisions

- Debug Mode is the canonical term. "Dev mode" is only one common way to run the server with Debug Mode enabled.
- The Debug Mode safety boundary is defined by ADR 0001: debug routes, debug UI, and debug-only client scripts should be absent outside Debug Mode.
- Debug Mode controls are host-only. The server must authorize debug actions by resolving the current room player and confirming host status.
- Datastar may be used for interactivity and live updates, but it is not the authorization boundary.
- The existing backup-oriented debug panel should evolve into or be complemented by a Debug Control Surface rather than becoming player-facing UI.
- Start Override has two MVP modes: Start with Debug Players and Start As-Is.
- Start with Debug Players creates enough stable Debug Players to satisfy the selected table size or preset, then uses normal Role Assignment where possible.
- Debug Players persist in the room as visible active seats. They count for role assignment, targeting, Reveal, Elimination, win prompts, and View As Player.
- Debug Players should be clearly marked as debug seats in host/debug UI and normal player lists.
- Start As-Is starts with the current active players and uses best-effort random role assignment.
- Coup Start As-Is must include King whenever at least one active player exists, then randomly fill the remaining active players from the selected Coup preset pool.
- Debug Insights is host-only and should expose player roles, private information, Reveal state, Elimination state, Debug Player markers, Rules Mode/config summary, Coup state, and current Advisory Win Prompt.
- View As Player is the player-facing label for Debug Impersonation.
- View As Player is read-only in MVP. It must not allow the host to call player actions such as Inquisition or Royal Guard as the selected player.
- Direct player role assignment is not part of the MVP. It should be treated as a follow-up slice after Start Override, Debug Insights, and View As Player are in place.
- Action-capable Debug Impersonation is not part of the MVP and should require a later explicit decision.

## Testing Decisions

- Good tests should verify external behavior at the highest available seam, not private helper details. The key behaviors are route absence, host authorization, rendered privacy, start outcomes, and room state after debug actions.
- Debug boundary tests should verify that debug routes are not registered or reachable when Debug Mode is disabled, and that debug UI/scripts are not rendered in non-debug pages.
- Host authorization tests should verify that host clients can use debug endpoints in Debug Mode and non-host clients cannot, even when crafting requests.
- Start with Debug Players tests should verify that missing seats are filled with stable Debug Players and normal Role Assignment runs against the resulting active player set.
- Start As-Is tests should verify best-effort assignment with only real active players, including the Coup King guarantee.
- Debug Player tests should verify that Debug Players are active, visible, stable after room updates, and distinguishable from hosts and real players.
- Debug Insights tests should verify that the host sees hidden role/private information and regular players do not.
- View As Player tests should verify that the host can render selected player perspectives read-only and that player actions are not exposed through that read-only surface.
- Existing prior art includes handler-level tests for start flow and Coup actions, template rendering tests for privacy boundaries, SSE-equivalent render tests for multi-client privacy, and host dashboard rendering tests.
- Browser tests should be added for UI changes when feasible. If Playwright MCP is used, follow the project requirement to use a sub-agent for browser automation.

## Out of Scope

- Direct role assignment by host.
- Action-capable Debug Impersonation where the host performs actions as a selected player.
- Production admin/moderator features.
- Public player-facing debug controls.
- Debug capabilities when Debug Mode is disabled.
- Replacing normal Role Assignment or Game State Tracking with a separate debug-only game engine.
- Raw backup inspection beyond the existing backup/debug persistence controls.

## Further Notes

- This PRD applies across Treacherest, not only Coup. Coup has the sharpest MVP requirements because its private information and win logic are more complex.
- Treachery/legacy Start As-Is anchor-role behavior was not explicitly resolved. A conservative implementation should preserve existing assignment semantics where possible and avoid adding Treachery-specific role guarantees unless separately decided.
- The current code already has a `DebugModeEnabled` configuration field, a development config with Debug Mode enabled, an SSE debug signal, and a small backup-oriented debug panel. The new work should harden the boundary and expand the host-only surface rather than creating a parallel debug concept.
