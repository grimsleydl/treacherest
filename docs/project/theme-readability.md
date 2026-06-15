# Theme Readability And Contrast

Treacherest uses DaisyUI v5 themes plus two project themes: `treacherest` and `treacherest-day`. Theme support is token-driven. Components should remain readable across selectable themes by using DaisyUI's paired semantic tokens instead of hard-coded color assumptions.

## Current Theme Set

The app exposes all 35 DaisyUI v5 built-in themes plus the two Treacherest themes:

- `treacherest`
- `treacherest-day`
- `light`
- `dark`
- `cupcake`
- `bumblebee`
- `emerald`
- `corporate`
- `synthwave`
- `retro`
- `cyberpunk`
- `valentine`
- `halloween`
- `garden`
- `forest`
- `aqua`
- `lofi`
- `pastel`
- `fantasy`
- `wireframe`
- `black`
- `luxury`
- `dracula`
- `cmyk`
- `autumn`
- `business`
- `acid`
- `lemonade`
- `night`
- `coffee`
- `winter`
- `dim`
- `nord`
- `sunset`
- `caramellatte`
- `abyss`
- `silk`

## Readability Rules

- Prefer paired DaisyUI tokens: `bg-base-100 text-base-content`, `bg-primary text-primary-content`, `bg-warning text-warning-content`, etc.
- Avoid cross-token pairs such as `text-primary` on `bg-base-100` for body text unless the component has been tested across the theme matrix.
- Dense text, small badges, form labels, and table rows should use `base-content` on a `base-*` surface unless the semantic color is essential.
- Semantic colors are best for short status labels, icons, borders, and buttons that also use the paired `*-content` text token.
- NoticeCard content uses a neutral readable surface with semantic border/accent treatment. Body copy inside a NoticeCard should normally inherit the NoticeCard text color instead of using `text-success`, `text-error`, `text-warning`, or muted base-content utilities on top of semantic alert backgrounds.
- Do not communicate game state with color alone. Keep text labels such as `revealed`, `eliminated`, `operator`, `debug`, and role names.
- Privy Panel, Operator Dashboard, Player View, and Debug Control Surface components should avoid custom color math unless it is backed by an audit case.

## Audit Command

Run the static token contrast audit after changing the theme list or broad component color rules:

```sh
cd nix/app
npm run build:css
npm run audit:themes
```

The audit reads `static/css/output.css` and reports contrast ratios for DaisyUI semantic foreground/background token pairs. It is report-only by default because some built-in themes may miss the WCAG AA normal-text target for semantic component pairs. Use the report to decide whether to:

- adjust a Treacherest custom token,
- change a component to use a safer paired token,
- hide or demote a built-in theme,
- add a local component-specific override, or
- accept a low-contrast pair only for large text/icons where appropriate.

For strict experiments:

```sh
cd nix/app
npm run audit:themes -- --fail-on-aa
```

## Theme Readability Lab

Use a separate static HTML design QA artifact for component-level theme review. The artifact should not be part of the live app or production route tree. The generator should be tracked, but the generated review file should live under `.scratch/theme-readability/` and remain untracked. It should render representative app UI examples with real app classes and compiled CSS, expose a theme switcher, and label every sample with a stable copyable identifier so visual feedback can name the exact failing case.

Generate it with:

```sh
cd nix/app
npm run build:css
npm run lab:themes
```

Then open `.scratch/theme-readability/index.html` in a browser.

The lab complements the token audit. Token checks answer whether a theme's semantic color pairs are mathematically plausible. The lab answers whether actual Treacherest UI combinations are readable once component nesting, opacity, font size, disclosure states, and surface layering are involved.

Label review targets at the text/background pair level rather than only at the component level. For example, Royal Guard should expose separate targets for title, body text, warning text, and button text because those regions may use different tokens inside one visual card.

Phase the lab. Phase 1 should be a curated review sheet for redesigned, high-risk surfaces: Player View, Player Lobby, Operator Dashboard, Debug Control Surface, Privy Panel, NoticeCards, Confirm-Twice Buttons, StateChips, PlayerRows, role cards, and role-count controls. Keep an automated raw inventory of discovered text/background class usage as a backstop, but do not make the human reviewer sign off every incidental template line in the first pass. Older ability, metamorph, transformation, and modal surfaces can be added in a later sweep.

Each lab target should show both visual and numeric evidence: a copyable target ID, the rendered sample, the foreground/background token guess, the browser-computed foreground/background colors, the computed contrast ratio for the current theme, and a status label such as `AA`, `Large only`, `Fail`, or `Manual`. Numeric contrast is evidence, not final signoff; the human reviewer still decides whether font size, opacity, spacing, and visual hierarchy are acceptable.

Prefer browser-computed contrast in the lab. The static token audit remains useful for broad theme health, but the lab should use `getComputedStyle()` on each target plus its effective background so nested classes, opacity, inherited colors, and real rendered font size are reflected in the displayed result.

For effective background detection, walk up from the target to the nearest non-transparent background by default. Allow explicit background scoping for tricky cases, such as `data-contrast-bg`, when the meaningful surface is a parent card, alert, overlay, or translucent panel that automatic walking would misread.

Phase 1 lab samples may be hand-authored representative markup using the same classes as the app. Do not require full Templ rendering with mock rooms and players for the first pass. Real component rendering can be added later for stateful surfaces whose readability depends on generated data or conditional structure.

The lab should include every selectable theme in its theme switcher, defaulting to `treacherest`. It should also provide filters so the reviewer can narrow the sheet to failures, manual-review targets, core surfaces, debug surfaces, role cards, notices, and other useful groups.

The lab should support click-to-toggle review accumulation. Each labeled target should let the reviewer mark `OK`, `Bad`, or clear the mark for the currently selected theme. Marks are scoped to the selected theme because the same target may be readable in one theme and unreadable in another. A sticky review-notes panel should accumulate copyable text grouped by status and theme so the reviewer can paste the results back into chat or into a follow-up file. This state may be in-memory for the first pass; optional artifact-local `localStorage` persistence can be added later if multi-session review becomes useful.

## Accommodation Strategy

Base page surfaces should be held to a stricter standard than decorative or short semantic accents. If `base-content` fails on `base-100`, `base-200`, or `base-300`, the theme is not suitable as a general app theme without token overrides. If a semantic pair fails, prefer changing the component usage before changing the whole theme.

Manual review is still required for:

- font readability at real sizes,
- focus rings,
- hover and disabled states,
- translucent overlays,
- screenshots of Privy Panel concealed/open states,
- theme behavior in debug mode,
- mobile density and overflow.

The release sweep in `docs/project/redesign-qa-checklist.md` should include at least `treacherest`, `treacherest-day`, `dracula`, and any newly concerning theme from the audit report.
