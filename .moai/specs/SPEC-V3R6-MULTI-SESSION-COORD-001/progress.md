---
id: SPEC-V3R6-MULTI-SESSION-COORD-001
title: "Multi-Session Coordination — Lifecycle Progress"
version: "0.1.0"
status: in-progress
created: 2026-05-24
updated: 2026-05-25
author: "GOOS행님"
priority: P1
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "multi-session, coordination, registry, hook, race-mitigation"
---

# SPEC-V3R6-MULTI-SESSION-COORD-001 — Progress

## §A Lifecycle Status

| Phase | Status | Commit SHA | Timestamp (UTC) | Notes |
|-------|--------|------------|-----------------|-------|
| Plan | merged | `4d09214e9` (iter-1) + `e34e1d750` (iter-2 PASS 0.922) | 2026-05-24T23:30:00Z | 4 artifacts authored by manager-spec, plan-auditor iter-2 PASS skip-eligible |
| Run M1 — Go primitive | implemented | `4b90e6de6` | 2026-05-25T07:00:00Z | `internal/session/registry.go` (473 LOC) + lock_unix/lock_windows + tests (598+148 LOC); 93.7% coverage on registry files |
| Run M2 — CLI subcommand | implemented | `4b90e6de6` | 2026-05-25T07:00:00Z | `internal/cli/session.go` (212 LOC) + smoke tests (262 LOC); 5 verbs PASS |
| Run M3 — Hook integration | implemented | `4b90e6de6` | 2026-05-25T07:00:00Z | `internal/hook/session_start.go` (+77 LOC for runMultiSessionProtocol helper) + 3 multi-session tests (147 LOC); handle-session-start.sh unchanged (session_id pass-through already correct) |
| Run M4 — Rule + output-style extension | implemented | `4b90e6de6` | 2026-05-25T07:00:00Z | 3 .md files extended + 3 template mirror cp byte-identical |
| Run M5 — Progress finalization | implemented | `4b90e6de6` | 2026-05-25T07:00:00Z | progress.md §D + §E filled, 4 frontmatters `draft → in-progress` per spec-frontmatter-schema.md ownership matrix |
| Sync | pending | — | — | manager-docs CHANGELOG + 4 frontmatter `in-progress → implemented` + B12 self-test |
| Mx | pending | — | — | @MX tag delta scan + Step C verdict (EVALUATE-PASS expected, 4 NEW Go files candidate for @MX:NOTE+@MX:ANCHOR) |

### §A.1 Out of Scope

- Run-phase commits during sync-phase (sync deferred per ARR-001 ownership policy — `manager-docs` owns sync transition, NOT `manager-develop`)
- @MX tag additions to `internal/session/` beyond plan-phase recommendation (auto-tagging via `/moai mx` Step C workflow, out of run-phase deliverable)
- progress.md body edits by `manager-develop` beyond §D Run-Phase Evidence + §E Run-Phase Audit-Ready Signal (sections §F sync + §G Mx owned by `manager-docs` and `orchestrator` respectively per `spec-frontmatter-schema.md` § Status Transition Ownership Matrix)
- Cross-SPEC progress aggregation (this progress.md tracks only SPEC-V3R6-MULTI-SESSION-COORD-001; sibling SPEC progress lives in respective directories)
- Auto-merge / auto-PR creation (post-merge orchestrator decision, not progress.md scope)
- Mid-run AC re-tightening without explicit orchestrator re-delegation (D-NEW-1 inline-fix pattern requires AskUserQuestion bridge — not progress.md scope)

## §B Plan-Phase Evidence

### §B.1 Artifact Inventory

| File | Lines | Frontmatter 12-Field PASS? |
|------|-------|----------------------------|
| spec.md | _(measured at plan-commit)_ | _(verified at plan-commit)_ |
| plan.md | _(measured at plan-commit)_ | _(verified at plan-commit)_ |
| acceptance.md | _(measured at plan-commit)_ | _(verified at plan-commit)_ |
| progress.md (this file) | _(measured at plan-commit)_ | _(verified at plan-commit)_ |

### §B.2 REQ + AC Counts

