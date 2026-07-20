---
id: SPEC-HARNESS-EVOLVE-003
title: "Curator production wiring — Tier-Surface mapping + validation gates + re-proposal suppression"
version: "0.1.0"
status: in-progress
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness/safety, internal/harness/curator, internal/harness, internal/config"
lifecycle: spec-anchored
tags: "harness-evolve-epic, curator-wiring, tier-surface, l2-canary, l3-contradiction, negative-evidence, frozen-guard, re-proposal-suppression, glm-observe-only"
era: V3R6
tier: L
depends_on: [SPEC-HARNESS-EVOLVE-002]
---

# SPEC-HARNESS-EVOLVE-003 — Progress

> Plan-phase artifact. The §E.1 section carries the plan-phase audit-ready
> signal. §E.2 / §E.3 / §E.4 are placeholder headings emitted at plan-phase
> per the canonical progress.md §E skeleton contract; manager-develop
> populates §E.2/§E.3 at run-phase and manager-docs populates §E.4 at
> sync-phase. The `§E.2`-`§E.4` literal headings are parser-load-bearing
> (`internal/spec/era.go` `hasAnyProgressMarker` greps for them) — do NOT
> rename.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-12
- artifact_count: 6 (spec.md + plan.md + acceptance.md + design.md + research.md + progress.md)
- tier: L
- era: V3R6 (explicit frontmatter `era: V3R6`)
- depends_on: [SPEC-HARNESS-EVOLVE-002] — status: completed (origin fa2a3086a)
- REQ count: 35 (REQ-HEV3-001 … REQ-HEV3-035)
- AC count: see acceptance.md (AC-HEV3-001 … AC-HEV3-NNN, behavior-verifiable per gate)
- NEEDS CLARIFICATION: 4 items resolved (plan.md §I H-1..H-4 — all encoded in spec.md v0.1.1)

## §E.2 Run-phase Evidence

### M0 — A1 Frozen expansion + Frozen-rule registry skeleton

