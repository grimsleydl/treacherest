# Coup Role Design Rationale

Status: draft.

This document explains the Coup roles, the design thinking behind them, and the main tradeoffs they are meant to hold in tension. It complements `docs/coup-prd.md` and `CONTEXT.md`; it is not a separate rules authority.

## Design North Star

Coup is a Treacherest Rules Mode for Magic: The Gathering Commander:

> Commander free-for-all, except every player has a secret incentive structure that makes some alliances temporarily more profitable than others.

The role design should preserve the feel of a normal Commander table. Players still negotiate, bluff, make temporary deals, punish threats, and respond to board state. Coup adds hidden incentives and private information, but it should not collapse the table into two obvious teams or ask players to mentally rewrite normal Magic card text.

## What The Roles Are For

Coup uses roles to create political pressure, not to hand out a large package of Magic rules changes. A good Coup role should do at least one of these things:

- Give a player a reason to bargain differently than their public board state suggests.
- Make a temporary alliance plausible without making it permanently trustworthy.
- Create suspicion around who benefits from a given death or reveal.
- Give the table reasons to delay, accelerate, or redirect aggression.
- Keep the hidden-role layer legible enough that players can make social reads.

A Coup role should avoid these patterns:

- Becoming a full Magic teammate by default.
- Needing kill credit or complicated combat attribution.
- Depending on hidden information that cannot be explained to a player privately.
- Producing a hostage pattern where one role can force another player into an obviously bad reveal.
- Making the Commander game feel secondary to role-card powers.

## Core Shape

Coup has one public center, one loyal protector role, two anti-King roles with conflicting endgames, one opportunist, and an optional solo chaos role.

| Role | Table function | Main design job |
| --- | --- | --- |
| King | Revealed political center | Creates a visible axis of conflict |
| Blue Knight | Hidden protector | Makes loyalty valuable but not fully provable |
| Black Knight | Assassin | Wants King dead but also wants Red gone |
| Red Knight | Usurper | Wants King dead but must eliminate Black |
| Green Knight | Blue-hunter / opportunist | Hunts Blue Knights and can share a win only under bounded Hunt eligibility |
| Wasteland Knight | Solo chaos role | Adds larger-table pressure without sharing victory |

The Red/Black split is central. Coup should not be "King team versus anti-King team." Black and Red both need the King dead, but they do not share the same endgame. That creates a temporary conspiracy with an expiration date.

## King

### Role Concept

The King starts revealed and acts as the political center of the table. The King wins if alive when Black, Red, and Wasteland threats are eliminated.

### Why The Role Exists

The revealed King gives the game a visible reference point. Without a revealed center, the hidden-role layer can become too diffuse: everyone knows there are secret goals, but no one has a stable axis around which early politics can form.

The King also gives non-King roles something concrete to react to:

- Blue has someone to protect.
- Black and Red have a shared short-term objective.
- Green has a center to bargain around.
- Wasteland has a high-value public obstacle.

### Design Pressure

The King should be powerful politically, not mechanically. The role starts revealed, so the King can openly negotiate and accuse. That public position is already a meaningful advantage and liability. The app should avoid compensating with broad extra powers by default.

### Failure Modes To Avoid

- Making the King a hard team captain.
- Letting the King prove every ally too easily in default settings.
- Making the King's survival the only thing anyone cares about.

## Blue Knight

### Role Concept

Blue Knight protects the King, wins with the King, and loses when the King loses. Blue can use explicit Coup abilities such as Royal Guard and Inquisition.

### Why The Role Exists

Blue gives the King's side hidden strength without making the whole table structure obvious. A hidden protector lets players ask:

- Who is defending the King because they are Blue?
- Who is defending the King because it is politically useful?
- Who is pretending to be Blue to manipulate the King?
- When should Blue reveal, and what is that reveal worth?

This preserves Commander-style table talk. Blue is a source of suspicion and bargaining, not just a visible teammate.

### Royal Guard Rationale

Royal Guard is intentionally an explicit Coup ability instead of making Blue and King Magic teammates by default. The default model is:

- Every other player remains an opponent for Magic card text.
- Blue receives a narrow permission to block for the King.
- Normal blocking restrictions still apply.

This avoids broad card-text complexity while still giving Blue a tangible protector identity.

### Inquisition Rationale

