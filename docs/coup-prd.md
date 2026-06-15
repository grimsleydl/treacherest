# Coup Rules Mode PRD

## Problem Statement

Treacherest currently supports a Treachery-inspired hidden-role Commander experience. Coup is a distinct rules mode for Magic: The Gathering Commander groups that want hidden incentives, temporary alliances, and political suspicion without turning the table into hard teams or rewriting normal Commander card text.

Coup needs a browser-assisted setup and game aid that assigns secret roles, distributes private information, explains role goals, supports public/manual state changes, and helps the table notice possible wins without fully adjudicating every Magic edge case.

## Target Users

- Commander playgroups of roughly five or more players.
- In-person tables where players can pass phones or use their own devices.
- Groups that want a political hidden-role variant that still feels like Commander free-for-all.
- Hosts who want configurable role presets and rules variants without managing role cards manually.

## Product Positioning

Treacherest is the umbrella app. Coup is one rules mode in Treacherest. Treachery remains historical inspiration and may remain a separate or legacy rules mode, but Treachery rules are not the authority for Coup.

For role-by-role design rationale, see `docs/coup-role-design.md`.

## Goals

- Preserve normal Commander free-for-all semantics by default.
- Assign Coup roles secretly and show each player their role privately.
- Reveal the King at setup.
- Give role-specific private information based on selected information policies.
- Provide readable role goals and rules reference text.
- Support manual Reveal and Elimination tracking in the first milestone.
- Handle Inquisition in-app, including notice confirmation and result policy.
- Show Advisory Win Prompts when tracked state appears to satisfy a win condition, but require table confirmation before ending the game.
- Keep rules variants explicit and configurable.

## Non-Goals

- Do not make Coup “Treachery but with fewer abilities.”
- Do not make King and Blue actual Magic teammates by default.
- Do not globally rewrite Magic card text.
- Do not rely on kill credit for default win logic.
- Do not fully automate Commander rules adjudication.
- Do not prioritize role-card art or prompt tooling before the core rules-mode UX.
- Do not implement a large abstract role engine before the Coup rules model stabilizes.

## Design North Star

Commander free-for-all, except every player has a secret incentive structure that makes some alliances temporarily more profitable than others.

## Existing Codebase Fit

The current app already has useful building blocks:

- Room creation and join flow.
- Lobby, countdown, playing, and ended states.
- Player records with hidden role, public reveal, face-up state, and elimination state.
- Role presets and configurable role counts.
- Pre-start role-count tuning similar to the existing Treachery mode.
- Host mode.
- Server-rendered Templ UI with Datastar/SSE updates.

The current user-facing role taxonomy is Treachery-oriented: Leader, Guardian, Assassin, Traitor. Coup should introduce its own user-facing role taxonomy instead of presenting those existing names as aliases.

## Primary User Flows

### Create Coup Game

1. User creates a Treacherest room.
2. User selects Coup as the rules mode.
3. User selects player count.
4. User adds or confirms player names.
5. User selects a Coup role preset.
6. User optionally tunes the number of each Coup role before assignment.
7. User selects rules variants.
8. App validates that the role count configuration and settings match the player count.

### Assign Roles Secretly

1. Game starts from the lobby.
2. App assigns roles according to the selected Coup preset.
3. King starts revealed to all players.
4. Non-King roles start hidden.
5. Each player can privately view their own role, goals, and relevant private information.
6. Players are instructed not to prove roles by showing their device unless a Coup rule reveals that role.

### Show Special Information

1. King receives Blue Knight information according to the King-to-Blue information policy.
2. Red receives Black Knight information according to the Red-to-Black information policy.
3. Black receives Red information only if the selected variant enables it.
4. Black Knights receive other Black identities only if the selected Network variant enables it.

### Use Inquisition

1. A revealed or unrevealed Blue Knight chooses to call Inquisition in-app.
2. Blue names the suspected Red Knight. King is not a valid target.
3. Blue is revealed when calling Inquisition.
4. App broadcasts an Inquisition Notice popup.
5. Any one living non-Blue player confirms Blue announced the Inquisition.
6. App resolves the result:
   - If correct, Inquisition succeeds and Red is revealed according to the selected result policy.
   - If wrong, the named player's role remains hidden and Blue loses half their current life total, rounded up.
7. App records Inquisition success for eligibility and Advisory Win Prompt logic.

### Manual Reveal And Elimination

1. A player can reveal their role when a Coup rule or table decision says it is public.
2. A player can be marked eliminated when they lose or die.
3. Eliminated players reveal their role.
4. Concessions count as Elimination for victory purposes unless the table explicitly defines a separate kill-credit rule.

### Advisory Win Prompt

