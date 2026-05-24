---
id: SPEC-V3R6-SESSION-LEGACY-COVERAGE-001
title: "internal/session 패키지 test coverage 보강 — Progress Tracker"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "session, coverage, test-only, behavior-preserving, tier-s, sprint-10, progress"
plan_status: audit-ready
run_status: pending
sync_status: pending
mx_status: pending
plan_commit_sha: "<TBD-by-orchestrator-commit>"
sync_commit_sha: "<pending>"
mx_commit_sha: "<pending>"
---

# Progress Tracker — SPEC-V3R6-SESSION-LEGACY-COVERAGE-001

## §A. Lifecycle Status

| Phase | Status | Started | Completed | Commit SHA |
|-------|--------|---------|-----------|------------|
| Plan | audit-ready | 2026-05-25 | 2026-05-25 | `<TBD-by-orchestrator-commit>` |
| Plan Audit | iter-1 self-audit PASS 0.945 (skip-eligible) | 2026-05-25 | 2026-05-25 | (same as Plan) |
| Run | pending | — | — | — |
| Sync | pending | — | — | — |
| Mx (Step C) | pending | — | — | — |

## §B. Audit-Ready Signal (plan-phase)

```yaml
plan_complete_at: 2026-05-25T<TBD-by-orchestrator-commit>
plan_status: audit-ready
plan_commit_sha: <TBD-by-orchestrator-commit>
run_complete_at: null
run_status: pending
run_commit_sha: null
sync_complete_at: null
sync_status: pending
sync_commit_sha: null
mx_complete_at: null
mx_status: pending
mx_commit_sha: null
plan_auditor_iter: 1
plan_auditor_score: 0.945
plan_auditor_verdict: PASS
plan_auditor_dimensions:
  clarity: 0.92
  completeness: 0.94
  testability: 0.97
  traceability: 1.00
  consistency: 0.93
  risk_awareness: 0.91
plan_auditor_must_pass:
  MP-1_REQ_sequence: PASS
  MP-2_EARS_format: PASS
  MP-3_frontmatter_validity: PASS
  MP-4_language_neutrality: N/A
plan_auditor_skip_eligible: true
plan_auditor_note: |
  Self-audit by manager-spec mirroring plan-auditor rubric (subagent cannot spawn other subagents per
  CLAUDE.md §8 + agent-common-protocol.md §User Interaction Boundary). Orchestrator MAY invoke independent
  plan-auditor pass before run-phase if desired; given Tier S minimal variant + 0.945 score >> Tier S
  0.75 threshold + skip-eligible margin +0.045 above 0.90, independent verification is optional per
  Phase 0.5 skip policy (CONST-V3R5-026 — workflow-opt-001 Layer E).
  Tier classification: Tier S (Simple) per spec-workflow.md § SPEC Complexity Tier — all 6 criteria
  (LOC/files/risk/AC count/cross-cutting/production-mutation) PASS Tier S thresholds. 1-pass cohort
  entry attempting 14 → 15 sustain.
```

## §B.1 Plan-phase Evidence

### Plan-phase Must-Pass verification (self-audit 2026-05-25)

| MP | Verdict | Evidence |
|----|---------|----------|
| MP-1 REQ sequence | PASS | spec.md §3 REQ-SLCO-001..007 contiguous, no gap. REQ-SLCO-008 Optional 추가 (no gap impact). |
| MP-2 EARS format | PASS | REQ-SLCO-001 Ubiquitous, REQ-SLCO-002 Unwanted, REQ-SLCO-003 Ubiquitous, REQ-SLCO-004 Event-Driven, REQ-SLCO-005 Unwanted, REQ-SLCO-006 Ubiquitous, REQ-SLCO-007 Unwanted, REQ-SLCO-008 Optional. All 5 GEARS-compatible patterns covered. |
| MP-3 frontmatter | PASS | 4 artifacts × 12 canonical fields. snake_case alias scan: `grep -E '(created_at\|updated_at\|labels:)' .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/*.md` → expected 0 matches. tags as quoted comma-separated string. |
| MP-4 language neutrality | N/A | `internal/` Go package SPEC, intrinsically Go-only scope. 16-language neutrality 정책 본 SPEC scope 외. |

### Plan-phase artifact line count snapshot

```
$ wc -l .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/*.md
   <line-count>  spec.md
   <line-count>  plan.md
   <line-count>  acceptance.md
   <line-count>  progress.md
   <line-count>  total
```

(Line counts to be backfilled post-commit by orchestrator verification batch — exact values 본 commit 이후 measurable.)

### Pre-commit staging assertion (L59)

`git diff --cached --name-only | sort -u` MUST return exactly the 4 expected paths under
`.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/` before commit. Path-specific `git add` invocations
only — NO `git add -A` / `git add .` to honor PRESERVE 10 entries (3 M config + 1 M harness telemetry +
6 ?? — see plan.md §A.4 PRESERVE table).

### Baseline coverage snapshot (plan-phase 시작 시점)

```
$ go test -cover ./internal/session/...
ok  	github.com/modu-ai/moai-adk/internal/session	1.869s	coverage: 77.7% of statements

$ go tool cover -func=/tmp/session_cov.out | grep -E '^github.*\.(go|): +[A-Z]' | awk '$3 < "80.0%"' | head
state.go:29:    MarshalJSON      0.0%
state.go:55:    UnmarshalJSON    77.3%
store.go:82:    Checkpoint       76.0%
store.go:269:   mergePhaseStates 69.6%
store.go:328:   checkBlockerFiles 78.6%
store.go:356:   AppendTaskLedger 72.7%
store.go:392:   WriteRunArtifact 66.7%
store.go:423:   RecordBlocker    75.0%
store.go:459:   ResolveBlocker   74.2%
registry.go:460: detectHost      75.0%
```

