#!/usr/bin/env node
// Canonical headless-Chromium renderer for SVG infographics.
//
// Runs on the Node 18+ standard library alone: no package install and no
// bundled browser. It locates a Chromium-family executable already present on
// the machine, discloses exactly which one it used and that browser's version,
// renders the SVG at an integer multiple of its viewBox, then verifies the
// written PNG by reading the dimensions out of the file's own IHDR header.
//
// Usage:
//   node render.mjs <file.svg> [options]
//
// Options:
//   --out <file.png>   output path (default: the input path with a .png suffix)
//   --scale <n>        integer scale factor applied to the viewBox (default 2)
//   --browser <path>   explicit browser executable, skipping discovery
//   --transparent      keep the page background transparent (default is white)
//   --no-sandbox       pass --no-sandbox; needed inside some containers
//   --timeout <ms>     virtual time budget handed to the browser (default 15000)
//   --json             emit machine-readable JSON instead of text
//   --help             print this usage block
//
// Exit codes:
//   0  PNG written and its IHDR dimensions match the requested target
//   1  render failed, or the written PNG did not match the target
//   2  no headless browser could be found  (the graceful-degradation signal:
//      deliver the editable SVG alone and state that no PNG was produced)
//   3  usage error, or the input could not be read

import { readFileSync, writeFileSync, existsSync, mkdtempSync, rmSync, openSync, readSync, closeSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { tmpdir, platform } from 'node:os';
import { join, resolve, dirname, basename, extname } from 'node:path';
import { pathToFileURL } from 'node:url';

// ---------------------------------------------------------------------------
// Argument handling
// ---------------------------------------------------------------------------

const USAGE = [
  'Usage: node render.mjs <file.svg> [--out <file.png>] [--scale <n>] [--browser <path>]',
  '                      [--transparent] [--no-sandbox] [--timeout <ms>] [--json]',
  '',
  'Renders an SVG to PNG through a headless Chromium-family browser and verifies',
  'the result against the requested target size.',
  'Exit 0 = verified, 1 = render/verify failure, 2 = no browser found, 3 = usage error.',
].join('\n');

function parseArgs(argv) {
  const opts = {
    input: null,
    out: null,
    scale: 2,
    browser: null,
    transparent: false,
    noSandbox: false,
    timeout: 15000,
    json: false,
    help: false,
  };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--help' || a === '-h') opts.help = true;
    else if (a === '--json') opts.json = true;
    else if (a === '--transparent') opts.transparent = true;
    else if (a === '--no-sandbox') opts.noSandbox = true;
    else if (a === '--out') opts.out = argv[++i];
    else if (a === '--browser') opts.browser = argv[++i];
    else if (a === '--scale') opts.scale = Number(argv[++i]);
    else if (a === '--timeout') opts.timeout = Number(argv[++i]);
    else if (a.startsWith('-')) return { error: `unknown option: ${a}` };
    else if (opts.input === null) opts.input = a;
    else return { error: `unexpected extra argument: ${a}` };
  }
  if (opts.help) return opts;
  if (opts.input === null) return { error: 'missing <file.svg>' };
  if (!Number.isInteger(opts.scale) || opts.scale < 1) return { error: '--scale expects a positive integer' };
  if (!Number.isFinite(opts.timeout) || opts.timeout < 1) return { error: '--timeout expects a positive number of milliseconds' };
  return opts;
}

// ---------------------------------------------------------------------------
// Browser discovery and version disclosure
// ---------------------------------------------------------------------------

const CANDIDATE_PATHS = {
  darwin: [
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    '/Applications/Chromium.app/Contents/MacOS/Chromium',
    '/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge',
    '/Applications/Brave Browser.app/Contents/MacOS/Brave Browser',
  ],
  linux: [
    '/usr/bin/google-chrome',
    '/usr/bin/google-chrome-stable',
    '/usr/bin/chromium',
    '/usr/bin/chromium-browser',
    '/usr/bin/microsoft-edge',
    '/snap/bin/chromium',
  ],
  win32: [
    'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe',
  ],
};

const PATH_NAMES = ['google-chrome', 'google-chrome-stable', 'chromium', 'chromium-browser', 'microsoft-edge', 'msedge'];

function browserVersion(exe) {
  const probe = spawnSync(exe, ['--version'], { encoding: 'utf8', timeout: 10000 });
  if (probe.status !== 0 || !probe.stdout) return null;
  return probe.stdout.trim();
}