1. App observes tracked state: role reveals, eliminations, Inquisition success, and selected settings.
2. If a win condition appears satisfied, app shows an Advisory Win Prompt.
3. Table confirms or rejects the prompt.
4. Confirmed Win ends the game.

## Core Roles

### King

- Starts revealed.
- Political center of the table.
- Wins if alive when Black, Red, and Wasteland threats are eliminated.
- Receives Blue Knight information according to selected information policy.

### Blue Knight

- Protects King.
- Wins with King.
- Loses when King loses.
- Can use Royal Guard after revealing.
- Can call Inquisition.

### Black Knight

- Assassin / hired killer.
- Wins if King is dead, at least one Black Knight is alive, and Red is dead or eliminated.
- Does not know Red by default.
- Does not know other Black Knights by default.

### Red Knight

- Usurper.
- Wins if King is dead, Red is alive, and all Black Knights are dead.
- Knows all Black Knights by default.
- Thematically hired Black through cutouts and does not realize Black must eventually betray Red.

### Green Knight

- Opportunist.
- May share King-side or Red-side victory only if eligible.
- Uses Strict Green Eligibility by default.
- Must not gain Red-side eligibility merely because Blue loses when King falls.

### Wasteland Knight

- Optional large-table or chaos role.
- Wants everyone else eliminated.
- Wins alone and never shares victory.
- Recommended for 8+ players or chaos variants.

## Default Rules

- Normal Commander free-for-all rules remain in force.
- Every other player remains an opponent for Magic card text.
- Coup roles affect victory conditions, private information, and explicit Coup abilities only.
- Players may claim any role at any time, truthfully or falsely.
- Players should not prove hidden roles by showing the app screen unless a Coup rule reveals them.
- King starts revealed.
- Other roles start hidden unless a selected variant says otherwise.
- Eliminated players reveal their role.
- Concessions count as Elimination for victory purposes.
- Default rules do not use kill credit.

## Role Presets

Treat these as configurable presets, not permanent hard-coded rules. The room creator should be able to start from a recommended preset and tune the number of each Coup role before the game starts, similar to Treachery mode. The final role count configuration must still match the number of participating players.

Default structural validation:

- A normal Coup role count configuration requires exactly one King.
- A normal Coup role count configuration requires exactly one Red Knight.
- Blue, Black, Green, and Wasteland counts may be tuned within player-count validation.
- The room creator may explicitly enable an Unsafe Role Count Override to start without a King or Red Knight, with the understanding that core Coup win logic, Inquisition, and table balance may be broken.

| Players | Default roles |
| --- | --- |
| 5 | King, Blue, Black, Red, Green |
| 6 | King, Blue, Black, Black, Red, Green |
| 7 | King, Blue, Blue, Black, Black, Red, Green |
| 8 | King, Blue, Blue, Black, Black, Black, Red, Green |
| 8 chaos | King, Blue, Blue, Black, Black, Red, Green, Wasteland |
| 9 | King, Blue, Blue, Black, Black, Black, Red, Green, Wasteland |

## Information Policies

### King-to-Blue

Default: Full Knowledge.

- King knows all Blue Knights.

Variants:

- Softened Knowledge: King receives a small candidate set containing true Blue and decoy information.
- No Knowledge: King receives no Blue identity information.

### Red-to-Black

Default: Red knows all Black Knights.

Variants:

- Red knows one Black Knight.
- Red knows no Black Knights.

### Black-to-Red

Default: Black Knights do not know Red.

Variants:

- One Black knows Red.
- All Black Knights know Red.
- Full conspiracy mode where Red and Black know each other.

### Black Network

Default: Black Knights do not know each other.

Variant:

- Black Knights know each other.

## Inquisition

Default:

- Each Blue Knight may call Inquisition once per game.
- Blue reveals when calling Inquisition.
- Blue names the suspected Red Knight.
- King is excluded from possible Inquisition targets.
- One living non-Blue player must confirm the Inquisition Notice before result display.
- If Blue is correct:
  - Inquisition succeeds.
  - Blue loses no life.
  - Red is revealed to all players under Public Inquisition Result.
- If Blue is wrong:
  - The named player's role remains hidden.
  - Blue loses half their current life total, rounded up.

Variant:

- Private Inquisition Result: if Blue is correct, only the inquisitor learns the result. Red remains hidden to the table, but the app records Inquisition success.

Open configuration candidates:

- Once per Blue per game, once per game total, configurable attempts, or disabled.
- Public result default versus private result variant.

## Royal Guard

Default wording:

> Once each combat, a revealed Blue Knight may have any number of untapped creatures they control block creatures attacking the King as though those creatures were attacking the Blue Knight. Normal blocking restrictions apply.

Default settings:

