---
id: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001
title: "V3R4 Self-Evolving Harness Loop Closure — Progress Tracker"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Self-Evolution Closure"
module: "internal/harness/proposalgen, internal/cli/harness"
lifecycle: spec-anchored
tags: "harness, proposal, progress, tier-m"
---

# Progress Tracker — SPEC-V3R6-HARNESS-PROPOSAL-GEN-001

## §A. Lifecycle Status

| Phase | Status | Started | Completed | Commit SHA |
|-------|--------|---------|-----------|------------|
| Plan | audit-ready | 2026-05-24 | 2026-05-24 | e5b2859a9 |
| Plan Audit | PASS 0.935 (skip-eligible) | 2026-05-24 | 2026-05-24 | e5b2859a9 |
| Run (M1) | complete | 2026-05-24 | 2026-05-24 | 24cb6ad4b (see §B.2 attribution note) |
| Sync | complete | 2026-05-24 | 2026-05-24 | TBD |
| Mx (Step C) | pending | — | — | — |

## §B. Audit-Ready Signal

```yaml
plan_complete_at: 2026-05-24T21:15:00Z
plan_status: audit-ready
plan_commit_sha: e5b2859a9fa23f00c53ad3af74115235834009c0
run_complete_at: 2026-05-24T22:30:00Z
run_status: complete
run_commit_sha: 24cb6ad4b5c11cd5dfab56429fef674f93f4d062
sync_complete_at: null
sync_commit_sha: pending
mx_complete_at: null
mx_commit_sha: pending
```

## §B.1 Run-phase Evidence (M1)

| AC | Verdict | Verification Command | Output |
|----|---------|---------------------|--------|
| AC-PGN-001 | PASS | `go test -run TestReader_LiveFixture ./internal/harness/proposalgen/...` | `ok` (8 records, 4 unique pattern_keys) |
| AC-PGN-002 | PASS | `go test -run TestMapper_CurrentDataNoOp ./internal/harness/proposalgen/...` | `ok` (0 candidates from system-event-only data) |
| AC-PGN-003 | PASS | `go test -run TestScaffolder_NoOpSkipsCreation ./internal/harness/proposalgen/...` | `ok` (.moai/proposals/ not created on no-op) |
| AC-PGN-004 | PASS | `go test -run TestPropose_DryRun_BaselineFixture ./internal/cli/harness/...` | `ok` (JSON matches REQ-PGN-014 exact shape) |
| AC-PGN-005 | PASS | `grep -rn 'AskUserQuestion(' internal/cli/harness/ \| grep -v '_test.go'` | empty (0 matches) |
| AC-PGN-006 | PASS | `go test -run TestPropose_NoAskUserQuestion ./internal/cli/harness/...` | `ok` (scans propose.go) |
| AC-PGN-007 | PASS | `go test -coverprofile=cover.out ./internal/harness/proposalgen/... ./internal/cli/harness/...` | total: 87.7% (≥85% target) |
| AC-PGN-008 | PASS | `go vet ./... && golangci-lint run --timeout=2m` | both exit 0; lint 0 issues |

Additional artefacts:
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- CLI smoke against live data: `go run ./cmd/moai harness propose --dry-run` emits `{"proposals":[],"reason":"no-actionable-patterns","malformed_lines":0,"evaluated_patterns":4,"auto_delegate":false}` (REQ-PGN-014 empirical PASS).
- Testdata fixture created at `internal/harness/proposalgen/testdata/tier-promotions-current-baseline.jsonl` (verbatim 8-record snapshot of live `.moai/harness/learning-history/tier-promotions.jsonl`).
- `internal/cli/harness_route.go` modified with exactly +6 lines (1 import + 1 registration block + 4-line comment) to register `propose` under `newHarnessRouterCmd()`; no other modifications.

## §B.2 Multi-Session Attribution Note (L52)

Per CLAUDE.local.md §23.8 (Multi-Session Race Mitigation), the run-phase deliverables for this SPEC landed on `main` under commit `24cb6ad4b` whose canonical commit subject reads `chore(SPEC-V3R6-MULTI-SESSION-COORD-001): progress.md §C plan-auditor signal backfill (iter-1 PASS-WITH-DEBT 0.812)`. This commit subject is misattributed to the COORD-001 SPEC due to a concurrent parallel session indiscriminately staging all working-tree changes (both COORD-001 progress.md + PROPOSAL-GEN-001 implementation files).

Inspection of `git show --stat 24cb6ad4b` confirms the commit contains BOTH:

