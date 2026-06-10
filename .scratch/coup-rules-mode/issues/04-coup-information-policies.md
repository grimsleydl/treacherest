# Coup Information Policies

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Implement Coup private information policies for setup and private role views. Defaults are: King has Full Knowledge of Blue Knights, Red has Conspiracy Knowledge of all Black Knights, Black Knights do not know Red, and Black Knights do not have Network knowledge of each other.

Add the documented variants in a way that keeps secret information scoped only to the intended recipient.

## Acceptance criteria

- [ ] King sees Blue Knight information according to the selected King-to-Blue Information Policy.
- [ ] Red sees Black Knight information according to the selected Red-to-Black Information Policy.
- [ ] Black Knights do not see Red or other Black Knights by default.
- [ ] The Black-to-Red and Network variants can be selected and are reflected in private information.
- [ ] Tests prove private information is not exposed to unintended players.

## Blocked by

- `.scratch/coup-rules-mode/issues/02-coup-5-player-happy-path.md`
