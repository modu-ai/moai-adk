# Design — SPEC-AUDIT-MULTI-MODEL-001

> Architecture for the multi-model audit convergence layer. Covers: the `ConvergenceResult` data model, the parallel-execution model (`errgroup` over active backends), the 4-step convergence algorithm with disagreement-as-advisory policy, Verification Matrix integration, the super-review independence mechanism, and the multi-review-gate Stop-hook extension. This is a WHAT/WHY design doc — implementation (HOW: exact struct field names, errgroup vs goroutine choice, state-file format) is deferred to run-phase per the SPEC scope boundary (spec.md §B). Integration points are verified in `research.md`.

## §1. Data model — `ConvergenceResult`

```
ConvergenceResult {
  per_backend_verdicts: [            // one entry per active backend (claude + codex + glm per their gates)
    {
      backend:    "claude" | "codex" | "glm",
      gate:       "off" | "advisory" | "required",
      verdict:    "pass" | "fail" | "inconclusive" | ...,   // existing review-output.schema.json values; NO new VerdictDisagreement
      summary:    string,
      findings:   [ ... ],          // the review-output.schema.json findings array
      next_steps: [ ... ],
    },
    ...
  ],
  overall_verdict:     "pass" | "fail",   // derived by the algorithm in §3
  disagreement_flag:   bool,              // true if any two backends conflict
  residual_risk_note:  string,            // human-readable description of the disagreement (empty when disagreement_flag is false)
  fail_open_backends:  [ "codex", "glm" ], // backends that returned VerdictInconclusive (missing/unauthenticated/error)
}
```

**Why this shape (design decisions locked at M0):**

1. **Per-backend entries preserve the raw verdicts.** The orchestrator (and the Verification Matrix) needs to see WHICH backend disagreed, not just that a disagreement happened. Collapsing to a single verdict upstream would lose this audit trail.
2. **`overall_verdict` is one of the existing `pass`/`fail` values — NO new `VerdictDisagreement` enum.** Disagreement is a boolean flag (`disagreement_flag`), not a verdict value. This preserves the `review-output.schema.json` SSOT and avoids a schema-breaking change (REQ-AMM-008 / AC-AMM-011).
3. **`residual_risk_note` is a plain string.** It is what the orchestrator surfaces in the Verification Matrix residual-risk row (REQ-AMM-007). Keeping it a plain string (not a structured object) lets the synthesis step describe the disagreement in natural language, which is what a human reader of the Completion Report actually needs.
4. **`fail_open_backends` is an explicit list.** Fail-open is mandatory (C2), but making the fail-open backends explicitly visible in the result (rather than implicit via their `verdict == inconclusive`) gives the orchestrator a clean surface to surface "codex was unavailable, fell back to claude + glm" without re-deriving it from `per_backend_verdicts[]`.

## §2. Parallel-execution model

```
audit_multi(claude_verdict, target, focus) invoked
  │
  ▼
read config: audit_model (MUST be multi), per-auditor audit_gate
  │
  ▼
filter active backends: those with gate != off
  │   (claude is ALWAYS active — its verdict is the in-session claude_verdict input)
  │
  ▼
errgroup.WithContext(ctx)                          ── already in go.mod
  │
  ├─ goroutine: codex_audit(target, focus, model, effort)   ── via existing mcp_codex.go handler
  ├─ goroutine: glm_audit(target, focus, model, effort)     ── via existing mcp_glm.go handler
  │
  ▼
errgroup.Wait()
  │   each goroutine recovers from panic / error and returns VerdictInconclusive (fail-open)
  │
  ▼
assemble per_backend_verdicts[] = [claude_verdict, codex_result, glm_result]
  │
  ▼
converge(per_backend_verdicts[], gates) → ConvergenceResult   ── §3 algorithm
```

**Why `errgroup` (design decision):**

- Already in `go.mod` (used by `internal/cli/` and `internal/web/` callers — no new dependency).
- Provides `errgroup.WithContext` for deadline propagation and early-cancel on the first hard error (though here we treat all backend errors as `VerdictInconclusive`, so the cancel-on-error semantics are not load-bearing — but the context-deadline path is).
- Idiomatic Go for "N independent operations in parallel, collect all results"; cleaner than hand-rolled goroutines + `sync.WaitGroup` + a results channel.

**Independence-preservation mechanism (§5 detail):**

The `claude_verdict` input parameter is consumed ONLY by the `converge(...)` synthesis step at the bottom of the diagram. It is NEVER passed as an argument to `codex_audit(...)` or `glm_audit(...)`. The goroutines receive `(target, focus, model, effort)` — the same arguments a single-backend codex/glm audit would receive. This is the mechanical enforcement of super-review independence (REQ-AMM-003 / C4): the secondary backends cannot read the claude analysis, so they produce uncorrelated second opinions.

