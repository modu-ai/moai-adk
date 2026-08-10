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
// Gate scope (honored via args.tier): Tier M and Tier L SPECs route through this 4-dimension gate.
//   Tier S SPECs do NOT — the caller (orchestrator) does not launch this workflow for a Tier S SPEC.
//   The gate is caller-side; args.tier is carried into the verdict for auditability.
//
// Determinism: spec_id / threshold / tier injected via `args`; no wall-clock read and no
//   random draw in the script body (resume-cache safe — any timestamp is stamped by the
//   orchestrator AFTER the run returns, per dynamic-workflows.md § How a Workflow Runs).
//
// Read-only: every agent (Context + all 4 Judges) is agentType 'Explore' — no Write/Edit is
//   granted. Judges gather evidence (Read/Grep/Glob/Bash read-only) and score; they never mutate.
//
// HARD constraints:
//   - No AskUserQuestion / no interactive surface — workflow agents cannot prompt the user
//     (agent-common-protocol.md § User Interaction Boundary); a judge lacking input returns its
//     evidence_gaps, never a question.
//   - No meta-judge agent — the aggregate is computed in SCRIPT JS below, never by a 5th LLM call
//     (a meta-judge would smooth dissent, defeating the harmonic mean's purpose).
//   - No LLM arithmetic — the harmonic mean is JS, deterministic and auditable.
//   - Gate on Tier M/L only (Tier S does not launch).
//
// Fail-honest semantics: ANY judge that fails to return (null / unparseable score) yields verdict
//   INCOMPLETE naming the missing dimension(s) — 3/4 is NOT a weaker verdict, it is NO verdict
//   (evidence absent != evidence of success, verification-claim-integrity.md §1). A score of 0 trips
//   the zero-score guard (a hard FAIL naming the dimension — never a divide-by-zero).
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
  description: 'Sync-phase 4-dimension quality read (Functionality/Security/Craft/Consistency) — parallel read-only judges + in-script harmonic-mean verdict; execution vehicle, NOT the binding sync-auditor verdict owner',
  phases: [
    { title: 'Context', detail: 'one read-only Explore agent extracts the SPEC audit surface (id, acceptance criteria, changed files, test command)' },
    { title: 'Judge', detail: 'four parallel read-only Explore judges, one per dimension, each scoring 0-1 with command+verbatim-output evidence under a skeptical-auditor stance' },
    { title: 'Verdict', detail: 'in-script harmonic mean of the four scores with a zero-score guard and an INCOMPLETE branch on any missing judge (no agent call)' },
  ],
}

// determinism: all inputs injected via args; no wall-clock, no random in body
const SPEC_ID = (args && args.spec_id) || 'SPEC-UNKNOWN'
const THRESHOLD = (args && typeof args.threshold === 'number') ? args.threshold : 0.85
const TIER = (args && args.tier) || 'M'

// The four audit dimensions. Verdict order below MUST match this array (judges[i] <-> DIMENSIONS[i]).
const DIMENSIONS = ['Functionality', 'Security', 'Craft', 'Consistency']

// Schema-forced output: the verdict computation consumes typed fields, so the Context + Judge
// outputs are schema-shaped (arithmetic needs structure). Explorer narrative in the sibling
// plan-research-fanout.js is markdown by contrast — that asymmetry is deliberate.
// These MUST be real JSON Schema. `agent({schema})` validates them under strict mode,
// where a bare field name at the top level is an unknown keyword and the call throws
// before any agent runs — so a shape-object sketch here disables the whole fan-out.
const CONTEXT_SCHEMA = {
  type: 'object',
  properties: {
    spec_id: { type: 'string', description: 'the audited SPEC id' },
    acceptance_criteria: {
      type: 'array',
      description: 'one AC statement per entry',
      items: { type: 'string' },
    },
    changed_files: {
      type: 'array',
      description: 'repo-relative paths this SPEC touches',
      items: { type: 'string' },
    },
    test_command: { type: 'string', description: 'the command that runs this SPEC test suite' },
  },
  required: ['spec_id', 'acceptance_criteria', 'changed_files', 'test_command'],
  additionalProperties: false,
}