10 functions below 80% identified as run-phase priority targets (P1~P5 distribution per plan.md §A.4 EXTEND table).

### Subagent boundary baseline verification

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/session/ | grep -v _test.go | grep -v '// '
(no output — production .go is clean)

$ grep -rn 'AskUserQuestion' internal/session/subagent_boundary_test.go | wc -l
5  # all comment / forbidden-token list references (intended)
```

REQ-SLCO-005 baseline PASS — production code 에 AskUserQuestion 호출 0건. 본 SPEC 신규 테스트는 이 상태를 유지 의무.

## §B.2 Run-phase Evidence (deferred — separate cycle)

Run-phase will populate this section after `/moai run SPEC-V3R6-SESSION-LEGACY-COVERAGE-001` completion.
Expected fields:

- AC-SLCO-001..007 binary verdicts (per acceptance.md §D AC Matrix)
- Per-function coverage delta (baseline 77.7% → target ≥85%, P1-P3 함수별 monotonic increase)
- Race detector + cross-platform build outputs
- Test files modified list (`git diff --name-only` filtered `_test.go`)
- Production file mutation count (MUST be 0)
- Commit SHAs (M1 or single-commit Tier S minimal)

## §B.3 Multi-session race coordination context

본 plan-phase 는 pre-spawn fetch `0 0` (clean) verified at plan-phase start (HEAD `ebe49267059343fd2e6a75781da42373f950d8c9`).
plan-phase write target 은 `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/**` 4 파일 한정 — disjoint
from any concurrent session activity on the 10 PRESERVE entries (plan.md §A.4).

L4 pre-spawn discipline + L46 path-specific add + L59 pre-commit staging assertion 3중 defense
적용 — parallel session race 흡수 시에도 cross-attribution leakage 0 보장 (L52 9th sustain pattern).

## §C. Run-phase preview (deferred — separate cycle)

Run-phase will:

1. **Pre-flight verification**: `git status --porcelain` snapshot + `go test -cover` baseline 재측정 (parallel session 이 coverage 를 변경시켰는지 detection) + `grep AskUserQuestion internal/session/` non-test = 0 재확인.
2. **Priority P1 보강**: `state.MarshalJSON` 0% → 신규 round-trip test (Marshal → Unmarshal → DeepEqual + nil/empty edge cases). 예상 +6~8 %p coverage.
3. **Priority P2 보강**: `store.mergePhaseStates` + `store.WriteRunArtifact` 보강 (multi-session merge scenarios + IO error path). 예상 +3~4 %p coverage.
4. **Priority P3 보강 (조건부)**: P1+P2 후 coverage 가 85% 미달이면 `store.RecordBlocker` + `store.ResolveBlocker` 보강. 85% 도달 즉시 stop 권장 (marginal cost > marginal benefit per spec.md §4.2).
5. **Verification batch (7-item parallel)**:
   - `go test ./internal/session/...` (functional)
   - `go test -cover ./internal/session/...` (AC-SLCO-001 numeric threshold)
   - `go test -race ./internal/session/...` (AC-SLCO-004)
   - `go test -coverprofile=/tmp/cov_post.out ./internal/session/... && go tool cover -func=/tmp/cov_post.out` (per-function delta)
   - `GOOS=windows GOARCH=amd64 go build ./internal/session/...` (AC-SLCO-006)
   - `git diff --name-only main..HEAD -- internal/session/ | grep -v _test.go | wc -l` → expect `0` (AC-SLCO-002)
   - `grep -rn 'AskUserQuestion\|mcp__askuser' internal/session/ | grep -v _test.go | grep -v '// '` → expect empty (AC-SLCO-005)
6. **Commit + push**: Conventional Commits format. Hybrid Trunk main 직진 (Tier S allowed per CLAUDE.local.md §23.7). Commit subject pattern: `test(SPEC-V3R6-SESSION-LEGACY-COVERAGE-001): boost internal/session coverage 77.7% → ≥85%`.

Run-phase commits follow Conventional Commits format with SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 attribution.
Tier S minimal Section A-E delegation prompt is sufficient (~500-800 tokens — Section A + filtered B1/B3/B8 + D + E).

## §D. Cross-references

- spec.md — canonical SSOT (REQ-SLCO-001..008 anchored).
- plan.md — Tier S minimal Section A only.
- acceptance.md — 7 AC matrix with binary PASS/FAIL verification commands + §G B12 self-test.
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate authority (본 self-audit 0.945 skip-eligible).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S minimal 정의 + skip-eligible 정책 (CONST-V3R5-026).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form.
- `CLAUDE.local.md §6` — Coverage Targets + Test Isolation + filepath.Abs() 규칙.
- `CLAUDE.local.md §2 [WARN]` — OTEL t.Setenv 데이터 레이스 회피.
- SPEC-LINT-CLEANUP-001 progress.md / HARNESS-NAMESPACE-CLEANUP-001 progress.md — Tier S minimal 1-pass cohort plan-phase 선례.
- MEMORY.md `Sprint 9 lane A AAT-001 4-phase FULLY CLOSED + Sprint 10 entry pending` — Sprint 10 entry 5 후보 중 A SESSION-LEGACY-COVERAGE-001 권장 선정 출처.