Inquisition gives Blue a way to turn suspicion into public or private information. It also gives Green a healthier path to cooperate with King-side outcomes without requiring Blue's death.

The cost for guessing wrong is half the Blue player's current life total, rounded up. That is severe enough to matter but should not usually kill the player outright.

### Failure Modes To Avoid

- Making Blue and King actual teammates by default.
- Letting Blue prove their role cheaply without a table-facing cost.
- Making Inquisition a secret action that the table cannot police.
- Making Blue's protection so strong that the Commander game stalls.

## Black Knight

### Role Concept

Black Knight is the assassin. Black wins if the King is dead, at least one Black Knight survives, and Red is dead or gone.

### Why The Role Exists

Black creates the direct assassination pressure that makes the King vulnerable. But Black is not simply "Red's teammate." The flavor is that Red hired Black to kill the King, while Black intends to eliminate Red and leave no witnesses.

This design gives Black two phases:

1. Work toward the King's death.
2. Survive the aftermath and make sure Red cannot claim the throne.

That second phase is what keeps Coup from becoming a clean anti-King team game.

### Information Rationale

By default, Black does not know Red and Black Knights do not know each other. This keeps the assassin network dangerous but uncertain. It also helps Red's information advantage matter.

Variants can let one Black know Red, all Blacks know Red, or Black Knights know each other, but those variants move the mode closer to explicit-team play and should be treated as higher-information options.

### Failure Modes To Avoid

- Making Black a normal teammate of Red.
- Requiring kill credit to determine whether Black succeeds.
- Letting multiple Black Knights operate with perfect coordination by default.

## Red Knight

### Role Concept

Red Knight is the usurper. Red wins if the King is dead, Red survives, and all Black Knights are dead.

### Why The Role Exists

Red creates a different kind of anti-King pressure from Black. Red wants the King dead, but Red also needs to survive the assassins they empowered. Red is politically ambitious, not merely murderous.

The role is designed to create unstable cooperation:

- Red benefits when Black pressures the King.
- Red eventually needs Black gone.
- Black eventually needs Red gone.
- Other players can exploit that split.

### Information Rationale

Red knows all Black Knights by default. This is a deliberate asymmetry. Red needs enough information to play the usurper role actively, and the table needs Red to be meaningfully dangerous.

That does not make Red and Black a team. Red knowing Black creates leverage, paranoia, and a temporary hiring relationship. It does not create shared victory.

### Failure Modes To Avoid

- Hiding so much from Red that Red cannot play their role.
- Letting Red and Black share victory by accident.
- Presenting Red/Black knowledge as teammate knowledge.

## Green Knight

### Role Concept

Green Knight is the Blue-hunter / conditional opportunist. Green serves neither crown by default. Green's Hunt is satisfied when at least one Blue Knight dies or is eliminated before the King falls; a harder variant can require all Blue Knights to die or be eliminated before the King falls.

Current default display should tell Green:

- You serve neither crown.
- You are hunting Blue Knights.
- Your Hunt is satisfied when at least one Blue Knight dies before the King falls.
- Blue dying with the King does not count.
- If Inquisition succeeds, you may share a King victory even without a Blue death.
- You may share a Red victory only if your Hunt was satisfied before the King fell.
- You do not share Black or Wasteland victories.

### Why The Role Exists

Green gives the table a flexible bargaining role that is not simply loyalist, assassin, or usurper. Green should be able to make temporary deals with more than one side, but only inside boundaries created by the Hunt.

The goal is not "Green can always win with whoever is winning." The goal is "Green wants the King's hidden protectors humbled before accepting either the King's continued rule or Red's usurpation."

### The Hostage-Pattern Problem

The main Green design trap is a hostage pattern:

1. Green demands that the King reveal Blue.
2. If the King refuses, Green helps kill the King.
3. Blue loses because the King dies.
4. Green then claims Red-side eligibility because Blue is now gone.

That pattern is bad because Green can convert the King's refusal into a forced Blue loss and then benefit from the very Blue death caused by King Fall.

Green Blue Hunt prevents this by locking Green Hunt Before King Fall. Blue dying because the King loses does not make Green newly eligible for Red's victory.

### Inquisition Relationship