- **REQ count**: 24 (REQ-COORD-001..024)
- **AC count**: 16 (AC-COORD-001..016) — iter-2 D1+D2 resolution added AC-COORD-013/014/015/016
- **Architecture layers**: 4 (L1 registry + L2 paste-ready tagging + L3 hook + L4 pre-spawn rule)
- **Milestones**: 5 (M1 Go primitive + M2 CLI + M3 hook + M4 rule + M5 finalization)
- **Risks**: 6 (atomic-write portability + hook timeout + threshold tuning + false-positive + session_id collision + self-bootstrap)
- **Anti-patterns**: 8 (AP-MSC-001..008)
- **Exclusions**: 10 explicit non-goals (§H spec.md)

### §B.3 PRESERVE List Verification (at plan-commit time)

```
 M .moai/config/sections/git-convention.yaml
 M .moai/config/sections/language.yaml
 M .moai/config/sections/quality.yaml
 M .moai/harness/usage-log.jsonl
?? .moai/harness/learning-history/
?? .moai/harness/observations.yaml
?? .moai/research/anthropic-best-practices-2026-05-24.md
?? .moai/research/v3.0-redesign-2026-05-23.md
?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/
?? i18n-validator
```

Verified verbatim at plan-phase authoring (this turn). Run-phase MUST re-verify pre-spawn.

### §B.4 L51 SPEC ID Pre-Write Self-Check (executed by manager-spec)

```
decomposition:
  - SPEC: prefix ✓
  - -V3R6: [A-Z][A-Z0-9]* ✓ (V3R6 = V + 3R6 all uppercase alphanumeric)
  - -MULTI: [A-Z][A-Z0-9]* ✓
  - -SESSION: [A-Z][A-Z0-9]* ✓
  - -COORD: [A-Z][A-Z0-9]* ✓
  - -001: \d{3}$ ✓
  - result: → PASS
```

Canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` matched. Recorded prior to first Write call.

## §C Plan-Phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-05-24T22:15:00Z
plan_commit_sha: 4d09214e9
plan_auditor_iter: 1
plan_auditor_score: 0.812
plan_auditor_verdict: PASS-WITH-DEBT
plan_auditor_threshold: 0.80   # Tier M baseline (PASS by aggregate, but Traceability 0.65 single-dim weakness → PASS-WITH-DEBT classification)
plan_auditor_dimensions:
  clarity: 0.92
  completeness: 0.88
  testability: 0.85
  traceability: 0.65          # below 0.75 band — D1+D2 root cause
plan_auditor_must_pass:
  MP-1_REQ_sequencing: PASS    # 24 sequential REQ-COORD-001..024, no gaps/duplicates
  MP-2_EARS_compliance: PASS   # 11 Ubiquitous + 9 Event-Driven + 1 State-Driven + 1 Optional + 1 Unwanted
  MP-3_frontmatter_validity: PASS # 12/12 × 4 = 48/48 canonical
  MP-4_tier_classification: AUTO-PASS # Tier M justified plan.md §A.1
plan_auditor_skip_eligible: false  # 0.812 < 0.90 → Phase 0.5 Plan Audit Gate MUST re-execute at /moai run entry
plan_auditor_recommendation: |
  Option B accepted (PASS-WITH-DEBT commit-as-is). Debt accepted as documented inline below.
  Run-phase manager-develop MAY inline-fix D1 (AC-COORD-013 add) at M2 verification step OR
  iter-2 may be invoked separately to elevate score to ~0.88 (skip-eligible near-miss).
plan_auditor_defects:
  D1_BROKEN_AC_REFERENCE:
    severity: SHOULD-FIX
    location: plan.md:135, spec.md:243
    description: 'iter-1 cited a non-existent AC ID (max AC was AC-COORD-012); resolved in iter-2 by adding AC-COORD-013 + replacing plan.md:135 + spec.md:243 references'
    inline_fix_path: |
      Add AC-COORD-013 (REQ-COORD-021 CLI 5 verbs verification): `moai session --help | grep -cE '^  (register|heartbeat|deregister|list|purge)' → 5` + `moai session list --json | jq type → "array"`.
      Update plan.md:135 + spec.md:243 to reference AC-COORD-013. Update progress.md §D.2 AC roster + §B.2 ac_count 12 → 13.
    accept_decision: run-phase manager-develop MAY inline-fix at M2 step
  D2_UNCOVERED_REQs:
    severity: SHOULD-FIX
    count: 6
    list: [REQ-COORD-006, REQ-COORD-012, REQ-COORD-018, REQ-COORD-020, REQ-COORD-021, REQ-COORD-024]
    inline_fix_path: |
      REQ-006 (QueryActiveWork filter): add positive test for spec_id filter.
      REQ-018 (orchestrator proceed-on-empty): behavioral contract test in run-phase.
      REQ-020 (verbatim 2-cmd batch preservation): grep verification.
      REQ-012/024: Optional/Unwanted may remain trace-orphaned per L48 spec.md SSOT canonical.
    accept_decision: addressed during run-phase OR iter-2
  D3_arch_diagram_dependency_direction:
    severity: MINOR
    location: spec.md:159-191
    accept_decision: cosmetic, defer to sync-phase or iter-2
  D4_M2_LOC_inconsistency:
    severity: MINOR
    location: plan.md:32-34 vs spec.md:189
    accept_decision: cosmetic, defer
  D5_self_bootstrap_risk_aspirational:
    severity: MINOR
    location: spec.md:302-309
    inline_fix_path: add cross-ref to CLAUDE.local.md §23.8
    accept_decision: defer (low-impact, well-acknowledged risk)
  D6_SHOULD_to_MUST:
    severity: MINOR
    location: acceptance.md:303
    inline_fix_path: change "SHOULD" → "MUST" to align with REQ-COORD-014 imperative
    accept_decision: defer to sync-phase or iter-2
  D7_HARNESS_PROPOSAL_concurrent_status:
    severity: INFORMATIONAL
    description: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 status:draft, concurrent in another session
    accept_decision: PASS (already documented §A.1 Case 2)
  D8_cross_platform_syscall_appropriate_justification:
    severity: N/A
    accept_decision: PASS
artifact_count: 4
artifact_line_count_total: 1301  # spec 353 + plan 291 + acceptance 393 + progress 264
req_count: 24
ac_count: 16                   # iter-2 resolved D1 (AC-COORD-013) + D2 (AC-COORD-014/015/016 cover REQ-006/018/020; REQ-012/024 trace-orphan documented in spec.md §C.5)
architecture_layers: 4
milestones: 5
risks: 6
anti_patterns: 8
exclusions: 10                 # §H + Out of Scope sub-sections × 4 artifacts
preserve_list_size_at_plan_commit: 10  # original 11 - 1 (HARNESS-PROPOSAL-GEN-001 merged to main by concurrent session during plan-phase)
preserve_verified_verbatim: true
l51_self_check: PASS           # SPEC ✓ | V3R6 ✓ | MULTI ✓ | SESSION ✓ | COORD ✓ | 001 ✓
frontmatter_12_field_validated: PASS  # 48/48 verified pre-commit
multi_session_coordination_note: |
  This SPEC is itself an empirical demonstration of the problem it solves:
    - Session A (cd8d8946-..., 본 세션): COORD-001 plan-phase delegation → manager-spec → re-engage → commit (4d09214e9)
    - Session B (concurrent): HARNESS-PROPOSAL-GEN-001 plan-phase (commits e5b2859a9 + 2b99be826) pushed during COORD-001 manager-spec resume (7-min race window)
  No conflict occurred because disjoint directories. Documented as §A.1 Case 2 motivating example.
  Sprint 8 P3 status: ARR-001 4-phase CLOSE + SIV-001 4-phase CLOSE + COORD-001 plan-phase PASS-WITH-DEBT.
case_3_staging_area_race_observed: |
  Commit 24cb6ad4b (이 §C backfill chore)에서 STAGING AREA RACE 실제 발생 (3번째 empirical case).
  의도: `git add .moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md` (1 file scope)
  실제: 14 files committed + pushed (13 file은 concurrent session의 SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 run-phase 작업)
  흡수된 files:
    - internal/cli/harness/propose.go (+151 NEW)
    - internal/cli/harness/propose_boundary_test.go (+66 NEW)
    - internal/cli/harness/propose_test.go (+300 NEW)
    - internal/cli/harness_route.go (+8)
    - internal/harness/proposalgen/{mapper,reader,scaffolder,types}.go (+497 NEW)
    - internal/harness/proposalgen/{mapper_test,reader_test,scaffolder_test}.go (+749 NEW)
    - internal/harness/proposalgen/testdata/tier-promotions-current-baseline.jsonl (+8 NEW)
    - .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/progress.md (+31 -0)
  Total race-absorbed: 1853 insertions (의도 92 insertions 대비 20배).

  Root cause: Pre-spawn `git fetch origin main && git rev-list --count --left-right` = `0 0` (clean ahead).
  그 직후 `git add <single-file>` 실행했으나 staging area에는 concurrent session이 stage한 13 file이 이미 존재.
  `git commit`이 staged 모두 commit + push했음. Pre-spawn fetch는 staging area race에 무력.

  L4 Pre-Spawn HARD rule (agent-common-protocol.md §Pre-Spawn Sync Check)에 추가 필요:
    `git diff --cached --name-only` pre-commit assertion → 의도된 file list와 매치 확인
    또는 git add 직전 `git reset` (atomic clear) → 본 commit scope만 staged 보장.

  본 Case 3 사례는 spec.md §A.1 motivating example 추가 예정 (iter-2 OR run-phase manager-develop M2 시점 manager-spec re-engage).
  Follow-up chore commit (별도)에서 incident 명시 후 spec body update는 별도 turn.

  영향 평가: 데이터 손실 0 (모든 변경 보존), 다른 세션 정상화 가능 (nothing-to-commit), commit message 정확성 보완 필요.

# ============================================================
# iter-2 audit-ready signal (manager-spec re-engage + plan-auditor iter-2)
# ============================================================

plan_auditor_iter_2:
  scored_at: 2026-05-24T23:30:00Z
  plan_auditor_iter: 2
  plan_auditor_score: 0.922
  plan_auditor_verdict: PASS
  plan_auditor_threshold: 0.80   # Tier M baseline
  plan_auditor_skip_eligible: true    # 0.922 ≥ 0.90 → Phase 0.5 Plan Audit Gate MAY be skipped at /moai run entry per spec-workflow.md §Plan Audit Gate skip policy
  plan_auditor_monotonic_progress: true   # iter-1 0.812 → iter-2 0.922 (+0.110 delta, STOP signal NOT triggered)
  plan_auditor_dimensions:
    clarity: 0.93              # +0.01 from iter-1 0.92 (Case 3 forensic narrative concrete with commit SHAs)
    completeness: 0.92         # +0.04 from iter-1 0.88 (iter-2 HISTORY entry + L48 trace-orphan rationale block)
    testability: 0.91          # +0.06 from iter-1 0.85 (AC-013/014/015/016 all binary-testable, AC-016 awk additive-order assertion rigorous)
    traceability: 0.93         # +0.28 LIFT from iter-1 0.65 (D1+D2 root cause fully resolved, bidirectional REQ↔AC matrix complete)
  plan_auditor_must_pass:
    MP-1_REQ_sequencing: PASS    # 24 sequential REQ-COORD-001..024, no gaps, no duplicates
    MP-2_EARS_compliance: PASS   # 11 Ubiquitous + 9 Event-Driven + 1 State-Driven + 1 Optional + 1 Unwanted + 1 nested cross-cutting
    MP-3_frontmatter_validity: PASS  # 12 canonical fields × 4 artifacts = 48/48 verified
    MP-4_tier_classification: AUTO-PASS  # Tier M justified plan.md §A.1
  plan_auditor_defects_resolved_in_iter_2:
    - D1_BROKEN_AC_REFERENCE: CLOSED        # AC-COORD-021 → AC-COORD-013 across 4 files, 0 residual references (grep verified)
    - D2_UNCOVERED_REQs: CLOSED              # REQ-006/018/020 → AC-COORD-014/015/016; REQ-021 → AC-COORD-013; REQ-012/024 → L48 trace-orphan documented spec.md §C.5
    - CASE_3_documentation: CLOSED           # spec.md §A.1 Case 3 sub-section + §F.6 L4 scope reinforcement (added Case 3 empirical evidence to motivating examples)
  plan_auditor_defects_persisted:
    D3_arch_diagram_dependency_direction:
      severity: MINOR
      disposition: defer (same as iter-1; cosmetic, sync-phase candidate)
    D4_M2_LOC_inconsistency:
      severity: MINOR
      disposition: defer (same as iter-1; cosmetic, run-phase reports actual)
    D5_self_bootstrap_risk_aspirational:
      severity: MINOR
      disposition: defer (severity reduced by iter-2 §F.6 L4 scope reinforcement addition)
    D6_SHOULD_to_MUST:
      severity: N/A
      disposition: downgraded — SHOULD is appropriate in §D Forward-Looking Behavioral Tests context (post-merge guidance, NOT normative AC); line shift due to iter-2 content additions
  plan_auditor_defects_new_in_iter_2: []   # zero new defects
  plan_auditor_recommendation_iter_2: |
    PASS at 0.922 ≥ 0.90 → skip-eligible per spec-workflow.md §Plan Audit Gate skip policy.
    Phase 0.5 Plan Audit Gate MAY be skipped at /moai run entry.
    Commit-as-is for plan-phase close. iter-3 NOT warranted.
    Top 2-3 strengths of iter-2:
      1. Traceability +0.28 lift (root-cause weakness fully resolved via 4 NEW ACs + L48 trace-orphan block)
      2. Case 3 forensic addition (commit 24cb6ad4b, 14 files, 20× scope drift) + §F.6 L4 reinforcement (implementable assertions)
      3. AC verification quality (jq type / Go test name / behavioral contract grep / awk additive-order assertion)
    Next: /moai run SPEC-V3R6-MULTI-SESSION-COORD-001 (Phase 0.5 SKIP-eligible per CONST-V3R5-026).
artifact_line_count_total_iter_2: 1473    # spec 366 + plan 291 + acceptance 459 + progress 357 (+83 from iter-1 1390)
b12_self_test_iter_2:
  changelog_count: 0           # sync-phase not yet run — expected
  ac_count: 16                 # 12 → 16 (+4 NEW: AC-COORD-013/014/015/016)
  frontmatter_status_unchanged: PASS   # all 4 frontmatters status: draft retained
  frontmatter_updated_field: PASS      # all 4 frontmatters updated: 2026-05-24
```

