# Coup Role Image Prompts

Status: first draft.

This document defines prompt guidance for generating Coup role-card art. Art is secondary to core rules UX, but prompts should produce assets that can later be imported through the existing Coup image pipeline.

## Import Targets

The importer accepts `.jpg`, `.jpeg`, `.png`, or `.webp` files named by slug, role name, or role ID. Canonical output files are ID-based.

| Role | Preferred source slug | Canonical ID |
| --- | --- | --- |
| King | `king` | `1001` |
| Blue Knight | `blue-knight` | `1002` |
| Black Knight | `black-knight` | `1003` |
| Red Knight | `red-knight` | `1004` |
| Green Knight | `green-knight` | `1005` |
| Wasteland Knight | `wasteland-knight` | `1006` |

Importer entry point: `nix/app/scripts/coup-images/import_coup_images.go`.

## Default Style Pack

Use Canonical Coup Art Direction by default: neutral court-intrigue fantasy role-card art. The goal is political tension, hidden incentives, and readable role identity without making the roles look like fixed hard teams.

Global prompt prefix:

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.
```

Recommended negative prompt:

```text
readable text, typography, logo, watermark, card border, UI, photorealistic celebrity, existing character, gore, meme, hard team emblem, modern gun, blurry face, extra limbs
```

## Canonical Role Prompts

### King

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.

The King, a revealed sovereign at the center of a dangerous council chamber. Gold crown, gold cloak, throne room map table, guarded confidence mixed with suspicion, courtiers and shadowed rivals implied at the edge of the scene. The mood should suggest political gravity and vulnerability rather than battlefield command. Dominant accent color: gold.
```

### Blue Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.

The Blue Knight, a royal guard and inquisitor sworn to protect the King. Blue enamel armor, shield angled toward an unseen threat, lantern or sealed warrant in one hand, palace corridor behind them, watchful and restrained. The image should imply defense, investigation, and a difficult choice about when to reveal loyalty. Dominant accent color: blue.
```

### Black Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.

The Black Knight, a hired assassin in dark ceremonial armor moving through a candlelit palace passage. Black cloak, hidden dagger, broken contract seal, face partly obscured but expressive. The scene should suggest a killer who intends to betray the person who hired them, with no explicit violence. Dominant accent color: black.
```

### Red Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.

The Red Knight, a charismatic usurper in red court armor, standing over a war map and a half-sealed bargain. Red cloak, signet ring, torchlight, ambitious expression, unaware of the betrayal hidden in the bargain. The image should imply a coup plot, confidence, and danger from supposed allies. Dominant accent color: red.
```

### Green Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.

The Green Knight, an opportunist envoy watching rival factions from a garden balcony beside the council hall. Green cloak, coin or treaty token balanced in one hand, vine-carved railings, expression that could be friendly or calculating. The image should imply temporary alliances and flexible incentives, not a fixed team. Dominant accent color: green.
```

### Wasteland Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, detailed costume and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no gore, no existing copyrighted character.

The Wasteland Knight, a solitary gray-armored exile standing amid the ruined outskirts of the kingdom. Ash-gray armor, cracked banner, abandoned crown fragment in the dust, storm-lit wasteland behind them. The image should imply a lone survivor who shares victory with no one. Dominant accent color: gray.
```

## Optional Style Packs

Optional style packs should transform the same six role concepts without changing rules text or default role identity.

### Frog/Scorpion Parable

- King: Frog King or great toad sovereign.
- Blue Knight: Bullfrog royal guard or frog inquisitor.
- Black Knight: Scorpion assassin.
- Red Knight: Frog who made the doomed bargain with the scorpion.
- Green Knight: Tree frog or reed frog opportunist.
- Wasteland Knight: Ruin-haunting marsh predator, cane toad warlord, or desert-adapted scorpion revenant.

### Treachery-Like Fantasy

High-fantasy identity-card energy, dramatic role portrait, darker hidden-role mood. Do not copy existing Treachery card art or exact layouts.

### Classic Court Intrigue

Stylized court politics, masks, daggers, seals, letters, nobles, guards, and conspirators. Less battlefield fantasy, more palace maneuvering.

### Sci-Fi Resistance

Futuristic conspiracy board-game tone with resistance cells, alien courts, encrypted contracts, surveillance, and neon faction accents.

### Investigative Satire

Conspiracy-board style with documents, red string, financial ledgers, cameras, and absurdly serious investigators. Keep the role readable and avoid making the joke overpower the game asset.

### White Rabbit Conspiracy Sci-Fi

Alien-aristocrat court intrigue, surreal white-rabbit symbolism, sterile palaces, impossible clocks, and conspiracy-board imagery.