A run-phase test asserts this by instrumenting the backend calls (or by inspecting the handler call signatures) — `claude_verdict` MUST NOT appear in the codex/glm call argument paths (AC-AMM-003, EC-6).

## §3. Convergence algorithm (disagreement = advisory, NOT block)

```
converge(per_backend_verdicts, gates):
  required_backends    = [v for v in per_backend_verdicts if gates[v.backend] == "required"]
  required_passes      = [v for v in required_backends    if v.verdict == "pass"]
  required_fails       = [v for v in required_backends    if v.verdict == "fail"]
  required_inconclusiv = [v for v in required_backends    if v.verdict == "inconclusive"]

  # Step 1: overall_verdict derivation (REQ-AMM-006)
  if required_fails:
    overall_verdict = "fail"
  elif required_passes and len(required_passes) == len(required_backends):
    overall_verdict = "pass"          # all required PASS, no required FAIL
  else:
    # required_backends all inconclusive (all non-Claude missing + claude not yet produced)
    # OR mixed pass/inconclusive — fail-open to claude (AC-AMM-021)
    claude_v = find(per_backend_verdicts, backend == "claude")
    overall_verdict = claude_v.verdict    # fall back to claude

  # Step 2: disagreement_flag derivation
  distinct_required_verdicts = set(v.verdict for v in required_backends if v.verdict in {"pass", "fail"})
  disagreement_flag = len(distinct_required_verdicts) > 1

  # also flag advisory-vs-required conflicts (AC-AMM-009)
  for v in per_backend_verdicts:
    if gates[v.backend] == "advisory" and v.verdict in {"pass", "fail"}:
      if v.verdict not in distinct_required_verdicts and distinct_required_verdicts:
        disagreement_flag = True

  # Step 3: residual_risk_note (plain string — empty when no disagreement)
  residual_risk_note = describe_disagreement(per_backend_verdicts, gates) if disagreement_flag else ""

  return ConvergenceResult{per_backend_verdicts, overall_verdict, disagreement_flag, residual_risk_note, fail_open_backends}
```