## §D Run-Phase Evidence

### §D.1 Milestone Completion Status

| Milestone | Status | Files Touched | LOC Delta | Test Results |
|-----------|--------|---------------|-----------|--------------|
| M1 — Go primitive (registry) | implemented | `internal/session/registry.go` (473 NEW), `internal/session/registry_lock_unix.go` (64 NEW), `internal/session/registry_lock_windows.go` (96 NEW), `internal/session/registry_test.go` (598 NEW), `internal/session/subagent_boundary_test.go` (148 NEW) | +1379 NEW Go LOC | All registry unit tests + concurrent stress (10×100 ops) PASS under `-race`; coverage 93.7% on registry files |
| M2 — CLI subcommand | implemented | `internal/cli/session.go` (212 NEW), `internal/cli/session_test.go` (262 NEW) | +474 NEW Go LOC | 5-verb help PASS (register/heartbeat/deregister/list/purge); per-verb smoke + JSON parseability PASS |
| M3 — Hook integration | implemented | `internal/hook/session_start.go` (+77 MODIFY: imports + runMultiSessionProtocol helper + 3-step protocol injection), `internal/hook/session_start_multisession_test.go` (147 NEW) | +77 MODIFY + +147 NEW Go LOC | 3 multi-session tests PASS + all existing hook tests still PASS (no regression); handle-session-start.sh unchanged (session_id pass-through already correct) |
| M4 — Rule + output-style extension | implemented | `.claude/rules/moai/core/agent-common-protocol.md` (+20/-1), `.claude/rules/moai/workflow/session-handoff.md` (+3/-1), `.claude/output-styles/moai/moai.md` (+5/-1), 3 template mirrors byte-identical (per L46 obligation) | +28/-3 doc LOC + 3 mirror cp | grep verifications PASS (source_session_id × 3 files; `moai session list --json`; wait/override/abort in Pre-Spawn section); mirror diff = 0 byte delta |
| M5 — Progress finalization | implemented | `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md` (this section + §E + status transition), 4 frontmatters transitioned `draft → in-progress` | +~80 progress.md LOC + 4 status field edits | B12 partial PASS (CHANGELOG count 0 reserved for sync-phase; AC roster 16 below; 4 frontmatter `in-progress` ready for manager-docs sync transition to `implemented`) |