**AC PASS/FAIL matrix (M0 scope)**:

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-HEV3-023a | PASS | `grep -A8 'var frozenPrefixes' internal/harness/safety/frozen_guard.go \| grep -c 'settings\.json'` | `1` (≥1) |
| AC-HEV3-023a | PASS | `grep -A8 'var frozenPrefixes' internal/harness/safety/frozen_guard.go \| grep -c 'settings\.local\.json'` | `1` (≥1) |
| AC-HEV3-023b | PASS | `grep -A8 'var frozenPrefixes' internal/harness/safety/frozen_guard.go \| grep -c 'frozen_guard'` | `2` (≥1) |
| AC-HEV3-024a | PASS | `go test -run TestIsFrozen_PermissionSurfaces ./internal/harness/safety/` | `--- PASS` (settings.json FROZEN) |
| AC-HEV3-024b | PASS | (same test, settings.local.json subcase) | `--- PASS` |
| AC-HEV3-025 | PASS | `go test -run TestIsFrozen_GuardSelfProtection ./internal/harness/safety/` | `--- PASS` (guard source FROZEN) |
| AC-HEV3-031 (M0 scope) | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/safety/frozen_rules.go internal/harness/safety/frozen_guard.go \| grep -v _test.go \| grep -v '//'` | exit 1 (0 matches — clean) |

**Build/vet**:
```
$ go build ./...                     → exit 0
$ go vet ./internal/harness/safety/  → exit 0
$ go test ./internal/harness/safety/ → ok  github.com/modu-ai/moai-adk/internal/harness/safety  0.284s
```

**Files**:
- `internal/harness/safety/frozen_guard.go` — expanded frozenPrefixes (3 → 7 entries)
- `internal/harness/safety/frozen_rules.go` (NEW) — FrozenRuleRegistry + FindFrozenRule
- `internal/harness/safety/frozen_guard_test.go` — added A1 permission-surface + self-protection tests
- `internal/harness/safety/frozen_rules_test.go` (NEW) — registry seeding + lookup tests

### M1 — A7 Negative-evidence registry (commit 6d98f4b2e)

| AC | Status | Evidence |
|----|--------|----------|
| AC-HEV3-018a/b | PASS | NegativeEvidence struct + AppendNegativeEvidence + ReadNegativeEvidence compile + round-trip test green |
| AC-HEV3-022 | PASS | `grep -c 'cooldown_until.*null\|cooldown_until.*9999'` = 0 (no permanent suppression) |

### M2 — Tier↔Surface mapping + value-range validation (commit 99f6f8cb9)

| AC | Status | Evidence |
|----|--------|----------|
| AC-HEV3-001 | PASS | EvolvableSurfaces[auto_detection] registered (Tier 4, IsConfigEdit=true) |
| AC-HEV3-002a/b | PASS | ValidateAutoDetectionConditions rejects file_count<=0 + file_count<=99999 |
| AC-HEV3-003a/b | PASS | SurfaceForTier(3)→CLAUDE.local.md, SurfaceForTier(4)→CLAUDE.md |
| AC-HEV3-004 | PASS | ValidateSurfaceForTier rejects cross-surface pairing |

### M3 — L3 Contradiction activation (commit 46e9be16c)

| AC | Status | Evidence |
|----|--------|----------|
| AC-HEV3-015a | PASS | `sed -n '68,74p' pipeline.go \| grep -c 'return harness.ContradictionReport{}'` = 0 (no-op gone) |
| AC-HEV3-015b | PASS | DetectFrozenRuleContradictions wired in contradiction.go + pipeline.go |
| AC-HEV3-016 | PASS | TestPipeline_L3FrozenRule_RejectedBy3: RejectedBy==3 (behavior) |
| AC-HEV3-017 | PASS | L3 Reason cites rule identifier "frozen-moai-rules" |
| AC-HEV3-033 | PASS | L3 reachability proven per cycle |

### M4 — L2 Canary activation (commit d5b7891a9)

| AC | Status | Evidence |
|----|--------|----------|
| AC-HEV3-012 | PASS | ShadowApplyCanary evaluates in-memory without touching file |
| AC-HEV3-013 | PASS | TestShadowApplyCanary_ExceedsBudget: budget-exceeding block rejected |
| AC-HEV3-014 | PASS | RegisterVetoAsNegativeEvidence: veto→A7 outcome="rolled-back" |
| AC-HEV3-034 | PASS | L2 consult referenced in canary_shadow.go + pipeline.go |

### M5 — Production wiring breakthrough (commit 3b93542af)

| AC | Status | Evidence |
|----|--------|----------|
| AC-HEV3-005 | PASS | WriteManagedBlockGated(10) + ApprovalDecision(6) + RejectionRecorder(7) in curator_dispatch.go |
| AC-HEV3-006 | PASS | No direct WriteManagedBlock bypass (grep clean) |
| AC-HEV3-007a/b | PASS | Rejection path: file untouched + recorder "rejected" + A7 entry; Approval path: file written |
| AC-HEV3-008 | PASS | curator.TierGatedWrite production caller — 1 caller: `internal/harness/curator_dispatch.go:176` (criterion is `0 → ≥1` per acceptance.md; the earlier "4 matches" prose was inaccurate, corrected at sync) |
| AC-HEV3-009 | PASS | curator.WriteManagedBlockGated production caller — 1 caller: `internal/harness/curator_dispatch.go:187` (criterion is `0 → ≥1` per acceptance.md; the earlier "4 matches" prose was inaccurate, corrected at sync) |
| AC-HEV3-010 | PASS | PrepareTierDispatch sets content.Tier explicitly |
| AC-HEV3-011 | PASS | 0 dead-code/feature-flag guards |
| AC-HEV3-026 | PASS | GLM observe-only: IsGLMSession gate in dispatch layer |
| AC-HEV3-027 | PASS | model_class/IsGLM in curator_dispatch.go (12 matches) |
| AC-HEV3-035 | PASS | A7 registry consulted in Prepare (early-block test green) |

### M6 — Integration + quality gate (this commit)

| AC | Status | Evidence |
|----|--------|----------|
| AC-HEV3-028 | PASS | ValidatePatternKey rejects SPEC/REQ/AC tokens (anti-fabrication) |
| AC-HEV3-029a | PASS | auto_detection in template harness.yaml (unchanged, neutral empty schema) |
| AC-HEV3-029b | PASS | TestTemplateNoInternalContentLeak exit 0 |
| AC-HEV3-030a | PASS | curator 91.4%, safety 86.8% (new files 88-100%), harness 87.3% |
| AC-HEV3-030b | PASS | go build exit 0 + GOOS=windows go build exit 0 |
| AC-HEV3-031 | PASS | 0 AskUserQuestion refs in all new production code |
| AC-HEV3-032 | PASS | 0 new hook files (git diff empty) |

**Full run-phase build/test**:
```
$ go test ./...                          → exit 0 (all packages green)
$ golangci-lint run --timeout=3m         → 0 issues
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ go test -cover ./internal/harness/curator/ → coverage: 91.4% of statements
```

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-07-13
- run_commit_sha: pending-backfill-m6
- ac_pass_count: 35 (all Must-pass ACs verified with observed evidence)
- ac_fail_count: 0
- new_warnings_or_lints_introduced: 0
- cross_platform_build:
  darwin_amd64: PASS
  windows_amd64: PASS
- total_run_phase_files: 16 (6 SPEC artifacts + 10 Go source/test files)
- m1_to_mN_commit_strategy: per-milestone feat commits (M0=fc6e49a1c, M1=6d98f4b2e, M2=99f6f8cb9, M3=46e9be16c, M4=d5b7891a9, M5=3b93542af)

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-07-13
- sync_commit_sha: ca92344d3 (backfilled in follow-up commit — D3 self-referential-hazard exemption: the sync commit ca92344d3 could not reference its own hash in-commit)
- changelog_entry_added: true (standalone EVOLVE-003 bullet under `### Added`; `grep -n 'SPEC-HARNESS-EVOLVE-003' CHANGELOG.md` = 1 standalone entry + 1 prose mention inside the EVOLVE-002 bullet)
- frontmatter_transition: draft → completed (the intermediate `in-progress` step was skipped during the worktree-recovery integration of run commit `d61399742`; the `draft → in-progress` transition that manager-develop should have performed on the M1 run commit was missed, so this single sync commit carries `draft → completed` in one step — ownership gap recorded honestly; may surface as an OwnershipTransitionInvalid lint finding, which is an acceptable recovery artifact)
- readme_changed: false (curator production wiring is internal self-evolving-harness infrastructure with no user-facing CLI surface; same disposition as EVOLVE-001/002)
- docs_site_changed: false (same rationale)
- ac_reverification_at_sync: orchestrator independently re-ran `go build ./...` (exit 0 darwin+windows), `go test ./internal/harness/{curator,safety,harness}/` (all ok), `golangci-lint run --timeout=3m` (`0 issues.`), AC-HEV3-008/009 grep (1 caller each in `curator_dispatch.go`), AC-HEV3-031 subagent-boundary grep (0 matches), coverage (curator 91.4% / safety 86.8% / harness 87.3%) BEFORE the sync commit
- evidence_correction: §E.2 AC-HEV3-008/009 rows corrected from the inaccurate "4 matches" prose to the actual 1-caller-each evidence (both in `internal/harness/curator_dispatch.go`)
- manager_docs_delegation: attempted manager-docs subagent sync delegation; agent failed with autocompact-thrash (CHANGELOG.md 43KB context overflow) — orchestrator-direct fallback per runtime-recovery rung 1 (in-turn self-correction), resume message's "stale worktree 시 orchestrator-direct" fallback pattern
