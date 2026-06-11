# Role-Card Art Direction Decision

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Decide whether role-card art and prompt tooling should be in near-term scope, and if so which visual style pack should be treated as the initial canonical direction. Candidate styles include Treachery-like fantasy, classic Coup court intrigue, Resistance: Coup sci-fi, investigative satire, White Rabbit conspiracy sci-fi, and the frog/scorpion parable theme.

This is a HITL decision issue. Do not build art tooling until the scope and initial art direction are approved.

## Acceptance criteria

- [x] The project has a documented decision on whether role-card art/prompt tooling is near-term scope.
- [x] If in scope, the project has a documented initial style direction.
- [x] If out of scope, the project has a documented revisit trigger.
- [x] Follow-up AFK implementation work can proceed without re-asking the art direction question.

## Decision

Role-card prompt generation and generated art direction are not near-term scope.

Near-term scope is tooling and repository structure for user-provided Coup role images. The user will provide role images later; the app should be able to consume those files, place them at canonical role-image paths, embed them into the binary like Treachery card images, and attach them to the Coup role cards.

No initial generated-art style pack is canonical at this time. Revisit prompt-generation styles only if the user later asks for generated role-card prompts or generated image production.

## Follow-up

- Implement image-ingest/embed tooling in `.scratch/coup-rules-mode/issues/14-role-card-prompt-tooling.md`.

## Blocked by

- `.scratch/coup-rules-mode/issues/05-coup-rules-reference-and-role-goals.md`