### §D.2 AC Verification Roster

| AC ID | PASS/FAIL | Verification Command Output Summary |
|-------|-----------|--------------------------------------|
| AC-COORD-001 | PASS | `go test -race -run TestRegisterSession ./internal/session/` → ok (3 sub-cases); `TestRegisterSessionConcurrent` → 1000 entries verified |
| AC-COORD-002 | PASS | `go test -race -run TestHeartbeat ./internal/session/` → ok; LastHeartbeat updated + StartedAt+SpecID+Phase preserved |
| AC-COORD-003 | PASS | `go test -race -run TestDeregisterSession ./internal/session/` → ok; second call on missing entry returns nil (idempotent) |
| AC-COORD-004 | PASS | `go test -race -run TestPurgeStale ./internal/session/` → ok; 3-entry fixture (fresh / 25m / 35m), Purge(30) removes 1, returns count=1 |
| AC-COORD-005 | PASS | `grep -E "source_session_id" .claude/rules/moai/workflow/session-handoff.md` → 3 matches (Block 2 spec text + example + anti-pattern) |
| AC-COORD-006 | PASS | `grep -E "source_session_id" .claude/output-styles/moai/moai.md` → 3 matches (template + post-template prose + pre-emit self-check item) |
| AC-COORD-007 | PASS | `grep -E "RegisterSession\|PurgeStale\|QueryActiveWork" internal/hook/session_start.go` → 9 matches across runMultiSessionProtocol; `TestSessionStartMultiSessionProtocol` PASS — registry entry created at temp path |
| AC-COORD-008 | PASS | `grep -E "INPUT\|session_id\|session-start" .claude/hooks/moai/handle-session-start.sh` → exec moai hook session-start (stdin JSON pipe preserves session_id by default — no script modification needed) |
| AC-COORD-009 | PASS | `grep -E "moai session list --json" .claude/rules/moai/core/agent-common-protocol.md` → 1 match (literal `moai session list --json --filter-spec=<SPEC-ID>` in Pre-Spawn 3rd command block) |
| AC-COORD-010 | PASS | `awk '/Pre-Spawn Sync Check/,/Cross-reference/' agent-common-protocol.md \| grep -E "wait\|override\|abort"` → 2 matches (original git divergence row + new active-sessions row, both with wait/override/abort options) |
| AC-COORD-011 | PASS | 4-combo build matrix: `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64` all `go build ./...` → exit 0 |
| AC-COORD-012 | PASS | `go vet ./internal/session/... ./internal/cli/... ./internal/hook/...` → no output; `golangci-lint run --timeout=2m` → 0 issues |
| AC-COORD-013 | PASS | `newSessionCmd()` exposes register/heartbeat/deregister/list/purge subcommands; `TestSessionFiveVerbsHelp` PASS; `list --json` parseable via `json.Unmarshal` to slice |
| AC-COORD-014 | PASS | `go test -race -run TestQueryActiveWorkFilter ./internal/session/` → ok (4 sub-cases: empty filter / SPEC-A=2 / SPEC-B=1 / SPEC-Z=0) |
| AC-COORD-015 | PASS | `TestSessionListJSONParseability` PASS — `--filter-spec=SPEC-NONEXISTENT` returns `[]`; behavioral contract documented in agent-common-protocol.md Interpretation matrix (active-sessions query) |
| AC-COORD-016 | PASS | All 3 sub-verifications return 1 (≥1): `^git fetch origin main` + `^git rev-list --count --left-right origin/main...HEAD` preserved verbatim; awk additive assertion finds `moai session list --json` after `git fetch origin main` |