Successful Inquisition gives Green a non-toxic path to cooperate with a King victory. It creates King-Side Inquisition Amnesty without requiring Blue to die. That is useful because it lets Blue and Green sometimes align after Blue takes a meaningful public risk.

Broad Amnesty, where successful Inquisition before the King falls can also let Green share a Red victory, is an optional settings variant. It is not the active default, and the Green role card should explain the active Red-victory condition directly rather than using the Broad Amnesty label.

### Failure Modes To Avoid

- Letting Green always join the winning side.
- Letting Green profit from Blue dying only because the King fell.
- Letting Blue reveal or exposure satisfy the Hunt.
- Making Green's in-game win condition vague.
- Turning Green into a hidden hard teammate of either King or Red.

## Wasteland Knight

### Role Concept

Wasteland Knight is an optional large-table or chaos role. Wasteland wins alone when every other player is eliminated and never shares victory.

### Why The Role Exists

Wasteland adds pressure for larger tables where the default political triangle can become too stable. A solo role creates a reason to distrust unusually destructive or isolationist play.

The role is intentionally optional because solo-win roles can distort smaller Commander tables. In a five-player game, a solo chaos role can consume too much oxygen. Around eight or more players, it can help keep the table from settling into predictable blocks.

### Failure Modes To Avoid

- Including Wasteland by default at small tables.
- Letting Wasteland share victory.
- Giving Wasteland enough special mechanics that the table becomes about Wasteland instead of Commander.

## Why These Roles Are Distinct From Treachery

Coup is inspired by hidden-role Commander play, but it should not be treated as Treachery with renamed roles.

The main differences:

- Coup keeps normal Commander free-for-all semantics by default.
- Coup avoids default Magic teammate rules.
- Coup emphasizes incentives, suspicion, and temporary alliances over role-card power packages.
- Coup uses Red/Black conflict to prevent a clean anti-King team.
- Coup treats information policies as configurable pressure knobs, not fixed team reveals.

The Treachery names also should not leak into Coup UI. Coup roles are King, Blue Knight, Black Knight, Red Knight, Green Knight, and Wasteland Knight. They are not Leader, Guardian, Assassin, and Traitor under the hood from the user's point of view.

## Role Count Reasoning

The recommended presets are shaped around preserving the political triangle:

- 5 players: one of each core role.
- 6 players: add a second Black Knight to increase anti-King pressure.
- 7 players: add a second Blue Knight to keep King-side protection viable.
- 8 players: add a third Black Knight to make the anti-King network more dangerous.
- 8 chaos or 9 players: add Wasteland when the table can absorb a solo role.

These are presets, not immutable rules. The Room Operator can tune role counts before start, but normal Coup assumptions expect exactly one King and one Red Knight unless Unsafe Role Count Override is enabled.

## Information Policy Reasoning

Information policies are balance knobs:

- King-to-Blue controls how strong the royal center starts.
- Red-to-Black controls how actively Red can play the usurper game.
- Black-to-Red controls how coordinated the assassination conspiracy becomes.
- Black Network controls whether Black Knights operate as a cell or as independent threats.

The default philosophy is asymmetric information without hard teams:

- King knows Blue by default.
- Red knows all Black Knights by default.
- Black does not know Red by default.
- Black Knights do not know each other by default.

That shape gives King and Red enough information to act, while keeping Black more dangerous, uncertain, and betrayal-oriented.

## App Implications

The role design implies these product behaviors:

- Role text must explain active win conditions directly.
- The app should show private information only to the player entitled to receive it.
- The Operator Dashboard should not become an omniscient hidden-role view outside Debug Mode.
- Advisory Win Prompts should be phrased as suggestions, not final rulings.
- Green eligibility and King Fall state must be tracked explicitly enough to avoid recomputing the wrong result later.
- Inquisition should be public as an event even when the result is private under a variant.
- Royal Guard should be displayed as a public action because using it reveals or exercises a public Coup ability.

## Design Summary

Coup works when every role has a reason to talk, lie, delay, accuse, and bargain without the table becoming a fixed team game.

King creates the public axis. Blue makes loyalty real but hidden. Black creates lethal anti-King pressure. Red turns that pressure into an unstable coup. Green rewards opportunism without allowing hostage patterns. Wasteland adds optional large-table chaos.

The roles are intentionally light on Magic rules rewrites. Their job is to reshape incentives around a Commander game, not replace the Commander game.