function discoverBrowser(explicit) {
  if (explicit) {
    if (!existsSync(explicit)) return { error: `--browser path does not exist: ${explicit}` };
    const version = browserVersion(explicit);
    if (version === null) return { error: `--browser path is not an executable browser: ${explicit}` };
    return { executable: explicit, version, source: 'explicit --browser flag' };
  }

  for (const key of ['CHROME_PATH', 'CHROMIUM_PATH', 'BROWSER_PATH']) {
    const fromEnv = process.env[key];
    if (fromEnv && existsSync(fromEnv)) {
      const version = browserVersion(fromEnv);
      if (version !== null) return { executable: fromEnv, version, source: `environment variable ${key}` };
    }
  }

  for (const candidate of CANDIDATE_PATHS[platform()] ?? []) {
    if (!existsSync(candidate)) continue;
    const version = browserVersion(candidate);
    if (version !== null) return { executable: candidate, version, source: 'well-known install location' };
  }

  const lookup = platform() === 'win32' ? 'where' : 'which';
  for (const name of PATH_NAMES) {
    const found = spawnSync(lookup, [name], { encoding: 'utf8', timeout: 10000 });
    if (found.status !== 0 || !found.stdout) continue;
    const exe = found.stdout.split(/\r?\n/)[0].trim();
    if (exe === '' || !existsSync(exe)) continue;
    const version = browserVersion(exe);
    if (version !== null) return { executable: exe, version, source: 'PATH lookup' };
  }

  return { notFound: true };
}

// ---------------------------------------------------------------------------
// SVG intrinsic size
// ---------------------------------------------------------------------------