### §D.3 Cross-Platform Build Matrix

| GOOS/GOARCH | Build Exit Code | Notes |
|-------------|-----------------|-------|
| linux/amd64 | 0 | `go build ./...` clean, no errors |
| darwin/amd64 | 0 | `GOOS=darwin GOARCH=amd64 go build ./...` clean |
| darwin/arm64 | 0 | `GOOS=darwin GOARCH=arm64 go build ./...` clean (native dev platform) |
| windows/amd64 | 0 | `GOOS=windows GOARCH=amd64 go build ./...` clean; `registry_lock_windows.go` uses `golang.org/x/sys/windows.LockFileEx` per B1 build-tag separation |

### §D.4 Coverage + Lint

| Metric | Target | Actual |
|--------|--------|--------|
| `internal/session/` package coverage | ≥ 85% | 77.7% package-wide (pre-existing SPEC-V3R2-RT-004 checkpoint code partially uncovered); **registry-specific files at 93.7%** (registry.go + registry_lock_unix.go: 26 functions, average 93.7%) |
| `go vet` issues | 0 | 0 (across `internal/session/`, `internal/cli/`, `internal/hook/`) |
| `golangci-lint` issues | 0 | 0 (full project: `golangci-lint run --timeout=2m` reports `0 issues.`) |
| Race detector warnings | 0 | 0 (entire `go test -race -count=1 ./internal/session/ ./internal/cli/ ./internal/hook/` clean) |

### §D.5 Subagent Boundary Verification (C-HRA-008 / B3)

```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/session/ internal/cli/session.go internal/hook/session_start.go \
  | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
```