- Blue must reveal to use Royal Guard.
- Royal Guard protects only the King player.
- Royal Guard does not protect planeswalkers, battles, or other permanents King controls.
- Royal Guard does not make Blue and King actual Magic teammates.

Configuration candidates:

- Any number of blockers by default.
- One blocker variant.
- Numeric blocker limit variant.
- Expanded protection for King-controlled planeswalkers/battles/permanents.
- High-complexity Treachery-like teammate variant.

## Win Conditions

### King-Side

King wins if:

- King is alive, and
- all Black Knights, Red Knights, and Wasteland Knights are eliminated.

Blue wins with King.

Green may win with King if:

- Green is alive when King wins, and
- either no Blue Knight is alive or Inquisition has succeeded.

### Black

Black wins if:

- King is dead or eliminated, and
- at least one Black Knight is alive, and
- Red is dead or eliminated.

### Red

Red wins if:

- King is dead or eliminated, and
- Red is alive, and
- all Black Knights are dead or eliminated.

Green may share Red victory only under Green Eligibility rules.

### Green

Default Strict Green Eligibility:

- Green can share King-side victory if alive and either no Blue Knight is alive or Inquisition has succeeded.
- Green can share Red-side victory only if all Blue Knights were already dead before King Fall.
- Blue Knights who lose because of King Fall do not count as having been dead for purposes of making Green eligible to share Red victory.

Broad Amnesty variant:

- If Inquisition succeeded before King Fall, Green may share either King-side or Red-side victory.

Simple Green variant:

- Green wins with King if alive and loses if Black or Wasteland wins.

### Wasteland

Wasteland wins only if they are the sole surviving player. Wasteland never shares victory.

## Functional Requirements

### Rules Mode Selection

- The app must support Coup as a rules mode under Treacherest.
- Existing Treachery-oriented docs and code should not be treated as Coup rules authority.

### Configuration

- Player count must determine available/recommended role presets.
- Role presets must be visible before game start.
- The room creator must be able to tune the number of each Coup role before the game starts.
- The app must validate that the final role count configuration equals the number of participating players.
- By default, the app must require exactly one King and exactly one Red Knight before starting Coup.
- The room creator must be able to explicitly acknowledge and enable an Unsafe Role Count Override to bypass the normal King/Red structural requirement.
- Rules variants must be explicit and reviewable before game start.
- Defaults must be selected for teachability.

### Assignment And Privacy

- Role assignment must exclude host-only/non-participating users.
- Each player must see only their own hidden role and allowed private information.
- King must be revealed publicly after assignment.
- Red and King special information must be shown only to the correct recipients.
- Private Inquisition Result must not leak Red to other players.

### Reveal And Elimination Tracking

- The app must support manual Reveal.
- The app must support manual Elimination.
- Eliminated players reveal their role.
- The app should distinguish private role viewing from public Reveal.

### Inquisition Tracking

- The app must require a witness confirmation before result display.
- The app must exclude King from possible Inquisition targets and reject attempts to name King.
- The app must support Public Inquisition Result and Private Inquisition Result.
- The app must track whether Inquisition succeeded.
- The app must record failed Inquisition life-loss guidance.

### Advisory Win Prompts

- The app may show potential winner prompts based on tracked state.
- The app must require human/table confirmation before ending the game.
- The app should explain which tracked facts caused the prompt.

## Suggested Data Structures

These names are candidates and should be adapted to local code patterns:

- `RulesMode`
- `RoleDefinition`
- `RoleAssignment`
- `Player`
- `Game`
- `RulePreset`
- `InformationPolicy`
- `InquisitionState`
- `GreenEligibilityState`
- `RevealState`
- `EliminationState`
- `WinCheckResult`
- `AdvisoryWinPrompt`
- `ConfirmedWin`

Potential state values:

- `kingAlive`
- `kingFallen`
- `unsafeRoleCountOverrideEnabled`
- `inquisitionSucceeded`
- `greenEligibleBeforeKingFall`
- `broadAmnestyEnabled`
- `blueRevealRequiredForRoyalGuard`
- `blueRevealRequiredForInquisition`
- `redKnowsOneBlack`
- `redKnowsAllBlack`
- `kingKnowsBlue`
- `kingGetsBlueCandidates`
- `blackKnowsRed`
- `blackNetworkEnabled`
- `inquisitionResultPolicy`
- `royalGuardBlockerLimit`

## Possible UI Screens And Components

- Rules mode selector.
- Coup setup panel.
- Player count and preset selector.
- Variant settings panel.
- Secret role view.
- King public reveal banner.
- Private information panel.
- Rules reference tab.
- Manual Reveal control.
- Manual Elimination control.
- Inquisition action modal.
- Inquisition Notice confirmation popup.
- Inquisition result display.
- Advisory Win Prompt banner/modal.
- Confirmed Win screen.