**The 4 policy cases (REQ-AMM-006 #1–#4):**

| Case | required_backends | overall_verdict | disagreement_flag | gate block? (REQ-AMM-014) |
|---|---|---|---|---|
| #1 — all required PASS | all `pass` | `pass` | `false` | no (ALLOW) |
| #2 — any required FAIL (no split) | all `fail` or `fail`+`inconclusive` | `fail` | `false` (if all agree on fail) or `true` (if mixed) | yes (BLOCK) |
| #3 — required split (disagreement) | some `pass`, some `fail` | `fail` (conservative) | `true` | yes (BLOCK — conservative per REQ-AMM-006 #2/#3) |
| #4 — advisory-only conflict | all required `pass`, one advisory `fail` | `pass` | `true` | no (ALLOW — advisory never blocks) |

**Key point — disagreement = advisory, NOT block (the fixed user decision, C3):**

Case #3 (required split) DOES block — but it blocks because the required-gate contract per-backend already mandates it (one required backend FAILing blocks), NOT because "disagreement" was promoted to a new block category. The disagreement is surfaced as `disagreement_flag = true` + `residual_risk_note` so the Verification Matrix can show it as residual-risk + advisory.

Case #4 (advisory-only conflict) NEVER blocks — `overall_verdict = pass`, the advisory `fail` is recorded, `disagreement_flag = true`, and the orchestrator surfaces the residual-risk + advisory without interrupting the autonomous flow (AC-AMM-009, AC-AMM-010).

This is the operationalization of the user decision: cross-model disagreement is INFORMATION (surfaced as residual-risk), not a GATE (block). The per-backend `audit_gate: required` remains the sole gate authority.

## §4. Verification Matrix integration

The orchestrator's Completion Report / Verification Matrix (per `verification-claim-integrity.md` §1.1 surface 1 + `.claude/output-styles/moai/moai.md` §8) gains a residual-risk row when `disagreement_flag == true`:

```
| Dimension | Claim | Evidence | Baseline | Gaps | Residual-risk |
|---|---|---|---|---|---|
| Cross-model audit | overall_verdict = PASS | [ConvergenceResult JSON] | [command + output] | [none] | ⚠ disagreement: codex=FAIL (advisory), claude+glm=PASS (required) — surfaced as advisory, NOT a block |
```

**Why this is the right integration surface:**

- The Verification Matrix is ALREADY the orchestrator's structured report for per-claim evidence + gaps + residual-risk.
- Adding a residual-risk row for disagreement reuses the existing surface — no new report format, no new banner.
- The residual-risk row's "advisory, NOT a block" framing is exactly what the user decision (C3) requires: the disagreement is visible to the human reader without interrupting the autonomous flow.

## §5. Super-review independence mechanism (REQ-AMM-003 / C4)

The Drew Hyde [R3] super-review pattern requires: Claude primary → independent secondary (Codex) → orchestrator synthesis, where the secondary does NOT see the primary's analysis (else it becomes a correlated re-sample, not a second opinion).

**Mechanical enforcement in this design:**

1. The `audit_multi` MCP tool takes `claude_verdict` as an INPUT parameter — it is provided BY the auditor agent (the claude-in-session verdict), not produced by a separate claude MCP call.
2. `claude_verdict` is passed ONLY to the `converge(...)` synthesis function (§3).
3. The `codex_audit(...)` and `glm_audit(...)` goroutines receive `(target, focus, model, effort)` — they do NOT receive `claude_verdict`.
4. A run-phase test (`TestConvergence_IndependenceClaudeVerdictNotInSecondaryPayload`) asserts this by inspecting the backend call payloads.

This is stronger than a prose rule — it is a structural property of the call graph. A future edit that accidentally threaded `claude_verdict` into the codex/glm arguments would fail the test.

## §6. Multi-review-gate Stop-hook extension (Path C)

The `moai hook multi-review-gate` Stop hook extends the codex-review-gate pattern to multi-model:

```
Stop hook fires (workflow.multi_review_gate.enabled = true)
  │
  ├─ previous turn = NO code edit / status report / review-result
  │     └─→ ALLOW immediately (self-gate — REUSED from codex_review_gate.go via withChangeDetector)
  │
  └─ previous turn = code edit
        │
        ▼
     read most recent ConvergenceResult   ── from .moai/state/audit-multi/<session>.json (written by audit_multi tool)
        │
        ▼
     apply convergence policy (§3):
        ├─ all required PASS                                   ──→ ALLOW
        ├─ any required FAIL (case #2 or #3)                   ──→ BLOCK
        ├─ advisory-only conflict (case #4)                    ──→ ALLOW (advisory, NOT block)
        └─ all non-Claude missing (fail-open to claude)        ──→ ALLOW iff claude_verdict == pass, else BLOCK
```

**What is REUSED from `codex_review_gate.go` (no new heuristic):**

- The `withChangeDetector` seam (the self-gate that detects "did the previous turn produce a code edit?").
- The ALLOW/BLOCK contract.
- The 900 s timeout override mechanism.
- The opt-in `workflow.<name>.review_gate.enabled` BranchGuard pattern.

**What is NEW:**

- The handler reads a `ConvergenceResult` instead of a single backend verdict.
- The convergence policy (§3) replaces the single-verdict ALLOW/BLOCK decision.
- The `fail-open to claude` branch (REQ-AMM-015) is new — when all non-Claude backends are missing, the gate falls back to claude-only.

## §7. Open design questions (none blocking — recorded for transparency)

- **DQ-1**: Should the `audit_multi` tool write the `ConvergenceResult` to `.moai/state/audit-multi/<session>.json` on every call (so M5's Stop hook can read it), OR should the Stop hook re-invoke convergence? (Lean: write on every call — the Stop hook reads the most recent result; re-invoking would double the audit cost and risk divergence between the in-session verdict and the gate-time verdict.) Confirmed at M0.
- **DQ-2**: When the `claude_verdict` input is absent (the auditor agent did not produce one), should `audit_multi` refuse, or proceed with codex/glm only and synthesize a placeholder? (Lean: refuse with a structured error — `claude_verdict` is the always-available anchor per the fail-open identity; proceeding without it would invert the fail-open direction. The tool returns a structured `ConvergenceResult` with `overall_verdict = fail` and a `residual_risk_note` explaining the missing claude anchor.) Confirmed at M0.

## §8. Cross-references

- `spec.md` — REQ-AMM-001..019 (requirements), §C (verbatim deferral quote from MOAI-MCP-SERVER), §E (constraints C1-C9), §G (risks R1-R5).
- `plan.md` — milestones M0-M7 (M0 design lock; M1 engine + sentinel flip; M2 algorithm tests; M3 MCP tool; M4 cross-model Skill; M5 Stop hook; M6 template mirror; M7 closure), anti-patterns AP-AMM-1..10.
- `acceptance.md` — AC-AMM-001..025 + edge cases EC-1..EC-7 + Definition of Done.
- `research.md` — super-review pattern [R3], AgentOrchestra [R5], cross-model adversarial review literature, verified integration-point inventory.
- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 (Codex 감사 위임), §3.6 v3 extension, Q1 routing paths.
- Hard dependency: SPEC-MOAI-MCP-SERVER-001 (completed, PR #1378) — `design.md` §3 (full tool table), §4 (auth/gate/fail-open state machine — the fail-open identity this SPEC inherits).
- Verification Matrix surface: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 + `.claude/output-styles/moai/moai.md` §8.
