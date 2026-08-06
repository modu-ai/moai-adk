# Progress — SPEC-AUDIT-MULTI-MODEL-001

> Plan-phase artifact. Run/sync-phase evidence (§E.2/§E.3/§E.4) is populated by manager-develop and manager-docs respectively. This file is the era.go-parseable lifecycle surface — the `§E.2`/`§E.3`/`§E.4` heading literals and the `sync_commit_sha` field name are parser-load-bearing (see `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map).

## §A. Status

- **Phase**: plan (iter-1)
- **Tier**: L (5-artifact plan set + progress.md)
- **Branch**: `feat/spec-audit-multi-model` (worktree `.claude/worktrees/audit-multi-model`)
- **HEAD at plan-start**: `f85ff4c3e`
- **Hard dependency**: SPEC-MOAI-MCP-SERVER-001 (status: completed, PR #1378)
- **Plan-auditor verdict**: _<pending — plan-phase audit not yet run>_

## §B. Plan-phase artifact set (Tier L)

| Artifact | Status | Notes |
|---|---|---|
| `spec.md` | authored (iter-1) | 19 REQ (REQ-AMM-001..019), 9 constraints (C1-C9), 5 risks (R1-R5), §C verbatim deferral quote |
| `plan.md` | authored (iter-1) | M0-M7 milestones, 10 anti-patterns (AP-AMM-1..10), Section A-E delegation template ready |
| `acceptance.md` | authored (iter-1) | 25 AC (AC-AMM-001..025) + 7 edge cases (EC-1..EC-7) + Definition of Done + traceability matrix |
| `design.md` | authored (iter-1) | ConvergenceResult data model, parallel-execution model, convergence algorithm, Verification Matrix integration, Stop-hook extension |
| `research.md` | authored (iter-1) | Super-review [R3] + AgentOrchestra [R5] + cross-model adversarial review literature + 18-point integration inventory verified at `f85ff4c3e` |
| `progress.md` | authored (iter-1) | this file — §E skeleton with placeholder headings |

## §C. Pre-flight baseline (plan-phase)

- `moai spec lint` baseline: _<pending — to be captured before plan-audit>_
- `go test ./...` baseline: _<pending — run-phase pre-flight>_
- `golangci-lint run` baseline: _<pending — run-phase pre-flight>_
- §25 template-neutrality CI guard baseline: _<pending — run-phase pre-flight>_

## §D. Decisions locked (M0 design lock)

M0 design-lock recorded 2026-08-06 (run-phase). The five design commitments per `plan.md` §F M0 are confirmed and referenced from here:

1. **`ConvergenceResult` shape** (`design.md` §1) — carries `per_backend_verdicts[]` (one `PerBackendVerdict` per active backend), `overall_verdict` (∈ {pass, fail} — NO new enum), `disagreement_flag` (bool), `residual_risk_note` (string, empty when no disagreement), `fail_open_backends` ([]string of backend names that returned `VerdictInconclusive`). LOCKED.
2. **4-step convergence algorithm** (`design.md` §3 / REQ-AMM-006 #1–#4) — required-FAIL → FAIL; all-required-PASS → PASS; required-split → conservative FAIL + disagreement_flag (NOT a new block category); advisory-only conflict → PASS + disagreement_flag. LOCKED.
3. **Independence-preservation contract** (`design.md` §5 / REQ-AMM-003 / C4) — `claude_verdict` is consumed ONLY by the `converge(...)` synthesis step; the codex/glm fan-out goroutines receive `(target, focus, model, effort)` and NEVER `claude_verdict`. Enforced structurally by the call graph + a run-phase test (EC-6 / AC-AMM-003). LOCKED.
4. **Disagreement = advisory, NOT block policy** (`design.md` §3 / REQ-AMM-007 / C3) — cross-model disagreement is surfaced as `disagreement_flag` + `residual_risk_note` for the Verification Matrix; it NEVER hard-blocks the flow on its own. A required-split resolves conservatively to `overall_verdict = FAIL` via the per-backend required-gate contract (NOT via a new "disagreement-block"). LOCKED.
5. **Sentinel-flip same-commit rule** (R3 / Definition of Done) — `multiConvergenceImplemented` at `internal/cli/mcp_audit.go:31` flips `false → true` IN THE SAME COMMIT as the M1 engine lands; no window where the sentinel lies. LOCKED.

### DQ-1 + DQ-2 resolution (`design.md` §7)

- **DQ-1 (write ConvergenceResult on every call)**: CONFIRMED. The `audit_multi` fan-out writes `ConvergenceResult` to `.moai/state/audit-multi/<session>.json` on every call, so the M5 multi-review-gate Stop hook reads the most recent result rather than re-invoking convergence (re-invoking would double audit cost + risk divergence between the in-session verdict and the gate-time verdict). The persistence helper lands at M1 alongside the engine; the state path is session-scoped + gitignored (`.moai/state/` is already in the local-only list per CLAUDE.local.md §2).
- **DQ-2 (claude_verdict absent → refuse)**: CONFIRMED. When `claude_verdict` is absent, the engine REFUSES to proceed by returning a structured `ConvergenceResult{overall_verdict: fail, residual_risk_note: "claude_verdict anchor missing — fail-open direction preserved (cannot synthesize over secondary-only verdicts)"}`. `claude_verdict` is the always-available anchor per the fail-open identity; proceeding without it would invert the fail-open direction (synthesizing over secondary-only verdicts when the anchor that guarantees a claude fallback is missing). The refusal is a structured result, NEVER a hard error — fail-open identity preserved both ways.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-06
plan_artifact_count: 6
plan_tier: L
plan_req_count: 19
plan_ac_count: 25
plan_iter: 1
hard_dependency: SPEC-MOAI-MCP-SERVER-001
sentinel_flip_target: internal/cli/mcp_audit.go:31 (multiConvergenceImplemented: false → true, in the same commit as the M1 engine — Definition of Done sentinel-flip same-commit invariant)
fixed_user_decisions:
  - tier: L single SPEC (full scope: convergence engine + policy + plan/sync wiring + goal gate)
  - disagreement_policy: advisory + residual-risk (NOT hard block); fail-open identity preserved
boundary: spec.md §C quotes AC-MCP-012 + AC-MCP-017 verbatim (the deferral this SPEC closes)
```

## §E.2 Run-phase Evidence

### AC matrix (M3–M7 round; M0–M2 AC-AMM-001..011 covered by the prior commit `a537d28e8`)

| AC | Status | Verification command | Observed output (tail) |
|----|--------|----------------------|------------------------|
| AC-AMM-012 (audit_multi registered with schema) | PASS | `go test -run TestAuditMulti_RegisteredWithSchema_AC_AMM_012 ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli` — tool name + claude_verdict/target/focus in inputSchema |
| AC-AMM-013 (thin wrapper, no backend reimpl) | PASS | `go test -run TestAuditMulti_DelegatesToRunMultiAudit_AC_AMM_013 ./internal/cli/` | `ok` — recordingCaller recorded ≥1 secondary call ⇒ delegation through runMultiAudit verified |
| AC-AMM-014 (per-auditor audit_gate respected) | PASS | `go test -run TestAuditMulti_RespectsCodexGateOff_AC_AMM_014 ./internal/cli/` | `ok` — codex not invoked + no codex entry in per_backend_verdicts when gate=off |
| AC-AMM-015 (plan-auditor Skill path) | PASS | `ls .claude/skills/moai-ref-cross-model-audit/SKILL.md` | file present + canonical `mcp__moai__audit_multi` reference in body |
| AC-AMM-016 (sync-auditor Skill path, same skill) | PASS | (single skill — both audit entry points load it; body declares "the single skill both audit entry points load — no duplication") | same file as AC-AMM-015 |
| AC-AMM-017 (independence rule verbatim) | PASS | `grep -A2 "Independence rule" .claude/skills/moai-ref-cross-model-audit/SKILL.md` | rule quoted verbatim ("Pass only the synthesized claude_verdict object...") |
| AC-AMM-018 (opt-in + 900s timeout + BranchGuard pattern) | PASS | `grep DefaultMultiReviewGateTimeout internal/config/defaults.go` + `grep MultiConfig internal/config/types.go` + reader `readMultiReviewGateEnabled` fail-CLOSED | `DefaultMultiReviewGateTimeout = 900 * time.Second` + `Multi.ReviewGate.Enabled` default false |
| AC-AMM-019 (self-gate prevents false blocks) | PASS | `go test -run TestMultiReviewGate_NoEditTurnAllows ./internal/cli/` | `ok` — no-edit turn ALLOWs, no state file read |
| AC-AMM-020 (gate verdict follows policy) | PASS | `go test -run "TestMultiReviewGate_(AllRequiredPassAllows|RequiredFailBlocks|AdvisoryDisagreementNeverBlocks)" ./internal/cli/` | `ok` — all-pass ALLOW, required-fail BLOCK, advisory NEVER BLOCK |
| AC-AMM-021 (fail-open to claude) | PASS | `go test -run TestMultiReviewGate_AllSecondariesInconclusive_FailOpenClaude ./internal/cli/` | `ok` — all-secondaries-inconclusive + claude pass ⇒ ALLOW |
| AC-AMM-022 (Skill mirrored + make build + §25) | PASS | `make build` + `grep moai-ref-cross-model-audit internal/template/catalog.yaml` + `go test -run TestTemplateNeutralityAudit ./internal/template/` | catalog hash `edbe85ce8a4f55a694d8d47065c20ad4ba8b7154cd59beb9efcf259d2f5ca0b4`; neutrality PASS |
| AC-AMM-023 (canonical MCP tool reference + verbatim rule) | PASS | `grep mcp__moai__audit_multi .claude/skills/moai-ref-cross-model-audit/SKILL.md` | canonical name present; independence rule quoted verbatim |
| AC-AMM-024 (subagent boundary, no AskUserQuestion) | PASS | `go test -run TestAuditMulti_NoHardErrorPath_AC_AMM_024 ./internal/cli/` + `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_audit_multi.go internal/cli/mcp_convergence.go internal/cli/multi_review_gate.go` | `ok` + grep returns 0 matches |
| AC-AMM-025 (hardcoding prevention) | PASS | `grep -n "DefaultMultiReviewGateTimeout\|MultiConfig\|MultiReviewGateConfig" internal/config/defaults.go internal/config/types.go` | thresholds in defaults.go, config-block in types.go reusing the `workflow.codex.review_gate` structural pattern (sibling `multi.review_gate` key) |

### Cross-platform build

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### Subagent boundary grep (C-HRA-008 family)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_audit_multi.go internal/cli/mcp_convergence.go internal/cli/multi_review_gate.go
(no matches — handler returns structured results only; orchestrator translates)
```

### Independence regression guard (EC-6)

```
$ go test -run TestAuditMulti_ClaudeVerdictNeverInSecondaryPayload_AC_AMM_003 ./internal/cli/
ok  github.com/modu-ai/moai-adk/internal/cli
```

### Sentinel-flip same-commit invariant (R3)

`multiConvergenceImplemented` at `internal/cli/mcp_audit.go:31` was flipped `false → true` IN THE SAME COMMIT as the M1 engine (commit `a537d28e8`, prior round). No further sentinel work needed in this round; the invariant holds.

### §25 template-neutrality (CI guard)

```
$ go test -run TestTemplateNeutralityAudit ./internal/template/
ok  github.com/modu-ai/moai-adk/internal/template   1.372s
```

The new skill body carries NO SPEC-ID, NO REQ-AMM token, NO commit SHA, NO internal date — verified by manual scan (`grep -lE 'SPEC-|REQ-AMM|a537d28e8|2026-08-0'` returns nothing) AND by the in-repo `TestTemplateNeutralityAudit` CI guard.

### RED failure output (TDD evidence — M3 + M5)

M3 RED (captured BEFORE `handleAuditMulti` was implemented):
```
$ go test -run TestAuditMulti ./internal/cli/
internal/cli/mcp_audit_multi_test.go:63:9: undefined: handleAuditMulti
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

M5 RED (captured BEFORE `HandleMultiReviewGate` was implemented):
```
$ go test -run TestMultiReviewGate ./internal/cli/
internal/cli/multi_review_gate_test.go:125:12: undefined: HandleMultiReviewGate
internal/cli/multi_review_gate_test.go:138:12: undefined: HandleMultiReviewGate
internal/cli/multi_review_gate_test.go:153:12: undefined: HandleMultiReviewGate
... (8 references total)
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

### Pre-existing flake isolation (not a regression)

`TestNavigatorEnrich_AtomicWriteBarrier` (internal/cli/navigator_enrich_test.go) flaked once in the first full-suite run (`barrier file not created (goroutine did not reach barrier)`) and passed on re-run with `-count=3`. The test uses a goroutine race that is timing-sensitive and unrelated to multi-model audit work; no production code in this round touches the navigator/astx package.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-07"
run_commit_sha: "42c7bcfdd"   # backfilled (self-referential-hazard workaround — a commit does not know its own SHA until it lands)
run_status: "PASS"
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: true
l44_post_push_fetch: pending-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  linux_amd64: pending-ci
  windows_amd64: pass  # local: GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 12  # 5 new Go files + 1 wrapper + 1 SKILL.md + catalog + 5 test pin updates
m1_to_mN_commit_strategy: "single squash commit carrying M3–M7 (sentinel already flipped in prior commit)"
```

Full-suite green: `go test ./... → exit 0`. Lint clean: `golangci-lint run → 0 issues`. `go vet ./... → exit 0`.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates the sync-phase close signal (sync_commit_sha, implemented → completed transition, 3-phase close) here>_

## §F. Phase 4 Mode Selection

_<pending run-phase — orchestrator logs the Phase 4 mode-selection decision here before the first implementation `Agent()` spawn, per `.claude/rules/moai/workflow/orchestration-mode-selection.md` §D. (Tier L + multi-domain Go/Skill/CI-guard scope → likely Mode 5 sequential sub-agent per Anthropic's coding-task parallelism caveat; the orchestrator decides at run-phase entry.)>_

## §G. Open items / Blockers

- _<none at plan-phase — the two DQs in `design.md` §7 are M0 design-lock decisions, not blockers; the two fixed user decisions settle the scope and policy questions>_

## §H. Recursive Self-Diagnosis Log

_<pending run-phase — populated if the run-phase hits a mechanical failure that triggers the DIAGNOSE-PATCH-VERIFY loop (per `manager-develop-prompt-template.md` §E)>_

## §I. Token Accounting

_<pending sync-close — per-SPEC token spend measurement populated by the token-accounting mechanism at sync-close>_
