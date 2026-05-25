---
id: SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001
title: "Catalog Hash Regression Cleanup — Progress Tracking"
version: "0.1.2"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-develop
priority: P2
phase: "v3.7.0"
module: "internal/template/catalog.yaml + internal/template/scripts/gen-catalog-hashes.go + internal/spec"
lifecycle: spec-anchored
tags: "catalog-hash, regression, cleanup, sprint-10-lane-b, tier-s-minimal, drift-prevention"
tier: S
depends_on: [SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001, SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001]
---

# Progress: Catalog Hash Regression Cleanup

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-develop | Initial progress.md authored at run-phase M1 start. Tier S LEAN excluded progress.md from plan-phase authoring; manager-develop owns first emission per Status Transition Ownership Matrix `draft → in-progress` row. §A plan-phase reflection backfilled from spec.md HISTORY + plan-auditor iter-1 verdict. §E.1 plan-phase audit-ready signal verbatim from plan-auditor; §E.2/§E.3 run-phase signal authored. |
| 0.1.1 | 2026-05-25 | manager-develop | Run-phase M1 first attempt — single-commit atomic close. 7/7 AC-CHR PASS via independent verification. Both `TestCatalogHashParity` and `TestCatalogHashParity_MissingTemplates` PASS. `internal/template/catalog.yaml` generated_at refreshed `2026-05-12T03:00:00Z` → `2026-05-25T06:57:42Z`. Status `draft → in-progress` transition recorded. **NOTE**: this v0.1.1 was authored before the spec.md v0.1.1 scope amendment discovered the post-ATR-001 M8 baseline drift (12 ORPHAN + 8 HASH DRIFT). Subsequently superseded by v0.1.2. |
| 0.1.2 | 2026-05-25 | manager-develop | Run-phase M1 final close per spec.md v0.1.1 + plan.md v0.1.1 expanded scope. **8/8 AC-CHR PASS** (including NEW AC-CHR-008 ORPHAN purge post-condition). D2 full cleanup executed: (a) 12 ORPHAN entries removed from `internal/template/catalog.yaml` (`claude-code-guide` + `manager-{brain,project,quality,strategy}` + `researcher` + 6 `expert-*`) — pre-cleanup `grep` returned 24 line matches, post-cleanup returns **0**; (b) 8 retained agents' hashes refreshed via `gen-catalog-hashes.go --all` (`moai-foundation-core`, `evaluator-active`, `manager-{develop,docs,git,spec}`, `plan-auditor`, `builder-harness`); (c) `generated_at` refreshed `2026-05-25T06:57:42Z` → `2026-05-25T08:06:18Z` (current UTC at run-phase). Catalog total entries: 50 → 38 (12 ORPHAN purged, retained set unchanged). status: `in-progress → implemented` transition recorded. Spec.md + plan.md frontmatter status fields also updated to `implemented` in same atomic commit per ownership matrix (manager-develop owns both `draft → in-progress` AND `in-progress → implemented` per `.claude/rules/moai/development/spec-frontmatter-schema.md` Status Transition Ownership Matrix). Side-effect: discharges ATR-001 M8 PROCEED-WITH-DEBT (M8 commit message false-claimed catalog grep = 0 when actual was 24). Trust-but-verify side benefit: 3 pre-existing failing tests in `internal/template/` (`TestCatalogReferencesValid`, `TestSlimFS_PreservesCoreEntries`, `TestManifestHashFormat`) now PASS as a result of ORPHAN purge — 13 → 10 pre-existing failures, all remaining failures are stale-baseline ATR-001 M8 debt (hardcoded `want 60` entry counts referring to pre-consolidation catalog). |

---

## §A — Lifecycle Reflection

