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

**Rules Reference**:
The table-facing explanation of a rules mode's roles, goals, and optional variants.
_Avoid_: Rules engine, automated adjudication

**Game State Tracking**:
Recording play events after roles are assigned, such as reveals, deaths, Inquisition outcomes, and victory eligibility.
_Avoid_: Role assignment, rules reference

**Debug Mode**:
A non-production operating mode for playtesting and inspecting Treacherest games with privileged aids that bypass or expose normal hidden-role flow.
_Avoid_: Dev mode, admin mode, moderator mode

**Debug Control Surface**:
A host-only set of Debug Mode controls for inspecting or mutating a room outside normal player-facing flows.
_Avoid_: Player controls, public overlay

**Debug Impersonation**:
A Debug Mode aid where the host views the game from a selected player's perspective without making that player a real host.
_Avoid_: Player login, ownership transfer, acting as player

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

**Inquisition**:
A Coup ability where a Blue Knight names a suspected Red Knight and may reveal Red if correct.
_Avoid_: Investigation, vote

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
The default Green Eligibility rule where Inquisition can help Green share King-side victory, but Red-side sharing requires Blue to have died before the King fell.
_Avoid_: Default Green team

**Broad Amnesty**:
A Green Eligibility variant where successful Inquisition before the King falls can qualify Green to share either King-side or Red-side victory.
_Avoid_: Default Green rule

**King Fall**:
The event where the King loses or dies. Blue losses caused by King Fall do not make Green newly eligible for Red-side victory.
_Avoid_: Blue death as Green eligibility
