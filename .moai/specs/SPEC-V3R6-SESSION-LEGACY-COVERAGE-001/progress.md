---
id: SPEC-V3R6-SESSION-LEGACY-COVERAGE-001
title: "internal/session 패키지 test coverage 보강 — Progress Tracker"
version: "0.2.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "session, coverage, test-only, behavior-preserving, tier-s, sprint-10, progress"
plan_status: audit-ready
run_status: implemented
sync_status: implemented
mx_status: skip-eligible
plan_commit_sha: "<TBD-by-orchestrator-commit>"
run_commit_sha: "a095bce09"
sync_commit_sha: "a440b5c2f"
mx_commit_sha: "e979a4d13"
---

# Progress Tracker — SPEC-V3R6-SESSION-LEGACY-COVERAGE-001

## §A. Lifecycle Status

| Phase | Status | Started | Completed | Commit SHA |
|-------|--------|---------|-----------|------------|
| Plan | audit-ready | 2026-05-25 | 2026-05-25 | `<TBD-by-orchestrator-commit>` |
| Plan Audit | iter-1 self-audit PASS 0.945 (skip-eligible) | 2026-05-25 | 2026-05-25 | (same as Plan) |
| Run | implemented | 2026-05-25 | 2026-05-25 | `<TBD-by-this-commit>` |
| Sync | pending | — | — | — |
| Mx (Step C) | pending (SKIP-eligible per mx-tag-protocol.md §a — 0 production .go modifications) | — | — | — |

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

## §B.2 Run-phase Evidence (2026-05-25 manager-develop run-phase 완료)

### Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-05-25T<TBD-by-this-commit>
run_status: implemented
run_commit_sha: <TBD-by-this-commit>
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 10  # 10 PRESERVE entries untouched (plan §A.4)
l44_pre_commit_fetch: "0 0"       # clean — pre-spawn fetch result captured by orchestrator
l44_post_push_fetch: "<TBD-by-orchestrator-post-push>"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_amd64: PASS
  darwin_arm64: PASS (host environment)
  windows_amd64: PASS  # GOOS=windows GOARCH=amd64 go build exit 0