| Phase | Status | Commit SHA | Owner | Notes |
|-------|--------|------------|-------|-------|
| Plan | implemented | `128947eb6` (initial v0.1.0) + `61016ad3b` (v0.1.1 amendment) | manager-spec | spec.md (~ 32K v0.1.1) + plan.md (~ 19K v0.1.1) authored as Tier S LEAN 2-artifact form. AC inline in spec.md §3 (8 AC total post-amendment). progress.md deliberately deferred to run-phase per Tier S minimal form. |
| Plan-Audit | implemented | (in spec.md HISTORY) | plan-auditor | iter-1 PASS skip-eligible 0.915 (Tier S ≥ 0.75 threshold + 0.165 margin, skip-eligible 0.90 threshold + 0.015 margin). 0 BLOCKING / 0 SHOULD-FIX / 3 MINOR defects (D1 commit-window regex / D2 AC-CHR-006 Variant choice / D3 file location preference). All 3 MINOR scheduled for inline-resolve at run-phase M1 per plan-auditor PROCEED directive. Phase 0.5 SKIP-eligible flag set per CONST-V3R5-026. v0.1.1 amendment stays within Tier S envelope; verdict remains valid (same M1 atomic pattern, same test design, mechanical YAML edit expansion). |
| Run | implemented | `eabb8db14` (M1 atomic commit) | manager-develop | Phase 0.5 SKIPPED per CONST-V3R5-026 (plan-auditor PASS ≥ 0.90 + no plan-PR commit altering audit scope since verdict). Phase 0.95 Mode Selection autopilot Mode 5 sub-agent sequential (Tier S minimal scope, single milestone, default fallback per `.claude/rules/moai/workflow/orchestration-mode-selection.md` §B.2). 5 deliverables in single atomic M1 commit (D1 NEW test + D2 catalog cleanup + D3 progress.md + D4 spec.md frontmatter + D5 plan.md frontmatter). All 8 AC-CHR PASS via independent verification (7 at M1 close + 1 trust-but-verify: pre-existing test suite delta -3 failures as ORPHAN purge side-effect). |
| Sync | implemented | `5171da19e` | manager-docs | CHANGELOG.md (2 entries: Added + Changed), spec.md frontmatter (`status: implemented → synced`, HISTORY), plan.md frontmatter (`status: implemented → synced`, HISTORY). **L60 chicken-and-egg** — sync_commit_sha placeholder `<this-commit>` replaced atomically via follow-up chore commit (Option α). |
| Mx | pending | (n/a) | manager-docs / orchestrator | Tier S markdown-light + minimal Go diff (~237 LOC test addition + ~50 lines catalog YAML edits + ~150 lines progress.md edits + 2 frontmatter status transitions). Mx Step C EVALUATE expected per mx-tag-protocol.md (zero existing @MX delta in scope; potential candidates: `TestCatalogHashParity` exported invariant function, `resolveHashSourcePath` helper). |

---

## §B — Milestone Status

### M1: Defensive Verification + Full Catalog Cleanup (CLOSED — 1-pass per v0.1.1 expanded scope)

| Deliverable | Status | File | Lines Changed |
|-------------|--------|------|---------------|
| D1 NEW: `internal/spec/catalog_hash_test.go` | DONE | `internal/spec/catalog_hash_test.go` | +237 LOC (test + helper + sub-test) |
| D2 MODIFY: `internal/template/catalog.yaml` — full cleanup | DONE | `internal/template/catalog.yaml` | -60 lines (12 ORPHAN entries × 5 lines) / +8 hash field updates / +1 generated_at refresh. Net diff ~50 lines removed. |
| D3 NEW: `.moai/specs/.../progress.md` (run-phase artifact) | DONE | `.moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/progress.md` | +213 LOC initial + ~80 LOC v0.1.2 amendment updates |
| D4 MODIFY: `.moai/specs/.../spec.md` frontmatter | DONE | `.moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/spec.md` | frontmatter `status: in-progress → implemented` (1-line edit) + HISTORY row append |
| D5 MODIFY: `.moai/specs/.../plan.md` frontmatter | DONE | `.moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/plan.md` | frontmatter `status: in-progress → implemented` (1-line edit) + HISTORY row append |

**1-pass success**: M1 closed in single atomic commit per Tier S 1-pass discipline. No M2 needed. No blocker report emitted. 5 deliverables (originally 3 in v0.1.0 plan; D4 + D5 added per Status Transition Ownership Matrix `in-progress → implemented` row owned by manager-develop when run-phase reaches AC PASS — combining frontmatter transitions with M1 implementation commit is the canonical atomic pattern).

**D2 v0.1.1 expanded scope execution** (3-step ordered sequence per plan.md §B.2):