function intrinsicSize(source) {
  const viewBoxMatch = /<svg\b[^>]*\bviewBox\s*=\s*["']([^"']+)["']/i.exec(source);
  if (viewBoxMatch) {
    const parts = viewBoxMatch[1].trim().split(/[\s,]+/).map(Number);
    if (parts.length === 4 && parts.every(Number.isFinite) && parts[2] > 0 && parts[3] > 0) {
      return { width: parts[2], height: parts[3], from: 'viewBox' };
    }
  }
  const widthMatch = /<svg\b[^>]*\bwidth\s*=\s*["']([0-9.]+)/i.exec(source);
  const heightMatch = /<svg\b[^>]*\bheight\s*=\s*["']([0-9.]+)/i.exec(source);
  if (widthMatch && heightMatch) {
    const w = Number(widthMatch[1]);
    const h = Number(heightMatch[1]);
    if (w > 0 && h > 0) return { width: w, height: h, from: 'width/height attributes' };
  }
  return null;
}

// The SVG is inlined into a wrapper document and stretched by CSS, so the
// viewBox drives the scaling regardless of any width/height on the root element.
function wrapperDocument(svgSource, targetWidth, targetHeight, transparent) {
  const body = svgSource
    .replace(/<\?xml[^>]*\?>/gi, '')
    .replace(/<!DOCTYPE[^>]*>/gi, '')
    .trim();
  const background = transparent ? 'transparent' : '#ffffff';
  return [
    '<!DOCTYPE html>',
    '<html><head><meta charset="utf-8"><style>',
    `html,body{margin:0;padding:0;background:${background};}`,
    `#frame{width:${targetWidth}px;height:${targetHeight}px;overflow:hidden;}`,
    '#frame > svg{width:100%;height:100%;display:block;}',
    '</style></head><body>',
    `<div id="frame">${body}</div>`,
    '</body></html>',
  ].join('');
}

// ---------------------------------------------------------------------------
// PNG header verification
// ---------------------------------------------------------------------------

const PNG_SIGNATURE = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);

function readIHDR(pngPath) {
  const header = Buffer.alloc(24);
  const fd = openSync(pngPath, 'r');
  try {
    const read = readSync(fd, header, 0, 24, 0);
    if (read < 24) return { error: 'file is shorter than a PNG header' };
  } finally {
    closeSync(fd);
  }
  if (!header.subarray(0, 8).equals(PNG_SIGNATURE)) return { error: 'file does not carry the PNG signature' };
  if (header.subarray(12, 16).toString('ascii') !== 'IHDR') return { error: 'first PNG chunk is not IHDR' };
  return { width: header.readUInt32BE(16), height: header.readUInt32BE(20) };
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

function runBrowser(exe, args) {
  return spawnSync(exe, args, { encoding: 'utf8', timeout: 120000 });
}

function render(opts, browser, targetWidth, targetHeight, wrapperPath, outPath) {
  const baseArgs = [
    '--disable-gpu',
    '--hide-scrollbars',
    '--disable-extensions',
    '--disable-lcd-text',
    `--screenshot=${outPath}`,
    `--window-size=${targetWidth},${targetHeight}`,
    `--virtual-time-budget=${opts.timeout}`,
  ];
  if (opts.transparent) baseArgs.push('--default-background-color=00000000');
  if (opts.noSandbox) baseArgs.push('--no-sandbox');

  const profileDir = mkdtempSync(join(tmpdir(), 'svg-render-profile-'));
  baseArgs.push(`--user-data-dir=${profileDir}`);
  const target = pathToFileURL(wrapperPath).href;

  try {
    // Newer builds require the explicit "new" headless mode; older builds only
    // accept the bare flag. Try the modern form first, then fall back.
    for (const headlessFlag of ['--headless=new', '--headless']) {
      const result = runBrowser(browser.executable, [headlessFlag, ...baseArgs, target]);
      if (result.status === 0 && existsSync(outPath)) return { headlessFlag };
      if (result.error && result.error.code === 'ETIMEDOUT') {
        return { error: 'the browser did not exit within the render timeout' };
      }
    }
    return { error: 'the browser exited without writing a screenshot' };
  } finally {
    rmSync(profileDir, { recursive: true, force: true });
  }
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

function fail(json, code, payload, text) {
  if (json) process.stdout.write(`${JSON.stringify({ status: 'failed', ...payload }, null, 2)}\n`);
  else process.stderr.write(`render: ${text}\n`);
  process.exit(code);
}

function main() {
  const opts = parseArgs(process.argv.slice(2));
  if (opts.error) {
    process.stderr.write(`render: ${opts.error}\n\n${USAGE}\n`);
    process.exit(3);
  }
  if (opts.help) {
    process.stdout.write(`${USAGE}\n`);
    process.exit(0);
  }

  const inputPath = resolve(opts.input);
  let source;
  try {
    source = readFileSync(inputPath, 'utf8');
  } catch (err) {
    fail(opts.json, 3, { reason: 'unreadable-input', detail: err.message }, `cannot read ${opts.input}: ${err.message}`);
  }

  const intrinsic = intrinsicSize(source);
  if (intrinsic === null) {
    fail(
      opts.json,
      3,
      { reason: 'no-intrinsic-size' },
      'the SVG declares neither a usable viewBox nor numeric width/height attributes',
    );
  }

  const targetWidth = Math.round(intrinsic.width * opts.scale);
  const targetHeight = Math.round(intrinsic.height * opts.scale);

  const browser = discoverBrowser(opts.browser);
  if (browser.error) {
    fail(opts.json, 3, { reason: 'bad-browser-argument', detail: browser.error }, browser.error);
  }
  if (browser.notFound) {
    const message =
      'no headless Chromium-family browser was found; deliver the editable SVG alone and state that no PNG was produced';
    if (opts.json) {
      process.stdout.write(`${JSON.stringify({ status: 'no-browser', reason: message }, null, 2)}\n`);
    } else {
      process.stderr.write(`render: ${message}\n`);
    }
    process.exit(2);
  }

  const outPath = resolve(
    opts.out ?? join(dirname(inputPath), `${basename(inputPath, extname(inputPath))}.png`),
  );

  const workDir = mkdtempSync(join(tmpdir(), 'svg-render-'));
  const wrapperPath = join(workDir, 'wrapper.html');
  let outcome;
  try {
    writeFileSync(wrapperPath, wrapperDocument(source, targetWidth, targetHeight, opts.transparent), 'utf8');
    outcome = render(opts, browser, targetWidth, targetHeight, wrapperPath, outPath);
  } finally {
    rmSync(workDir, { recursive: true, force: true });
  }

  if (outcome.error) {
    fail(
      opts.json,
      1,
      { reason: 'render-failed', detail: outcome.error, browser: browser.executable, version: browser.version },
      outcome.error,
    );
  }

  const ihdr = readIHDR(outPath);
  if (ihdr.error) {
    fail(
      opts.json,
      1,
      { reason: 'unverifiable-png', detail: ihdr.error, output: outPath },
      `wrote ${outPath} but could not verify it: ${ihdr.error}`,
    );
  }

  const verified = ihdr.width === targetWidth && ihdr.height === targetHeight;
  const record = {
    status: verified ? 'verified' : 'mismatch',
    input: inputPath,
    output: outPath,
    browser: browser.executable,
    browserVersion: browser.version,
    browserSource: browser.source,
    headlessMode: outcome.headlessFlag,
    intrinsic: { width: intrinsic.width, height: intrinsic.height, from: intrinsic.from },
    scale: opts.scale,
    target: { width: targetWidth, height: targetHeight },
    ihdr: { width: ihdr.width, height: ihdr.height },
  };

  if (opts.json) {
    process.stdout.write(`${JSON.stringify(record, null, 2)}\n`);
  } else {
    process.stdout.write(
      [
        `browser    ${browser.executable}`,
        `version    ${browser.version}   (found via ${browser.source}, ${outcome.headlessFlag})`,
        `source     ${inputPath}  ${intrinsic.width}x${intrinsic.height} from ${intrinsic.from}`,
        `target     ${targetWidth}x${targetHeight}  (scale ${opts.scale}x)`,
        `png IHDR   ${ihdr.width}x${ihdr.height}`,
        `output     ${outPath}`,
        verified ? 'verified   dimensions match the target' : 'MISMATCH   the PNG does not match the requested target',
        '',
      ].join('\n'),
    );
  }

  process.exit(verified ? 0 : 1);
}

main();