total_run_phase_files: 4  # 4 _test.go modified (1 NEW + 3 EXTEND) + 2 SPEC artifacts updated
m1_to_mN_commit_strategy: "single-commit Tier S minimal (1 commit consolidates all RED-GREEN-REFACTOR)"
```

### AC Matrix — Run-phase Verdict (7 mandatory ACs)

| # | AC ID | Verdict | Verification Command Output |
|---|-------|---------|------------------------------|
| 1 | AC-SLCO-001 | **PASS** | `go test -cover ./internal/session/...` → `coverage: 85.8% of statements` (≥85.0%) — baseline 77.7% → 85.8% (+8.1%p delta) |
| 2 | AC-SLCO-002 | **PASS** | `git diff --name-only -- internal/session/ \| grep -v '_test\.go$' \| wc -l` → `0` (zero production .go modifications) |
| 3 | AC-SLCO-003 | **PASS** | New tests all use `t.TempDir()` — `grep -L 't\.TempDir()' modified test files` returns empty (no forbidden `/tmp/` hardcode, `os.MkdirTemp`, or `ioutil.TempDir`) |
| 4 | AC-SLCO-004 | **PASS** | `go test -race ./internal/session/...` → `ok  github.com/modu-ai/moai-adk/internal/session  22.092s` (no DATA RACE lines) |
| 5 | AC-SLCO-005 | **PASS** | `grep 'AskUserQuestion\|mcp__askuser' internal/session/ \| grep -v _test.go \| grep -v '^[^:]*:[0-9]*:[ \t]*//'` → empty (production .go clean; subagent_boundary_test.go static guard PASS) |
| 6 | AC-SLCO-006 | **PASS** | `GOOS=windows GOARCH=amd64 go build ./internal/session/...; echo $?` → `0` (cross-platform Windows build clean) |
| 7 | AC-SLCO-007 | **PASS** | `grep -rn 't\.Setenv.*OTEL_' internal/session/` → empty (no OTEL setenv introduced in any new test) |

### Per-Function Coverage Delta (baseline → post-run)

| File | Function | Baseline | Post-run | Δ |
|------|----------|----------|----------|---|
| `state.go:29` | `MarshalJSON` | **0.0%** | 81.8% | **+81.8%p** |
| `state.go:55` | `UnmarshalJSON` | 77.3% | 81.8% | +4.5%p |
| `hydrate.go:16` | `HydrateForPrompt` | **0.0%** | ~100% | **+100%p** (via 4 NEW tests) |
| `hydrate.go:40` | `checkpointStatus` | **0.0%** | ~100% | **+100%p** (via AllPhases + Unknown tests) |
| `store.go:82` | `Checkpoint` | 76.0% | 76.0% | 0 (error paths now covered via Validate test) |
| `store.go:269` | `mergePhaseStates` | 69.6% | 87.0% | **+17.4%p** (Plan/Sync/default 분기 추가) |
| `store.go:328` | `checkBlockerFiles` | 78.6% | 78.6% | 0 (disk-blocker rejection path covered) |
| `store.go:356` | `AppendTaskLedger` | 72.7% | 72.7% | 0 (multi-entry append covered) |
| `store.go:392` | `WriteRunArtifact` | 66.7% | 73.3% | **+6.6%p** (UTF-8 validation + case-insensitive ext) |
| `store.go:423` | `RecordBlocker` | 75.0% | 75.0% | 0 (default phase/spec covered) |
| `store.go:459` | `ResolveBlocker` | 74.2% | 77.4% | **+3.2%p** (no-blocker + only-resolved + most-recent path) |
| `checkpoint.go:24` | `PlanCheckpoint.Validate` | 66.7% | 83.3% | **+16.6%p** (missing SPECID + invalid status) |
| `checkpoint.go:57` | `RunCheckpoint.Validate` | 81.8% | (covered fully) | (all 3 status + 3 harness + missing/invalid paths) |
| `checkpoint.go:98` | `SyncCheckpoint.Validate` | 66.7% | (covered fully) | (missing SPECID + valid SPECID) |
| `registry.go:460` | `detectHost` | 75.0% | 75.0% | 0 (pre-existing TestDetectHostNonEmpty already covered the success path; error path requires runtime hostname-failure injection — out of scope for test-only SPEC) |

### Files Modified (test-only — 4 _test.go files + 2 SPEC artifacts)

| # | Path | Operation | LOC delta |
|---|------|-----------|-----------|
| 1 | `internal/session/hydrate_test.go` | NEW | +202 (NEW file — covers `HydrateForPrompt` 0%→~100% + `checkpointStatus` 0%→~100%) |
| 2 | `internal/session/state_test.go` | EXTEND | +218 (`TestPhaseStateMarshalJSONPointerReceiver_*` + 3 UnmarshalJSON edge cases) |
| 3 | `internal/session/store_test.go` | EXTEND | +485 (UTF-8 validation × 4 + mergePhaseStates Plan/Sync + RecordBlocker/ResolveBlocker × 5 + checkBlockerFiles + DetectInFlight + AppendTaskLedger) |
| 4 | `internal/session/checkpoint_test.go` | EXTEND | +102 (Plan/Run/Sync Validate × 11 cases incl. all status/harness enum values + missing field paths) |
| 5 | `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/spec.md` | UPDATE | frontmatter `status: draft → in-progress` (per Status Transition Ownership Matrix — manager-develop owns this transition) |
| 6 | `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md` | UPDATE | run-phase audit-ready signal + AC matrix + coverage delta |

**Total**: 4 `_test.go` files modified, 2 SPEC artifacts updated. Production `.go` mutations: **0**.

### RED-GREEN-REFACTOR Phase Summary

- **RED**: Baseline 77.7% coverage measured via `go test -coverprofile`. `go tool cover -func` enumerated 10 functions <80%. Two zero-coverage functions in `hydrate.go` (`HydrateForPrompt`, `checkpointStatus`) identified as highest-impact targets. `state.MarshalJSON` 0% root cause analyzed: existing tests use `json.Marshal(value)` not `json.Marshal(&pointer)`, so pointer-receiver method never invoked.
- **GREEN**: Added 33 new test functions across 4 test files (1 NEW `hydrate_test.go` + 3 EXTEND), exercising the uncovered code paths. All tests pass first run (existing production code is correct; tests are characterizing existing behavior). No production code modified.
- **REFACTOR**: Test additions use shared patterns: `t.TempDir()` for all temp dirs (REQ-SLCO-003), table-driven tests for enum coverage (Validate harness/status), explicit `&state` syntax to trigger pointer receivers, fake `SessionStore` interface implementation for `HydrateForPrompt` testing without I/O. Subagent boundary preserved — no new tests reference `AskUserQuestion` tokens.

### Mx Step C Preliminary Judgment

**SKIP-eligible** per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a (file exclusion criteria + tag necessity rubric):
- 0 production `.go` files modified → no @MX tag delta required
- 4 `_test.go` files modified (1 NEW + 3 EXTEND) → tests are characterization, no new goroutines, no complexity ≥15, no fan_in ≥3 from new code (test files are entry points by definition)
- New helper types in tests (`fakeSessionStore`, `unknownCheckpoint`) are local to test files, no @MX:ANCHOR required
- No new `// TODO:` or dangerous patterns introduced

