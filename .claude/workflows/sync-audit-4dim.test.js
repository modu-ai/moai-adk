// sync-audit-4dim.test.js — unit tests for the PURE verdict functions embedded in
// sync-audit-4dim.js. Lightweight, dependency-free (Node built-in `node:test`).
//
// The workflow script cannot be imported directly (its body uses dynamic-workflows runtime
// globals `agent` / `parallel` / `phase` / `args` plus top-level `return`, which are not legal
// under a plain Node ESM import). The verdict decision logic is therefore extracted as a PURE
// function block delimited by sentinel comments, and this test harness loads that block via
// `new Function` — exercising the verdict logic WITHOUT the workflow runtime.
//
// SECURITY RATIONALE (test-only `new Function`, not eval-style dynamic dispatch):
//   This `new Function` is TEST-ONLY. Its input is NOT untrusted/external input — it is a
//   sentinel-delimited substring of THIS REPO'S OWN committed workflow source
//   (sync-audit-4dim.js, between `// === VERDICT PURE FUNCTIONS START/END ===`). The source is
//   read at test time from the local file and bounds-checked by the assertions above. It exists
//   because the dynamic-workflows runtime injects globals + the script top-level `return`s,
//   making a plain Node ESM `import` a SyntaxError (no workflow script under .claude/workflows/
//   uses import/require). The production sync-audit-4dim.js contains NO eval / new Function /
//   Math.random / Date.now — the verdict path is deterministic and resume-cache safe.
//
// Run: `node --test .claude/workflows/sync-audit-4dim.test.js`
//
// DEV-ONLY — not mirrored to internal/template/templates/.claude/workflows/ (tests are not a
// distributed template asset).

import { test } from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const workflowSrc = fs.readFileSync(path.join(__dirname, 'sync-audit-4dim.js'), 'utf8')

// --- extract the pure-verdict block between sentinel comments -----------------------------
const START = '// === VERDICT PURE FUNCTIONS START ==='
const END = '// === VERDICT PURE FUNCTIONS END ==='
const startIdx = workflowSrc.indexOf(START)
const endIdx = workflowSrc.indexOf(END)
assert.ok(startIdx !== -1, 'VERDICT PURE FUNCTIONS START marker not found in workflow source')
assert.ok(endIdx !== -1, 'VERDICT PURE FUNCTIONS END marker not found in workflow source')
assert.ok(endIdx > startIdx, 'VERDICT markers are out of order')

const pureBlock = workflowSrc.slice(startIdx + START.length, endIdx)
// Compile the pure block in a sandbox that returns the exported bindings.
// TEST-ONLY `new Function`: pureBlock is a sentinel-extracted substring of THIS repo's own
// committed sync-audit-4dim.js (verified above), NOT untrusted input — see file header for
// the full security rationale. Production sync-audit-4dim.js is eval/free.
const harness = new Function(pureBlock + '\n; return { DIMENSIONS, scoreOf, detectPhantomMechanisms, computeVerdict }')
const { DIMENSIONS, scoreOf, detectPhantomMechanisms, computeVerdict } = harness()

// --- fixtures ------------------------------------------------------------------------------
const fourFiniteJudges = [
  { dimension: 'Functionality', score: 0.9, findings: [], evidence_gaps: [] },
  { dimension: 'Security', score: 0.9, findings: [], evidence_gaps: [] },
  { dimension: 'Craft', score: 0.9, findings: [], evidence_gaps: [] },
  { dimension: 'Consistency', score: 0.9, findings: [], evidence_gaps: [] },
]
const baseOpts = () => ({
  judges: fourFiniteJudges.map((j) => ({ ...j })),
  probeResults: [],
  dimensions: DIMENSIONS,
  threshold: 0.85,
  tier: 'M',
  specId: 'SPEC-TEST-001',
})

