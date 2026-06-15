#!/usr/bin/env node
import { mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { selectableThemes } from "./theme_data.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const appRoot = resolve(__dirname, "..");
const repoRoot = resolve(appRoot, "../..");
const defaultOutputDir = resolve(repoRoot, ".scratch/theme-readability");
const defaultOutputPath = resolve(defaultOutputDir, "index.html");

export { selectableThemes };

export const phaseOneSamples = [
  {
    id: "player-view.royal-guard",
    title: "Royal Guard NoticeCard",
    groups: ["core", "notices", "controls", "player-view"],
    description: "Known nested-color case: info alert surface with base-content body text, warning text, and a primary action.",
    targets: [
      target("player-view.royal-guard.title", "Title", "alert-info text on info alert surface"),
      target("player-view.royal-guard.body", "Body", "text-base-content/80 inside alert-info"),
      target("player-view.royal-guard.warning", "Warning", "text-warning inside alert-info"),
      target("player-view.royal-guard.button", "Button", "btn-primary inside alert-info"),
    ],
    markup: `
      <section class="notice-card alert flex items-start alert-info" data-lab-surface="royal-guard">
        <div class="min-w-0">
          <h3 class="font-bold" data-review-target="player-view.royal-guard.title">Royal Guard</h3>
          <div class="mt-1 text-sm">
            <p class="text-sm text-base-content/80" data-review-target="player-view.royal-guard.body">A revealed Blue Knight may have any number of untapped creatures they control block creatures attacking the King.</p>
            <p class="text-xs text-warning" data-review-target="player-view.royal-guard.warning">Using Royal Guard publicly reveals you as Blue Knight.</p>
            <div class="mt-3">
              <button type="button" class="btn min-h-11 btn-primary" data-review-target="player-view.royal-guard.button">Use Royal Guard</button>
            </div>
          </div>
        </div>
      </section>
    `,
  },
  {
    id: "player-view.inquisition",
    title: "Inquisition Notices",
    groups: ["core", "notices", "controls", "player-view"],
    description: "Public and result notices for the Coup Inquisition flow.",
    targets: [
      target("player-view.inquisition.notice-title", "Notice title", "alert-info content"),
      target("player-view.inquisition.notice-body", "Notice body", "text-base-content/80 inside alert-info"),
      target("player-view.inquisition.waiting", "Waiting note", "text-base-content/60 inside alert-info"),
      target("player-view.inquisition.success", "Success result", "text-success inside alert-success"),
      target("player-view.inquisition.failure", "Failure result", "text-error inside alert-error"),
      target("player-view.inquisition.confirm-button", "Witness button", "btn-outline btn-primary inside alert-info"),
    ],
    markup: `
      <div class="space-y-3">
        <section class="notice-card alert flex items-start alert-info">
          <div class="min-w-0">
            <h3 class="font-bold" data-review-target="player-view.inquisition.notice-title">Inquisition Notice</h3>
            <div class="mt-1 text-sm">
              <p class="text-sm text-base-content/80" data-review-target="player-view.inquisition.notice-body">Blue Knight has called an Inquisition and named a suspected Red Knight.</p>
              <p class="text-xs text-base-content/60" data-review-target="player-view.inquisition.waiting">Waiting for one living non-Blue witness to confirm the table was notified.</p>
              <button type="button" class="btn btn-sm btn-outline btn-primary mt-3" data-review-target="player-view.inquisition.confirm-button">I witnessed this</button>
            </div>
          </div>
        </section>
        <section class="notice-card alert flex items-start alert-success">
          <div class="min-w-0">
            <h3 class="font-bold">Inquisition Result</h3>
            <p class="text-sm text-success" data-review-target="player-view.inquisition.success">Inquisition succeeded. Red Knight was revealed.</p>
          </div>
        </section>
        <section class="notice-card alert flex items-start alert-error">
          <div class="min-w-0">
            <h3 class="font-bold">Inquisition Result</h3>
            <p class="text-sm text-error" data-review-target="player-view.inquisition.failure">Inquisition failed. Blue should lose half their life total.</p>
          </div>
        </section>
      </div>
    `,
  },
  {
    id: "player-view.advisory-win",
    title: "Advisory Win Prompt",
    groups: ["core", "notices", "controls", "player-view"],
    description: "Advisory-tier notice and confirm/reject actions.",
    targets: [
      target("player-view.advisory-win.kicker", "Kicker", "opacity-70 inside alert-warning"),
      target("player-view.advisory-win.title", "Title", "alert-warning content"),
      target("player-view.advisory-win.body", "Facts", "base text inside alert-warning"),
      target("player-view.advisory-win.confirm", "Confirm button", "btn-primary inside alert-warning"),
      target("player-view.advisory-win.reject", "Reject button", "btn-outline btn-warning inside alert-warning"),
    ],
    markup: `
      <section class="notice-card alert flex items-start alert-warning">
        <div class="min-w-0">
          <p class="font-mono text-[10px] font-bold uppercase tracking-[0.16em] opacity-70" data-review-target="player-view.advisory-win.kicker">ADVISORY - NOT A RULING</p>
          <h3 class="font-bold" data-review-target="player-view.advisory-win.title">Looks like Red might have just won</h3>
          <div class="mt-1 text-sm">
            <p data-review-target="player-view.advisory-win.body">King is dead, Red is alive, and all Black Knights are eliminated. Confirm with the table before ending the game.</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <button type="button" class="btn min-h-11 btn-primary" data-review-target="player-view.advisory-win.confirm">Confirm Win</button>
              <button type="button" class="btn min-h-11 btn-outline btn-warning" data-review-target="player-view.advisory-win.reject">Reject Prompt</button>
            </div>
          </div>
        </div>
      </section>
    `,
  },
  {
    id: "player-view.eliminated",
    title: "Eliminated Player Notice",
    groups: ["core", "notices", "player-view"],
    description: "Error-tier read-only notice for a player out of the game.",
    targets: [
      target("player-view.eliminated.title", "Title", "alert-error content"),
      target("player-view.eliminated.body", "Body", "base text inside alert-error"),
    ],
    markup: `
      <section class="notice-card alert flex items-start alert-error">
        <div class="min-w-0">
          <h3 class="font-bold" data-review-target="player-view.eliminated.title">You are out of the game</h3>
          <div class="mt-1 text-sm">
            <p data-review-target="player-view.eliminated.body">Your actions are removed. The Room Operator can undo this from the Operator Dashboard if it was marked by mistake.</p>
          </div>
        </div>
      </section>
    `,
  },
  {
    id: "player-view.privy-panel",
    title: "Privy Panel",
    groups: ["core", "privy", "player-view", "role-cards"],
    description: "Concealed and open private-information surfaces using the project privy tokens.",
    targets: [
      target("player-view.privy-panel.header", "Header", "privy-content on privy-bg"),
      target("player-view.privy-panel.hold-button", "Hold button", "inline privy inverse button"),
      target("player-view.privy-panel.open-button", "Open button", "btn-ghost on privy surface"),
      target("player-view.privy-panel.auto-note", "Auto-conceal note", "opacity-60 on privy surface"),
    ],
    markup: `
      <section class="privy w-full max-w-md overflow-hidden shadow-xl open">
        <header class="flex items-center justify-between gap-3 border-b px-4 py-2" style="border-color: var(--privy-border)">
          <span class="font-mono text-[10px] font-bold uppercase tracking-[0.16em]" data-review-target="player-view.privy-panel.header">PRIVATE - ONLY YOU</span>
          <span class="font-mono text-[10px] uppercase tracking-[0.12em]">OPEN</span>
        </header>
        <div class="grid">
          <div class="privy-veil p-6 text-center">
            <button type="button" class="btn min-h-11 w-full" style="background: var(--privy-content); color: var(--privy-bg)" data-review-target="player-view.privy-panel.hold-button">Hold to peek</button>
            <button type="button" class="btn btn-ghost mt-2 min-h-11 w-full" data-review-target="player-view.privy-panel.open-button">Open until concealed</button>
          </div>
          <div class="p-4">
            <p class="mt-2 text-center text-xs opacity-60" data-review-target="player-view.privy-panel.auto-note">Conceals on its own in 30 seconds</p>
          </div>
        </div>
      </section>
    `,
  },
  {
    id: "player-view.role-card",
    title: "Role Card Presentation",
    groups: ["core", "role-cards", "player-view"],
    description: "Hero and compact role-card surfaces with badges, goal callout, and disclosure rows.",
    targets: [
      target("player-view.role-card.eyebrow", "Eyebrow", "opacity-70 on role card"),
      target("player-view.role-card.name", "Role name", "base text on bg-base-100"),
      target("player-view.role-card.type-badge", "Type badge", "badge-primary"),
      target("player-view.role-card.rarity-badge", "Rarity badge", "badge-secondary"),
      target("player-view.role-card.goal-label", "Goal label", "text-primary on bg-primary/10"),
      target("player-view.role-card.goal-body", "Goal body", "base text on bg-primary/10"),
      target("player-view.role-card.disclosure", "Disclosure summary", "base text on bg-base-200"),
    ],
    markup: `
      <article class="role-card role-card-hero card bg-base-100 shadow-xl border-4 border-info">
        <div class="card-body gap-4">
          <header class="space-y-2">
            <p class="font-mono text-[10px] font-bold uppercase tracking-[0.16em] opacity-70" data-review-target="player-view.role-card.eyebrow">Private role</p>
            <h2 class="font-display text-3xl font-semibold leading-tight" data-review-target="player-view.role-card.name">Blue Knight</h2>
            <div class="flex flex-wrap gap-2">
              <span class="badge badge-primary" data-review-target="player-view.role-card.type-badge">Coup Role</span>
              <span class="badge badge-secondary" data-review-target="player-view.role-card.rarity-badge">Coup</span>
            </div>
          </header>
          <section class="rounded-box border border-primary/40 bg-primary/10 p-3">
            <p class="font-mono text-[10px] font-bold uppercase tracking-[0.16em] text-primary" data-review-target="player-view.role-card.goal-label">Win Condition:</p>
            <p class="mt-1 text-sm font-semibold" data-review-target="player-view.role-card.goal-body">Win with the King. Lose if the King loses.</p>
          </section>
          <details class="collapse collapse-arrow border border-base-300 bg-base-200" open>
            <summary class="collapse-title min-h-11 py-3 text-sm font-semibold" data-review-target="player-view.role-card.disclosure">Rulings</summary>
            <div class="collapse-content text-xs">Royal Guard and Inquisition details live here.</div>
          </details>
        </div>
      </article>
    `,
  },
  {
    id: "player-view.roster",
    title: "PlayerRow And StateChip",
    groups: ["core", "player-view", "controls"],
    description: "Stable roster row states, chips, public role expansion, and face-down copy.",
    targets: [
      target("player-view.roster.name", "Player name", "base text on bg-base-100"),
      target("player-view.roster.you-chip", "You chip", "badge-primary"),
      target("player-view.roster.operator-chip", "Operator chip", "badge-primary badge-outline"),
      target("player-view.roster.leader-chip", "Leader chip", "badge-warning"),
      target("player-view.roster.facedown-chip", "Face-down chip", "badge-ghost"),
      target("player-view.roster.eliminated-chip", "Eliminated chip", "badge-error badge-outline"),
      target("player-view.roster.facedown-copy", "Face-down copy", "text-base-content/60"),
      target("player-view.roster.public-role-copy", "Public role copy", "text-base-content/80 on bg-base-200"),
    ],
    markup: `
      <div class="space-y-3">
        <div class="player-row rounded-box border border-base-300 bg-base-100">
          <div class="flex min-h-11 items-center justify-between gap-3 px-3 py-2">
            <div class="min-w-0">
              <p class="truncate font-semibold" data-review-target="player-view.roster.name">Debug Player 2</p>
              <div class="mt-1 flex flex-wrap gap-1">
                <span class="state-chip badge badge-xs font-bold uppercase tracking-[0.12em] badge-primary" data-review-target="player-view.roster.you-chip">You</span>
                <span class="state-chip badge badge-xs font-bold uppercase tracking-[0.12em] badge-primary badge-outline" data-review-target="player-view.roster.operator-chip">Operator</span>
                <span class="state-chip badge badge-xs font-bold uppercase tracking-[0.12em] badge-warning" data-review-target="player-view.roster.leader-chip">Leader</span>
                <span class="state-chip badge badge-xs font-bold uppercase tracking-[0.12em] badge-ghost" data-review-target="player-view.roster.facedown-chip">Face Down</span>
                <span class="state-chip badge badge-xs font-bold uppercase tracking-[0.12em] badge-error badge-outline" data-review-target="player-view.roster.eliminated-chip">Eliminated</span>
              </div>
            </div>
            <span class="btn btn-ghost btn-xs">Details</span>
          </div>
          <div class="border-t border-base-300 px-3 py-3 text-sm">
            <p class="text-base-content/60" data-review-target="player-view.roster.facedown-copy">Card is face down.</p>
            <div class="mt-2 space-y-2 rounded-box bg-base-200 p-3">
              <p class="font-semibold">King</p>
              <div class="text-xs text-base-content/80" data-review-target="player-view.roster.public-role-copy">Win if alive when Black, Red, and Wasteland are eliminated.</div>
            </div>
          </div>
        </div>
      </div>
    `,
  },
  {
    id: "player-lobby.summary",
    title: "Player Lobby Summary",
    groups: ["core", "player-lobby"],
    description: "Non-operator lobby hero, settings summary, open seats, and rules reference copy.",
    targets: [
      target("player-lobby.summary.room-label", "Room label", "text-base-content/60 on bg-base-100"),
      target("player-lobby.summary.room-code", "Room code", "text-primary on bg-base-100"),
      target("player-lobby.summary.status", "Status line", "text-base-content/70 on bg-base-100"),
      target("player-lobby.summary.settings", "Settings summary", "text-base-content/80 on bg-base-100"),
      target("player-lobby.summary.open-seat", "Open seat", "text-base-content/60 on bg-base-200/50"),
    ],
    markup: `
      <div class="space-y-3">
        <div class="rounded-box border border-base-300 bg-base-100 p-5 shadow-sm">
          <p class="text-xs font-bold uppercase tracking-[0.16em] text-base-content/60" data-review-target="player-lobby.summary.room-label">Room Code</p>
          <h1 class="font-mono text-5xl font-bold tracking-widest text-primary" data-review-target="player-lobby.summary.room-code">7HN2N</h1>
          <p class="mt-2 text-sm text-base-content/70" data-review-target="player-lobby.summary.status">Waiting for host - 3 of 5 seats filled</p>
        </div>
        <div class="rounded-box border border-base-300 bg-base-100 px-4 py-3 text-sm text-base-content/80" data-review-target="player-lobby.summary.settings">Coup - 5 players - Public Inquisition - Full King knowledge</div>
        <div class="rounded-box border border-dashed border-base-300 bg-base-200/50 px-3 py-3 text-sm text-base-content/60" data-review-target="player-lobby.summary.open-seat">Open seat</div>
      </div>
    `,
  },
  {
    id: "operator-dashboard.setup",
    title: "Operator Dashboard Setup",
    groups: ["core", "operator-dashboard", "controls"],
    description: "Room setup cards, QR placeholder, public facts, validation text, and start controls.",
    targets: [
      target("operator-dashboard.setup.title", "Dashboard title", "text-base-content on bg-base-200"),
      target("operator-dashboard.setup.room-code", "Room code", "text-base-content on bg-base-200"),
      target("operator-dashboard.setup.card-title", "Card title", "base text on bg-base-100"),
      target("operator-dashboard.setup.qr-placeholder", "QR placeholder", "text-base-content/60 on bg-base-300"),
      target("operator-dashboard.setup.public-note", "Public note", "text-base-content/70 on bg-base-100"),
      target("operator-dashboard.setup.validation-ok", "Validation OK", "text-success on bg-base-100"),
      target("operator-dashboard.setup.validation-warning", "Validation warning", "text-warning on bg-base-100"),
      target("operator-dashboard.setup.start-button", "Start button", "btn-primary"),
    ],
    markup: `
      <div class="space-y-3 rounded-box bg-base-200 p-4">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <h1 class="text-3xl font-bold text-base-content" data-review-target="operator-dashboard.setup.title">Operator Dashboard</h1>
          <p class="font-mono text-2xl font-bold tracking-[0.18em] text-base-content" data-review-target="operator-dashboard.setup.room-code">7HN2N</p>
        </div>
        <div class="card border border-base-300 bg-base-100 shadow-lg p-4">
          <h2 class="card-title text-base-content" data-review-target="operator-dashboard.setup.card-title">Scan to Join</h2>
          <div class="mt-3 flex h-32 w-32 items-center justify-center rounded bg-base-300 text-base-content/60" data-review-target="operator-dashboard.setup.qr-placeholder">QR</div>
          <p class="mt-3 text-sm text-base-content/70" data-review-target="operator-dashboard.setup.public-note">Public state only. Hidden roles and private instructions are not shown here.</p>
          <p class="mt-3 text-sm text-success" data-review-target="operator-dashboard.setup.validation-ok">Role counts match current player count.</p>
          <p class="text-sm text-warning" data-review-target="operator-dashboard.setup.validation-warning">Unsafe override is active for this required Coup role.</p>
          <button type="button" class="btn btn-primary btn-lg mt-3 w-full text-xl" data-review-target="operator-dashboard.setup.start-button">Start Game</button>
        </div>
      </div>
    `,
  },
  {
    id: "operator-dashboard.role-counts",
    title: "Role Count Configuration",
    groups: ["core", "operator-dashboard", "controls"],
    description: "Stepper rows, required-role copy, and Unsafe Role Count Override warning surface.",
    targets: [
      target("operator-dashboard.role-counts.heading", "Heading", "text-base-content on bg-base-100"),
      target("operator-dashboard.role-counts.count", "Count", "text-base-content on bg-base-100"),
      target("operator-dashboard.role-counts.detail", "Detail", "text-base-content/80 on bg-base-100"),
      target("operator-dashboard.role-counts.stepper", "Stepper", "btn-neutral"),
      target("operator-dashboard.role-counts.required-copy", "Required copy", "text-base-content/70 on bg-base-100"),
      target("operator-dashboard.role-counts.unsafe-warning", "Unsafe warning", "text-warning on bg-warning/10"),
    ],
    markup: `
      <section class="role-configuration card border border-base-300 bg-base-100 shadow-lg">
        <div class="card-body gap-3">
          <h2 class="card-title" data-review-target="operator-dashboard.role-counts.heading">Role Count Configuration</h2>
          <div class="config-row rounded-box border border-base-300 bg-base-100 px-4 py-3">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="font-semibold">Blue Knight</p>
                <p class="mt-1 text-xs text-base-content/70" data-review-target="operator-dashboard.role-counts.required-copy">Required for Coup baseline. Enable Unsafe Role Count Override to edit this row.</p>
              </div>
              <div class="join">
                <button type="button" class="btn btn-square btn-sm btn-neutral join-item" data-review-target="operator-dashboard.role-counts.stepper">-</button>
                <span class="grid min-w-12 place-items-center px-3 text-xl font-extrabold leading-none text-base-content" data-review-target="operator-dashboard.role-counts.count">1</span>
                <button type="button" class="btn btn-square btn-sm btn-neutral join-item">+</button>
              </div>
            </div>
            <div class="collapse-content text-sm text-base-content/80" data-review-target="operator-dashboard.role-counts.detail">Protects the King, wins with the King, and may use Royal Guard and Inquisition.</div>
          </div>
          <div class="rounded-box border border-warning/40 bg-warning/10 px-4 py-3">
            <p class="text-warning" data-review-target="operator-dashboard.role-counts.unsafe-warning">Unsafe Role Count Override lets this room start with structurally broken role counts.</p>
          </div>
        </div>
      </section>
    `,
  },
  {
    id: "debug-control-surface",
    title: "Debug Control Surface",
    groups: ["debug", "controls"],
    description: "Debug-only right rail with warning/error affordances and hidden-role spoiler redaction.",
    targets: [
      target("debug-control-surface.header", "Header", "text-error on bg-base-100/95"),
      target("debug-control-surface.mode-chip", "Mode chip", "badge-warning"),
      target("debug-control-surface.helper", "Helper text", "text-base-content/70 on bg-base-100/95"),
      target("debug-control-surface.spoiler-label", "Spoiler label", "text-base-content/70 on bg-base-200"),
      target("debug-control-surface.redacted", "Redacted role", "bg-base-content/20 on bg-base-200"),
      target("debug-control-surface.warning-button", "Warning button", "btn-warning"),
      target("debug-control-surface.clear-button", "Clear button", "btn-outline btn-error"),
    ],
    markup: `
      <aside class="w-full max-w-sm rounded-lg border-2 border-dashed border-error bg-base-100/95 p-4 shadow-2xl">
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="font-mono text-[10px] font-bold uppercase tracking-[0.16em] text-error" data-review-target="debug-control-surface.header">DEBUG BUILD</div>
            <h2 class="text-sm font-bold">Debug Control Surface</h2>
          </div>
          <span class="badge badge-warning" data-review-target="debug-control-surface.mode-chip">Debug Mode</span>
        </div>
        <p class="mt-3 text-xs text-base-content/70" data-review-target="debug-control-surface.helper">Host-only controls are available from the Room Operator session.</p>
        <label class="mt-3 flex items-center justify-between gap-3 rounded-box border border-base-300 bg-base-200 p-2 text-xs">
          <span class="font-semibold uppercase text-base-content/70" data-review-target="debug-control-surface.spoiler-label">Show hidden roles</span>
          <input type="checkbox" class="toggle toggle-warning toggle-sm"/>
        </label>
        <div class="mt-3 w-full rounded bg-base-200 p-2 space-y-1 border text-left">
          <span class="inline-flex rounded bg-base-content/20 px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.12em]" data-review-target="debug-control-surface.redacted">Hidden role</span>
        </div>
        <div class="mt-3 grid grid-cols-2 gap-2">
          <button type="button" class="btn btn-xs btn-warning w-full" data-review-target="debug-control-surface.warning-button">Dump</button>
          <button type="button" class="btn btn-xs min-h-8 btn-outline btn-error" data-review-target="debug-control-surface.clear-button">Clear Room</button>
        </div>
      </aside>
    `,
  },
];

export function buildThemeReadabilityLab({
  cssHref = "../../nix/app/static/css/output.css",
  generatedAt = new Date().toISOString(),
  inventory = collectColorUsageInventory(),
} = {}) {
  const themeOptions = renderThemeOptions();
  const samples = phaseOneSamples.map(renderSample).join("\n");
  const inventoryPatternList = inventory.patterns.map((pattern) => `<code>${escapeHTML(pattern)}</code>`).join("\n");

  return `<!doctype html>
<html lang="en" data-theme="treacherest">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Treacherest Theme Readability Lab</title>
  <link rel="stylesheet" href="${escapeAttribute(cssHref)}"/>
  <style>
    :root { color-scheme: light dark; }
    body { min-height: 100vh; }
    .lab-shell { display: grid; grid-template-columns: minmax(0, 1fr) minmax(18rem, 24rem); gap: 1rem; align-items: start; }
    .lab-controls { position: sticky; top: 1rem; max-height: calc(100vh - 2rem); overflow: auto; }
    .lab-samples { display: grid; gap: 1rem; }
    .lab-sample[hidden] { display: none; }
    .lab-target { outline: 2px dashed color-mix(in oklab, var(--color-primary) 45%, transparent); outline-offset: 2px; cursor: pointer; }
    .lab-target[data-review-state="ok"] { outline-color: var(--color-success); }
    .lab-target[data-review-state="bad"] { outline-color: var(--color-error); }
    .target-row[data-review-state="ok"] { border-color: var(--color-success); }
    .target-row[data-review-state="bad"] { border-color: var(--color-error); }
    .target-row[data-status="fail"] { border-color: var(--color-error); }
    .target-row[data-status="large"] { border-color: var(--color-warning); }
    .target-row[data-status="aa"] { border-color: var(--color-success); }
    .target-row[data-status="manual"] { border-color: var(--color-info); }
    .lab-swatch { display: inline-block; width: 0.75rem; height: 0.75rem; border-radius: 999px; border: 1px solid color-mix(in oklab, currentColor 30%, transparent); vertical-align: -0.1rem; }
    .inventory-grid { display: flex; flex-wrap: wrap; gap: 0.35rem; }
    @media (max-width: 960px) {
      .lab-shell { grid-template-columns: 1fr; }
      .lab-controls { position: static; max-height: none; }
    }
  </style>
</head>
<body class="bg-base-200 text-base-content">
  <main class="mx-auto max-w-7xl p-4">
    <header class="mb-4 rounded-box border border-base-300 bg-base-100 p-4 shadow-sm">
      <p class="text-xs font-bold uppercase tracking-[0.16em] text-base-content/60">Generated ${escapeHTML(generatedAt)}</p>
      <div class="mt-2 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-3xl font-bold">Theme Readability Lab</h1>
          <p class="mt-1 max-w-3xl text-sm text-base-content/70">Static design QA artifact for reviewing actual Treacherest text/background combinations across every selectable theme. Click targets or use OK/Bad buttons to accumulate copyable notes for the selected theme.</p>
        </div>
        <label class="form-control w-full max-w-xs">
          <span class="label-text">Theme</span>
          <select id="theme-select" class="select select-bordered select-sm">${themeOptions}</select>
        </label>
      </div>
    </header>

    <div class="lab-shell">
      <section class="lab-samples" id="lab-samples">
        <div class="rounded-box border border-base-300 bg-base-100 p-4 shadow-sm">
          <h2 class="text-lg font-bold">Filters</h2>
          <div class="mt-3 flex flex-wrap gap-2" id="filter-controls">
            ${renderFilterButton("all", "All")}
            ${renderFilterButton("fail", "Show fails")}
            ${renderFilterButton("manual", "Manual")}
            ${renderFilterButton("core", "Core surfaces")}
            ${renderFilterButton("debug", "Debug")}
            ${renderFilterButton("role-cards", "Role cards")}
            ${renderFilterButton("notices", "Notices")}
            ${renderFilterButton("controls", "Controls")}
          </div>
        </div>
        ${samples}
        <section class="rounded-box border border-base-300 bg-base-100 p-4 shadow-sm">
          <h2 class="text-lg font-bold">Raw Inventory Backstop</h2>
          <p class="mt-1 text-sm text-base-content/70">Scanned ${inventory.fileCount} source files and found ${inventory.occurrenceCount} color-bearing utility occurrences. Phase 1 uses curated representative samples; these patterns guide later expansion.</p>
          <div class="inventory-grid mt-3 text-xs">${inventoryPatternList}</div>
        </section>
      </section>

      <aside class="lab-controls rounded-box border border-base-300 bg-base-100 p-4 shadow-sm">
        <h2 class="text-lg font-bold">Review Notes</h2>
        <p class="mt-1 text-sm text-base-content/70">Marks are scoped to the selected theme. Copy this section back into chat or save it into a follow-up file.</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <button type="button" id="copy-review-notes" class="btn btn-primary btn-sm">Copy notes</button>
          <button type="button" id="clear-review-notes" class="btn btn-outline btn-error btn-sm">Clear notes</button>
        </div>
        <textarea id="review-notes-output" class="textarea textarea-bordered mt-3 min-h-96 w-full font-mono text-xs" readonly>No marks yet.</textarea>
      </aside>
    </div>
  </main>

  <script>
${labClientScript()}
  </script>
</body>
</html>
`;
}

export function collectColorUsageInventory({
  roots = [resolve(appRoot, "internal/views"), resolve(appRoot, "static/css/input.css")],
} = {}) {
  const files = [];
  for (const root of roots) {
    collectFiles(root, files);
  }

  const pattern = /\b(?:text|bg|border|btn|badge|alert)-(?:base|primary|secondary|accent|neutral|info|success|warning|error)(?:-[A-Za-z0-9/]+)?\b/g;
  const patterns = new Map();
  let occurrenceCount = 0;

  for (const file of files) {
    const content = readFileSync(file, "utf8");
    for (const match of content.matchAll(pattern)) {
      occurrenceCount += 1;
      patterns.set(match[0], (patterns.get(match[0]) ?? 0) + 1);
    }
  }

  return {
    fileCount: files.length,
    occurrenceCount,
    patterns: [...patterns.entries()].sort((a, b) => a[0].localeCompare(b[0])).map(([name, count]) => `${name} (${count})`),
  };
}

export function writeThemeReadabilityLab(outputPath = defaultOutputPath) {
  mkdirSync(dirname(outputPath), { recursive: true });
  const html = buildThemeReadabilityLab();
  writeFileSync(outputPath, html);
  return outputPath;
}

function renderThemeOptions() {
  const groups = new Map();
  for (const theme of selectableThemes) {
    if (!groups.has(theme.family)) {
      groups.set(theme.family, []);
    }
    groups.get(theme.family).push(theme);
  }

  return [...groups.entries()]
    .map(([family, themes]) => {
      const options = themes
        .map((theme) => `<option value="${escapeAttribute(theme.value)}"${theme.value === "treacherest" ? " selected" : ""}>${escapeHTML(theme.label)}</option>`)
        .join("");
      return `<optgroup label="${escapeAttribute(family)}">${options}</optgroup>`;
    })
    .join("");
}

function renderSample(sample) {
  const groups = sample.groups.join(" ");
  const targetRows = sample.targets.map((sampleTarget) => renderTargetRow(sampleTarget)).join("\n");
  return `
    <section class="lab-sample rounded-box border border-base-300 bg-base-100 p-4 shadow-sm" data-sample-id="${escapeAttribute(sample.id)}" data-groups="${escapeAttribute(groups)}">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="font-mono text-[10px] font-bold uppercase tracking-[0.16em] text-base-content/60">${escapeHTML(sample.id)}</p>
          <h2 class="text-xl font-bold">${escapeHTML(sample.title)}</h2>
          <p class="mt-1 max-w-2xl text-sm text-base-content/70">${escapeHTML(sample.description)}</p>
        </div>
        <div class="flex flex-wrap gap-1">
          ${sample.groups.map((group) => `<span class="badge badge-outline">${escapeHTML(group)}</span>`).join("")}
        </div>
      </div>
      <div class="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,24rem)]">
        <div class="rounded-box border border-base-300 bg-base-200 p-4">
          ${sample.markup}
        </div>
        <div class="space-y-2">
          ${targetRows}
        </div>
      </div>
    </section>
  `;
}

function renderTargetRow(sampleTarget) {
  return `
    <div class="target-row rounded-box border border-base-300 bg-base-100 p-3 text-sm" data-target-row="${escapeAttribute(sampleTarget.id)}" data-status="manual">
      <div class="flex flex-wrap items-start justify-between gap-2">
        <div class="min-w-0">
          <button type="button" class="font-mono text-left text-xs font-bold text-primary" data-copy-target="${escapeAttribute(sampleTarget.id)}">${escapeHTML(sampleTarget.id)}</button>
          <p class="mt-1 font-semibold">${escapeHTML(sampleTarget.label)}</p>
          <p class="mt-1 text-xs text-base-content/70">${escapeHTML(sampleTarget.guess)}</p>
        </div>
        <span class="badge badge-info badge-sm" data-target-status="${escapeAttribute(sampleTarget.id)}">Manual</span>
      </div>
      <dl class="mt-2 grid grid-cols-2 gap-x-3 gap-y-1 text-xs">
        <dt class="text-base-content/60">Contrast</dt>
        <dd data-target-ratio="${escapeAttribute(sampleTarget.id)}">not measured</dd>
        <dt class="text-base-content/60">Text</dt>
        <dd><span class="lab-swatch" data-target-fg="${escapeAttribute(sampleTarget.id)}"></span> <code data-target-fg-value="${escapeAttribute(sampleTarget.id)}">n/a</code></dd>
        <dt class="text-base-content/60">Background</dt>
        <dd><span class="lab-swatch" data-target-bg="${escapeAttribute(sampleTarget.id)}"></span> <code data-target-bg-value="${escapeAttribute(sampleTarget.id)}">n/a</code></dd>
      </dl>
      <div class="mt-3 grid grid-cols-3 gap-2">
        <button type="button" class="btn btn-success btn-xs" data-review-action="ok" data-review-id="${escapeAttribute(sampleTarget.id)}">OK</button>
        <button type="button" class="btn btn-error btn-xs" data-review-action="bad" data-review-id="${escapeAttribute(sampleTarget.id)}">Bad</button>
        <button type="button" class="btn btn-ghost btn-xs" data-review-action="clear" data-review-id="${escapeAttribute(sampleTarget.id)}">Clear</button>
      </div>
    </div>
  `;
}

function renderFilterButton(filter, label) {
  return `<button type="button" class="btn btn-sm ${filter === "all" ? "btn-primary" : "btn-outline"}" data-filter="${escapeAttribute(filter)}">${escapeHTML(label)}</button>`;
}

function target(id, label, guess) {
  return { id, label, guess };
}

function collectFiles(path, files) {
  const info = statSync(path);
  if (info.isFile()) {
    if (path.endsWith(".templ") || path.endsWith("input.css")) {
      files.push(path);
    }
    return;
  }

  for (const entry of readdirSync(path)) {
    collectFiles(join(path, entry), files);
  }
}

function labClientScript() {
  return String.raw`
const marks = new Map();
let activeFilter = "all";

const themeSelect = document.getElementById("theme-select");
const notesOutput = document.getElementById("review-notes-output");

for (const target of document.querySelectorAll("[data-review-target]")) {
  target.classList.add("lab-target");
  target.addEventListener("click", (event) => {
    event.stopPropagation();
    toggleMark(target.dataset.reviewTarget, "bad");
  });
}

for (const button of document.querySelectorAll("[data-review-action]")) {
  button.addEventListener("click", () => {
    const action = button.dataset.reviewAction;
    const id = button.dataset.reviewId;
    if (action === "clear") {
      setMark(id, null);
    } else {
      setMark(id, action);
    }
  });
}

for (const button of document.querySelectorAll("[data-copy-target]")) {
  button.addEventListener("click", async () => {
    await copyText(button.dataset.copyTarget);
  });
}

for (const button of document.querySelectorAll("[data-filter]")) {
  button.addEventListener("click", () => {
    activeFilter = button.dataset.filter;
    for (const filterButton of document.querySelectorAll("[data-filter]")) {
      filterButton.classList.toggle("btn-primary", filterButton === button);
      filterButton.classList.toggle("btn-outline", filterButton !== button);
    }
    applyFilters();
  });
}

themeSelect.addEventListener("change", () => {
  document.documentElement.setAttribute("data-theme", themeSelect.value);
  refreshContrast();
  refreshReviewState();
});

document.getElementById("copy-review-notes").addEventListener("click", async () => {
  await copyText(notesOutput.value);
});

document.getElementById("clear-review-notes").addEventListener("click", () => {
  marks.clear();
  refreshReviewState();
});

requestAnimationFrame(() => {
  refreshContrast();
  refreshReviewState();
});

function setMark(id, state) {
  const key = markKey(themeSelect.value, id);
  if (!state) {
    marks.delete(key);
  } else {
    marks.set(key, { theme: themeSelect.value, id, state });
  }
  refreshReviewState();
}

function toggleMark(id, state) {
  const key = markKey(themeSelect.value, id);
  const existing = marks.get(key);
  setMark(id, existing?.state === state ? null : state);
}

function markKey(theme, id) {
  return theme + "::" + id;
}

function refreshReviewState() {
  const theme = themeSelect.value;
  for (const target of document.querySelectorAll("[data-review-target]")) {
    const mark = marks.get(markKey(theme, target.dataset.reviewTarget));
    target.dataset.reviewState = mark?.state || "";
  }
  for (const row of document.querySelectorAll("[data-target-row]")) {
    const mark = marks.get(markKey(theme, row.dataset.targetRow));
    row.dataset.reviewState = mark?.state || "";
  }
  renderNotes();
  applyFilters();
}

function renderNotes() {
  if (marks.size === 0) {
    notesOutput.value = "No marks yet.";
    return;
  }

  const sorted = [...marks.values()].sort((a, b) =>
    a.state.localeCompare(b.state) || a.theme.localeCompare(b.theme) || a.id.localeCompare(b.id)
  );
  const lines = [];
  for (const state of ["bad", "ok"]) {
    const entries = sorted.filter((mark) => mark.state === state);
    if (entries.length === 0) {
      continue;
    }
    lines.push(state + ":");
    for (const entry of entries) {
      lines.push(entry.theme + " :: " + entry.id);
    }
    lines.push("");
  }
  notesOutput.value = lines.join("\n").trim();
}

function refreshContrast() {
  const sampleFailures = new Map();
  const sampleManuals = new Map();

  for (const target of document.querySelectorAll("[data-review-target]")) {
    const id = target.dataset.reviewTarget;
    const foreground = parseColor(getComputedStyle(target).color);
    const background = effectiveBackgroundColor(target);
    const ratio = foreground && background ? contrastRatio(foreground, background) : null;
    const status = contrastStatus(target, ratio);
    const sample = target.closest(".lab-sample");

    updateTargetResult(id, foreground, background, ratio, status);

    if (sample) {
      if (status === "fail") sampleFailures.set(sample.dataset.sampleId, true);
      if (status === "manual") sampleManuals.set(sample.dataset.sampleId, true);
    }
  }

  for (const sample of document.querySelectorAll(".lab-sample")) {
    sample.dataset.hasFail = sampleFailures.has(sample.dataset.sampleId) ? "true" : "false";
    sample.dataset.hasManual = sampleManuals.has(sample.dataset.sampleId) ? "true" : "false";
  }
  applyFilters();
}

function updateTargetResult(id, foreground, background, ratio, status) {
  const row = document.querySelector('[data-target-row="' + cssEscape(id) + '"]');
  const statusBadge = document.querySelector('[data-target-status="' + cssEscape(id) + '"]');
  const ratioNode = document.querySelector('[data-target-ratio="' + cssEscape(id) + '"]');
  const fgSwatch = document.querySelector('[data-target-fg="' + cssEscape(id) + '"]');
  const bgSwatch = document.querySelector('[data-target-bg="' + cssEscape(id) + '"]');
  const fgValue = document.querySelector('[data-target-fg-value="' + cssEscape(id) + '"]');
  const bgValue = document.querySelector('[data-target-bg-value="' + cssEscape(id) + '"]');

  if (row) row.dataset.status = status;
  if (statusBadge) {
    statusBadge.textContent = statusLabel(status);
    statusBadge.className = "badge badge-sm " + statusBadgeClass(status);
  }
  if (ratioNode) ratioNode.textContent = ratio ? ratio.toFixed(2) + ":1" : "manual";
  if (fgSwatch && foreground) fgSwatch.style.backgroundColor = rgbString(foreground);
  if (bgSwatch && background) bgSwatch.style.backgroundColor = rgbString(background);
  if (fgValue && foreground) fgValue.textContent = rgbString(foreground);
  if (bgValue && background) bgValue.textContent = rgbString(background);
}

function contrastStatus(target, ratio) {
  if (!ratio) return "manual";
  if (ratio >= 4.5) return "aa";
  if (ratio >= 3 && isLargeText(target)) return "large";
  return "fail";
}

function isLargeText(element) {
  const style = getComputedStyle(element);
  const size = Number.parseFloat(style.fontSize);
  const weight = Number.parseInt(style.fontWeight, 10) || 400;
  return size >= 24 || (size >= 18.66 && weight >= 700);
}

function statusLabel(status) {
  if (status === "aa") return "AA";
  if (status === "large") return "Large only";
  if (status === "fail") return "Fail";
  return "Manual";
}

function statusBadgeClass(status) {
  if (status === "aa") return "badge-success";
  if (status === "large") return "badge-warning";
  if (status === "fail") return "badge-error";
  return "badge-info";
}

function effectiveBackgroundColor(element) {
  const explicit = element.dataset.contrastBg ? document.querySelector(element.dataset.contrastBg) : null;
  const colors = [];
  let current = explicit || element;

  while (current && current.nodeType === Node.ELEMENT_NODE) {
    const color = parseColor(getComputedStyle(current).backgroundColor);
    if (color && color.a > 0) {
      colors.push(color);
      if (color.a >= 0.999) break;
    }
    current = current.parentElement;
  }

  if (colors.length === 0) {
    colors.push({ r: 255, g: 255, b: 255, a: 1 });
  }

  return colors.reverse().reduce((backdrop, color) => blend(color, backdrop));
}

function parseColor(value) {
  if (!value || value === "transparent") return null;

  const rgb = value.match(/rgba?\(\s*([0-9.]+)\s*,?\s+([0-9.]+)\s*,?\s+([0-9.]+)(?:\s*[\/,]\s*([0-9.]+%?))?\s*\)/i);
  if (rgb) {
    return {
      r: Number(rgb[1]),
      g: Number(rgb[2]),
      b: Number(rgb[3]),
      a: parseAlpha(rgb[4]),
    };
  }

  const oklch = value.match(/oklch\(\s*([+-]?\d*\.?\d+)%?\s+([+-]?\d*\.?\d+)\s+([+-]?\d*\.?\d+)(?:deg)?(?:\s*\/\s*([0-9.]+%?))?\s*\)/i);
  if (oklch) {
    const converted = oklchToSRGB(Number(oklch[1]), Number(oklch[2]), Number(oklch[3]));
    return { ...converted, a: parseAlpha(oklch[4]) };
  }

  return null;
}

function parseAlpha(value) {
  if (!value) return 1;
  if (value.endsWith("%")) return Number(value.slice(0, -1)) / 100;
  return Number(value);
}

function oklchToSRGB(lightnessInput, chroma, hueDegrees) {
  const lightness = lightnessInput > 1 ? lightnessInput / 100 : lightnessInput;
  const hue = hueDegrees * (Math.PI / 180);
  const a = chroma * Math.cos(hue);
  const b = chroma * Math.sin(hue);
  const lPrime = lightness + 0.3963377774 * a + 0.2158037573 * b;
  const mPrime = lightness - 0.1055613458 * a - 0.0638541728 * b;
  const sPrime = lightness - 0.0894841775 * a - 1.2914855480 * b;
  const l = lPrime ** 3;
  const m = mPrime ** 3;
  const s = sPrime ** 3;
  return {
    r: Math.round(clamp01(+4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s) * 255),
    g: Math.round(clamp01(-1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s) * 255),
    b: Math.round(clamp01(-0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s) * 255),
  };
}

function blend(foreground, background) {
  const alpha = foreground.a + background.a * (1 - foreground.a);
  if (alpha <= 0) return { r: 255, g: 255, b: 255, a: 1 };
  return {
    r: (foreground.r * foreground.a + background.r * background.a * (1 - foreground.a)) / alpha,
    g: (foreground.g * foreground.a + background.g * background.a * (1 - foreground.a)) / alpha,
    b: (foreground.b * foreground.a + background.b * background.a * (1 - foreground.a)) / alpha,
    a: alpha,
  };
}

function contrastRatio(foreground, background) {
  const fg = relativeLuminance(foreground);
  const bg = relativeLuminance(background);
  const lighter = Math.max(fg, bg);
  const darker = Math.min(fg, bg);
  return (lighter + 0.05) / (darker + 0.05);
}

function relativeLuminance(color) {
  return 0.2126 * channelLuminance(color.r) + 0.7152 * channelLuminance(color.g) + 0.0722 * channelLuminance(color.b);
}

function channelLuminance(value) {
  const normalized = value / 255;
  return normalized <= 0.03928 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
}

function rgbString(color) {
  return "rgb(" + Math.round(color.r) + " " + Math.round(color.g) + " " + Math.round(color.b) + ")";
}

function clamp01(value) {
  return Math.min(1, Math.max(0, value));
}

function applyFilters() {
  for (const sample of document.querySelectorAll(".lab-sample")) {
    const groups = (sample.dataset.groups || "").split(/\s+/);
    let visible = true;
    if (activeFilter === "fail") visible = sample.dataset.hasFail === "true";
    else if (activeFilter === "manual") visible = sample.dataset.hasManual === "true";
    else if (activeFilter !== "all") visible = groups.includes(activeFilter);
    sample.hidden = !visible;
  }
}

async function copyText(value) {
  try {
    await navigator.clipboard.writeText(value);
  } catch {
    const textArea = document.createElement("textarea");
    textArea.value = value;
    document.body.append(textArea);
    textArea.select();
    document.execCommand("copy");
    textArea.remove();
  }
}

function cssEscape(value) {
  if (window.CSS && typeof window.CSS.escape === "function") {
    return window.CSS.escape(value);
  }
  return value.replace(/"/g, '\\"');
}
`;
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function escapeAttribute(value) {
  return escapeHTML(value).replaceAll("'", "&#39;");
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const outputPath = writeThemeReadabilityLab();
  const relativePath = relative(repoRoot, outputPath);
  console.log(`Wrote ${relativePath}`);
}
