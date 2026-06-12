# Audience-Split Redesign Surfaces

Treacherest's redesign uses audience-specific surfaces rather than one mixed room page. A Player View is for a participant's own play experience. An Operator Dashboard is for public room and table management. A Debug Control Surface is for privileged local development aids and remains governed by the debug-mode absence boundary.

The Operator Dashboard is not an omniscient moderator view. Outside explicit debug-only controls, it must show public table state and room-management controls without revealing other players' hidden roles or private rules-mode information. A Room Operator may also be a player; before game start they can reach the Operator Dashboard to configure and start the room, but after game start the normal path for a playing Room Operator is their Player View with a low-emphasis navigation affordance back to the Operator Dashboard. A non-playing Host remains on the Operator Dashboard.

Private information uses Privy Panels. A Privy Panel is visually concealed by default for shoulder-surfing protection, but privacy still depends on the server only rendering the current player's own private information. Private capability UI, such as a Blue Knight's Inquisition caller form, belongs inside that player's Privy Panel. Public notices and public results belong in public table notice areas.

The redesign keeps Treacherest server-rendered. It should be implemented with Go, Templ, Datastar attributes, SSE fragments, CSS, and small transient Datastar signals. It should not introduce React, a SPA router, a client-side game store, or client-side authority logic. SSE patch targets should be stable zones that avoid reinitializing the outer SSE connection wrapper.

Consequences:

- `treacherest` is the product-default theme, with `treacherest-day` as the light counterpart and legacy DaisyUI themes retained for regression checks.
- Non-operator Player Views must not render room configuration forms and hide them with CSS; those controls belong on Operator Dashboard surfaces.
- Playing Room Operators need explicit navigation between Player View and Operator Dashboard, not inline operator controls in the ordinary player surface.
- Confirm-Twice Buttons replace browser-native confirmation dialogs for consequential product and operator actions.
- Debug Insights may redact hidden-role spoilers by default for screen-sharing safety, but that redaction is not a security boundary and must not be copied into normal privacy handling.
- Browser and DOM privacy tests should verify that hidden roles and private rules-mode information are absent from unauthorized surfaces, not merely hidden visually.
