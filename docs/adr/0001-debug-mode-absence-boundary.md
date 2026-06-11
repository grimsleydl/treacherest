# Debug Mode Surfaces Are Absent Outside Debug Mode

Treacherest has privileged debug aids for playtesting hidden-role flows, but those aids can expose hidden information or bypass normal table constraints. We decided that debug routes, rendered debug UI, and debug-only client scripts should be registered or served only when Debug Mode is enabled; handler-level checks may remain as defense-in-depth, but non-debug runs should not expose reachable debug surfaces at all. This trades a little conditional wiring complexity for a clearer safety boundary: outside Debug Mode, there should be no functional path into debug behavior.

When Debug Mode is enabled, debug controls remain host-only. Client-side Datastar behavior may keep the interface responsive, but authorization must be checked on the server by resolving the current room player and confirming host status. Debug Impersonation is a host capability, not a way for a non-host player client to gain debug privileges.

Debug Impersonation starts as view-only: the host may inspect a selected player's perspective, but player actions such as Inquisition or Royal Guard should still require the real player flow unless a later decision explicitly introduces action-capable impersonation.

Start Override may either fill missing seats with Debug Players or start with the current active players as-is. When starting Coup as-is, the assignment should still include the King when at least one player exists, then randomly fill the remaining real players from the selected role pool.
