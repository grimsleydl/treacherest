# Treacherest

Treacherest is the umbrella context for browser-assisted hidden-role variants layered onto Magic: The Gathering Commander. It includes multiple possible rules modes without treating any single mode as the whole product.

## Language

**Treacherest**:
The umbrella app and project for running hidden-role Commander variants.
_Avoid_: Coup when referring to the whole app, Treachery when referring to the whole app

**Rules Mode**:
A named variant of hidden-role Commander with its own roles, information rules, and victory conditions.
_Avoid_: Game type, ruleset, module

**Coup**:
A rules mode in Treacherest where Commander remains a free-for-all, but each player has a secret incentive structure that makes some alliances temporarily more profitable than others.
_Avoid_: Treachery, Treachery-lite

**Treachery**:
The existing inspiration and legacy comparison point for Treacherest. Treachery is not the rules authority for Coup.
_Avoid_: Coup, default rules

**Role Assignment**:
The act of giving each player a hidden role and any private information that follows from that role.
_Avoid_: Role engine, game enforcement

**Role Count Configuration**:
The Pre-Start Game Settings selection of how many copies of each role are included in a rules mode's assignment pool, usually starting from a recommended Role Preset.
_Avoid_: Hard-coded role list, post-assignment role editing

**Unsafe Role Count Override**:
An explicit pre-start opt-in that allows a room to start with a Role Count Configuration that violates normal structural assumptions, such as missing King or Red.
_Avoid_: Default setup, supported preset, silent invalid configuration

**Rules Reference**:
The table-facing explanation of a rules mode's roles, goals, and optional variants.
_Avoid_: Rules engine, automated adjudication

**Pre-Start Game Settings**:
Room setup choices made before a game starts that define role assignment, rules-mode structure, information policies, variants, and unsafe setup overrides. These settings are mutable only while the room is in lobby state.
_Avoid_: Live table state, player action, debug mutation

**Role Image Prompt Pack**:
A named set of visual prompt guidance for generating role-card art across a rules mode's roles. A prompt pack does not imply separate imported image sets unless a future feature explicitly adds that.
_Avoid_: Required game rules, role ability text, image-set selector

**External Image Generation Workflow**:
The out-of-app process of using Role Image Prompt Packs to generate role art, then importing finished image files into Treacherest.
_Avoid_: In-app AI generation, runtime prompt execution

**Canonical Coup Art Direction**:
The default Role Image Prompt Pack for Coup: neutral court-intrigue fantasy role-card art with clear role identity, political tension, and role colors used as accents rather than full-image washes. Frog/scorpion parable imagery is optional and not part of this default.
_Avoid_: Frog/scorpion default, Treachery clone, hard-team faction art, monochrome role art

**Game State Tracking**:
Recording play events after roles are assigned, such as reveals, deaths, Inquisition outcomes, and victory eligibility.
_Avoid_: Role assignment, rules reference

**Debug Mode**:
A non-production operating mode for playtesting and inspecting Treacherest games with privileged aids that bypass or expose normal hidden-role flow.
_Avoid_: Dev mode, admin mode, moderator mode

**Host**:
A non-playing room participant or display surface used to manage or present a room without receiving a hidden role.
_Avoid_: Room creator, room operator, debug operator

**Room Operator**:
The room-authorized person who can manage a room. The room creator is a Room Operator whether they are playing in the game or using a non-playing Host surface. Operator authority is creator-only unless a future co-host feature explicitly defines delegation.
_Avoid_: Host, first player, active player, co-host

**Operator Session**:
A browser session that has Room Operator authority for a specific room. Operator authority is established when the room is created and is not inferred from Host status, player order, or room participation.
_Avoid_: Player cookie, host cookie, viewed player

**Debug Operator**:
A Room Operator using Debug Mode authority for a room.
_Avoid_: Impersonated player, non-host player, public user

**Room Management Control**:
A room-level action such as configuring variants or starting the game. Room Management Controls require Room Operator authority.
_Avoid_: Player action, debug control, first-player control

**Operator Dashboard**:
The Room Operator-facing surface for managing a room, including join instructions, public table state, setup controls, start controls, and Debug Control Surface access when Debug Mode is enabled. The Operator Dashboard is not a hidden-role viewer; outside explicit debug-only controls, it must not reveal other players' private roles or private rules-mode information. A non-playing Host may use the Operator Dashboard, but the dashboard belongs to Room Operator authority rather than Host status.
_Avoid_: Host dashboard when referring to authority, player lobby, hidden-role admin view

**Treacherest Theme**:
The default visual theme for Treacherest: candlelit court surfaces, restrained gold accents, and parchment material for private information. Legacy DaisyUI themes may remain selectable for regression and accessibility checks, but the Treacherest Theme is the product default.
_Avoid_: Dracula as product default, role-color theme, decorative fantasy skin