// --- AC-FP-001: phantom mechanism yields FAIL with named phantom in payload ----------------
test('AC-FP-001: a claimed mechanism with actual_matches == 0 yields verdict FAIL naming it in phantom_mechanisms[], and does NOT fall through to the harmonic mean', () => {
  const opts = baseOpts()
  opts.probeResults = [{
    name: 'input-validation-guard',
    probe_command: "grep -r 'ValidateInput' $(git diff --name-only)",
    expected_match_substring: 'ValidateInput',
    actual_matches: 0,
  }]
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'FAIL')
  assert.equal(v.phantom_mechanisms.length, 1)
  assert.equal(v.phantom_mechanisms[0].name, 'input-validation-guard')
  assert.equal(v.phantom_mechanisms[0].actual_matches, 0)
  assert.equal(v.phantom_mechanisms[0].probe_command, "grep -r 'ValidateInput' $(git diff --name-only)")
  assert.equal(v.phantom_mechanisms[0].expected_match_substring, 'ValidateInput')
  // The phantom guard sat BEFORE the harmonic mean — no harmonic_mean field on a phantom-FAIL.
  assert.equal(v.harmonic_mean, undefined,
    'phantom-FAIL must NOT fall through to the harmonic-mean computation (no harmonic_mean field)')
})

// --- AC-FP-002: happy-path passthrough -----------------------------------------------------
test('AC-FP-002: a claimed mechanism with actual_matches > 0 passes the guard unchanged and reaches the harmonic mean', () => {
  const opts = baseOpts()
  opts.probeResults = [{
    name: 'input-validation-guard',
    probe_command: "grep -r 'ValidateInput' $(git diff --name-only)",
    expected_match_substring: 'ValidateInput',
    actual_matches: 3,
  }]
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'PASS')
  assert.equal(typeof v.harmonic_mean, 'number')
  assert.ok(v.harmonic_mean >= 0.85)
  // No phantom_mechanisms entry for the verified mechanism — the guard fell through silently.
  assert.equal(v.phantom_mechanisms, undefined,
    'a verified mechanism must NOT produce a phantom_mechanisms[] entry')
})

// --- AC-FP-003: precedence — null-judge INCOMPLETE fires before phantom guard ---------------
test('AC-FP-003: null-judge (INCOMPLETE) fires BEFORE the phantom-mechanism guard even when a probe returned actual_matches == 0', () => {
  const opts = baseOpts()
  opts.judges[2] = { dimension: 'Craft', score: null, findings: [], evidence_gaps: [] } // null score
  opts.probeResults = [{
    name: 'phantom-mech',
    probe_command: 'grep -r NoSuchThing anywhere',
    expected_match_substring: 'NoSuchThing',
    actual_matches: 0,
  }]
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'INCOMPLETE')
  assert.deepEqual(v.missing, ['Craft'])
  assert.equal(v.phantom_mechanisms, undefined,
    'phantom guard was never reached — null-judge guard fires first')
})

// --- AC-FP-004: precedence — zero-score FAIL fires before phantom guard --------------------
test('AC-FP-004: zero-score (FAIL) fires BEFORE the phantom-mechanism guard even when a probe returned actual_matches == 0', () => {
  const opts = baseOpts()
  opts.judges[1] = { dimension: 'Security', score: 0, findings: [], evidence_gaps: [] } // zero score
  opts.probeResults = [{
    name: 'phantom-mech',
    probe_command: 'grep -r NoSuchThing anywhere',
    expected_match_substring: 'NoSuchThing',
    actual_matches: 0,
  }]
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'FAIL')
  assert.deepEqual(v.zero_scored, ['Security'])
  assert.equal(v.phantom_mechanisms, undefined,
    'phantom guard was never reached — zero-score guard fires first')
})

// --- AC-FP-005: per-mechanism baseline attribution in the FAIL payload ---------------------
test('AC-FP-005: every phantom_mechanisms[] entry carries {name, probe_command, expected_match_substring, actual_matches: 0} for baseline attribution', () => {
  const opts = baseOpts()
  opts.probeResults = [
    { name: 'mech-a', probe_command: 'cmd-a', expected_match_substring: 'SubA', actual_matches: 0 },
    { name: 'mech-b', probe_command: 'cmd-b', expected_match_substring: 'SubB', actual_matches: 0 },
  ]
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'FAIL')
  assert.equal(v.phantom_mechanisms.length, 2)
  for (const p of v.phantom_mechanisms) {
    assert.ok('name' in p)
    assert.ok('probe_command' in p)
    assert.ok('expected_match_substring' in p)
    assert.equal(p.actual_matches, 0)
  }
})

