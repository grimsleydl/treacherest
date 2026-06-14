# Room Operator Debug Authority

The room creator's browser session is the Room Operator for that room, independent of whether the creator is a playing participant or a non-playing Host surface. Room management controls and Operator Dashboard SSE streams require Room Operator authority; debug controls additionally require server Debug Mode to be enabled. This replaces earlier shortcuts that treated `Player.IsHost`, the first active player, or a host cookie as sufficient authority.

Debug Impersonation is a full Debug Mode capability for Room Operators: the operator may render and act as a selected Viewed Player, including synthetic Debug Players, while retaining operator-only access to the Debug Control Surface. The impersonated player never gains operator authority, non-operator players are not served debug controls, and debug perspective overrides have no effect when Debug Mode is disabled.

Consequences:

- New rooms must record operator-session metadata at creation time.
- Rooms without operator metadata should not infer authority from player order or Host status.
- Player actions under Debug Impersonation must resolve against the Viewed Player, while room-management and debug actions must resolve against the Operator Session.
- Removed Viewed Players clear the debug perspective override; eliminated players remain valid targets for testing eliminated-player UI.
