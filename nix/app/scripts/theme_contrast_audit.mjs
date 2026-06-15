#!/usr/bin/env node
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { selectableThemes } from "./theme_data.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const args = new Set(process.argv.slice(2));
const cssPath = resolve(__dirname, "../static/css/output.css");
const css = readFileSync(cssPath, "utf8");

const themeOrder = selectableThemes.map((theme) => theme.value);

const tokenPairs = [
  ["base-content", "base-100", 4.5, "page text"],
  ["base-content", "base-200", 4.5, "raised surface text"],
  ["base-content", "base-300", 4.5, "muted surface text"],
  ["primary-content", "primary", 4.5, "primary component text"],
  ["secondary-content", "secondary", 4.5, "secondary component text"],
  ["accent-content", "accent", 4.5, "accent component text"],
  ["neutral-content", "neutral", 4.5, "neutral component text"],
  ["info-content", "info", 4.5, "info component text"],
  ["success-content", "success", 4.5, "success component text"],
  ["warning-content", "warning", 4.5, "warning component text"],
  ["error-content", "error", 4.5, "error component text"],
];

const themeBlocks = parseThemeBlocks(css);
const results = [];
const missing = [];

for (const theme of themeOrder) {
  const tokens = themeBlocks.get(theme);
  if (!tokens) {
    missing.push(theme);
    continue;
  }

  for (const [foreground, background, minimum, label] of tokenPairs) {
    const foregroundColor = parseOKLCH(tokens.get(foreground));
    const backgroundColor = parseOKLCH(tokens.get(background));
    if (!foregroundColor || !backgroundColor) {
      missing.push(`${theme}:${foreground}/${background}`);
      continue;
    }

    results.push({
      theme,
      foreground,
      background,
      label,
      minimum,
      ratio: contrastRatio(foregroundColor, backgroundColor),
    });
  }
}

const failures = results.filter((result) => result.ratio < result.minimum);
const worst = [...results].sort((a, b) => a.ratio - b.ratio).slice(0, 20);

console.log(`Theme contrast audit (${cssPath})`);
console.log(`${themeOrder.length} selectable themes, ${results.length} checked token pairs`);

if (missing.length > 0) {
  console.log("");
  console.log("Missing theme blocks or tokens:");
  for (const item of missing) {
    console.log(`- ${item}`);
  }
}

console.log("");
console.log(`Pairs below WCAG AA normal-text target (${failures.length}):`);
if (failures.length === 0) {
  console.log("- none");
} else {
  for (const result of failures) {
    console.log(formatResult(result));
  }
}

console.log("");
console.log("Lowest contrast pairs:");
for (const result of worst) {
  console.log(formatResult(result));
}

if (args.has("--fail-on-aa") && (failures.length > 0 || missing.length > 0)) {
  process.exitCode = 1;
}

function parseThemeBlocks(source) {
  const blocks = new Map();
  const blockPattern = /([^{}]+)\{([^{}]*--color-[^{}]+)\}/g;
  let blockMatch;

  while ((blockMatch = blockPattern.exec(source)) !== null) {
    const selector = blockMatch[1];
    const body = blockMatch[2];
    const themeNames = [...selector.matchAll(/\[data-theme="?([a-z0-9-]+)"?\]/g)].map((match) => match[1]);

    if (themeNames.length === 0) {
      continue;
    }

    const tokens = new Map();
    for (const tokenMatch of body.matchAll(/--color-([a-z0-9-]+):\s*([^;]+);/g)) {
      tokens.set(tokenMatch[1], tokenMatch[2].trim());
    }

    if (!tokens.has("base-100") || !tokens.has("base-content")) {
      continue;
    }

    for (const themeName of themeNames) {
      blocks.set(themeName, tokens);
    }
  }

  return blocks;
}

function parseOKLCH(value) {
  if (!value) {
    return null;
  }

  const match = value.match(/oklch\(\s*([+-]?\d*\.?\d+)%?\s+([+-]?\d*\.?\d+)\s+([+-]?\d*\.?\d+)(?:deg)?(?:\s*\/\s*[^)]+)?\s*\)/i);
  if (!match) {
    return null;
  }

  const lightness = Number(match[1]) > 1 ? Number(match[1]) / 100 : Number(match[1]);
  const chroma = Number(match[2]);
  const hue = Number(match[3]) * (Math.PI / 180);
  const a = chroma * Math.cos(hue);
  const b = chroma * Math.sin(hue);

  const lPrime = lightness + 0.3963377774 * a + 0.2158037573 * b;
  const mPrime = lightness - 0.1055613458 * a - 0.0638541728 * b;
  const sPrime = lightness - 0.0894841775 * a - 1.2914855480 * b;

  const l = lPrime ** 3;
  const m = mPrime ** 3;
  const s = sPrime ** 3;

  return {
    r: clamp01(+4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s),
    g: clamp01(-1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s),
    b: clamp01(-0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s),
  };
}

function contrastRatio(foreground, background) {
  const foregroundLuminance = relativeLuminance(foreground);
  const backgroundLuminance = relativeLuminance(background);
  const lighter = Math.max(foregroundLuminance, backgroundLuminance);
  const darker = Math.min(foregroundLuminance, backgroundLuminance);
  return (lighter + 0.05) / (darker + 0.05);
}

function relativeLuminance(color) {
  return 0.2126 * color.r + 0.7152 * color.g + 0.0722 * color.b;
}

function clamp01(value) {
  return Math.min(1, Math.max(0, value));
}

function formatResult(result) {
  const ratio = result.ratio.toFixed(2).padStart(5, " ");
  const pair = `${result.foreground} on ${result.background}`.padEnd(36, " ");
  return `- ${result.theme.padEnd(16, " ")} ${pair} ${ratio}:1  ${result.label}`;
}