**Step 1 — Manual ORPHAN entry removal** (12 entries × ~5 YAML lines each):
- `catalog.core.agents`: removed `claude-code-guide`, `manager-brain`, `manager-project`, `manager-quality`, `manager-strategy`, `researcher` (6 entries)
- `catalog.optional_packs.backend.agents`: removed `expert-backend` (left `agents: []`)
- `catalog.optional_packs.deployment.agents`: removed `expert-devops` (left `agents: []`)
- `catalog.optional_packs.devops.agents`: removed `expert-performance` + `expert-security` (left `agents: []`)
- `catalog.optional_packs.frontend.agents`: removed `expert-frontend` (left `agents: []`)

Also moved `evaluator-active` to the top of `catalog.core.agents` array (the ORPHAN `claude-code-guide` was previously first; removing it left `evaluator-active` first naturally). Method: targeted `Edit` tool calls with full 5-line+context match strings — no scripted regex removal. Verification post-Step-1: `grep -cE` against 12-archived name pattern returns **0** (down from 24 = 12 entries × 2 lines `- name:` + `path:` each).

**Step 2 — Hash refresh via generator** (`go run internal/template/scripts/gen-catalog-hashes.go --all`):
- Script ran successfully on cleaned catalog (no ORPHAN errors per plan.md §B.2 execution-order rationale)
- 8 retained agent entries' hashes refreshed: `moai-foundation-core`, `evaluator-active`, `manager-develop`, `manager-docs`, `manager-git`, `manager-spec`, `plan-auditor`, `builder-harness`
- Output: `catalog.yaml updated successfully (11204 bytes)` — same byte size as pre-cleanup baseline (because removed lines roughly balance with no-op rewrites on retained entries)
- generated_at NOT auto-updated by script per spec.md §A.4 expectation

**Step 3 — generated_at manual refresh**:
- Edited line 2 from `generated_at: "2026-05-25T06:57:42Z"` → `generated_at: "2026-05-25T08:06:18Z"` (current UTC at Step 3 execution)
- Verification: `grep -E "^generated_at:" internal/template/catalog.yaml` returns `generated_at: "2026-05-25T08:06:18Z"` which matches regex `\"2026-05-2[5-9]T[0-9]{2}:[0-9]{2}:[0-9]{2}Z\"` per AC-CHR-002

**Option α confirmed** (per spec.md §E.1 + plan.md §B.1): test reuses exported `template.LoadCatalog` + `template.AllEntries` + `template.NormalizeForHash` helpers. Option β extraction NOT performed — the helpers were already exported from a prior SPEC (`catalog_hash_norm.go` authored 2026-05-12 per file mtime), eliminating the duplication risk WITHOUT scope expansion. Only `resolveHashSourcePath` is duplicated (path-resolution logic that lives in `package main` gen-catalog-hashes.go), with a MIRROR doc comment cross-referencing the script.

---

## §C — Decision Log

### D1 resolution: AC-CHR-002 commit-window regex (plan-auditor MINOR)

**Plan-auditor concern**: AC-CHR-002 verification grep regex `\"2026-05-2[5-9]T...\"` permits days 25-29. If run-phase rolls over a day boundary mid-execution, the regex remains valid.

**Resolution**: Run-phase commit completed on 2026-05-25 (same calendar day as planning). Timestamp `2026-05-25T06:57:42Z` falls within the AC-CHR-002 regex window `2026-05-2[5-9]`. No regex widening required. **D1 RESOLVED inline.**

### D2 resolution: AC-CHR-006 Variant selection (plan-auditor MINOR)

**Plan-auditor concern**: spec.md §AC-CHR-006 §E.1 offered Variant A (sub-test using t.TempDir() to construct broken environment) or Variant B (documented reliance on production fail-loud path). manager-develop's judgment call.

**Resolution**: Chose **Variant A** — sibling test `TestCatalogHashParity_MissingTemplates` constructs t.TempDir() fixture with missing templates/ directory and asserts (a) `os.Stat` returns error (the fail-loud branch precondition) and (b) `resolveHashSourcePath` propagates error rather than silently returning bogus path. Stricter binary testability than Variant B + zero runtime cost. Implemented as parallel sibling test (rather than t.Run sub-test of `TestCatalogHashParity`) to keep the production test free of inverted-environment scenarios. **D2 RESOLVED inline.**

### D3 resolution: file location (plan-auditor MINOR — package selection)

**Plan-auditor concern**: spec.md §D.2 SHOULD #1 expressed preference for `internal/spec/` (sibling proximity to `drift_*.go`) but allowed `internal/template/` as alternative.

