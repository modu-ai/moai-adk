// sync-audit-4dim.js — 4-dimension sync-phase quality verdict (Context → Judge → Verdict)
//
// VERDICT SCOPING (what this workflow IS and is NOT):
//   This is an EXECUTION VEHICLE for a skeptical 4-dimension quality read. SPEC-AUDIT-SNAPSHOT-001
//   (A3) PROMOTED its verdict to BINDING on the happy path: where the verdict is PASS with all
//   four dims above their floor, not INCOMPLETE, and no contested finding, the orchestrator treats
//   this workflow's harmonic-mean verdict as the binding sync-phase verdict and does NOT spawn the
//   cold `sync-auditor` subagent. The cold auditor remains the FALLBACK verdict owner for the
//   failure modes (INCOMPLETE / dim-0 / contested finding) — see sync.md FO-SYNC-1 "Binding
//   promotion" and internal/runtime.FourDimVerdict.IsBinding() for the mechanical predicate. Four
//   dimensions are judged in parallel (Functionality / Security / Craft / Consistency); the verdict
//   is the HARMONIC MEAN of the four scores, chosen deliberately so that ONE low dimension drags
//   the whole verdict down (the arithmetic mean would let a strong dimension mask a weak one).
//
// PHANTOM-MECHANISM GUARD (composes with the existing guards; does NOT replace any of them and
//   does NOT add a 5th dimension). A SPEC under audit MAY declare defensive / structural mechanisms
//   it claims to have implemented (an input-validation guard, a migration call-site, an invariant
//   anchor). Each declared mechanism carries a literal probe_command + expected_match_substring.
//   The Context agent executes each probe against the ACTUAL write surface (the diff.patch paths
//   plus the produced files in the working tree — NOT the declared target_surface intent alone,
//   because a declaration is a claim, not evidence) and returns probe_results. The Verdict phase
//   then applies a deterministic JS rule: any claimed mechanism whose probe returned
//   actual_matches == 0 (and did NOT error) is a PHANTOM mechanism — a hard FAIL naming it, never
//   absorbed into a softer verdict. A probe that errored (no count produced) routes to
//   evidence_gaps — it is neither a phantom (which requires a zero COUNT) nor a verified mechanism.
//   The guard has a TIGHTER trust boundary than the harmonic mean: a literal command + a
//   deterministic integer count, re-runnable by any reviewer against the diff. This closes the
//   happy-path falsification surface (a PASS bypasses the cold auditor entirely, so a declared-
//   but-absent mechanism would otherwise ride a falsified PASS through the harmonic mean).
//
// Gate scope (honored via args.tier): Tier M and Tier L SPECs route through this 4-dimension gate.
//   Tier S SPECs do NOT — the caller (orchestrator) does not launch this workflow for a Tier S SPEC.
//   The gate is caller-side; args.tier is carried into the verdict for auditability.
//
// Determinism: spec_id / threshold / tier injected via `args`; no wall-clock read and no
//   random draw in the script body or in the verdict pure functions (resume-cache safe — any
//   timestamp is stamped by the orchestrator AFTER the run returns, per dynamic-workflows.md
//   § How a Workflow Runs).
//
// Read-only: every agent (Context + all 4 Judges) is agentType 'Explore' — no Write/Edit is
//   granted. The Context agent executes probe_commands via read-only Bash (grep / cat / git diff
//   against the working tree); it never mutates. Judges gather evidence and score; they never mutate.
//
// HARD constraints:
//   - No AskUserQuestion / no interactive surface — workflow agents cannot prompt the user
//     (agent-common-protocol.md § User Interaction Boundary); a judge lacking input returns its
//     evidence_gaps, never a question.
//   - No meta-judge agent — the aggregate is computed in SCRIPT JS below, never by a 5th LLM call
//     (a meta-judge would smooth dissent, defeating the harmonic mean's purpose).
//   - No LLM arithmetic — the harmonic mean AND the actual_matches == 0 comparison are JS,
//     deterministic and auditable.
//   - Gate on Tier M/L only (Tier S does not launch).
//
// Fail-honest semantics: ANY judge that fails to return (null / unparseable score) yields verdict
//   INCOMPLETE naming the missing dimension(s) — 3/4 is NOT a weaker verdict, it is NO verdict
//   (evidence absent != evidence of success, verification-claim-integrity.md §1). A score of 0 trips
//   the zero-score guard (a hard FAIL naming the dimension — never a divide-by-zero). A phantom
//   mechanism trips the phantom-mechanism guard (a hard FAIL naming the mechanism — never a silent
//   pass-through to the harmonic mean).
//
// Distribution: this is a MoAI-shipped generic fan-out script — it is template-managed, so `moai
//   update` overwrites the local copy. Edit it in the template source, not in the local project.
//   User-owned Runner Workflows (the `hns-*` / `harness-*` prefixes) are preserved instead.
//
// Cross-refs: dynamic-workflows.md — the workflow primitive (16-concurrent cap, determinism,
//   resume caching) and the shipped-vs-user-owned split under `.claude/workflows/`.
//
// Usage:
//   Workflow({ scriptPath: ".claude/workflows/sync-audit-4dim.js",
//              args: { spec_id: "SPEC-FOO-001", threshold: 0.85, tier: "L" } })

