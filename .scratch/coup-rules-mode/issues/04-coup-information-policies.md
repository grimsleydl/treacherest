# Coup Information Policies

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Implement Coup private information policies for setup and private role views. Defaults are: King has Full Knowledge of Blue Knights, Red has Conspiracy Knowledge of all Black Knights, Black Knights do not know Red, and Black Knights do not have Network knowledge of each other.

Add the documented variants in a way that keeps secret information scoped only to the intended recipient.

## Acceptance criteria

- [x] King sees Blue Knight information according to the selected King-to-Blue Information Policy.
- [x] Red sees Black Knight information according to the selected Red-to-Black Information Policy.
- [x] Black Knights do not see Red or other Black Knights by default.
- [x] The Black-to-Red and Network variants can be selected and are reflected in private information.
- [x] Tests prove private information is not exposed to unintended players.

## Completion notes

- Added explicit Coup information policy types for King-to-Blue, Red-to-Black, Black-to-Red, and Black Network settings.
- Default policy is King full Blue knowledge, Red all Black knowledge, no Black-to-Red knowledge, and no Black Network.
- Added assignment-time private information notes scoped to the recipient's own role card.
- Added setup selectors and Datastar handlers/routes for choosing information policies.
- Wired room-selected policy through the Coup start path.
- Added tests for default information, candidate knowledge, Red variants, Black-to-Red/Network variants, setup update routes, start-path propagation, and view-level privacy.

## Blocked by

- `.scratch/coup-rules-mode/issues/02-coup-5-player-happy-path.md`
