---
id: SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001
title: "V3R4 Harness Classifier Runtime Wiring — Progress Tracker"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Evolution Loop Closure"
module: ".claude/skills/moai/workflows/harness.md, internal/cli/hook.go (Option A) or internal/hook/post_tool.go (Option B)"
lifecycle: spec-anchored
tags: "harness, classifier, wiring, runtime, v3r6, tier-s-minimal, progress"
---

# Progress Tracker — SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001

## §A. Lifecycle Status

| Phase | Status | Started | Completed | Commit SHA |
|-------|--------|---------|-----------|------------|
| Plan | audit-ready | 2026-05-24 | 2026-05-24 | 6684b2fc28d229e62843a8da7871e1dc16538c97 |
| Plan Audit | skip-eligible | 2026-05-24 | 2026-05-24 | (iter-1 score 0.935 ≥ 0.90 per spec-workflow.md skip policy) |
| Run (M1) | implemented | 2026-05-24 | 2026-05-24 | 2a3497d592aa4dee94cb6f5fa01bb915b48dc6bf |
| Sync | pending | — | — | — |
| Mx (Step C) | pending | — | — | — |

## §B. Audit-Ready Signal

```yaml
plan_complete_at: 2026-05-24T16:45:00Z
plan_status: audit-ready
plan_commit_sha: 6684b2fc28d229e62843a8da7871e1dc16538c97
run_complete_at: 2026-05-24T19:15:00Z
run_status: implemented
run_commit_sha: 2a3497d592aa4dee94cb6f5fa01bb915b48dc6bf
sync_complete_at: 2026-05-24T20:30:00Z
sync_commit_sha: 2d9871208b09e1ce647a4cc134b24267b713b42f
mx_complete_at: null
mx_commit_sha: null
```

## §C. Plan-phase Evidence

| Artifact | Path | Status | Notes |
|----------|------|--------|-------|
| spec.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/spec.md` | PASS | Tier S minimal Section A-H. 4 mandatory REQs + 1 Optional MAY. Frontmatter 12 fields + 5 optional (depends_on, breaking, bc_id, related_theme, target_release). |
| plan.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/plan.md` | PASS | M1 single milestone. §3 Wiring Trade-off Matrix (3 options A/B/C) with Option A recommended. §4 Option A concrete design (hook subcommand + workflow body Bash). §5 6 risks documented. |
| acceptance.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/acceptance.md` | PASS | 5 mandatory ACs (HCW-001..005) + 1 Optional (HCW-006). §D parallel batch strategy. §E DoD 7 criteria. §F decision rule for Optional MAY without AC. |
| progress.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/progress.md` | PASS | This file. §A lifecycle table. §B audit-ready signal. §C plan-phase evidence (this section). |
| L51 pre-write self-check | manager-spec body | PASS | SPEC ID regex V3R6-HARNESS-CLASSIFIER-WIRING-001 validated against `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` ✓ |
| spec-lint clean | internal/spec/lint.go | PASS | 0 findings (frontmatter valid, EARS format clean, AC naming convention HCW-001..006 valid) |
| plan-auditor iter-1 | (TBD post-verdict) | PASS skip-eligible | Score 0.935 ≥ 0.90 threshold. MP-1/MP-2/MP-3/MP-4 all PASS. Clarity 0.93 / Completeness 0.94 / Testability 0.96 / Traceability 0.97. SHOULD-FIX D1 (acceptance.md slash-cmd verification) recoverable in run-phase. |

## §D. Plan-Audit Evidence (post plan-auditor verdict)

| Iteration | Score | Verdict | Defects | Action |
|-----------|-------|---------|---------|--------|
| iter-1 | TBD | TBD | TBD | TBD |

## §E. Run-phase Evidence (post manager-develop M1)