// `score` is nullable on purpose: the judge prompt instructs a judge that cannot evaluate
// its dimension to return null rather than fabricate a number, and the Verdict phase reads
// a non-numeric score as a missing judge (INCOMPLETE). A non-nullable number here would
// force the fabrication the prompt forbids. anyOf rather than a union `type` keeps the
// range constraint unambiguous under strict validation.
const JUDGE_SCHEMA = {
  type: 'object',
  properties: {
    dimension: {
      type: 'string',
      enum: ['Functionality', 'Security', 'Craft', 'Consistency'],
    },
    score: {
      anyOf: [
        { type: 'number', minimum: 0, maximum: 1 },
        { type: 'null' },
      ],
      description: 'quality score for this dimension (0 = hard fail, 1 = flawless); null when the dimension could not be evaluated at all',
    },
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          severity: { type: 'string', enum: ['critical', 'major', 'minor'] },
          summary: { type: 'string' },
          file: { type: 'string' },
          evidence: { type: 'string', description: 'the command run PLUS its verbatim output — never a summary' },
        },
        required: ['severity', 'summary', 'file', 'evidence'],
        additionalProperties: false,
      },
    },
    evidence_gaps: {
      type: 'array',
      description: 'a check the judge could NOT run, and why (evidence absent != pass)',
      items: { type: 'string' },
    },
  },
  required: ['dimension', 'score', 'findings', 'evidence_gaps'],
  additionalProperties: false,
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

Report only what you can VERIFY from the artifacts. If a field cannot be determined, return it empty
rather than guessing.`

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

// SCRIPT JS ONLY — no agent call sits between judge collection and the returned verdict.
// A judge is "missing" if it did not return or its score is not a finite number.
const scoreOf = (j) => (j && typeof j.score === 'number' && Number.isFinite(j.score)) ? j.score : null

// Null-judge guard FIRST, before any mean computation: 4 dimensions are the contract; 3/4 is no verdict.
const missing = DIMENSIONS.filter((dim, i) => scoreOf(judges[i]) === null)
if (missing.length > 0) {
  return { verdict: 'INCOMPLETE', missing, tier: TIER, threshold: THRESHOLD, spec_id: SPEC_ID }
}

// All four judges returned a finite score. Aggregate their findings/gaps (null-filtered) for the report.
const scores = DIMENSIONS.map((dim, i) => judges[i].score)
const findings = DIMENSIONS.flatMap((dim, i) => (judges[i].findings || []).filter(Boolean).map((f) => ({ dimension: dim, ...f })))
const evidenceGaps = DIMENSIONS.flatMap((dim, i) => (judges[i].evidence_gaps || []).filter(Boolean).map((g) => ({ dimension: dim, gap: g })))

// Zero-score guard: the harmonic mean divides by each score, so a 0 dimension is a hard FAIL naming
// the dimension — never a division by zero, never Infinity.
const zeroScored = DIMENSIONS.filter((dim, i) => scores[i] <= 0)
if (zeroScored.length > 0) {
  return { verdict: 'FAIL', zero_scored: zeroScored, tier: TIER, threshold: THRESHOLD, spec_id: SPEC_ID, findings, evidence_gaps: evidenceGaps }
}

// Harmonic mean n / Σ(1/sᵢ) — in-script, deterministic, auditable. One low dimension drags it down.
const reciprocalSum = scores.reduce((acc, s) => acc + 1 / s, 0)
const harmonicMean = DIMENSIONS.length / reciprocalSum

return {
  verdict: harmonicMean >= THRESHOLD ? 'PASS' : 'FAIL',
  harmonic_mean: harmonicMean,
  threshold: THRESHOLD,
  tier: TIER,
  spec_id: SPEC_ID,
  scores: DIMENSIONS.map((dim, i) => ({ dimension: dim, score: scores[i] })),
  findings,
  evidence_gaps: evidenceGaps,
}