**Player View**:
The participant-facing room surface for a player. A Player View may include that player's own Privy Panel and player actions, but it must not include Room Management Controls unless those controls are intentionally reached through the Operator Dashboard.
_Avoid_: Operator Dashboard, Host view, debug view

**Privy Panel**:
A player-private UI region for that player's own hidden role, win condition, and private rules-mode information. A Privy Panel may be visually concealed for shoulder-surfing protection, but authorization still comes from the server deciding whose private information is rendered.
_Avoid_: Public card panel, security boundary, role-proof screen

**Confirm-Twice Button**:
An in-app confirmation control for consequential actions. The first activation changes the button into a short-lived confirmation state; the second activation performs the action.
_Avoid_: Browser confirm dialog, accidental single-tap destructive action, modal for every minor confirmation

**Debug Control Surface**:
A Debug Operator-only set of controls for inspecting or mutating a room outside normal player-facing flows.
_Avoid_: Player controls, public overlay

**Debug Impersonation**:
A Debug Mode aid where the Debug Operator uses a selected player's perspective and normal player actions without granting that player operator authority. Its player-facing label should be "View As Player" unless a more explicit action label is needed.
_Avoid_: Player login, ownership transfer, operator transfer

**Viewed Player**:
The player identity currently being rendered and acted as by a Room Operator during Debug Impersonation.
_Avoid_: Current operator, host, authenticated player

**Debug Perspective Override**:
A per-room Operator Session selection that makes player-facing room surfaces render as a Viewed Player until cleared. It only affects rendering and actions while Debug Mode is active. Eliminated players remain valid Viewed Players; removed players do not.
_Avoid_: Global impersonation, role reassignment, player transfer

**Operator View**:
The Debug Mode perspective with no Viewed Player selected. A playing Room Operator sees their own player-facing room surface with debug controls; a non-playing Host sees the host dashboard with debug controls.
_Avoid_: Self, default player, selected player

**Start Override**:
A Debug Mode aid that starts a room outside normal start validation for playtesting incomplete or unusual table states.
_Avoid_: Normal start, production bypass

**Debug Player**:
A stable synthetic active player used in Debug Mode to fill a visible table seat without requiring a real connected player.
_Avoid_: Dummy player, host, spectator, bot

**Debug Insights**:
A Debug Operator-only view of normally hidden or derived room facts used to verify hidden-role setup and state tracking. Debug Insights may also act as the operator's entry point for Debug Impersonation because each visible player entry represents a selectable Viewed Player. Debug Insights may redact hidden-role spoilers by default for screen-sharing safety, but this redaction is not a security boundary.
_Avoid_: Public rules reference, player view, public player list

**Debug Role Color Coding**:
Debug-only visual grouping of player entries by hidden role to speed playtesting and inspection. King is gold, roles with a color in the role name use that color, and Wasteland is gray. It must not imply public table knowledge.
_Avoid_: Public role reveal, player-facing team color, permanent faction marker

**Reveal**:
A public transition where a hidden role becomes known to the table.
_Avoid_: Private view, screen peek

**Elimination**:
A player losing or dying for purposes of Coup victory conditions.
_Avoid_: Kill credit, damage event

**Advisory Win Prompt**:
A non-final app hint that tracked state may satisfy a victory condition.
_Avoid_: Automatic win enforcement

**Confirmed Win**:
A table-approved conclusion that ends the game under the selected rules mode.
_Avoid_: App-enforced win

**Commander Free-for-All**:
The baseline Magic game structure Coup preserves by default: every other player remains an opponent for Magic card text.
_Avoid_: Team game, Two-Headed Giant, shared team

**Coup Ability**:
An explicit permission or restriction created by the Coup rules mode, separate from normal Magic card text.
_Avoid_: Magic teammate rule, card errata

**King**:
The revealed political center of a Coup game, opposed by the anti-King factions.
_Avoid_: Leader

**Blue Knight**:
A hidden Coup role aligned with protecting the King.
_Avoid_: Guardian

**Black Knight**:
A hidden Coup role that wants the King dead while also outliving or eliminating Red.
_Avoid_: Assassin

**Red Knight**:
A hidden Coup role that wants the King dead and Black eliminated.
_Avoid_: Usurper as a canonical role name, Traitor

**Green Knight**:
A hidden Coup role whose victory depends on eligibility to share another faction's outcome.
_Avoid_: Wild card as a canonical role name

**Wasteland Knight**:
An optional Coup role for larger or more chaotic games that wins alone and does not share victory.
_Avoid_: Chaos role as a canonical role name

**Information Policy**:
A rules-mode setting that controls which players receive private knowledge about other roles during role assignment.
_Avoid_: Reveal rule, team assignment

**Full Knowledge**:
An information policy where a player is told the exact player or players matching a relevant role.
_Avoid_: Hard team

**Softened Knowledge**:
An information policy where a player is given a small candidate set containing truth plus decoy information.
_Avoid_: Partial reveal, rumor