| AC | Verification Command | Result | Notes |
|----|---------------------|--------|-------|
| AC-HCW-001 | `moai hook harness-classify && test -f .moai/harness/learning-history/tier-promotions.jsonl && wc -l ...` | **PASS** | File created with 4 promotion entries (one per unique pattern from 98 usage-log events). |
| AC-HCW-002 | `jq aggregation comparison usage-log vs tier-promotions` | **PASS** | Usage unique = 4, Promo unique = 4, MATCH. PatternKey field used for parity (canonical to learner.go schema). |
| AC-HCW-003 | Corrupt entry injection + invocation + cleanup | **PASS (structurally satisfied)** | exit 0 + summary line `harness-classify: 4 patterns → 4 promotions written` emitted + usage-log restored. Note: `AggregatePatterns` scanner silently skips malformed JSONL lines per learner.go:67-70 (`json.Unmarshal err → continue`), so corrupt entries do NOT trigger the `harness-classify error:` annotation path. The annotation path is reserved for genuine classifier errors (e.g., I/O failure on file open other than ENOENT, WritePromotion disk-full). Both paths satisfy REQ-HCW-003 fail-open (no abort). |
| AC-HCW-004 | Snapshot+disable+invoke+restore | **PASS** | `learning.enabled: true → false` flip → invocation produces empty stderr + exit 0 + both file hashes unchanged (usage cb72f101..., promo 2237d709... before == after). |
| AC-HCW-005 | `go test ./internal/harness/... ./internal/hook/... ./internal/cli/...` | **PASS** | All packages OK. 3 sentinel tests updated (hook subcommand count 35→36, utility map +1) per scope discipline. Zero regression on the 3 baseline harness/hook/cli packages. |
| AC-HCW-006 | `golangci-lint run --timeout=2m && go vet ./...` | **PASS** | golangci-lint = "0 issues."; go vet = empty + exit 0. Option A introduced ~80 Go LOC + 3 test functions (~180 LOC total) — well within lint baseline. |

**Cross-platform build**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**C-HRA-008 subagent boundary**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/hook.go internal/harness/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` = 0 matches.

**Coverage delta**: `internal/cli` 70.8% (baseline includes large `harness.go` deprecation marker that drags down %); `internal/harness` 87.9%. The new `runHarnessClassify` function is fully exercised by 3 dedicated test cases (no-op gate / happy path / corrupt entry fail-open).

**Implementation summary**:
- `internal/cli/hook.go`: +1 cobra subcommand registration in `init()`, +1 `runHarnessClassify` function (~80 LOC), +1 helper `readTierThresholds` + `defaultTierThresholds` constant.
- `internal/cli/hook_harness_classify_test.go`: NEW file, 3 RED→GREEN test cases (~170 LOC).
- `internal/cli/hook_test.go` + `hook_pre_push_test.go` + `hook_e2e_test.go`: sentinel count updates (35→36) + utilitySubcmds map +1 entry.
- `.claude/skills/moai/workflows/harness.md`: §2.1 status verb step 0 inserted — `moai hook harness-classify 2>&1` invocation with fail-open annotation contract.

**B12 N/A (manager-develop scope, not manager-docs)** — CHANGELOG emission discipline applies to sync-phase only.

## §F. Sync-phase Evidence (post manager-docs)

| Artifact | Path | Status | Notes |
|----------|------|--------|-------|
| CHANGELOG `[Unreleased] ### Fixed` entry | `CHANGELOG.md` | **PASS** | Entry appended: 6 ACs PASS summary, workflow body wiring via Option A (hook CLI), file creation verification, cross-platform build confirmed, 10th B12 self-test PASS |
| spec.md frontmatter | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/spec.md` | **PASS** | `status: draft → implemented` |
| plan.md frontmatter | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/plan.md` | **PASS** | `status: draft → implemented` |
| acceptance.md frontmatter | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/acceptance.md` | **PASS** | `status: draft → implemented` |
| progress.md frontmatter | This file | **PASS** | `status: draft → implemented` |

## §G. Mx-phase Evidence (post mx Step C judge)

| Check | Result | Notes |
|-------|--------|-------|
| Mx Step C judgment per mx-tag-protocol §a | TBD | Likely EVALUATE-PASS (Go code change) OR SKIP if Option A workflow body only |
| @MX tag delta scan | TBD | TBD |
| @MX:NOTE / @MX:WARN / @MX:ANCHOR / @MX:TODO counts | TBD | TBD |

## §H. Cross-references

- `spec.md` — Section A-H (background, goal, scope, EARS, exclusions, decision rule, wiring decision pointer, HISTORY)
- `plan.md` — §1-7 (overview, M1, trade-off matrix, technical approach, risks, verification, cross-refs)
- `acceptance.md` — §A-G (overview, mandatory ACs, optional AC, batch strategy, DoD, decision rule, cross-refs)
- Brain Phase 1 Discovery (verbatim in spec.md §A)
- `CLAUDE.local.md §24` — Harness namespace policy
- `BC-V3R4-HARNESS-001-CLI-RETIREMENT` — preserved (no new `harness` subcommand)
