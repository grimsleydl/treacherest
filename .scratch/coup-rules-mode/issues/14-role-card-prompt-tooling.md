# Coup Role-Image Ingest And Embed Tooling

Status: ready-for-agent

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add secondary tooling and repository structure for user-provided Coup role images. The user will provide the images later. The tooling should consume local role-image files, copy them into canonical static paths, and let the app embed and attach them to Coup role cards the same way Treachery role images are embedded.

## Acceptance criteria

- [ ] Repository contains a stable place for Coup role images.
- [ ] A script can consume local image files for King, Blue Knight, Black Knight, Red Knight, Green Knight, and Wasteland Knight.
- [ ] The script copies supported image formats into canonical role-image filenames tied to Coup role IDs.
- [ ] The app embeds Coup role images and attaches them to Coup role cards when files are present.
- [ ] The tooling is clearly secondary and does not block core Coup gameplay when images are missing.
- [ ] Tests cover representative image-loading and import behavior.

## Blocked by

- Resolved by `.scratch/coup-rules-mode/issues/13-role-card-art-direction-decision.md`.

## Notes

- User-provided role images are the near-term path.
- Generated prompts and generated-art style packs are out of near-term scope.