**No Knowledge**:
An information policy where a player receives no private role-location information beyond their own role.
_Avoid_: Blind mode

**Conspiracy Knowledge**:
Private information that tells Red which players are Black Knights without making Red and Black a shared-victory team.
_Avoid_: Team assignment, teammate reveal

**Network**:
A variant information policy where members of a role faction know one another.
_Avoid_: Team, party, alliance

**Green Eligibility**:
Whether Green is allowed to share a King-side or Red-side victory under the selected Coup rules.
_Avoid_: Green team membership

**Green Red-Share Lock**:
The Red-side Green Eligibility latch recorded when King Fall happens. Before King Fall it is pending; when King Fall is recorded it is computed from the current pre-King-Fall Blue elimination state and, when Broad Amnesty is enabled, pre-King-Fall Inquisition success. After King Fall it is either eligible or not eligible and is not recomputed later.
_Avoid_: Live Green eligibility, Blue death after King Fall

**Green Hunt Before King Fall**:
The user-facing label for the Green Red-Share Lock. It should read as pending before King Fall, satisfied when the Hunt was locked as satisfied, and not satisfied when King Fall happened before the Hunt was satisfied.
_Avoid_: Green Eligible Before King Fall

**Green Hunt Satisfied Before King Fall**:
The product meaning of the locked Red-side Green Eligibility state. It may temporarily map to an older implementation field, but product copy, tests, and new helpers should use Hunt-satisfaction language.
_Avoid_: Generic Green eligibility

**Green Blue Hunt**:
The active default Green Eligibility model where Green is a Blue-hunter. Green's Hunt is satisfied when at least one Blue Knight dies or is eliminated before King Fall; a harder variant may require all Blue Knights to die or be eliminated before King Fall. Blue reveal, role exposure, Royal Guard reveal, Inquisition reveal, or table-talk proof do not satisfy the Hunt.
_Avoid_: Green wins with whoever is winning, Blue exposure is enough

**Green Hunt Requirement**:
The pre-start Coup variant setting that controls whether Green Blue Hunt requires one Blue Knight to die before King Fall or all Blue Knights to die before King Fall.
_Avoid_: Green difficulty, Blue exposure threshold

**King-Side Inquisition Amnesty**:
The default Green Blue Hunt exception where successful Inquisition lets Green share a King-side victory even if no Blue Knight died before King Fall. By default this does not let Green share a Red-side victory.
_Avoid_: Inquisition makes Red legitimate by default

**Inquisition**:
A Coup ability where a Blue Knight names a suspected Red Knight and may reveal Red if correct. The revealed King is not a valid Inquisition target. The caller form is private capability UI and belongs inside the Blue Knight's Privy Panel; public notices and public results belong in public table notices.
_Avoid_: Investigation, vote, naming the King

**Inquisition Notice**:
A public acknowledgement that an Inquisition has been called before its result is shown.
_Avoid_: Secret Inquisition

**Public Inquisition Result**:
The default Inquisition result policy where a correct Inquisition reveals Red to the table.
_Avoid_: Private reveal

**Private Inquisition Result**:
A variant Inquisition result policy where the result is shown only to the Blue Knight who called it.
_Avoid_: Default Inquisition

**Royal Guard**:
A Coup ability where a revealed Blue Knight can directly block for the King under narrow combat permissions.
_Avoid_: Teammate blocking, shared combat

**Strict Green Eligibility**:
The earlier Green Eligibility model where Green's default Red-side sharing required all Blue Knights to be dead before King Fall. This is superseded by Green Blue Hunt as the default model.
_Avoid_: Active default Green rule

**Green Win Condition Display**:
The canonical Green Knight role-card win condition explanation for the active Green Blue Hunt model. It should explicitly tell Green: "You serve neither crown"; "You are hunting Blue Knights"; "Your Hunt is satisfied when at least one Blue Knight dies before King Fall"; "Blue dying with the King does not count"; "If Inquisition succeeds, you may share a King-side victory even without a Blue death"; and "You do not share Black or Wasteland victories."
_Avoid_: Vague selected Green rules copy

**Public Green Hunt Summary**:
The public/revealed Green Knight rules summary. It should explain the same Hunt rules as the private Green role card, but without second-person private-role flavor.
_Avoid_: Different public Green rules

**Broad Amnesty**:
A Green Eligibility variant where successful Inquisition before the King falls can qualify Green to share either King-side or Red-side victory. It is not the default because exposing Red does not normally make Red's coup legitimate to Green.
_Avoid_: Default Green rule

**Inquisition Amnesty**:
The pre-start Coup variant setting that controls whether successful Inquisition gives Green King-side amnesty only or Broad Amnesty for King-side and Red-side sharing.
_Avoid_: Inquisition result policy

**King Fall**:
The event where the King loses or dies. Blue losses caused by King Fall do not make Green newly eligible for Red-side victory.
_Avoid_: Blue death as Green eligibility