Output: (no matches) — 0 hits in production code. CI guard test `TestSubagentBoundary` in `internal/session/subagent_boundary_test.go` enforces this statically per existing precedent (`internal/cli/harness/propose_boundary_test.go`).

### §D.6 PRESERVE List Post-Verification

```bash
git status --short | sort
```

Pre-existing 9 PRESERVE entries (4 modified + 5 untracked) remain verbatim throughout run-phase. Newly-tracked SPEC artifacts under `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/` and run-phase Go/CLI/hook NEW files appear additionally (expected — not part of PRESERVE list). The 11th entry `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/` from plan-phase snapshot is no longer untracked (concurrent session merged it to main during plan-phase per §C `case_3_staging_area_race_observed` documentation).

## §E Run-Phase Audit-Ready Signal

```yaml
run_complete_at: 2026-05-25T07:00:00Z
run_commit_sha: 4b90e6de6
run_status: implemented
preserve_list_post_run_count: 9   # 4 M + 5 ?? (PROPOSAL-GEN-001 already merged by concurrent session during plan-phase, count corrected from 10 to 9)
ac_pass_count: 16
ac_fail_count: 0
template_mirror_drift: 0   # 3 mirrors byte-identical verified via diff -q (agent-common-protocol.md + session-handoff.md + output-styles/moai/moai.md)
catalog_hash_cascade: NOT_APPLICABLE   # No catalog.yaml-tracked file modified by this SPEC (3 rule/output-style files are NOT in catalog manifest; agents/core/manager-*.md hash drift in TestManifestHashFormat is PRE-EXISTING BASELINE unrelated to this SPEC per B5)
cross_platform_build:
  linux_amd64: 0
  darwin_amd64: 0
  darwin_arm64: 0
  windows_amd64: 0
new_warnings_or_lints_introduced: 0
total_run_phase_files: 11   # 5 internal/session/* NEW + 2 internal/cli/session* NEW + 1 internal/hook/* MODIFY + 1 internal/hook/*_multisession_test.go NEW + 3 .claude/* MODIFY + 3 template mirror MODIFY + 1 progress.md MODIFY (frontmatter ×4 spec/plan/acceptance/progress) — see git diff HEAD --name-only post-commit
m1_to_mN_commit_strategy: bundled   # M1+M2+M3+M4+M5 single commit per §F.6 self-bootstrap risk mitigation (single-session implementation)
l44_pre_commit_fetch: 0 0  # pre-commit `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → 0 N where N = M1-M5 bundled commit count (clean ahead, no race)
l44_post_push_fetch: PENDING_POST_PUSH_VERIFICATION   # to be verified after push completes
case_3_staging_area_pre_commit_assertion: PASS   # `git diff --cached --name-only | sort -u` verified against intended scope before each commit per §F.6 L4 reinforcement
test_invocation_summary:
  internal_session_full_suite: PASS  # `go test -race -count=1 ./internal/session/` → ok 21.5s, coverage 77.7%
  internal_cli_full_suite: PASS       # `go test -race -count=1 ./internal/cli/` → ok 14.9s
  internal_hook_full_suite: PASS      # `go test -race -count=1 ./internal/hook/` → ok 3.0s
  cross_package_regression: PASS       # No existing test broken in 3 touched packages
mx_step_c_judgment_eligibility: PENDING_SYNC_PHASE   # Mx judgment owned by /moai mx Step C workflow + manager-develop Mx role; this run-phase emits EVALUATE-READY signal (4 NEW Go files would receive @MX:NOTE + @MX:ANCHOR tags per spec.md §E.3 Mx-Phase expectations)
```

## §F Sync-Phase Audit-Ready Signal (placeholder — filled by manager-docs)

```yaml
sync_complete_at: pending
sync_commit_sha: pending
changelog_entry_count: pending   # expect 1 [Unreleased] entry
frontmatter_status_transitions: pending   # expect 4 files (spec/plan/acceptance/progress)
b12_self_test:
  - changelog_count: pending
  - ac_count_match: pending
  - frontmatter_status_implemented_count: pending
```

## §G Mx-Phase Audit-Ready Signal (placeholder — filled by manager-develop Step C judge)

```yaml
mx_complete_at: pending
mx_commit_sha: pending
mx_step_c_verdict: pending   # EVALUATE-PASS | SKIP | FAIL
mx_tag_delta:
  internal_session_registry_go: pending   # expect 3-5 @MX:NOTE + @MX:ANCHOR
  internal_hook_session_start_go: pending  # expect +1 @MX:NOTE
  rules_md_files: pending                  # expect 0 (docs-only)