// --- AC-FP-006: deterministic — no Date.now / Math.random in the verdict path --------------
test('AC-FP-006: the pure verdict block references neither Date.now nor Math.random (deterministic; resume-cache safe)', () => {
  // A call in the verdict path would break resume caching per dynamic-workflows.md.
  // Mentioning them in comments is allowed by the runtime; a real CALL is the hazard. Scan for
  // call-form specifically (identifier followed by '(').
  const callPattern = /\b(Date\.now|Math\.random)\s*\(/
  assert.equal(callPattern.test(pureBlock), false,
    'verdict pure block must not CALL Date.now() or Math.random() — found a call site')
})

test('AC-FP-006 (determinism): identical inputs produce byte-identical verdicts across two invocations', () => {
  const opts = baseOpts()
  opts.probeResults = [{ name: 'm', probe_command: 'c', expected_match_substring: 's', actual_matches: 0 }]
  const v1 = computeVerdict(structuredClone(opts))
  const v2 = computeVerdict(structuredClone(opts))
  assert.deepEqual(v1, v2)
})

// --- AC-FP-008: 4-dimension enum FROZEN ----------------------------------------------------
test('AC-FP-008: DIMENSIONS is exactly [Functionality, Security, Craft, Consistency] — the phantom guard did NOT add a 5th dimension', () => {
  assert.deepEqual(DIMENSIONS, ['Functionality', 'Security', 'Craft', 'Consistency'])
  assert.equal(DIMENSIONS.length, 4)
})

// --- AC-FP-011: empty claimed_mechanisms → phantom guard is a no-op ------------------------
test('AC-FP-011: absent probeResults makes the phantom guard a no-op (verdict falls through to harmonic mean; no phantom_mechanisms field)', () => {
  const opts = baseOpts()
  opts.probeResults = undefined
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'PASS')
  assert.equal(v.phantom_mechanisms, undefined)
})

test('AC-FP-011: empty probeResults array makes the phantom guard a no-op (verdict falls through to harmonic mean; no phantom_mechanisms field)', () => {
  const opts = baseOpts()
  opts.probeResults = []
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'PASS')
  assert.equal(v.phantom_mechanisms, undefined)
})

// --- AC-FP-012: probe execution error routes to evidence_gaps, NOT phantom-FAIL ------------
test('AC-FP-012: an errored probe (no actual_matches) routes to evidence_gaps and does NOT trigger phantom-FAIL nor a spurious PASS', () => {
  const opts = baseOpts()
  opts.probeResults = [{
    name: 'mech-that-errored',
    probe_command: 'grep -r Sub missing-dir',
    expected_match_substring: 'Sub',
    error: 'exit status 2: directory not found',
    // actual_matches OMITTED — the probe could not produce a count.
  }]
  const v = computeVerdict(opts)
  // NOT a phantom-FAIL — phantom requires an observed zero COUNT.
  assert.equal(v.phantom_mechanisms, undefined,
    'an errored probe must NOT be treated as a phantom (phantom requires actual_matches == 0)')
  // The errored mechanism is reported in evidence_gaps.
  const gap = (v.evidence_gaps || []).find((g) => g.dimension === 'PhantomProbe')
  assert.ok(gap, 'an errored probe MUST appear in evidence_gaps[]')
  assert.ok(gap.gap.includes('mech-that-errored'))
  assert.ok(gap.gap.includes('grep -r Sub missing-dir'))
  // The verdict is NOT spurious — the four judges still produce a real harmonic mean verdict
  // (the errored probe is reported as a gap, not silently absorbed). With all judges at 0.9 and
  // threshold 0.85, the verdict is PASS *with the gap surfaced*, not a silent clean PASS.
  assert.equal(v.verdict, 'PASS') // the gap does not override the four finite scores
  assert.ok(v.evidence_gaps.some((g) => g.dimension === 'PhantomProbe'))
})