1. The COORD-001 progress.md backfill (the subject's intended scope, 92 lines).
2. The full PROPOSAL-GEN-001 M1 implementation (14 files, 1881 line insertions, including: `internal/cli/harness/{propose.go,propose_test.go,propose_boundary_test.go}`, `internal/harness/proposalgen/{types,reader,mapper,scaffolder}.{go,_test.go}` + testdata fixture, `internal/cli/harness_route.go` 8-line registration delta, and this very file's `status: in-progress` + §B.1 Run-phase Evidence block).

This note ESTABLISHES THE CANONICAL PROVENANCE for the PROPOSAL-GEN-001 run-phase content within `24cb6ad4b`. No history rewrite was attempted (Hybrid Trunk + multi-session race prevent safe rewind per §23.5). The misattribution is an audit-trail artifact only; the implementation itself is byte-identical to what was authored and verified.

Followup (sync-phase): manager-docs SHOULD reference this §B.2 note in the `[Unreleased]` CHANGELOG entry to surface the attribution clarification publicly.

## §C. Plan-phase Evidence

| Artifact | Path | Status | Notes |
|----------|------|--------|-------|
| spec.md | `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/spec.md` | PASS | Tier M. Frontmatter 12 canonical fields + 5 optional (depends_on, breaking, bc_id, related_theme, target_release, tier). 14 REQs (REQ-PGN-001..014). §7 Out of Scope with 5 sub-sections (h3 per spec-lint MissingExclusions requirement). |
| plan.md | `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/plan.md` | PASS | Tier M Section A-E template. §3 trade-off matrix with 3 architectural decisions (package placement / CLI placement / output format), each with 3 alternatives + recommendation. §A.4 PRESERVE list (6 entries). §A.5 EXTEND list (7 NEW + 1 MODIFY = 8 files, ~1351 LOC estimate) — iter-2 D1 fix: removed `internal/cli/harness/cmd.go` (would have caused cobra duplicate-parent panic; `propose` is registered as subcommand under existing `newHarnessRouterCmd()` instead). |
| acceptance.md | `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/acceptance.md` | PASS | 8 mandatory ACs (AC-PGN-001..008) + 1 Optional (AC-PGN-009). §C edge cases (7 entries) covered within AC bodies. §D Definition of Done (7 conditions). §E Optional MAY decision rule. §F parallel batch strategy (7 commands). |
| progress.md | `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/progress.md` | PASS | This file. §A lifecycle table. §B audit-ready signal (TBD SHA filled post-commit). §C plan-phase evidence (this section). |

## §D. Decisions Recorded (User-Confirmed, see plan.md §A.3)

1. **Delegation mode**: auto-delegate + AskUserQuestion gate (Approve/Modify/Reject, 3 options).
2. **Git env**: main direct push (Hybrid Trunk, no plan/* or feat/* branches, no worktree).
3. **Tier**: M (Medium) — ~1351 LOC, 8 files (iter-2 D1 fix: 7 NEW + 1 MODIFY, see §C plan.md row), plan-auditor threshold 0.80, TDD, coverage 85%.
4. **16-language neutrality**: all proposal output language-neutral.

## §E. Trade-off Decisions (see plan.md §3)

| Decision area | Chosen option | Rationale summary |
|---------------|---------------|-------------------|
| Package placement | A — `internal/harness/proposalgen/` | Co-locates with `internal/harness.Promotion` source struct; clear harness-subdomain semantic |
| CLI placement | A — `propose` subcommand under existing `newHarnessRouterCmd()` in `internal/cli/harness_route.go` | Reuses already-registered V3R4 harness command surface (no new top-level cobra parent); avoids cobra duplicate-subcommand panic; semantically consistent with existing V3R4 sibling verbs (`status`/`apply`/`rollback`/`disable`/`mute`/etc.) — iter-2 D1 architectural fix |
| Output format | A — structured JSON with metadata | Single payload drives orchestrator AskUserQuestion gate; explicit `reason` field surfaces no-op semantics |

## §F. Critical Scope Reminder

[HARD] With current `tier-promotions.jsonl` data (8 records, 4 unique system-event pattern_keys as of 2026-05-24), this SPEC produces a generator that emits **zero actionable proposals** — this is the correct behavior per REQ-PGN-014 and AC-PGN-002/AC-PGN-004. The generator is "future-data ready"; its value materializes when richer pattern types (code_change, error_pattern, tool_failure, repeated_edit) appear in `tier-promotions.jsonl` via subsequent V3R6+ learning sources.

Run-phase implementer MUST NOT interpret the no-op as a defect. AC-PGN-002 explicitly verifies the no-op.

## §G. Pre-Spawn Sync Verification (per agent-common-protocol.md)

Before run-phase manager-develop spawn, orchestrator MUST run:

```bash
git fetch origin main
git rev-list --count --left-right origin/main...HEAD
```

Expected: `0 0` (clean) or `0 N` (local ahead). `N 0` or `N M` triggers AskUserQuestion gate (rebase/inspect/abort) before run-phase.

## §H. Open Issues

None at plan completion. All architectural and policy decisions resolved (see §D and §E).