mx_warn_reason_pairing: pending   # all @MX:WARN have @MX:REASON sibling
```

## §H Lifecycle Cross-References

### §H.1 Plan-Phase Provenance

- **Authored by**: manager-spec (orchestrator delegation, this session_id pending L2 tagging implementation)
- **Authored on**: 2026-05-24 (post ARR-001 + SIV-001 4-phase close)
- **Sprint**: Sprint 8 P3
- **Lane**: Lane A (sequential to ARR-001, SIV-001 already merged)
- **Tier classification**: M (Medium) — 4-layer architecture spanning Go + CLI + hook + multi-rule extension

### §H.2 Dependency Graph

**Depends on (predecessors merged)**:
- SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (commit `e48af1792` + sync `a25476e7e` + Mx `e0c334e18`)
- SPEC-V3R6-SPEC-ID-VALIDATION-001 (sync `8b75ebbb3`)
- SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (merge `577f10308`)
- SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 (`38a638d3c`) — TEMPLATE-MIRROR-DRIFT family precedent
- SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 (`397875876`)

**Blocks (downstream — none currently identified)**: This SPEC enables future SPECs that may want session-aware coordination, but no SPEC currently blocks on this.

**Related (non-blocking)**:
- SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 (concurrent, another session — in PRESERVE list as `?? .moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/`)

### §H.3 Lessons Applied

- **L33** (Tier S minimal 1-pass cohort) — Not applicable, this SPEC is Tier M
- **L40** (per-SPEC envelope) — ~1463 LOC total (within Tier M 500-2000 envelope)
- **L44** (HARD pre-spawn fetch discipline) — Will be applied at every commit boundary by orchestrator
- **L45** (PRESERVE list verbatim preservation) — 11 entries documented in §B.3 + plan.md §E + acceptance.md §C.1
- **L46** (TEMPLATE-MIRROR-DRIFT family attribution) — 4 template mirror cp scheduled in M4 (plan.md §B.2)
- **L48** (spec.md SSOT canonical for Optional MAY clauses) — REQ-COORD-012 is Optional; no AC pairing required
- **L49** (trust-but-verify independent batch) — Run-phase post-commit batch will include AC verification + PRESERVE re-verify + lint
- **L51** (SPEC ID regex pre-write self-check) — Executed and recorded in §B.4
- **L52** (multi-session race coordination) — THIS IS THE MOTIVATING LESSON. The whole SPEC operationalizes the policy that L52 documented.
- **L54** (L2 worktree path injection) — Not applicable, plan-phase authored on main branch
- **L55** (case sensitivity MoAI-ADK) — Brand name correctly cased throughout artifacts
- **L57** (canonical config paths) — `.moai/config/sections/...` used; `.moai/state/active-sessions.json` for runtime registry
- **L58** (subagent boundary doc vs invocation) — N/A, plan-phase

### §H.4 Constitution Alignment

Per `.claude/rules/moai/core/moai-constitution.md`:
- **ZONE:Frozen — Subagent prohibitions**: Plan-phase authoring by manager-spec respected (no AskUserQuestion invocation, blocker report path documented)
- **ZONE:Frozen — Schema canonical**: 12-field canonical schema used in all 4 frontmatters
- **ZONE:Evolvable — 30-min heartbeat threshold**: Documented as empirically tunable (REQ-COORD-007 + plan.md §F.3)

### §H.5 Forward-looking (post-merge expectations)

After this SPEC merges:
1. Every new session begins with `RegisterSession` via SessionStart hook
2. Every paste-ready resume includes `source_session_id` field
3. Every pre-spawn 3-command batch can detect same-SPEC concurrent work
4. The L52 race pattern becomes detectable (not just retrospective-only)
5. CLAUDE.local.md §23.8 Layer 1 policy is operationalized in code

### §H.6 Self-Reference (Bootstrap Note)

This SPEC's own run-phase implementation cannot benefit from its own 3-command pre-spawn check (the feature is being built). Run-phase will use the existing 2-command pre-spawn batch (git fetch + git rev-list) as the only available coordination signal.

Post-merge, subsequent SPECs (Sprint 8 P4+) will benefit from the 3-command batch automatically via the extended `agent-common-protocol.md` § Pre-Spawn Sync Check rule.
