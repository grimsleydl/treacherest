# Redesign QA Checklist

Use this checklist for the Treacherest audience-split redesign release sweep. Automated checks can prove DOM and server-rendered behavior; visual quality and accessibility sign-off still require human review.

## Automated Checks

- Game page stable zones render in order: `#zone-status`, `#zone-privy`, `#zone-notices`, `#zone-actions`, `#zone-roster`.
- `#zone-notices` uses `aria-live="polite"`.
- Countdown renders with `role="timer"`.
- SyncPill renders with `role="status"` for live, reconnecting, and stale states.
- Privy Panel renders concealed by default, has keyboard-operated open/conceal buttons, auto-conceals, and uses local Datastar signals only.
- Privy Panel hold-to-peek uses pointer events only; no polling, mousemove loop, animation frame loop, or network request.
- Confirm-Twice actions replace browser `confirm()` on redesigned player reveal, elimination, Royal Guard, Inquisition, advisory win, and Debug Clear flows.
- Non-operator lobby DOM contains no configuration forms, config endpoint hooks, unsafe override copy, start controls, or debug controls.
- Player roster rows show only public role state for other players; hidden roles render as face down without hidden role names, text, colors, or role-specific classes.
- Operator Dashboard live view is public-state-only outside the Debug Control Surface.
- Private Coup information-policy text is rendered only for the entitled player.
- Private Inquisition result details render only inside the inquisitor's Privy Panel; other clients receive only the neutral public notice.
- Debug Control Surface is absent outside debug mode and absent for non-operator sessions.
- Debug hidden-role spoilers are redacted by default and suppress role-color presentation until explicitly shown.
- Datastar/SSE patches include stable target wrappers and do not reinitialize wrappers that own `data-init` SSE connections.
- Selectable legacy themes have generated DaisyUI color-token blocks and do not inherit the Treacherest bespoke palette.
- All selectable themes have generated DaisyUI color-token blocks, and `npm run audit:themes` has been reviewed for low-contrast token pairs.
- The Theme Readability Lab has been generated with `npm run lab:themes`, and copied OK/Bad notes have been reviewed for selected themes.

Suggested focused commands:

```sh
cd nix/app && npm run build:css && npm run audit:themes && npm run lab:themes && npm run test:theme-lab
env TMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./internal/views/pages ./internal/views/components ./internal/views/layouts -count=1
env TMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./internal/handlers -run 'TestStreamGame|TestStreamHost|TestRenderGameContent_CoupPrivacyIsScopedPerClientLikeSSE|TestPreStartSettingsFreeze|TestUpdateCoup|TestUpdateTreacheryPlayerCount|TestDebugModeRoutes' -count=1
env TMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 build
```

## Human Visual And Accessibility Review

Capture and inspect screenshots for:

- Mobile Player View in `treacherest`.
- Mobile Player View in `treacherest-day`.
- Mobile Player View in one legacy theme, preferably `dracula`.
- Mobile lobby in `treacherest` and `treacherest-day`.
- Desktop or tablet Operator Dashboard pre-game setup.
- Desktop or tablet live Operator Dashboard.
- Debug Control Surface right rail with spoilers hidden and shown.
- Privy Panel concealed, peeking, open, and compact public-role states.
- Advisory win prompt and confirmed win notice.

Review manually:

- Focus rings are visible on all interactive controls.
- Keyboard navigation reaches and activates the Privy Panel open/conceal controls without pointer hold.
- Native disclosures/details are keyboard-operable and their expanded state is visually clear.
- Debug panel minimize/expand is keyboard-operable and communicates state.
- `prefers-reduced-motion` removes or materially reduces redesign transitions.
- Chip-size text has acceptable contrast in `treacherest` and `treacherest-day`.
- Player-surface controls meet the 44px touch target.
- Long names, room codes, and button labels do not overlap or overflow at mobile widths.
- Operator Dashboard does not imply the operator can see hidden roles unless the explicit Debug Control Surface spoiler toggle is active.
- Any remaining browser `alert()` or `confirm()` in metamorph, transformation, ability-modal, backup/restore, or debug helper flows is recorded as an opportunistic follow-up rather than a redesign release blocker.

## Sign-Off Record

Record release-specific results below when running the sweep.

| Date | Reviewer | Build/Commit | Automated Checks | Screenshots | Notes |
| --- | --- | --- | --- | --- | --- |
| _pending_ | _pending_ | _pending_ | _pending_ | _pending_ | _pending_ |