export const meta = {
  name: 'sync-audit-4dim',
  description: 'Sync-phase 4-dimension quality read (Functionality/Security/Craft/Consistency) — parallel read-only judges + in-script harmonic-mean verdict with phantom-mechanism guard; execution vehicle, NOT the binding sync-auditor verdict owner',
  phases: [
    { title: 'Context', detail: 'one read-only Explore agent extracts the SPEC audit surface (id, acceptance criteria, changed files, test command) AND executes any declared phantom-mechanism probes against the actual write surface' },
    { title: 'Judge', detail: 'four parallel read-only Explore judges, one per dimension, each scoring 0-1 with command+verbatim-output evidence under a skeptical-auditor stance' },
    { title: 'Verdict', detail: 'in-script deterministic verdict — null-judge guard → zero-score guard → phantom-mechanism guard → harmonic mean (no agent call)' },
  ],
}

// determinism: all inputs injected via args; no wall-clock, no random in body
const SPEC_ID = (args && args.spec_id) || 'SPEC-UNKNOWN'
const THRESHOLD = (args && typeof args.threshold === 'number') ? args.threshold : 0.85
const TIER = (args && args.tier) || 'M'

// === VERDICT PURE FUNCTIONS START ===
// Deterministic verdict logic — no runtime globals (agent/parallel/phase/args), no wall-clock,
// no random draw. Unit-tested directly via node:test (see sync-audit-4dim.test.js). Extracted so
// the verdict decision is falsifiable without launching the workflow runtime.

// The four audit dimensions. FROZEN — the phantom-mechanism guard adds a VERDICT-PHASE guard,
// NOT a 5th dimension. Verdict order below MUST match this array (judges[i] <-> DIMENSIONS[i]).
const DIMENSIONS = ['Functionality', 'Security', 'Craft', 'Consistency']

// A judge is "missing" if it did not return or its score is not a finite number.
const scoreOf = (j) => (j && typeof j.score === 'number' && Number.isFinite(j.score)) ? j.score : null

// Phantom-mechanism detection. Operates on probe_results returned by the Context agent.
//
// Routing rules (deterministic):
//   - error present (no count produced) → evidence_gap (the mechanism is NEITHER phantom NOR
//     verified; the missing data is reported, never treated as a pass or a fail on its own).
//   - actual_matches == 0 (a real zero COUNT, no error) → phantom (hard FAIL).
//   - actual_matches > 0 → verified (falls through to the harmonic mean).
//   - absent / empty probe_results → no-op (the guard is inert; verdict falls through).
function detectPhantomMechanisms(probeResults) {
  const phantoms = []
  const errored = []
  if (!Array.isArray(probeResults) || probeResults.length === 0) {
    return { phantoms, errored }
  }
  for (const r of probeResults) {
    if (!r || typeof r !== 'object') continue
    const entry = {
      name: r.name,
      probe_command: r.probe_command,
      expected_match_substring: r.expected_match_substring,
    }
    // An errored probe produced NO count — route to evidence_gaps. This MUST NOT trigger a
    // phantom-FAIL (which requires an observed zero COUNT) and MUST NOT be treated as verified.
    if (r.error) {
      errored.push({ ...entry, error: r.error })
      continue
    }
    // A real zero count (no error) is a phantom — a claimed mechanism with zero on-disk evidence.
    if (typeof r.actual_matches === 'number' && r.actual_matches === 0) {
      phantoms.push({ ...entry, actual_matches: 0 })
    }
  }
  return { phantoms, errored }
}

