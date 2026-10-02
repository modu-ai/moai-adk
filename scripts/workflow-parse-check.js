#!/usr/bin/env node
// workflow-parse-check.js — parse gate for dynamic-workflow scripts (.claude/workflows/*.js).
//
// Why not `node --check` alone: the workflow scripts' load shape is neither plain CJS nor plain
// ESM — `export const meta` (ESM syntax) coexists with top-level `return` (function-body-only),
// and the runtime strips the export prefix and runs the body as an async function with injected
// globals (args / agent / parallel / phase). Measured on Node 22.14 (2026-10-02, card t1445):
// file-mode `node --check` PASSES files that are invalid in BOTH module systems (module syntax
// detection falls back silently), so a prose intrusion inside a template literal goes uncaught —
// exactly the #1747 defect class this gate exists for.
//
// Acceptance: a file passes when EITHER check passes —
//   1. runtime-shape parse: `export ` prefixes rewritten to `const `, body compiled as an
//      AsyncFunction (top-level return/await legal — parse only, never invoked);
//   2. plain ESM parse (for scripts that are genuine ES modules, e.g. runner workflows without
//      a top-level return).
// A file fails only when BOTH fail — the "prose / broken syntax landed in the script" signal.
// Exits 0 when all green, 1 with per-file diagnostics otherwise.

import { readFileSync, readdirSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import path from 'node:path'

const roots = process.argv.slice(2).length > 0
  ? process.argv.slice(2)
  : ['.claude/workflows', 'internal/template/templates/.claude/workflows']

const AsyncFunction = Object.getPrototypeOf(async function () {}).constructor

const fnWrapParse = (src) => {
  // The runtime treats `export const meta` as presentation metadata: strip the prefix and the
  // body parses as a plain function body, where top-level return is legal.
  const body = src.replace(/^export /gm, '')
  new AsyncFunction('args', 'agent', 'parallel', 'phase', body) // parse-only, never invoked
}

const esmParse = (src) => {
  // Genuine ES modules (no top-level return): delegate to Node's own parser via stdin.
  execFileSync(process.execPath, ['--input-type=module', '--check'], {
    input: src,
    stdio: ['pipe', 'pipe', 'pipe'],
  })
}

const firstErrorLine = (e) => (e.message || String(e)).split('\n')[0]

let failures = 0
for (const root of roots) {
  let entries
  try {
    entries = readdirSync(root)
  } catch {
    console.log(`SKIP ${root} (not present in this checkout)`)
    continue
  }
  for (const name of entries.filter((f) => f.endsWith('.js')).sort()) {
    const file = path.join(root, name)
    const src = readFileSync(file, 'utf8')
    let fnErr = null
    try {
      fnWrapParse(src)
    } catch (e) {
      fnErr = firstErrorLine(e)
    }
    if (fnErr === null) {
      console.log(`PASS ${file} (runtime-shape)`)
      continue
    }
    try {
      esmParse(src)
      console.log(`PASS ${file} (esm)`)
    } catch (e) {
      failures++
      console.log(`FAIL ${file}`)
      console.log(`  runtime-shape: ${fnErr}`)
      const esmErr = e.stderr
        ? e.stderr.toString().split('\n').find((l) => l.includes('Error')) || firstErrorLine(e)
        : firstErrorLine(e)
      console.log(`  esm: ${esmErr}`)
    }
  }
}

if (failures > 0) {
  console.log(`\n${failures} workflow script(s) failed the parse gate.`)
  process.exit(1)
}
console.log('\nAll workflow scripts parse in their load shape.')
