---
id: SPEC-CODEX-COVER-RESIDUAL-001
title: "Codex coverage residual — runCodexReviewGate RunE wiring tests + (codexSessionHandle).pid nil-guard arm"
version: "0.2.0"
status: draft
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, testing, coverage, review-gate, internal-cli, wiring"
related_specs: [SPEC-CODEX-TEST-GAPS-001, SPEC-MOAI-MCP-SERVER-001]
---

# SPEC-CODEX-COVER-RESIDUAL-001 — Codex coverage residual

## A. History

- 2026-09-07 — v0.1.0 — Card t519 plan-phase. Split out of the lane-7 t501 verdict (`.moai/reports/t501/verdict.md` §Gaps/§Residual-risk): two residual codex surfaces the t501 6-file scan never reached. Tier classified **M**, not S — see §A.1.
- 2026-09-07 — v0.2.0 — Plan-audit iteration 1 fix round (FAIL, score 0.775 against the Tier M threshold of 0.80; 4 blocking, 5 advisory). All four blocking findings were fixture-specification defects sharing one root cause: three Given clauses were written from the *intent* of a sibling test rather than from its *text*, dropping a seam the sibling actually carries. F1 — AC-CCR-004 gained the mandatory `codexLookPath` swap (its verdict was host-dependent without it). F2 — AC-CCR-003 gained `withChangeDetector(t, true)`, without which its sole adoption mutant M2 could not fire. F3 — the M1 ledger row was split, with the new M1b carrying AC-CCR-002's adoption (M1 flips only AC-CCR-001). F4 — the seam-location table was corrected against grep: the helpers live in four files, and `stubCodexRunner` is in `codex_rpc_error_test.go`, a file neither artifact had named. Advisories A2 (no `t.Parallel()`) and A3 (M5a's RED arrives as a panic) applied; A1 is a recommendation to change nothing and was honoured; A4 and A5 declined with reasons recorded in progress.md. Scope, requirements, and the ≥90.0% coverage threshold are unchanged.

### A.1 Tier classification note (deviation recorded)

The dispatch named Tier S in its identity block while the card text allowed "Tier S~M" and the deliverable list requested the **3-file** artifact set (spec.md + plan.md + acceptance.md). Tier M is the resolution, on two independent grounds, both mechanical:

1. **Artifact set.** Tier S is a 2-file set (spec.md + plan.md, AC inline in spec.md §3); Tier M is the 3-file set requested. `spec-workflow.md` § SPEC Complexity Tier binds tier to artifact set.
2. **REQ/AC budget.** This SPEC carries 10 requirements and 12 acceptance criteria. The Tier S ceilings are 8 and 8 (applied independently). Both are exceeded; the Tier M ceilings are 16 and 16.

This is the same correction the precedent SPEC recorded at its own plan-audit iteration 1 (SPEC-CODEX-TEST-GAPS-001 v0.2.0: "Tier S→M (artifact set + REQ/AC counts exceeded the Tier S ceiling)"). Classifying M at authoring time avoids re-deriving it under audit. Consequence: the Tier M plan-audit PASS threshold is 0.80, and the Section A-E delegation template is REQUIRED rather than optional.

## B. Baseline — the card's premise HOLDS on the execution axis

The precedent SPEC (SPEC-CODEX-TEST-GAPS-001 §B) recorded a card premise that was **refuted** on the execution axis: 62 of 114 functions were never named in a test file, yet only 4 of 118 measured at 0.0%. That refutation is a property of that card, not a general law. This card's premise was re-measured on the same two axes and **holds on the execution axis** — the two named surfaces really are under-covered, not merely un-named.

Recording the direction explicitly is the point of this section: an investigator who reads only the precedent could infer that a coverage card is presumptively refuted. Here it is not.

### B.1 Measurement provenance

Measured by the lane on tree `bf779ecf2` (this worktree, this run). **manager-spec did not re-execute this measurement**; the figures below are the lane's observation, cited verbatim, and the run-phase re-measurement (REQ-CCR-009) is what re-establishes them on the final tree.

