# Coup Role Image Prompts

Status: workflow ready.

This document defines prompt guidance for generating Coup role-card art. Art is secondary to core rules UX, but prompts should produce assets that can later be imported through the existing Coup image pipeline.

Image generation is an external workflow. Treacherest should not call an AI image model or expose in-app prompt generation by default; it should consume completed image files through the importer.

Prompt scope is final role-card art only. Do not create separate AI prompts or separate generated assets for thumbnails, setup lists, or debug panels unless a future UX pass proves those need distinct compositions. Small UI surfaces should crop or scale from the same imported final artwork.

Target a vertical card-art aspect ratio with MTG-like portrait composition. Keep the main character and role-defining props centered enough to survive thumbnail crops, but do not bake a card frame, rules text, title text, mana symbols, logo, or UI into the image.

Generated art should be purely illustrative. Role names, rules text, victory text, reveal state, and card layout must be rendered by Treacherest or a future print/export template so wording can change without regenerating art.

The current import pipeline supports one image per Coup role. Optional style packs are prompt guidance only for now; they should not create parallel image sets or an in-app style selector until a later UI need is proven.

## Asset Workflow

Use this workflow when adding or replacing the production Coup role images.

1. Generate final card-art images outside the app from the prompts in this document.
2. Save the six final images in a temporary source directory outside git-tracked app assets, for example `.scratch/coup-role-images/final/`.
3. Name the files by preferred slug:
   - `king.png`
   - `blue-knight.png`
   - `black-knight.png`
   - `red-knight.png`
   - `green-knight.png`
   - `wasteland-knight.png`
4. From the repository root, run the importer:

   ```sh
   just import-coup-role-images
   ```

5. Confirm the importer wrote canonical embedded assets under `nix/app/static/images/coup/`:
   - `king.*`
   - `blue-knight.*`
   - `black-knight.*`
   - `red-knight.*`
   - `green-knight.*`
   - `wasteland-knight.*`
6. Run the focused image tests:

   ```sh
   just test-coup-role-images
   ```

7. Run the normal build/test path before committing product assets.

The temporary source files in `.scratch/` are working material and should stay untracked. The canonical files copied into `nix/app/static/images/coup/` are product assets and should be committed when the image set is accepted.

The importer requires a complete six-role image set. This is deliberate: importing a partial set can leave the role-card art style inconsistent. The runtime still allows missing Coup images before assets exist, so core gameplay is not blocked.

When replacing an image, re-run the importer with all six source images. It removes stale canonical files for the same role across supported extensions, so switching King from `king.jpg` to `king.webp` does not leave both versions behind.

## Import Targets

The importer accepts `.jpg`, `.jpeg`, `.png`, or `.webp` files named by slug, role name, or role ID. Canonical output files are slug-based so the asset directory stays readable.

| Role | Preferred source slug | Accepted source ID |
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

Use role colors as accents rather than full-image color washes. King should read gold, Blue Knight blue, Black Knight black, Red Knight red, Green Knight green, and Wasteland Knight gray, but each image should still have a rich palette, environmental contrast, and readable material detail.

Avoid visual language that makes Blue/King or Red/Black look like obvious permanent teams. Role colors, heraldry, and props are useful, but the art should preserve suspicion and shifting incentives rather than visually declaring stable team membership.

Use one primary identifiable role figure per image. Anonymous silhouettes, courtiers, shadows, or partial figures are fine for atmosphere, but the art should not show multiple identifiable role characters together in ways that imply private information, fixed alliances, or role relationships.

Wasteland Knight should be visually distinct through gray ruin, exile, and lone-survivor cues, but it should still use the same painterly court-intrigue role-card language as the rest of the Coup set.

Global prompt prefix:

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition with one primary identifiable role figure, centered crop-safe character and symbolic props, rich lighting, clean silhouette, rich palette with role-color accents instead of monochrome color wash, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.
```

Recommended negative prompt:

```text
readable text, typography, title, rules text, logo, watermark, card border, card frame, UI, mana symbols, photorealistic celebrity, existing character, gore, meme, hard team emblem, matching team uniforms, multiple identifiable role characters, explicit alliance scene, monochrome color wash, single-color card, modern gun, blurry face, extra limbs
```

## Canonical Role Prompts

### King

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, centered crop-safe character and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.

The King, a revealed sovereign at the center of a dangerous council chamber. Gold crown, gold cloak, throne room map table, guarded confidence mixed with suspicion, courtiers and shadowed rivals implied at the edge of the scene. The mood should suggest political gravity and vulnerability rather than battlefield command. Dominant accent color: gold.
```

### Blue Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, centered crop-safe character and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.

The Blue Knight, a royal guard and inquisitor sworn to protect the King. Blue enamel armor, shield angled toward an unseen threat, lantern or sealed warrant in one hand, palace corridor behind them, watchful and restrained. The image should imply defense, investigation, and a difficult choice about when to reveal loyalty. Dominant accent color: blue.
```

### Black Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, centered crop-safe character and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.

The Black Knight, a hired assassin in dark ceremonial armor moving through a candlelit palace passage. Black cloak, hidden dagger, broken contract seal, face partly obscured but expressive. The scene should suggest a killer who intends to betray the person who hired them, with no explicit violence. Dominant accent color: black.
```

### Red Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, centered crop-safe character and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.

The Red Knight, a charismatic usurper in red court armor, standing over a war map and a half-sealed bargain. Red cloak, signet ring, torchlight, ambitious expression, unaware of the betrayal hidden in the bargain. The image should imply a coup plot, confidence, and danger from supposed allies. Dominant accent color: red.
```

### Green Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, centered crop-safe character and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.

The Green Knight, an opportunist envoy watching rival factions from a garden balcony beside the council hall. Green cloak, coin or treaty token balanced in one hand, vine-carved railings, expression that could be friendly or calculating. The image should imply temporary alliances and flexible incentives, not a fixed team. Dominant accent color: green.
```

### Wasteland Knight

```text
Vertical fantasy role-card illustration for a hidden-role Commander table variant, portrait card-art aspect ratio. Court intrigue, political tension, polished painterly board-game card art, dramatic but readable single-role composition, centered crop-safe character and symbolic props, rich lighting, clean silhouette, no readable text, no logo, no watermark, no card frame, no UI, no mana symbols, no gore, no existing copyrighted character.

The Wasteland Knight, a solitary gray-armored exile standing amid the ruined outskirts of the kingdom while still feeling like part of the same court-intrigue role-card set. Ash-gray armor, cracked banner, abandoned crown fragment in the dust, storm-lit wasteland behind them. The image should imply a lone survivor who shares victory with no one. Dominant accent color: gray.
```

## Optional Style Packs

Optional style packs should transform the same six role concepts without changing rules text or default role identity.

The frog/scorpion parable pack is explicitly optional and should not become the default visual language for Coup. The canonical default remains neutral court-intrigue fantasy.

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