// Pure verdict — total precedence order:
//   1. null-judge guard   → INCOMPLETE (4 dims are the contract; 3/4 is no verdict)
//   2. zero-score guard   → FAIL (a 0 dimension makes the harmonic mean undefined — hard FAIL)
//   3. phantom-mechanism guard → FAIL (any claimed mechanism with actual_matches == 0 — hard FAIL)
//   4. harmonic mean      → PASS / FAIL
// opts: { judges, probeResults, dimensions, threshold, tier, specId }
function computeVerdict(opts) {
  const judges = opts.judges
  const probeResults = opts.probeResults || []
  const DIM = opts.dimensions
  const threshold = opts.threshold
  const tier = opts.tier
  const specId = opts.specId

  // (1) null-judge guard FIRST — before any mean computation or probe inspection.
  const missing = DIM.filter((dim, i) => scoreOf(judges[i]) === null)
  if (missing.length > 0) {
    return { verdict: 'INCOMPLETE', missing, tier, threshold, spec_id: specId }
  }

  // All four judges returned a finite score. Aggregate findings/gaps for the report.
  const scores = DIM.map((dim, i) => judges[i].score)
  const findings = DIM.flatMap((dim, i) => (judges[i].findings || []).filter(Boolean).map((f) => ({ dimension: dim, ...f })))
  const evidenceGaps = DIM.flatMap((dim, i) => (judges[i].evidence_gaps || []).filter(Boolean).map((g) => ({ dimension: dim, gap: g })))

  // (2) zero-score guard — the harmonic mean divides by each score, so a 0 dimension is a hard FAIL.
  const zeroScored = DIM.filter((dim, i) => scores[i] <= 0)
  if (zeroScored.length > 0) {
    return { verdict: 'FAIL', zero_scored: zeroScored, tier, threshold, spec_id: specId, findings, evidence_gaps: evidenceGaps }
  }

  // (3) phantom-mechanism guard — AFTER zero-score (structural / cheaper), BEFORE harmonic mean.
  // The probe_command in each phantom entry IS the verbatim command that was run; actual_matches: 0
  // IS the observed output — per baseline-integrity attribution, the reviewer can re-run the exact
  // command against the diff and observe the zero directly.
  const { phantoms, errored } = detectPhantomMechanisms(probeResults)
  const aggregatedGaps = errored.length > 0
    ? evidenceGaps.concat(errored.map((e) => ({ dimension: 'PhantomProbe', gap: `probe for "${e.name}" returned no count (${e.error}); probe_command: ${e.probe_command}` })))
    : evidenceGaps
  if (phantoms.length > 0) {
    return {
      verdict: 'FAIL',
      phantom_mechanisms: phantoms,
      tier,
      threshold,
      spec_id: specId,
      findings,
      evidence_gaps: aggregatedGaps,
    }
  }

  // (4) harmonic mean — n / Σ(1/sᵢ), in-script, deterministic, auditable. One low dimension drags
  // it down. Reached only when every claimed mechanism was either verified (actual_matches > 0),
  // errored (routed to evidence_gaps), or absent (no claimed_mechanisms declared — the phantom
  // guard is a no-op).
  const reciprocalSum = scores.reduce((acc, s) => acc + 1 / s, 0)
  const harmonicMean = DIM.length / reciprocalSum

  return {
    verdict: harmonicMean >= threshold ? 'PASS' : 'FAIL',
    harmonic_mean: harmonicMean,
    threshold,
    tier,
    spec_id: specId,
    scores: DIM.map((dim, i) => ({ dimension: dim, score: scores[i] })),
    findings,
    evidence_gaps: aggregatedGaps,
  }
}
// === VERDICT PURE FUNCTIONS END ===