- Command:

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 700s -coverprofile=.moai/state/verify/t519/cover-baseline.out ./internal/cli/
  ```

- Observed output (verbatim): `ok  	github.com/modu-ai/moai-adk/internal/cli	473.665s	coverage: 80.9% of statements`, rc=0
- `go tool cover -func` total: **81.1%**. The two totals differ because they use different counting modes; both are recorded and neither is reconciled into the other.
- Package-wide, 76 functions sit at 0.0%. They are out of scope (§F).

### B.2 Target rows

Evidence: `.moai/reports/t519/coverage-baseline-targets.txt` (per-function extract), `.moai/reports/t519/coverage-baseline-run.log` (run log), `.moai/state/verify/t519/cover-baseline.out` (profile).

| Function | Location | Baseline | Disposition |
|---|---|---|---|
| `HandleCodexReviewGate` | codex_review_gate.go:66 | 100.0% | covered — out of scope |
| `isBlockVerdict` | codex_review_gate.go:116 | 100.0% | covered — out of scope |
| `hasReviewableChanges` | codex_review_gate.go:125 | 100.0% | covered — out of scope |
| `reviewableFromPorcelain` | codex_review_gate.go:145 | 83.3% | adjacent, NOT named by the card — §F |
| `isRuntimeManagedPath` | codex_review_gate.go:168 | 100.0% | covered — out of scope |
| **`runCodexReviewGate`** | **codex_review_gate.go:183** | **0.0%** | **Axis 1 — REQ-CCR-001..004** |
| `readHookInput` | codex_review_gate.go:207 | 85.7% | adjacent, NOT named by the card — §F |
| `emitHookOutput` | codex_review_gate.go:221 | 100.0% | covered — out of scope |
| `resolveProjectDirFromInput` | codex_review_gate.go:236 | 100.0% | covered — out of scope |
| **`(codexSessionHandle).pid`** | **mcp_codex.go:701** | **66.7%** | **Axis 2 — REQ-CCR-005** |

### B.3 Root cause of the 0.0% (Axis 1)

`internal/cli/multi_review_gate_wiring_test.go` carries a header comment stating its tests "mirror the codex-review-gate wiring tests one-for-one so the two gates stay symmetric". The lane verified that **no such codex wiring tests exist**: the codex side has only `TestCodexReviewGate_SubcommandRegistered` (codex_review_gate_test.go:324, a registration assertion that never invokes the RunE) plus pure-handler tests that call `HandleCodexReviewGate` directly.

The comment therefore describes a symmetry that was asserted but never built — the multi gate got the wiring tests, the codex gate did not, and the RunE has been at 0.0% since. The asymmetry is the finding; this SPEC closes it by building the codex side to the same shape.

## C. Requirements (GEARS)

### REQ-CCR-001 — RunE stdin fail-open arms [Priority High]

**When** the `codex-review-gate` cobra command is executed with a stdin payload `readHookInput` cannot parse — either malformed (`{not json`) or empty — the command shall emit an empty ALLOW (`{}`) on stdout, return no error from `Execute()`, and write a `codex-review-gate:`-prefixed diagnostic to stderr on the malformed arm (codex_review_gate.go:184-189).

### REQ-CCR-002 — RunE happy-path ALLOW without consulting codex [Priority High]

**Where** a temp project's `workflow.yaml` sets `workflow.codex.review_gate.enabled: false` (or the file is absent), **when** the command is executed with a well-formed payload naming that project directory, the command shall emit an empty ALLOW and shall **not** consult the codex binary — exercising the full stdin → `resolveProjectDirFromInput` → `readCodexReviewGateEnabled` → `HandleCodexReviewGate` → stdout chain (codex_review_gate.go:190-191, 201).

### REQ-CCR-003 — RunE handler-error fail-open [Priority High]

**When** `HandleCodexReviewGate` returns a non-nil error (gate enabled, reviewable change present, codex session start failing), the command shall emit an empty ALLOW, return no error from `Execute()`, and write a `codex-review-gate: error:`-prefixed diagnostic to stderr (codex_review_gate.go:193-197).

**While** authoring this test, the test shall assert the **stderr** diagnostic, because both the error arm and the success arm emit byte-identical `{}` on stdout — a stdout-only assertion cannot separate them (see §D.2).

### REQ-CCR-004 — RunE BLOCK propagation [Priority High]

**When** the gate is enabled, a reviewable change is present, and the injected codex session yields a failing review verdict, the command shall emit a JSON object on stdout decoding to `decision: "block"` with a non-empty `reason` — proving the RunE forwards the handler's BLOCK rather than substituting an ALLOW (codex_review_gate.go:201).

### REQ-CCR-005 — `(codexSessionHandle).pid` nil-guard arm [Priority Medium]

**When** a direct method-call test exercises `(*codexSessionHandle).pid` (mcp_codex.go:701-706), the test suite shall assert all three arms: a nil receiver → 0, a handle with a nil `conn` → 0, and a handle whose `conn` is a `*fakeCodexConn` → `fakeCodexConnPID`. The nil-receiver and nil-`conn` arms are a disjunction — `h == nil` must short-circuit before `h.conn` is dereferenced.

### REQ-CCR-006 — documented-skip record (no test) [Priority Medium]

**The SPEC artifact itself** (this file, §D) shall carry the skip record for the surface deliberately left untested, so the next investigator finds the reason rather than re-flagging it. No test is written for a recorded skip in this SPEC's run phase.

### REQ-CCR-007 — zero production diffs [HARD, Priority High]

**While** the run phase executes, the working tree shall carry zero diffs to non-test `.go` files. Every change lands in `*_test.go` files under `internal/cli/`. Raising a coverage number by editing the code under measurement is prohibited outright — including a refactor that would make a surface more testable.

### REQ-CCR-008 — mutant-first discipline [HARD, Priority High]

**When** any test item in this SPEC is authored, it shall name at least one mutant (a specific edit to the production source that the test is claimed to catch), be observed FAIL under that mutant with the verbatim failing output recorded, and be observed PASS with the mutant reverted. **When** a candidate mutant is found not to change the test's verdict, the test item shall record that mutant as vacuous and name a different one — a rising coverage number and a test that catches something are different facts.

**While** any commit lands, the mutant shall already be reverted: `git status --short` shall be clean of production-source modifications at every commit (consistent with REQ-CCR-007).

### REQ-CCR-009 — completion re-measurement [Priority High]

**When** the run phase completes, a fresh coverprofile over `./internal/cli/` — taken with the same env-scrubbed command as §B.1, on the final tree — shall show `runCodexReviewGate` at or above 90.0% and `(codexSessionHandle).pid` at 100.0%.

The `runCodexReviewGate` threshold is deliberately below 100%: the S1 arm (§D.1) is unreachable through the real handler, so a 100% target would be unsatisfiable by correct work. The package-wide figure is recorded for the record but is **not** a requirement — the inherited 80.9% sits below the project's 85% target for reasons this SPEC does not address (§F).

### REQ-CCR-010 — quality-gate constraint [Priority High]

**While** the run phase executes, the touched test files shall pass `go vet`, `golangci-lint run`, and `gofmt` clean, and the full diff shall remain tests-only.

## D. Documented-Skip Record

Deliberately NOT tested by this SPEC (REQ-CCR-006). A future investigator citing this surface as "uncovered" should find this section first.

| ID | Surface | Location | Reason |
|---|---|---|---|
| S1 | the `out == nil` defensive arm | codex_review_gate.go:198-200 | Unreachable through the real handler. `HandleCodexReviewGate` is called directly (not through an injectable seam) and returns a non-nil `*hook.HookOutput` on every one of its six paths — it never returns `(nil, nil)`. The arm is defensive against a future handler change, and reaching it would require introducing a production seam, which REQ-CCR-007 forbids. |

### D.1 Why S1 caps the achievable coverage (derivation, not a measurement)

Go's `-func` percentage is over counted statements. `runCodexReviewGate` (codex_review_gate.go:183-202) carries 13 statements by inspection, of which S1's `out = &hook.HookOutput{}` is one — so the expected ceiling is 12/13 ≈ 92.3%.

This is a **derivation from reading the source, not a measured value**; the same model reproduces the observed `(codexSessionHandle).pid` baseline (3 statements, 1 uncovered → 66.7%), which is why it is recorded rather than merely asserted. The run-phase measurement (REQ-CCR-009) is what establishes the real figure; the AC threshold is set at ≥90.0% so a correct implementation satisfies it whether or not the derivation is exact.

### D.2 The vacuous mutant for REQ-CCR-003 (recorded so it is not re-attempted)

The obvious mutant for the handler-error arm — emitting `out` instead of `&hook.HookOutput{}` at codex_review_gate.go:196 — **does not change stdout**. When `HandleCodexReviewGate` returns an error it returns its `allow` value alongside, so `out` is already the empty ALLOW and both branches serialize to `{}`. A test asserting only stdout would pass under that mutant, which is exactly the vacuous-green shape REQ-CCR-008 exists to catch. The detectable mutants are named in acceptance.md §D.

## E. Constraints

1. Tests only — zero diffs to non-test `.go` files (REQ-CCR-007).
2. All code and comments in English (repo standard, `code_comments: en`).
3. Reuse the existing hermetic seams; do **not** add new production seams or new shared test infrastructure. They are spread across **five** files, all package-level and reusable: `withChangeDetector` (codex_review_gate_test.go); `writeWorkflowYAML` and `assertAllowJSON` (multi_review_gate_wiring_test.go); `withCodexLookPath` / `withCodexRunner` / `withCodexSession` / `codexSessionScript` / `fakeCodexSession` / `fakeCodexConn` / `errFakeCodexCrash` (mcp_codex_test.go); `stubCodexRunner` (codex_rpc_error_test.go); `fakeCodexConnPID` (codex_jobs_test.go). Per-identifier line numbers are in plan.md §A.4.
4. `newGateCmd` (multi_review_gate_wiring_test.go:102) is hard-wired to `runMultiReviewGate` and MUST NOT be modified or reused; the codex side needs its own constructor under a distinct name.
5. Verification is scoped: `go test ./internal/cli/` only, with a timeout budget ≥600s (baseline measured 473.7s). NEVER `go test ./...` locally — CI runs the full suite.
6. The env scrub travels with the command in one compound invocation (`unset … && go test …`); a separate `unset` does not carry into the next Bash call.

## F. Out of Scope

### Out of Scope — adjacent functions the card does not name

- `reviewableFromPorcelain` (83.3%) and `readHookInput` (85.7%) are in the same file and below 100%, but neither is named by the card. Raising them is scope creep, not delivery. They are recorded here as residual risk so the next investigator sees they were considered and declined.

### Out of Scope — the package-wide coverage figure

- The inherited 80.9% package figure sits below the project's 85% package target, and 76 functions across `internal/cli` sit at 0.0%. Closing that gap is a different, much larger card. This SPEC's completion criterion is per-function (REQ-CCR-009), never package-wide.

### Out of Scope — production code changes

- No production code changes of any kind, including refactors that would make a surface more testable or a seam that would make S1 reachable. If a test reveals a production defect, it is reported as a blocker/finding, not fixed in this SPEC's run phase.

### Out of Scope — the multi-review-gate side

- `multi_review_gate_wiring_test.go` and `runMultiReviewGate` are not modified. The header comment there is factually wrong about the codex side existing (§B.3), but correcting a comment in another gate's test file is outside this card's scope; this SPEC closes the asymmetry by building the missing side, not by editing the claim.

### Out of Scope — duplicate registration coverage

- No new subcommand-registration test. `TestCodexReviewGate_SubcommandRegistered` (codex_review_gate_test.go:324) already pins `hook.go:221` (`Use: "codex-review-gate", RunE: runCodexReviewGate`); duplicating it would add a test that asserts nothing new.

## G. Cross-References

- Origin: `.moai/reports/t501/verdict.md` §Gaps / §Residual-risk (card t519's source)
- Baseline evidence: `.moai/reports/t519/coverage-baseline-targets.txt`, `.moai/reports/t519/coverage-baseline-run.log`, `.moai/state/verify/t519/cover-baseline.out`
- Structural precedent: `.moai/specs/SPEC-CODEX-TEST-GAPS-001/` (name-citation vs execution axis, documented-skip record, mutant-per-test discipline)
- Related SPECs: SPEC-MOAI-MCP-SERVER-001 (owns `runCodexReviewGate`, REQ-MCP-008), SPEC-CODEX-TEST-GAPS-001 (the sibling residue SPEC; its REQ-CTG-012 covers `(realCodexConn).pid`, this SPEC's REQ-CCR-005 covers the sibling `(codexSessionHandle).pid`)
- Doctrine: `.claude/rules/moai/development/verification-completeness.md` §1.1 (observed-failure completion, empty-sweep hazard), §2 (two-cell adoption + mutant probe)