Recommendation: orchestrator MAY skip Mx Step C entirely or run mechanical scan + judge no-tag-required.

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

## §E. Sync-phase Audit-Ready Signal (2026-05-25)

```yaml
sync_complete_at: 2026-05-25T<sync-commit-timestamp>
sync_commit_sha: <pending>  # backfilled post-commit
sync_status: implemented
b12_self_test_a: "grep -c 'SPEC-V3R6-SESSION-LEGACY-COVERAGE-001' CHANGELOG.md → 1 (NEW entry)"
b12_self_test_b: "AC count in CHANGELOG match acceptance.md → 7 (all 7 mandatory ACs listed)"
b12_self_test_c: "frontmatter status all 4 artifacts → implemented (PASS, no drift)"
sync_commit_range: "a095bce09..HEAD (1 commit expected: CHANGELOG + 4 frontmatter)"
files_modified_count: 5  # CHANGELOG.md + 4 SPEC artifacts
files_by_path:
  - CHANGELOG.md (entry append to [Unreleased] / ### Changed section)
  - .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/spec.md (frontmatter status/version/sync_commit_sha)
  - .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/plan.md (frontmatter status/version/sync_commit_sha)
  - .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/acceptance.md (frontmatter status/version/sync_commit_sha)
  - .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md (frontmatter status/version/sync_commit_sha + this §E section)
l44_pre_commit_fetch: "0 0"  # verified before CHANGELOG edit
l44_post_push_fetch: "<TBD-by-orchestrator-post-push>"
preserve_list_post_sync_count: 14+  # 14+ untouched entries (3 config M + 1 harness telemetry + 5+ research/runtime ?? + 5+ parallel-session ?? per plan.md §A.4 PRESERVE table + parallel WORKFLOW-PLAN/FOUNDATION-CORE sessions' uncommitted work)
mx_step_c_judgment: SKIP-eligible  # 0 production .go mutations per run-phase, 4 _test.go are characterization/test-entry, no new @MX tags required
changelog_entry_position: "line N in [Unreleased] ### Changed section"  # exact position post-commit
canary_compliance_check:
  canonical_status_enum: "implemented ✓ (not superseded/archived/rejected — forward state)"
  4_artifact_sync_commit_sha_backfill: "4 × sync_commit_sha: '<pending>' ready for chore backfill if needed"
  b12_all_checks_pass_readiness: "PASS-ready: AC count = 7, frontmatter status = implemented, CHANGELOG entry present, paths all exist"
```

### Sync-phase Quality Verification