// Schema-forced output: the verdict computation consumes typed fields, so the Context + Judge
// outputs are schema-shaped (arithmetic needs structure). Explorer narrative in the sibling
// plan-research-fanout.js is markdown by contrast — that asymmetry is deliberate.
const CONTEXT_SCHEMA = {
  spec_id: 'string — the audited SPEC id',
  acceptance_criteria: ['string — one AC statement per entry'],
  changed_files: ['string — repo-relative path touched by this SPEC'],
  test_command: 'string — the command that runs this SPEC test suite',
  // Optional: defensive / structural mechanisms the SPEC claims to have implemented. Each entry
  // carries a literal probe_command + the substring the probe is expected to match on disk. The
  // Context agent executes each probe_command against the ACTUAL write surface (diff.patch paths +
  // produced files) and returns probe_results. Absent or empty → the phantom-mechanism guard is
  // a no-op (verdict falls through to the harmonic mean unchanged).
  claimed_mechanisms: [{
    name: 'string — short identifier for the claimed mechanism',
    probe_command: 'string — literal read-only shell command whose stdout+file corpus is searched',
    expected_match_substring: 'string — literal substring (NOT regex) the probe counts occurrences of',
  }],
  // Probe execution results — one per claimed_mechanism entry, populated by the Context agent.
  // actual_matches == 0 (no error) → phantom mechanism (hard FAIL). error present → evidence_gap.
  probe_results: [{
    name: 'string — matches the claimed_mechanism.name',
    probe_command: 'string — the verbatim command that was run',
    expected_match_substring: 'string — the substring that was counted',
    actual_matches: 'number — observed count of expected_match_substring in the probe corpus',
    error: 'optional string — present when the probe command could not produce a count (file not found, non-zero exit, timeout); routes to evidence_gaps, NOT phantom-FAIL',
  }],
}

const JUDGE_SCHEMA = {
  dimension: 'string — one of Functionality/Security/Craft/Consistency',
  score: 'number 0..1 — quality score for this dimension (0 = hard fail, 1 = flawless)',
  findings: [{ severity: 'critical|major|minor', summary: 'string', file: 'string', evidence: 'string — command run + verbatim output' }],
  evidence_gaps: ['string — a check the judge could NOT run and why (evidence absent != pass)'],
}

// ---------------------------------------------------------------------------
phase('Context')

const CONTEXT_PROMPT = `You are a read-only audit-context extractor. Do NOT modify any file.

Analyze the SPEC "${SPEC_ID}" in this repository. Read its artifacts under .moai/specs/${SPEC_ID}/
(spec.md, plan.md, acceptance.md, progress.md) using Read/Grep/Glob.

Return the audit surface as an object with EXACTLY these fields:
- spec_id: the SPEC id ("${SPEC_ID}")
- acceptance_criteria: the list of acceptance-criterion statements (from acceptance.md, the SSOT)
- changed_files: the list of repo-relative source paths this SPEC touches (from plan.md scope + git)
- test_command: the single command that runs this SPEC's test suite (e.g. "go test ./internal/foo/...")
- claimed_mechanisms: the list of defensive / structural mechanisms this SPEC CLAIMS to have
  implemented, as declared in its plan.md / acceptance.md. Each entry MUST carry:
    * name           — short identifier for the mechanism
    * probe_command  — a literal READ-ONLY shell command (grep / cat / git diff — NO Write/Edit,
                       NO mutation) whose output the audit will search for the expected substring.
                       The command MUST probe the ACTUAL write surface: the diff.patch paths plus
                       the produced files in the working tree (e.g. the files changed by this SPEC,
                       discoverable via \`git diff --name-only\` against the merge-base). Do NOT
                       point the probe at the declared target_surface intent alone — a declaration
                       is a claim, not evidence; the probe must hit what was actually written.
    * expected_match_substring — the literal substring (NOT a regex) the probe counts occurrences
                       of in the combined stdout + file-contents corpus. Keep it literal so the
                       count is deterministic and auditable.
  If the SPEC declares no such mechanisms, return claimed_mechanisms as an empty array [].
- probe_results: for EACH entry in claimed_mechanisms, execute its probe_command via read-only
  Bash and report the result. Each entry MUST carry:
    * name, probe_command, expected_match_substring — copied from the claimed_mechanism entry.
    * actual_matches — the integer count of expected_match_substring occurrences found in the
                       combined stdout + walked-file-contents corpus the probe_command produced.
                       Count literal substring matches (NOT regex). Zero is a real observation
                       (the mechanism is absent on disk) — report 0, do NOT fudge.
    * error — present ONLY when the probe could not produce a count (the command exited non-zero
              for a reason other than grep-no-match, the file was not found, or the probe timed
              out). When error is present, OMIT actual_matches. An errored probe is reported as
              an evidence_gap, never as a phantom-FAIL and never as a verified mechanism.
  If claimed_mechanisms is empty, return probe_results as an empty array [].

Report only what you can VERIFY from the artifacts and the working tree. If a field cannot be
determined, return it empty rather than guessing. The verdict phase will apply the deterministic
rule (actual_matches == 0 with no error → phantom-FAIL; error → evidence_gap) in JS — your job is
to execute the probes honestly and return the counts you actually observed.`

