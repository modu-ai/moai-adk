# progress.md — SPEC-DECISION-AUTHORITY-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13T21:28:45+0900
plan_audit_verdict: PASS 0.95 (iteration 2/3, no blocking findings — .moai/reports/t692/plan-audit-iter2.md, tree 4c2ec29f1)
plan_audit_iterations: iter1 FAIL 0.85 (.moai/reports/t692/plan-audit-iter1.md, tree 6732d1461) → fix pass 4c2ec29f1 → iter2 PASS 0.95
advisory_notes: A1/A2/A3 recorded in the iter2 verdict, non-blocking, deliberately not taken at plan-phase
authoring_tree: 62fbd6baf (worktree .claude/worktrees/t692, branch WT-judgment-authority)
audited_artifact_set: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md (Tier L; hash-frozen since verdict 4c2ec29f1)
red_probes: 22 ledger cells (P1-P21, acceptance.md §E) — P1-P19 at 62fbd6baf, P11/P20/P21 re-measured at 6732d1461
kickoff_decision: Implementation Kickoff Approval GRANTED by operator 2026-09-13 (relayed via lead question channel)
operator_decisions_adopted: OD-1 (decision_gate distributed default off; local dogfood on) · OD-2 (product-level reconcile routes to manager-docs) · OD-3 (plan-auditor integration deferred beyond v1) — all three adopted as recommended in spec.md §E.3
run_entry_basis: plan-audit skip-eligible (verdict PASS 0.95 ≥ Tier L 0.85, artifact hash unchanged since 4c2ec29f1; decision record lives in progress.md — outside the ComputeHash subject set — so this record does not invalidate the cached verdict)

## §E.2 Run-phase Evidence

Run-phase executed M1→M5 serially on `WT-judgment-authority`, worktree `.claude/worktrees/t692`. Commits: M1 `3318b7775` (config axis + draft→in-progress), M2 `02c6959ae` (decision-index authoring flow, both trees), M3 `e29c2b535` (kickoff presentation step, both trees), M4 `13486b852` (embed regen + agent emission), M5 this commit (behavioral repro + sweep + this record). Measurement tree for the closing sweep: `13486b852` (M5 adds only this progress record).

### Milestone commit list (E6)

| M | SHA | Subject |
|---|---|---|
| M1 | `3318b7775` | feat(SPEC-DECISION-AUTHORITY-001): M1 config axis — decision_gate field, resolver, defaults |
| M2 | `02c6959ae` | feat(SPEC-DECISION-AUTHORITY-001): M2 decision-index authoring flow |
| M3 | `e29c2b535` | feat(SPEC-DECISION-AUTHORITY-001): M3 kickoff decision-index presentation step |
| M4 | `13486b852` | feat(SPEC-DECISION-AUTHORITY-001): M4 embed regeneration + agent emission |
| M5 | (this commit) | feat(SPEC-DECISION-AUTHORITY-001): M5 behavioral repros + acceptance sweep |

### AC matrix (E1) — all 19 ACs, flipping milestone named

