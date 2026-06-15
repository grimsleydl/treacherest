# Coup Role Images

This directory holds the canonical embedded Coup role images used by the app. Do not hand-copy generated files here. Put final generated art in a temporary source directory, then run the importer so filenames and stale-extension cleanup stay consistent.

Example source directory:

```text
.scratch/coup-role-images/final/
├── king.png
├── blue-knight.png
├── black-knight.png
├── red-knight.png
├── green-knight.png
└── wasteland-knight.png
```

Run from the repository root:

```sh
just import-coup-role-images
```

The importer also accepts source filenames by role ID, such as `1001.png`, but slug filenames are preferred while reviewing generated art.

Canonical embedded filenames use Coup role IDs and should be committed when the final image set is accepted:

- `1001.*` King
- `1002.*` Blue Knight
- `1003.*` Black Knight
- `1004.*` Red Knight
- `1005.*` Green Knight
- `1006.*` Wasteland Knight

Supported extensions are `.jpg`, `.jpeg`, `.png`, and `.webp`.

The importer requires all six role images in the source directory. Missing embedded images are allowed at runtime before art exists, but imports should be complete sets so the app does not ship a mixed or partial Coup art direction.

See [coup-role-image-prompts.md](../../../../../docs/coup-role-image-prompts.md) for generation prompts and the full review workflow.
