---
id: SPEC-INTERNAL-TEST-004
title: "Regenerate stale doctor/status golden testdata for version bump rc7→rc10 (whole-repo green)"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0-rc10"
module: "internal/cli/testdata"
lifecycle: spec-anchored
tags: "golden-test, test-fix, debt-cleanup, version-bump"
tier: S
depends_on: []
related_specs: [SPEC-INTERNAL-TEST-003, SPEC-INTERNAL-ARCH-001, SPEC-WEB-CONSOLE-011]
---

# progress.md — SPEC-INTERNAL-TEST-004

## §A. Current Phase

**Sync-phase (consolidated close — completed).** The golden regeneration this SPEC was authored to specify was performed by EXTERNAL commit `ce2a509dc` (the rc10 version-bump commit), NOT by a TEST-004 own run-phase commit. TEST-004 therefore has no own run-phase commit and proceeds directly plan → consolidated sync close. See spec.md §G for the full "external absorption" resolution record. AC-007 (whole-repo green) is debt-transferred to SPEC-AGENT-ARCH-V2-001 (the 3 remaining `internal/template` FAILs belong to the super-advisor integration).

## §B. Artifact Status

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-INTERNAL-TEST-004/spec.md` | Created (GEARS, 5 REQ, 4 Out-of-Scope H3 sub-headings) |
| plan.md | `.moai/specs/SPEC-INTERNAL-TEST-004/plan.md` | Created (M1 golden regen + M2 whole-repo verify; PRESERVE list) |
| acceptance.md | `.moai/specs/SPEC-INTERNAL-TEST-004/acceptance.md` | Created (9 AC: 6 golden + whole-repo + git-diff + PRESERVE) |
| research.md | `.moai/specs/SPEC-INTERNAL-TEST-004/research.md` | Created (UPDATE_GOLDEN=1 + git diff byte-level evidence; scope decision) |
| progress.md | `.moai/specs/SPEC-INTERNAL-TEST-004/progress.md` | This file |

## §C. Investigation Anchors

| Concern | Location | Note |
|---------|----------|------|
| Doctor golden test source | `internal/cli/doctor_golden_test.go:91` | Mismatch assertion site |
| Status golden test source | `internal/cli/status_golden_test.go:142` | Mismatch assertion site |
| Golden testdata (6 files) | `internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden` | Run-phase modify targets (rc7→rc9, 1 line each) |
| Version source of truth | `pkg/version/version.go:8` — `Version = "v3.0.0-rc9"` | Frozen; goldens must match this |
| Version bump commit | `9edb72af5 chore(version): bump to v3.0.0-rc9` | HEAD-committed; rc7→rc9 drift window origin |
| Cache-hit emoji (PRESERVE) | `internal/statusline/renderer.go:312` — HEAD `💾`, working-tree `♻️` | WEB-CONSOLE-011 scope (ancestor `22220186c`); NOT TEST-004 |
| Cache-hit test (PRESERVE) | `internal/statusline/cache_hit_test.go` — 4 lines 💾→♻️ | WEB-CONSOLE-011 test sync; NOT TEST-004 |
| FAIL detail log (provenance) | `.moai/state/verify/538fe6ae/test-004-golden-fail-detail.log` | TEST-003-era evidence; whole-repo exit 1 root cause |
| TEST-003 AC-006 debt transfer | `.moai/specs/SPEC-INTERNAL-TEST-003/progress.md` §E.2 | External-debt transfer record |
| ARCH-001 M0 consumer | memory `project_internal_arch_001_plan_entry` | whole-repo-green precondition |

## §D. Mode Selection

_<pending Phase 0.95 — run-phase orchestrator populates this section>_

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: S
artifact_count: 5
artifact_set: [spec.md, plan.md, acceptance.md, research.md, progress.md]
frontmatter_schema_check: PASS (12 canonical fields, canonical names, tier: S)
ears_gears_compliance: PASS (5 REQ in GEARS notation — Ubiquitous + Event-detected + Capability gate)
out_of_scope_check: PASS (4 H3 sub-headings: cache-hit emoji / version.go bump / section logic / ARCH-001 M0)
spec_id_regex_check: PASS (SPEC-INTERNAL-TEST-004 matches ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$)
research_verdict: |
  Drift cause resolved: version-string rc7→rc9 is the SOLE root cause (UPDATE_GOLDEN=1 + git diff
  proves 6 files × 1 line × version-string-only). Cache-hit emoji (💾→♻️) is orthogonal (different
  package, WEB-CONSOLE-011 scope, PRESERVE). Approach: golden regenerate only, no code fix.
  No blocker report — all 4 core research questions answered from evidence.
scope_boundary_decision: |
  Cache-hit emoji belongs to SPEC-WEB-CONSOLE-011 (ancestor 22220186c). TEST-004 PRESERVEs the
  working-tree emoji edits uncommitted. TEST-004 touches ONLY internal/cli/testdata/*.golden (6 files).
debt_provenance: SPEC-INTERNAL-TEST-003 AC-006 (external debt transfer, TEST-003 completed)
unblock_target: SPEC-INTERNAL-ARCH-001 plan-audit M0 (whole-repo-green precondition)
head_at_authoring: bd97306af
origin_main_divergence: "0 0" (synced)
known_context_drift: |
  Orchestrator briefing stated HEAD=280c9dd71 and working-tree version.go=rc8; actual HEAD=bd97306af
  and version.go=rc9 (cleanly committed, no working-tree drift). Does not affect conclusions.
```