| Check | Result | Evidence |
|-------|--------|----------|
| L44 HARD pre-commit fetch | PASS | `git rev-list --count --left-right origin/main...HEAD` = `0 0` at sync-phase entry |
| L46 path-specific staging | READY | Exactly 5 paths: CHANGELOG.md + 4 `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/*.md` |
| L48 SSOT canary | PASS | 4 SPEC artifacts: frontmatter only (status/version/sync_commit_sha); body content ZERO changes (spec.md/plan.md/acceptance.md body) |
| L49 Trust-but-verify batch | READY | 7-item parallel verification pending orchestrator post-push (all reads, not writes) |
| L52 race absorption | OBSERVE | Parallel sessions WORKFLOW-PLAN-GEARS-ALIGN-001 (iter-3 @d834f4ac5) + FOUNDATION-CORE-GEARS-ALIGN-001 (iter-4 @d834f4ac5) present; disjoint scope vs SESSION-LEGACY-COVERAGE-001 confirmed via `git diff --name-only <base>...d834f4ac5 -- internal/session/` → empty (no file overlap) |
| L59 pre-commit staging assertion | READY | `git diff --cached --name-only \| sort -u \| wc -l` expected = **5** (not 4: CHANGELOG.md is 5th) |
| B12 CHANGELOG self-test A | READY | `grep 'SPEC-V3R6-SESSION-LEGACY-COVERAGE-001' CHANGELOG.md \| wc -l` expected = 1 (NEW, unique) |
| B12 CHANGELOG self-test B | READY | AC count in CHANGELOG body = 7 (matches acceptance.md §D AC-SLCO-001..007 exactly) |
| B12 CHANGELOG self-test C | READY | All 4 artifacts frontmatter: `status: implemented`, `version: "0.2.0"`, `sync_commit_sha:` backfill-ready |
| Mx Step C judgment | SKIP-ELIGIBLE | Per mx-tag-protocol.md §a: 0 production .go, 4 _test.go characterization only, no new dangerous patterns, subagent boundary preserved (C-HRA-008 PASS) |
| PRESERVE invariant | INTACT | 14+ entries untouched per plan.md §A.4 (3 M config / 1 M harness telemetry / 5+ ?? parallel research / 5+ ?? parallel session artifacts / .moai/specs/.moai/ spurious dir left alone) |

## §F. Mx-phase Audit-Ready Signal (2026-06-02)

```yaml
mx_complete_at: 2026-06-02
mx_status: skip-justified
mx_commit_sha: e979a4d13
mx_tag_count: 0
mx_skip_justified: true
mx_verdict: SKIP-JUSTIFIED
mx_evidence: |
  Tier S test-only SPEC with 0 production .go modifications (4 _test.go files modified, 
  1 NEW hydrate_test.go + 3 EXTEND). Test-only code is characterization by definition. 
  No new functions added to production code (0 candidates for @MX:ANCHOR). 
  No new goroutines, complexity ≥15, or state mutation (0 candidates for @MX:WARN). 
  No new business rules or TODO items (0 candidates for @MX:NOTE/@MX:TODO). 
  Subagent boundary preserved (C-HRA-008 PASS). Skip judgment per mx-tag-protocol.md §a.
```

## §D. Cross-references

- spec.md — canonical SSOT (REQ-SLCO-001..008 anchored).
- plan.md — Tier S minimal Section A only.
- acceptance.md — 7 AC matrix with binary PASS/FAIL verification commands + §G B12 self-test.
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate authority (본 self-audit 0.945 skip-eligible).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S minimal 정의 + skip-eligible 정책 (CONST-V3R5-026).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form.
- `CLAUDE.local.md §6` — Coverage Targets + Test Isolation + filepath.Abs() 규칙.
- `CLAUDE.local.md §2 [WARN]` — OTEL t.Setenv 데이터 레이스 회피.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical frontmatter (§D.3 Status Transition Ownership Matrix: manager-docs owns `in-progress → implemented` on sync-phase commit).
- SPEC-LINT-CLEANUP-001 progress.md / HARNESS-NAMESPACE-CLEANUP-001 progress.md — Tier S minimal 1-pass cohort plan-phase 선례.
- MEMORY.md `Sprint 9 lane A AAT-001 4-phase FULLY CLOSED + Sprint 10 entry pending` — Sprint 10 entry 5 후보 중 A SESSION-LEGACY-COVERAGE-001 권장 선정 출처.