## Edge Cases

- Unsafe Role Count Override with no King.
- Unsafe Role Count Override with no Red Knight.
- Multiple Blue Knights and Inquisition attempts.
- Multiple Black Knights with no Network.
- Red eliminated before King.
- King Fall while Blue is still alive.
- Green alive at Red victory after Blue loses due to King Fall.
- Private Inquisition Result with advisory win prompts.
- Player concession versus elimination.
- Host/spectator clients seeing hidden information.
- Browser back/refresh during private role view.
- Multiple tabs for the same player session.

## Testing Strategy

- Unit tests for Coup preset distributions.
- Unit tests for information policy outputs.
- Unit tests for Green Eligibility rules.
- Unit tests for Inquisition result policy.
- Unit tests for Royal Guard setting serialization.
- Integration tests for create/join/start/assign role flow.
- Integration tests for privacy boundaries in rendered views.
- Browser tests for private role view, Inquisition Notice, and manual Reveal/Elimination flows.
- Multi-client browser tests for SSE updates that must not leak hidden information.

## Implementation Phases

### Phase 1: Coup Setup And Manual State Aid

- Add Coup as a rules mode.
- Add Coup role definitions and presets.
- Add variant configuration for default information policies.
- Assign roles secretly.
- Reveal King publicly.
- Show private role goals and special information.
- Add rules reference text.
- Support manual Reveal and Elimination.
- Implement in-app Inquisition flow with witness confirmation.
- Show Advisory Win Prompts with human confirmation.

### Phase 2: Stronger State Tracking

- Track Green Eligibility before King Fall.
- Track King Fall explicitly.
- Improve advisory win checks.
- Add richer audit/history of reveals, eliminations, and Inquisitions.
- Support more variants after table testing.

### Phase 3: Content And Art Tooling

- Role-card art and prompt generation using `docs/coup-role-image-prompts.md`.
- Keep AI image generation outside the app; Treacherest consumes completed imported images.
- Prompt scope is final role-card art only; thumbnails and compact UI art derive from imported final artwork.
- Art should use portrait card-art composition without baked-in frames, text, logos, mana symbols, or UI.
- Role names, rules text, victory text, reveal state, and card layout remain app/template-rendered.
- Canonical default visual theme: neutral Coup court-intrigue fantasy.
- Role colors should appear as accents, not full-image color washes.
- Wasteland art should feel gray, ruined, and exiled while still belonging to the same Coup court-intrigue set.
- Art should avoid visual language that makes Blue/King or Red/Black look like obvious permanent teams.
- Art should use one primary identifiable role figure, with only anonymous silhouettes or courtiers for atmosphere.
- Optional visual style packs are prompt guidance only for now, including frog/scorpion parable, Treachery-like fantasy, classic court intrigue, sci-fi resistance, investigative satire, and White Rabbit conspiracy sci-fi.
- The frog/scorpion parable style pack must remain optional and must not replace the neutral court-intrigue default.
- The import pipeline should keep one image per Coup role until a later UI need justifies multiple image sets.
- Printable/exportable role cards.

## Acceptance Criteria

- A host can create a Coup game for 5-9 players using default presets.
- Each player can privately see their assigned role and goals.
- King is public after assignment.
- King sees Blue information under Full Knowledge default.
- Red sees all Black Knights under the default policy.
- Black does not see Red or other Blacks by default.
- Blue can call Inquisition through the app.
- One living non-Blue player can confirm the Inquisition Notice.
- Public Inquisition Result reveals Red to all when correct.
- Private Inquisition Result informs only the inquisitor and does not publicly reveal Red.
- Failed Inquisition displays half-current-life-rounded-up life loss.
- Manual Reveal and Elimination are supported.
- Advisory Win Prompts require table confirmation.
- Hidden role information is not rendered to the wrong player views.

## Open Questions

- Should Inquisition attempts be configurable in Phase 1 or fixed as once per Blue per game until playtested?
- Should Blue be allowed to call Inquisition before revealing manually, with the action revealing them automatically?
- Should the app record life totals or only instruct players how much life to lose after failed Inquisition?
- What exact UI permission model should control manual Reveal and Elimination?
- Should any player be able to mark another player eliminated, or only the affected player/host?
- Should Advisory Win Prompts be shown to all players or host-only first?
- How should a rejected Advisory Win Prompt be recorded?
- Should Softened Knowledge candidate sets include only living players at setup, or can they include future decoys such as host/spectator labels? Recommended: only active players.
- Should Wasteland be available at 8 players by default as a selectable chaos preset or hidden behind advanced settings?
- Should Coup eventually replace the existing Treachery-first home flow, or live beside it as a mode selector?