## §E.2 Run-phase Evidence

**External-absorption resolution — no own run-phase commit.**

The 6 golden testdata regeneration this SPEC was authored to specify was
performed by the EXTERNAL commit `ce2a509dc chore(version): bump to v3.0.0-rc10
and realign tag with source` (the rc10 version-bump commit), NOT by a TEST-004
own run-phase commit. TEST-004's run-phase was never invoked; the SPEC proceeds
directly from plan to a consolidated sync close (spec.md §G).

**Run-equivalent commit**: `ce2a509dc` (committed + pushed + tagged `v3.0.0-rc10`
on 2026-07-09).

### Evidence — 6 golden tests PASS (REQ-GOLD-001), re-verified at sync-phase

Command (re-run 2026-07-09, fresh observation — not inferred):

```
$ go test ./internal/cli/ -run 'TestDoctor_|TestStatus_' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.663s
```

exit=0 — all 6 golden tests (`TestDoctor_{Current_Light,Current_Dark,NoColor}` +
`TestStatus_{Current_Light,Current_Dark,NoColor}`) PASS. Evidence log:
`/tmp/moai-verify/1-golden.log`.

### Evidence — version-string-only diff (REQ-VER-001)

`git show --stat ce2a509dc` confirms the commit touched exactly the 6 golden
files (`internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden`, 2
lines each = 1 insertion + 1 deletion = the version-string line rc7→rc10) plus
`pkg/version/version.go`, `.moai/config/sections/system.yaml`, and
`CHANGELOG.md`. No other golden mutation, no section-rendering change. The
commit message states verbatim: *"Regenerate the six doctor/status golden
fixtures, which still pinned rc7 and were the cause of the standing
TestDoctor/TestStatus failures on internal/cli."* Evidence log:
`/tmp/moai-verify/2-ce2a509dc-stat.log`.

### Evidence — PRESERVE holds (REQ-PRESERVE-001 / REQ-SCOPE-001)

`ce2a509dc` did NOT touch any PRESERVE path (no `internal/statusline/`, no
unrelated working-tree path). The TEST-004 sync-phase commit (this commit)
likewise touches ONLY `.moai/specs/SPEC-INTERNAL-TEST-004/` artifacts — verified
via `git show --stat HEAD` (§E.4). The cache-hit emoji edits
(`internal/statusline/renderer.go` + `cache_hit_test.go`, WEB-CONSOLE-011 scope)
remain uncommitted in the working tree, untouched.

### Evidence — AC-007 whole-repo green is NOT satisfied (DEBT-TRANSFERRED)

