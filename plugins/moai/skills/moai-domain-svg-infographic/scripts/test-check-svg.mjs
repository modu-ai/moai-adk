#!/usr/bin/env node
// Bipolar self-test for check-svg.mjs.
//
// Runs on the Node 18+ standard library alone: no package install, no network
// access, no browser. It spawns the real check-svg.mjs as a child process for
// every fixture and re-implements no SVG parsing of its own — the linter under
// test is the only measurement path.
//
// Each fixture declares its expectation in-file as a leading comment naming the
// exact diagnostic code set it must produce:
//
//   <!-- expect: SVG070 -->        one code
//   <!-- expect: SVG070 SVG073 --> several codes
//   <!-- expect: -->               the empty set (the only accepted spelling)
//
// The comparison is against the exact set of `code` values parsed from the
// linter's --json output. The child process exit code is never consulted: a
// fixture that emits the wrong codes can still exit 0, and a fixture that emits
// the right codes can still exit 1 under --strict, so the exit code decides
// nothing here.
//
// Usage:
//   node test-check-svg.mjs
//
// Exit codes:
//   0  every fixture's emitted code set matches its declaration
//   1  at least one fixture mismatched, or declared no expectation at all

import { readFileSync, readdirSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const checker = join(scriptDir, 'check-svg.mjs');
const fixtureDir = join(scriptDir, 'fixtures');

// The empty set has exactly one accepted spelling: `<!-- expect: -->`, with
// nothing between the colon and the closing marker. `{}` and `none` are
// deliberately NOT accepted — a placeholder token would drop the fixture out of
// the clean-fixture selection the no-false-positive check is driven off.
const EXPECT_RE = /<!--\s*expect:([^]*?)-->/;

function declaredCodes(source) {
  const m = EXPECT_RE.exec(source);
  if (m === null) return null; // no declaration at all
  const body = m[1].trim();
  if (body === '') return [];
  return body.split(/[\s,]+/).filter(Boolean);
}

function emittedCodes(file) {
  const run = spawnSync(process.execPath, [checker, file, '--json'], { encoding: 'utf8' });
  let parsed;
  try {
    parsed = JSON.parse(run.stdout);
  } catch {
    const detail = (run.stderr || run.stdout || '').trim();
    return { error: `check-svg.mjs produced no parseable JSON: ${detail}` };
  }
  const diagnostics = Array.isArray(parsed.diagnostics) ? parsed.diagnostics : [];
  return { codes: diagnostics.map((d) => d.code) };
}

function asSet(codes) {
  return [...new Set(codes)].sort();
}

function render(codes) {
  return codes.length === 0 ? '{}' : `{${codes.join(', ')}}`;
}

const fixtures = readdirSync(fixtureDir).filter((f) => f.endsWith('.svg')).sort();
let failures = 0;

for (const name of fixtures) {
  const path = join(fixtureDir, name);
  const declared = declaredCodes(readFileSync(path, 'utf8'));
  if (declared === null) {
    console.log(`FAIL  ${name}  no <!-- expect: ... --> declaration`);
    failures++;
    continue;
  }
  const result = emittedCodes(path);
  if (result.error !== undefined) {
    console.log(`FAIL  ${name}  ${result.error}`);
    failures++;
    continue;
  }
  const want = asSet(declared);
  const got = asSet(result.codes);
  if (want.join(' ') === got.join(' ')) {
    console.log(`PASS  ${name}  ${render(got)}`);
  } else {
    console.log(`FAIL  ${name}  expected ${render(want)} but got ${render(got)}`);
    failures++;
  }
}

console.log(`${fixtures.length - failures}/${fixtures.length} fixtures matched`);
process.exit(failures === 0 ? 0 : 1);