test('AC-FP-012 (mixed): a phantom (actual_matches==0) AND an errored probe route correctly — phantom triggers FAIL, errored probe is a gap', () => {
  const opts = baseOpts()
  opts.probeResults = [
    { name: 'real-phantom', probe_command: 'cmd-p', expected_match_substring: 'P', actual_matches: 0 },
    { name: 'errored-probe', probe_command: 'cmd-e', expected_match_substring: 'E', error: 'boom' },
  ]
  const v = computeVerdict(opts)
  assert.equal(v.verdict, 'FAIL')
  assert.equal(v.phantom_mechanisms.length, 1)
  assert.equal(v.phantom_mechanisms[0].name, 'real-phantom')
  assert.ok(v.evidence_gaps.some((g) => g.dimension === 'PhantomProbe' && g.gap.includes('errored-probe')))
})

// --- structural: verdict precedence order is total (AC-FP-003/004 + happy path) -----------
test('structural: verdict precedence is total — INCOMPLETE > zero-FAIL > phantom-FAIL > harmonic mean', () => {
  // INCOMPLETE wins over everything.
  let opts = baseOpts()
  opts.judges = [
    { score: null, findings: [], evidence_gaps: [] },
    { score: 0, findings: [], evidence_gaps: [] },
    { score: 0.9, findings: [], evidence_gaps: [] },
    { score: 0.9, findings: [], evidence_gaps: [] },
  ]
  opts.probeResults = [{ name: 'p', probe_command: 'c', expected_match_substring: 's', actual_matches: 0 }]
  assert.equal(computeVerdict(opts).verdict, 'INCOMPLETE')

  // zero-FAIL wins over phantom-FAIL.
  opts = baseOpts()
  opts.judges = [
    { score: 0.9, findings: [], evidence_gaps: [] },
    { score: 0, findings: [], evidence_gaps: [] },
    { score: 0.9, findings: [], evidence_gaps: [] },
    { score: 0.9, findings: [], evidence_gaps: [] },
  ]
  opts.probeResults = [{ name: 'p', probe_command: 'c', expected_match_substring: 's', actual_matches: 0 }]
  const zf = computeVerdict(opts)
  assert.equal(zf.verdict, 'FAIL')
  assert.ok(zf.zero_scored)
  assert.equal(zf.phantom_mechanisms, undefined)

  // phantom-FAIL wins over harmonic-mean FAIL (low score).
  opts = baseOpts()
  opts.judges = fourFiniteJudges.map((j) => ({ ...j, score: 0.2 })) // low → harmonic mean FAIL
  opts.probeResults = [{ name: 'p', probe_command: 'c', expected_match_substring: 's', actual_matches: 0 }]
  const pf = computeVerdict(opts)
  assert.equal(pf.verdict, 'FAIL')
  assert.ok(pf.phantom_mechanisms)
  assert.equal(pf.harmonic_mean, undefined, 'phantom-FAIL short-circuits before harmonic mean computation')
})

// --- detectPhantomMechanisms unit (helper) ------------------------------------------------
test('detectPhantomMechanisms: returns empty arrays for absent / empty / non-array input', () => {
  assert.deepEqual(detectPhantomMechanisms(undefined), { phantoms: [], errored: [] })
  assert.deepEqual(detectPhantomMechanisms(null), { phantoms: [], errored: [] })
  assert.deepEqual(detectPhantomMechanisms([]), { phantoms: [], errored: [] })
  assert.deepEqual(detectPhantomMechanisms('nope'), { phantoms: [], errored: [] })
})

test('detectPhantomMechanisms: scoreOf is null-safe on missing / malformed judge objects', () => {
  assert.equal(scoreOf(null), null)
  assert.equal(scoreOf(undefined), null)
  assert.equal(scoreOf({}), null)
  assert.equal(scoreOf({ score: NaN }), null)
  assert.equal(scoreOf({ score: Infinity }), null)
  assert.equal(scoreOf({ score: '0.9' }), null) // string, not number
  assert.equal(scoreOf({ score: 0 }), 0) // zero is a VALID finite number (triggers zero-score guard)
  assert.equal(scoreOf({ score: 0.9 }), 0.9)
})