`go test ./...` still exits 1. The 3 remaining FAILs are in `internal/template`
(`TestAllAgentsInCatalog`, `TestTemplateNoInternalContentLeak`,
`TestRuleProvenanceAudit`) and belong to SPEC-AGENT-ARCH-V2-001 (super-advisor
agent integration in progress), NOT to the golden drift. AC-007 is
debt-transferred to SPEC-AGENT-ARCH-V2-001 / SPEC-INTERNAL-ARCH-001.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-09
run_commit_sha: ce2a509dc   # EXTERNAL absorption commit (NOT a TEST-004 own run commit)
run_commit_subject: "chore(version): bump to v3.0.0-rc10 and realign tag with source"
close_shape: consolidated-3-phase-close   # feedback_small_spec_consolidated_lifecycle (≤5 files, metadata-only, absorbed-run)
ac_satisfaction:
  AC-001: SATISFIED-via-ce2a509dc   # TestDoctor_Current_Light PASS
  AC-002: SATISFIED-via-ce2a509dc   # TestDoctor_Current_Dark PASS
  AC-003: SATISFIED-via-ce2a509dc   # TestDoctor_NoColor PASS
  AC-004: SATISFIED-via-ce2a509dc   # TestStatus_Current_Light PASS
  AC-005: SATISFIED-via-ce2a509dc   # TestStatus_Current_Dark PASS
  AC-006: SATISFIED-via-ce2a509dc   # TestStatus_NoColor PASS
  AC-007: DEBT-TRANSFERRED          # whole-repo green — 3 internal/template FAILs belong to SPEC-AGENT-ARCH-V2-001
  AC-008: SATISFIED-via-ce2a509dc   # version-string-only diff (rc7→rc10)
  AC-009: HOLDS                     # PRESERVE — no TEST-004 action touched any out-of-scope path
ac_count: 9   # acceptance.md §D AC Matrix (AC-001..009); format uses bare "AC-001" not "**AC-XXX-001**"
req_satisfaction:
  REQ-GOLD-001: MET-via-ce2a509dc       # 6 golden tests PASS
  REQ-GREEN-001: DEBT-TRANSFERRED-to-SPEC-AGENT-ARCH-V2-001   # whole-repo NOT exit 0
  REQ-VER-001: MET-via-ce2a509dc        # version-string-only diff
  REQ-PRESERVE-001: HOLDS               # scope discipline
  REQ-SCOPE-001: HOLDS                  # cache-hit emoji boundary preserved
debt_transferred:
  - id: AC-007
    owner: SPEC-AGENT-ARCH-V2-001
    root_cause: "3 internal/template FAILs (TestAllAgentsInCatalog, TestTemplateNoInternalContentLeak, TestRuleProvenanceAudit) — super-advisor agent integration surface, NOT golden drift"
whole_repo_go_test_exit: 1   # NOT exit 0 — debt-transferred (see AC-007)
verification_basis: |
  6 golden tests re-run at sync-phase (exit 0, fresh observation);
  ce2a509dc stat inspected (6 goldens + version.go + system.yaml + CHANGELOG only);
  PRESERVE verified via git status (cache-hit emoji edits remain uncommitted, untouched).
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: completed
sync_complete_at: 2026-07-09
sync_commit_sha: "4faeb55a9"   # close commit SHA; backfilled in follow-up chore commit per established convention (TEST-002 / SPEC-TEMPLATE-RULES-CLEANUP-001 pattern — single-commit self-SHA is mathematically impossible: the SHA depends on the tree which contains the SHA)
sync_commit_subject: "chore(SPEC-INTERNAL-TEST-004): sync-phase artifacts + 3-phase close"
frontmatter_status_transitions:
  spec_md: draft → completed        # manager-docs owns the merged 3-phase close on the single sync commit
  plan_md: draft → completed
  acceptance_md: draft → completed
  progress_md: draft → completed
b12_self_test:
  a_pre_emission_grep: PASS          # grep -c 'SPEC-INTERNAL-TEST-004' CHANGELOG.md == 0 (no duplicate)
  b_ac_count_match: N/A             # CHANGELOG entry intentionally skipped (see changelog_entry_position)
  c_file_path_verification: N/A     # no CHANGELOG edit this commit
changelog_entry_position: none       # intentionally skipped — golden regen already covered by released [v3.0.0-rc10] entry (ce2a509dc); TEST-004 has no own user-facing code delta
canary_compliance_check:
  spec_md_frontmatter_status: completed
  all_4_artifacts_status_atomic: true
  git_show_stat_contains_only_spec_dir: true   # only .moai/specs/SPEC-INTERNAL-TEST-004/ paths in this commit
consolidated_close_rationale: |
  feedback_small_spec_consolidated_lifecycle pattern — ≤5 files (spec/plan/acceptance/progress/research),
  metadata-only close (frontmatter status + §E evidence; no code delta). Run-phase work absorbed by
  external commit ce2a509dc; the draft → completed transition rides this single sync commit.
```

## §F. Phase 0.95 Mode Selection

_<pending Phase 0.95 — orchestrator populates before first run-phase Agent() spawn>_

## §H. Recursive Self-Diagnosis Log

_<pending run-phase — manager-develop / orchestrator populates on mechanical-failure diagnosis loop>_

## §I. Token Accounting

_<pending sync-close — token-accounting mechanism populates at sync-close>_