| AC | Status | Evidence (command → observed at closing sweep, tree `13486b852`) | Flipped |
|---|---|---|---|
| AC-DA-001 | PASS | `grep -c decision_gate internal/config/types.go` → 1 (yaml tag); `grep -c DecisionGate internal/config/defaults.go` → 1 (default `DecisionGate: "off"`). Ledger note: the RED cell P1 grepped the yaml KEY on defaults.go (0 at base); post-M1 the yaml key lives in types.go and the Go field name in defaults.go — the same shapes ledger P8 documents as the false-positive-control surface pair. | M1 |
| AC-DA-002 | PASS | `go test -run TestDecisionGate -v ./internal/config/` → 6 test functions swept (floor ≥4), all pass, exit 0 | M1 |
| AC-DA-003 | PASS | `grep -c decision_gate internal/template/templates/.moai/config/sections/interview.yaml` → 1 (`off` + comment); absent-key resolution covered by TestDecisionGateAbsentKeyIsOff | M1 |
| AC-DA-004 | PASS | P5/P12 flip: 6-file swept set, every `decision_gate` occurrence sits in an enclosing block carrying Where-on/Where-off conditioning (manager-spec 1/1, clarity-interview 2/2, spec-assembly 1/1 — local+mirror) | M2 (spec-assembly half M3) |
| AC-DA-005 | PASS | `grep -c decision-index .claude/agents/moai/manager-spec.md` → 1; block states statelessness + gate-on-only authoring | M2 |
| AC-DA-006 | PASS | `POLICY-COVERED` present (manager-spec 3, clarity-interview 1); stray second-vocabulary sweep (`URGENT\|CRITICAL-FLAG\|BLOCKER` across 6 files) → no matches, exit 1 | M2 |
| AC-DA-007 | PASS | `grep -rc "authority anchor"` → manager-spec 1, clarity-interview 2; rows require committed file+section anchor | M2 |
| AC-DA-008 | PASS | `grep -rc "authority register"` → manager-spec 1, clarity-interview 1; register enumerated committed-only (product.md / completed-SPEC HISTORY+Amendments / config sections / constitution; untracked excluded) | M2 |
| AC-DA-009 | PASS | `grep -rc "Never downgrade"` → manager-spec 1, clarity-interview 2; unverifiable anchor → FOUNDER stated | M2 |
| AC-DA-010 | PASS | `grep -c decision-index spec-assembly.md` → 1 (Step 2.3.3b, both trees) | M3 |
| AC-DA-011 | PASS | `grep -c "HARD] The Implementation Kickoff Approval" spec-assembly.md` → exactly 1 (local AND mirror); clause text unchanged (P19 shape) | preserve |
| AC-DA-012 | PASS | `grep -c decision-index .claude/agents/moai/plan-auditor.md` → 0; `git diff --stat 62fbd6baf..HEAD -- .claude/agents/moai/plan-auditor.md` → empty (P21 re-measured at `13486b852`) | preserve |
| AC-DA-013 | PASS | `grep -c NEED_ANALYSIS spec-assembly.md` → 1; vocabulary DECIDE / NEED_ANALYSIS / NEED_EVIDENCE / DEFER in Step 2.3.3b | M3 |
| AC-DA-014 | PASS | `grep -c product.md spec-assembly.md` → 1; manager-docs ownership boundary named in Step 2.3.3b | M3 |
| AC-DA-015 | PASS | `grep -ic "zero judgment" spec-assembly.md` → 1; zero-rows ≠ approval clause present | M3 |
| AC-DA-016 | PASS | `grep -c "Detect → Explain"` → manager-spec 1, clarity-interview 2; no embedded preferred answer in either mode | M2 |
| AC-DA-017 | PASS | `grep -cE "decision_gate.*recommendation_mode\|recommendation_mode.*decision_gate" internal/config/types.go` → 0 (P20 still 0); `TestDecisionGateOrthogonalToRecommendationMode` passes (independence test exists) | M1 |
| AC-DA-018 | PASS | Token presence in BOTH trees per edited file (manager-spec / clarity-interview / spec-assembly × distinctive inserted tokens, local+mirror each); never `diff -q`, never `cp` | M4 |
| AC-DA-019 | PASS | M5 behavioral repro (below) | M5 |

### M5 behavioral repro (AC-DA-019, both passes under `/tmp/t692-repro`)

- **Off-mode** (gate key absent): project `/tmp/t692-repro/off/` with throwaway SPEC-REPRO-001 and an interview.yaml carrying no `decision_gate` key. The authoring precondition ("Where the setting is on") evaluates off → authoring branch skipped. Observed: `test ! -e decision-index.md` → PASS (file not created); no index rows exist to surface, kickoff composition unchanged.
- **On-mode** (gate `on`): project `/tmp/t692-repro/on/` with throwaway SPEC-REPRO-002 and `decision_gate: on` in its interview.yaml. decision-index.md authored per the M2 fixed row shape. Observed: `Label: DECIDED` ×1, `Label: POLICY-COVERED` ×1 (anchor: interview.yaml § decision_gate, verifiable), `Label: EVIDENCE-NEEDED` ×1, `Label: FOUNDER` ×1 — Q4's candidate anchor (`docs/pricing.md § Tiers`) does not exist, so the row escalates to FOUNDER (Never downgrade) instead of routing to DECIDED/POLICY-COVERED.

### E8 — M1 RED evidence (verbatim, before the resolver landed)

`go test -run 'TestDecisionGate' ./internal/config/` at tree `3318b7775^` (= `c388db2fa` + test file, pre-implementation):

```
# github.com/modu-ai/moai-adk/internal/config [github.com/modu-ai/moai-adk/internal/config.test]
internal/config/interview_decision_gate_test.go:26:9: def.DecisionGate undefined (type InterviewConfig has no field or method DecisionGate)
internal/config/interview_decision_gate_test.go:29:16: def.ResolvedDecisionGate undefined (type InterviewConfig has no field or method ResolvedDecisionGate)
internal/config/interview_decision_gate_test.go:46:16: cfg.ResolvedDecisionGate undefined (type *InterviewConfig has no field or method ResolvedDecisionGate)
internal/config/interview_decision_gate_test.go:78:9: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
exit=1
```

