# Coup Role-Image Ingest And Embed Tooling

Status: done

## Parent

`.scratch/coup-rules-mode/PRD.md`

## What to build

Add secondary tooling and repository structure for user-provided Coup role images. The user will provide the images later. The tooling should consume local role-image files, copy them into canonical static paths, and let the app embed and attach them to Coup role cards the same way Treachery role images are embedded.

## Acceptance criteria

- [x] Repository contains a stable place for Coup role images.
- [x] A script can consume local image files for King, Blue Knight, Black Knight, Red Knight, Green Knight, and Wasteland Knight.
- [x] The script copies supported image formats into canonical role-image filenames tied to Coup role IDs.
- [x] The app embeds Coup role images and attaches them to Coup role cards when files are present.
- [x] The tooling is clearly secondary and does not block core Coup gameplay when images are missing.
- [x] Tests cover representative image-loading and import behavior.

## Blocked by

- Resolved by `.scratch/coup-rules-mode/issues/13-role-card-art-direction-decision.md`.

## Notes

- User-provided role images are the near-term path.
- Generated prompts and generated-art style packs are out of near-term scope.

## Completed

- Added `nix/app/static/images/coup/` as the canonical image location.
- Added `go run ./scripts/coup-images -source <dir>` to copy supported user-provided role images into role-ID filenames.
- Added embedded Coup image loading so present images attach to Coup role cards as base64 images and public paths.
- Missing images are allowed, so core Coup setup remains usable before art exists.

## Verification

- `TMPDIR=/workspace/treacherest/.scratch/go-tmp GOTMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./internal/game -run 'TestLoadCoupRoleImages|TestAssignCoupRoles|TestCoup'`
- `TMPDIR=/workspace/treacherest/.scratch/go-tmp GOTMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go test ./scripts/coup-images`
- `TMPDIR=/workspace/treacherest/.scratch/go-tmp GOTMPDIR=/workspace/treacherest/.scratch/go-tmp GOCACHE=/workspace/treacherest/.scratch/go-cache CGO_ENABLED=0 go build -o /workspace/treacherest/.scratch/go-tmp/server-check ./cmd/server`

`go test ./cmd/server` was not used as verification because existing tests still call `handlers.New` with the old signature and fail before this change's server wiring is exercised.