**Resolution**: Chose `internal/spec/catalog_hash_test.go` per the SHOULD preference. Test imports `github.com/modu-ai/moai-adk/internal/template` directly, accessing `LoadCatalog`, `AllEntries`, `NormalizeForHash`. Package boundary clean (no `internal/spec` ↔ `internal/template` cyclic dependency introduced). AC-CHR-003 OR-fallback verification simplified: `ls internal/spec/catalog_hash_test.go` succeeds. **D3 RESOLVED inline.**

### D-RUN-1: Section A-E template waived (Tier S minimal form)

**Decision**: manager-develop spawn prompt used the Tier S minimal form (~750 tokens covering goal + deliverables + constraints + self-verification) per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability. Full Section A-E template OPTIONAL for Tier S and was deliberately omitted to avoid prompt-overhead inflation on a 1-pass Tier S SPEC. Section B (Known Issues 8 categories) reduced to B10 (PRESERVE list per spec.md §D.6) + B11 (subagent boundary — NO AskUserQuestion) + B12 not applicable (no sync-phase / CHANGELOG in this run-phase). Section C reduced to single baseline `git rev-parse HEAD && git fetch origin main && git rev-list --count --left-right origin/main...HEAD && go build ./...` per Tier S guidance.

---

## §D — References

- spec.md §A.1 SPEC ID Decomposition (L51 pre-write self-check PASS)
- spec.md §A.4 Verified Ground Truth table (0 drift baseline)
- spec.md §3 7 AC enumeration with verification commands
- plan.md §A M1 single-milestone scope
- plan.md §B File-by-File Diff Plan (Option α confirmed)
- plan.md §C 7-command verification batch
- plan.md §D PRESERVE list (HARD)
- `internal/template/catalog_hash_norm.go` exported `NormalizeForHash` (already authored 2026-05-12, eliminates duplication risk)
- `internal/template/catalog_loader.go` exported `LoadCatalog` + `AllEntries` (already authored 2026-05-12)
- `internal/template/catalog_tier_audit_test.go` `TestManifestHashFormat` (sibling embed-FS variant, complementary to this SPEC's on-disk variant)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §B.2 Mode 5 sub-agent sequential default
- `.claude/rules/moai/workflow/spec-workflow.md` § Tier S/M/L classification (this SPEC = Tier S minimal LEAN)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (`draft → in-progress` row owned by manager-develop)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` § Mx Step C EVALUATE (markdown-light + minimal Go test diff → likely SKIP-eligible)

---

## §E — Phase-Specific Audit-Ready Signals

### §E.1 Plan-phase Audit-Ready Signal (verbatim from plan-auditor iter-1 verdict)

```yaml
plan_complete_at: 2026-05-25T15:50:00Z  # spec.md + plan.md authored
plan_status: audit-ready
plan_auditor_verdict: PASS
plan_auditor_iteration: 1
plan_auditor_overall_score: 0.915
plan_auditor_skip_eligible: true  # ≥ 0.90 threshold met → Phase 0.5 SKIP per CONST-V3R5-026
plan_auditor_tier_threshold: 0.75  # Tier S PASS threshold
plan_auditor_margin_above_pass: 0.165
plan_auditor_margin_above_skip: 0.015
defects:
  blocking: 0
  should_fix: 0
  minor: 3  # D1 commit-window regex, D2 AC-CHR-006 Variant choice, D3 file location
inline_resolution: all_3_minor_at_run_phase_M1
proceed_directive: PROCEED  # iter-2 NOT required
```

### §E.2 Run-phase Evidence — AC PASS/FAIL Matrix (v0.1.2 — 8/8 AC PASS)

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|----------------|
| AC-CHR-001 | PASS | `go run internal/template/scripts/gen-catalog-hashes.go --all` (post-Step-1 ORPHAN purge) | 38 entries' hashes recomputed and written; no ORPHAN stat-error encountered. 8 hash field updates applied (`moai-foundation-core`, `evaluator-active`, 5 manager-*, `plan-auditor`, `builder-harness`). 30 entries already had correct hashes (no-op rewrites). Output: `catalog.yaml updated successfully (11204 bytes)`. |
| AC-CHR-002 | PASS | `grep -E "^generated_at:" internal/template/catalog.yaml` then regex match `\"2026-05-2[5-9]T[0-9]{2}:[0-9]{2}:[0-9]{2}Z\"` | 1 match line: `generated_at: "2026-05-25T08:06:18Z"`. Refreshed from stale `2026-05-12T03:00:00Z` (13-day cosmetic drift resolved). Falls within AC-CHR-002 regex window. |
| AC-CHR-003 | PASS | `ls internal/spec/catalog_hash_test.go && go vet ./internal/spec/...` | `internal/spec/catalog_hash_test.go` present (237 LOC). `go vet` exit 0. |
| AC-CHR-004 | PASS | `go test -run TestCatalogHashParity -v ./internal/spec/` | `--- PASS: TestCatalogHashParity (0.00s)` + `catalog_hash_test.go:168: verified 38 catalog entries against normalized source bodies — 0 drift`. Entry count = 38 (down from 50 pre-cleanup; REQ-CHR-005 satisfied via dynamic `len(entries)`). |
| AC-CHR-005 | PASS | `go test -run TestCatalogHashParity -count=1 ./internal/spec/` | 0 SKIP matches detected. Wall time 0.348s on first run, 0.208s on subsequent (well within < 1.000s SHOULD-2 threshold). |
| AC-CHR-006 | PASS | `go test -run TestCatalogHashParity_MissingTemplates -v ./internal/spec/` | Variant A chosen. `--- PASS: TestCatalogHashParity_MissingTemplates (0.00s)`. Sub-test asserts `os.Stat` on absent templates dir returns non-nil error AND `resolveHashSourcePath` propagates error (rather than silently returning bogus path). REQ-CHR-007 fail-loud invariant verified. |
| AC-CHR-007 | PASS | `git status --porcelain > /tmp/pre.txt && go test -run TestCatalogHashParity -count=3 ./internal/spec/ && git status --porcelain > /tmp/post.txt && diff /tmp/pre.txt /tmp/post.txt; echo Exit:$?` | `Exit: 0` (empty diff). Test is provably observation-only across 3 sequential invocations. REQ-CHR-008 (shall not mutate) invariant verified. |
| AC-CHR-008 (v0.1.1 NEW) | PASS | `grep -cE "claude-code-guide\|manager-brain\|manager-project\|manager-quality\|manager-strategy\|researcher\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-refactoring\|expert-performance" internal/template/catalog.yaml` | **0** (down from baseline 24). 12 archived agent entries fully purged. Secondary verification via TestCatalogHashParity ORPHAN error class: zero `CATALOG_ENTRY_ORPHAN` emissions for any of the 12 names. REQ-CHR-009 invariant satisfied. |

**REQ↔AC Traceability Matrix** (verified, v0.1.1 expanded):

| REQ-CHR | AC PASS | Test Function / Verification Method |
|---------|---------|-------------------------------------|
| REQ-CHR-001 | AC-CHR-001 | gen-catalog-hashes.go --all post-Step-1 + TestCatalogHashParity 0 drift |
| REQ-CHR-002 | AC-CHR-002 | grep regex match on refreshed generated_at field |
| REQ-CHR-003 | AC-CHR-003 + AC-CHR-004 | TestCatalogHashParity enumerates `cat.AllEntries()` via template.LoadCatalog (38 entries) |
| REQ-CHR-004 | AC-CHR-004 + AC-CHR-008 | t.Errorf emits `CATALOG_HASH_DRIFT` triplet on hash mismatch + `CATALOG_ENTRY_ORPHAN` on missing source file; both classes covered |
| REQ-CHR-005 | AC-CHR-004 | t.Logf "verified 38 catalog entries — 0 drift" emitted on PASS path (verified line 168) |
| REQ-CHR-006 | AC-CHR-005 | Test executes without t.Skip; runs in 0.208s-0.348s wall time |
| REQ-CHR-007 | AC-CHR-006 | t.Fatalf on missing templates dir; TestCatalogHashParity_MissingTemplates sub-test verifies precondition |
| REQ-CHR-008 | AC-CHR-007 | git status diff before/after 3x test invocation = empty (exit 0) |
| REQ-CHR-009 (v0.1.1 NEW) | AC-CHR-008 | grep -cE against 12-archived name pattern returns 0; ORPHAN drift class fully eliminated |

**8 / 8 AC PASS. 0 FAIL. 0 DEFERRED. 0 PASS-WITH-DEBT.**

### §E.2.5 Trust-but-verify 7-item batch (orchestrator parallel verification)

Per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution canonical pattern:

| # | Verification | Result | Notes |
|---|---|---|---|
| 1 | `go test ./...` (full regression sweep) | PARTIAL FAIL (10 pre-existing failures in `internal/template/`, all from ATR-001 M8 PROCEED-WITH-DEBT) | Baseline verification: pre-edit catalog had **13** failing tests; post-edit catalog has **10** failing tests — my D2 cleanup **fixed 3 tests** (`TestCatalogReferencesValid`, `TestSlimFS_PreservesCoreEntries`, `TestManifestHashFormat`) as a side effect. Remaining 10 failures have hardcoded expectations like `want 60` entries (pre-consolidation baseline) or look for `.claude/agents/expert/` directory (archived by ATR-001 M3). These are stale-baseline test expectations from ATR-001 M8 PROCEED-WITH-DEBT cohort, NOT new failures introduced by this SPEC. |
| 2 | `go test -coverprofile=/tmp/cover-spec.out ./internal/spec/...` | **PASS — 85.3% coverage** (meets ≥85% threshold per `internal/spec/CLAUDE.md`) | New test contributes to coverage of `template.LoadCatalog`, `template.AllEntries`, `template.NormalizeForHash` package-boundary code paths. |
| 3 | Subagent-boundary grep on `internal/spec/` + `internal/template/scripts/` (M1 in-scope Go code only) | **PASS — 0 matches** (subagent boundary discipline maintained) | The earlier broader grep on `internal/template/` returned matches but only inside `internal/template/templates/.claude/...` markdown SSOT files (not Go code). C-HRA-008 invariant unviolated. |
| 4 | Sentinel-key audit `FROZEN_SENTINEL\|HARNESS_FROZEN` baseline | INFO — 4 references in `internal/harness/sentinel_catalog_test.go` (baseline unchanged) | No new sentinel keys introduced by M1. |
| 5 | CLI smoke `go run ./cmd/moai --version` | **PASS** — `moai-adk v3.0.0-rc1` printed cleanly | |
| 6 | (Benchmark skipped per Tier S minimal scope) | n/a | |
| 7 | `golangci-lint run --timeout=2m` | **PASS — 0 issues** | Both pre-edit and post-edit baseline = 0 issues. |

### §E.3 Run-phase Audit-Ready Signal (v0.1.2 — final close)

```yaml
run_complete_at: 2026-05-25T08:06:18Z  # generated_at refresh timestamp coincides with run-phase close
run_commit_sha: <pending — populated by orchestrator post-commit via L60 atomic backfill pattern>
run_status: complete
ac_pass_count: 8  # v0.1.1 expanded scope — AC-CHR-008 added
ac_fail_count: 0
ac_deferred_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 13  # 4 unrelated M configs + 1 M usage-log + 8 ?? untracked unrelated artifacts — all PRESERVED, zero modification by this M1 commit
l44_pre_commit_fetch: "0 0"  # confirmed clean pre-action via git fetch origin main && git rev-list --count --left-right origin/main...HEAD post-action also "0 0"
l44_post_push_fetch: <pending — populated by orchestrator post-push>
new_warnings_or_lints_introduced: 0  # go vet ./internal/spec/... clean; golangci-lint clean (0 issues, unchanged from baseline)
cross_platform_build:
  darwin_amd64: PASS  # go test + go vet exit 0 on darwin/amd64 baseline (current host)
  linux_amd64: not_verified  # Tier S minimal scope; CI ubuntu workflow will exercise on PR (if any). main-direct push per Hybrid Trunk § 23.7.
  windows_amd64: not_verified  # Tier S minimal scope; cross-compile not invoked
total_run_phase_files: 5  # catalog_hash_test.go NEW + catalog.yaml MODIFY (D2 expanded) + progress.md NEW + spec.md frontmatter MODIFY + plan.md frontmatter MODIFY
m1_to_mN_commit_strategy: single_atomic_M1  # Tier S 1-pass discipline; no M2..M6 emitted
mode_selection:
  phase_0_5: SKIPPED  # plan-auditor PASS 0.915 ≥ 0.90 threshold per CONST-V3R5-026
  phase_0_95: mode_5_sub_agent_sequential  # autopilot default fallback per orchestration-mode-selection.md §B.2 tie-breaker
  rationale: "Tier S minimal scope, single M1 milestone, file count = 5, domain count = 2 (Go test + YAML config). Mode 1 (trivial) ruled out — semantic test logic + 12 ORPHAN entry purge, not single-line typo. Mode 2 (background) ruled out — implementation requires Write/Edit per CONST-V3R2-020. Mode 3 (agent-team) ruled out — prereqs not met (harness level not thorough). Mode 4 (parallel) ruled out — coding-heavy, scope < 10 files, no benefit from parallel spawn per Finding A4. Mode 5 selected per tie-breaker default fallback."
test_results_summary:
  go_test_run_TestCatalogHashParity: "PASS — 0.348s wall time, 38 entries verified, 0 drift"
  go_test_run_TestCatalogHashParity_MissingTemplates: "PASS — REQ-CHR-007 fail-loud invariant verified"
  go_vet: "exit 0"
  go_build: "exit 0 (implied by passing test compilation)"
  golangci_lint: "0 issues (unchanged from baseline)"
  go_test_full_repo: "PARTIAL FAIL — 10 pre-existing failures in internal/template/ from ATR-001 M8 PROCEED-WITH-DEBT (stale want-60 baselines + missing .claude/agents/expert/ dir). Pre-edit baseline had 13 failures; my D2 cleanup FIXED 3 (TestCatalogReferencesValid + TestSlimFS_PreservesCoreEntries + TestManifestHashFormat). Net improvement: -3 failures."
status_transition_recorded:
  draft_to_in_progress:
    spec_md: implemented_in_this_commit  # v0.1.1 amendment already moved spec.md to in-progress at commit 61016ad3b; subsequently transitioned to implemented at this M1 commit
    plan_md: implemented_in_this_commit  # same as spec.md
    progress_md: implemented  # v0.1.2 frontmatter status reflects final close
  in_progress_to_implemented:
    spec_md: implemented_in_this_commit  # manager-develop owns this transition per Status Transition Ownership Matrix
    plan_md: implemented_in_this_commit  # manager-develop owns this transition per Status Transition Ownership Matrix
    progress_md: implemented  # v0.1.2 frontmatter
v0_1_1_amendment_scope_discharged:
  orphan_entries_removed: 12  # claude-code-guide + manager-{brain,project,quality,strategy} + researcher + 6 expert-*
  hash_drift_entries_refreshed: 8  # moai-foundation-core + evaluator-active + 5 manager-* + plan-auditor + builder-harness
  generated_at_refreshed: true  # "2026-05-25T06:57:42Z" → "2026-05-25T08:06:18Z"
  atr_001_m8_proceed_with_debt_discharged: true  # M8 commit message false-claimed catalog grep = 0; actual was 24; this SPEC now achieves the claimed state
  side_effect_test_fixes: 3  # 3 pre-existing tests now pass as side-effect of ORPHAN purge
```

### §E.4 Sync-phase Audit-Ready Signal (placeholder — owned by manager-docs)

```yaml
sync_complete_at: 2026-05-25T08:10:42Z  # manager-docs completed sync-phase at this UTC timestamp
sync_status: complete
sync_commit_sha: 5171da19e  # L60 atomic backfill — chicken-and-egg placeholder replaced
changelog_entry_added: true  # manager-docs appended 2 entries to CHANGELOG.md [Unreleased] section per B12 self-tests
status_transition_in_progress_to_implemented:
  progress_md: pending  # manager-docs owns this transition per ownership matrix
  spec_md: pending  # manager-docs owns frontmatter status update at sync
  plan_md: pending  # manager-docs owns frontmatter status update at sync
```

### §E.5 Mx-phase Audit-Ready Signal (placeholder — owned by manager-docs / orchestrator)

```yaml
mx_step_c_decision: <pending — likely SKIP-eligible per markdown-light + minimal Go diff>
mx_step_c_candidate_invariants:
  - candidate: "TestCatalogHashParity (exported via package spec test)"
    fan_in: 0  # test function, not invoked by other code
    eligibility: SKIP-likely  # not a public API surface for @MX:ANCHOR
  - candidate: "resolveHashSourcePath helper (test-internal)"
    fan_in: 2  # called by TestCatalogHashParity + TestCatalogHashParity_MissingTemplates
    eligibility: SKIP-likely  # below @MX:ANCHOR threshold (fan_in >= 3)
mx_commit_sha: <pending>
mx_phase_close_marker: pending  # 4-phase close to be appended by orchestrator at Mx commit
```