(The run was red for the right reason: the five TestDecisionGate* tests reference the field/resolver this SPEC introduces; the build failure is the missing implementation, not a fixture defect.)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-13T22:10:00+0900
run_commit_sha: pending-backfill-m5
run_status: audit-ready
ac_pass_count: 19
ac_fail_count: 0
preserve_list_post_run_count: 2
l44_pre_commit_fetch: not-run (lane worktree; no push performed — lane never pushes per 2026-09-02 operator directive)
l44_post_push_fetch: not-run (same — push is the lead's batch act)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass (go build ./... exit 0)
  windows_amd64: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0)
coverage:
  command: go test -cover ./internal/config/...
  observed: "internal/config 82.0% | atomicfile 81.8% | toolpolicy 89.1%"
  gate_note: "package total 82.0% is BELOW the 85% gate figure and is PRE-EXISTING, not introduced: baseline measured at pre-change commit c388db2fa via a throwaway clone (/tmp/t692-baseline) returned the same 82.0%. The SPEC's new code is fully covered: ResolvedDecisionGate 100.0%, ResolvedRecommendationMode 100.0% (go tool cover -func). Reported as PASS-WITH-DEBT candidate for the auditor's judgment; no coverage regression attributable to this SPEC."
lint:
  command: golangci-lint run --timeout=2m ./internal/config/...
  new_issues: 0
  pre_existing_baseline: "1 — internal/config/cg_migration.go:159 QF1001 staticcheck (present at pre-flight, unchanged)"
total_run_phase_files: 11
m1_to_m5_commit_strategy: "one commit per milestone (M1-M5), conventional subjects, card id in body, staged by explicit pathspec; lane does not push"
```

Preserve-list verification at close (tree `13486b852`): `.claude/agents/moai/plan-auditor.md` byte-identical to `62fbd6baf` (P21 empty diff); `.claude/rules/moai/core/askuser-protocol.md` untouched (P18 control = 4); spec-assembly `[HARD]` gate clause exactly 1 occurrence, text unchanged (P19); `RecommendationMode` field + `ResolvedRecommendationMode()` unchanged and still 100% covered.

Template neutrality (REQ-DA-021): `git diff 3318b7775..13486b852` over the three template files — zero added lines match SPEC-ID / REQ-token / date / SHA / macOS-path / CLAUDE.local patterns (grep exit 1). `make build` green after `make agents-emit` regenerated the machine-emitted `.codex/agents/moai/manager-spec.toml` from the M2 mirror change; `internal/template/...` test suites green.



## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Mode Selection recorded 2026-09-13, after Implementation Kickoff Approval (operator decision relayed via lead), before the first run-phase Agent() spawn.

**Decision: serial**

**Justification summary**: coding-heavy doctrine+config work with a strict M1→M5 dependency chain — one manager-develop, sequential milestones (Anthropic coding-task parallelism caveat); fanout/sweep rejected for shared-surface races and semantic edits.

**Input parameters**: tier L; scope ~8 files (2 Go config + 3 doctrine skill/agent files + 3 template mirrors + 2 local config); domains = 3 (Go config, doctrine markdown, template mirrors); file language mix = markdown-dominant with a small Go config slice; concurrency benefit = LOW (single sequential flow, M1→M4 dependency chain, milestone ordering by decision-reversibility); Agent Teams prereqs = not requested.

**Mode evaluation table**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | not selected | multi-file, cross-domain — not trivial |
| serial | **selected** | coding-heavy + doctrine-edit work with strict milestone ordering (Anthropic coding-task parallelism caveat); one manager-develop carries M1→M5 |
| fanout | not selected | no independent research/read domain; milestones share surfaces and order matters |
| sweep | not selected | semantic doctrine edits, not mechanical-uniform bulk |

**Decision: serial**

**Justification**: The implementation is one ordered chain — M1's config resolver gates M2's flow text, M3's gate enrichment depends on M2's index shape, M4's mirrors apply M2/M3. Parallel spawns would race shared files (manager-spec.md, spec-assembly.md) for no wall-clock gain. Plan-audit skip-eligibility holds (verdict PASS 0.95 ≥ 0.85, hash unchanged since 4c2ec29f1), so Phase 1 re-execution is skipped; Implementation Kickoff Approval was granted by the operator (see §E.1 kickoff_decision).
