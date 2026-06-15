import assert from "node:assert/strict";
import test from "node:test";

import {
  buildThemeReadabilityLab,
  phaseOneSamples,
  selectableThemes,
} from "./theme_readability_lab.mjs";

test("theme readability lab exposes all selectable themes and review controls", () => {
  const html = buildThemeReadabilityLab({
    cssHref: "../../nix/app/static/css/output.css",
    generatedAt: "2026-06-15T00:00:00.000Z",
    inventory: {
      fileCount: 25,
      occurrenceCount: 669,
      patterns: ["alert-info", "text-warning"],
    },
  });

  assert.match(html, /<html[^>]+data-theme="treacherest"/);
  assert.match(html, /<select[^>]+id="theme-select"/);

  for (const theme of selectableThemes) {
    assert.match(html, new RegExp(`<option value="${theme.value}"`));
  }

  for (const sample of phaseOneSamples) {
    assert.match(html, new RegExp(`data-sample-id="${escapeRegExp(sample.id)}"`));
    for (const target of sample.targets) {
      assert.match(html, new RegExp(`data-review-target="${escapeRegExp(target.id)}"`));
    }
  }

  assert.match(html, /data-review-action="ok"/);
  assert.match(html, /data-review-action="bad"/);
  assert.match(html, /data-review-action="clear"/);
  assert.match(html, /id="copy-review-notes"/);
  assert.match(html, /id="review-notes-output"/);
  assert.match(html, /data-filter="fail"/);
  assert.match(html, /data-filter="debug"/);
  assert.match(html, /getComputedStyle/);
  assert.match(html, /contrastRatio/);

  assert.doesNotMatch(html, /data-on:/);
  assert.doesNotMatch(html, /data-signals/);
  assert.doesNotMatch(html, /\/sse\//);
});

test("phase one samples include the agreed high-risk surfaces", () => {
  const sampleIDs = phaseOneSamples.map((sample) => sample.id);
  const targetIDs = phaseOneSamples.flatMap((sample) => sample.targets.map((target) => target.id));

  for (const expected of [
    "player-view.royal-guard",
    "player-view.inquisition",
    "player-view.privy-panel",
    "player-view.role-card",
    "player-view.roster",
    "operator-dashboard.setup",
    "operator-dashboard.role-counts",
    "debug-control-surface",
  ]) {
    assert.ok(sampleIDs.includes(expected), `missing sample ${expected}`);
  }

  for (const expected of [
    "player-view.royal-guard.title",
    "player-view.royal-guard.body",
    "player-view.royal-guard.warning",
    "player-view.royal-guard.button",
    "debug-control-surface.header",
    "operator-dashboard.role-counts.unsafe-warning",
  ]) {
    assert.ok(targetIDs.includes(expected), `missing target ${expected}`);
  }
});

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