const context = await agent(CONTEXT_PROMPT, { label: `context:${SPEC_ID}`, phase: 'Context', agentType: 'Explore', effort: 'medium', schema: CONTEXT_SCHEMA })

// ---------------------------------------------------------------------------
phase('Judge')

// Skeptical-auditor stance: every score claim MUST be backed by a command that was actually run
// plus its verbatim output. Evidence absent is NOT evidence of a pass — it is an evidence_gap.
const JUDGE_PROMPT = (dimension) => `You are a read-only, skeptical quality auditor judging ONE dimension: ${dimension}.
Do NOT modify any file. You have Read/Grep/Glob and read-only Bash (test/lint/build) only.

Audit context for the SPEC under review:
${JSON.stringify(context, null, 2)}

Judge the "${dimension}" dimension of this SPEC's implementation. Score it 0..1 where:
  1.0 = flawless on this dimension, 0.0 = a hard failure on this dimension.

Skeptical stance (verification-claim-integrity §1): treat every claim as suspect until you have SHOWN
evidence. Each finding's "evidence" field MUST contain the exact command you ran AND its verbatim
output — never a summary, never an assumption. A check you could not run is an evidence_gap, NOT a pass.

Dimension focus for "${dimension}":
  - Functionality: do the acceptance criteria actually hold? Run the test_command; read the ACs; verify behavior.
  - Security: input validation at trust boundaries, secret handling, injection surfaces, OWASP-relevant defects.
  - Craft: readability, naming, simplicity, duplication, error handling — would a staff engineer accept it?
  - Consistency: does it match the existing codebase style, conventions, and neighbouring patterns?

Return an object with EXACTLY: dimension, score (0..1), findings[{severity,summary,file,evidence}],
evidence_gaps[]. If you cannot evaluate this dimension at all, return score as null (do NOT fabricate a score).`

// Four judge agent calls in parallel — ALL read-only (agentType 'Explore'), effort 'xhigh'. Each
// call site inlines the read-only opts so the read-only contract is pinned to the JUDGE site itself.
// Thunk order MUST match DIMENSIONS so judges[i] aligns with DIMENSIONS[i] in the Verdict phase.
const judges = await parallel([
  () => agent(JUDGE_PROMPT('Functionality'), { label: 'judge:Functionality', phase: 'Judge', agentType: 'Explore', effort: 'xhigh', schema: JUDGE_SCHEMA }),
  () => agent(JUDGE_PROMPT('Security'),      { label: 'judge:Security',      phase: 'Judge', agentType: 'Explore', effort: 'xhigh', schema: JUDGE_SCHEMA }),
  () => agent(JUDGE_PROMPT('Craft'),         { label: 'judge:Craft',         phase: 'Judge', agentType: 'Explore', effort: 'xhigh', schema: JUDGE_SCHEMA }),
  () => agent(JUDGE_PROMPT('Consistency'),   { label: 'judge:Consistency',   phase: 'Judge', agentType: 'Explore', effort: 'xhigh', schema: JUDGE_SCHEMA }),
])

// ---------------------------------------------------------------------------
phase('Verdict')

// SCRIPT JS ONLY — no agent call sits between judge collection and the returned verdict. The
// verdict decision is a PURE function (computeVerdict above): null-judge → zero-score → phantom →
// harmonic mean. probe_results come from the Context agent (the script body has no shell / fs
// access per dynamic-workflows.md § How a Workflow Runs); only the verdict DECISION is JS — probe
// EXECUTION is delegated to the Context agent because the script cannot run shell itself.
return computeVerdict({
  judges,
  probeResults: (context && context.probe_results) || [],
  dimensions: DIMENSIONS,
  threshold: THRESHOLD,
  tier: TIER,
  specId: SPEC_ID,
})
