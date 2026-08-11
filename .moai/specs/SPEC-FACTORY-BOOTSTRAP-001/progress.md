# progress.md — SPEC-FACTORY-BOOTSTRAP-001

> Phase evidence and audit-ready signals. §E.1 is populated at plan-phase; §E.2–§E.4 are placeholder headings to be populated by manager-develop (run-phase) and manager-docs (sync-phase) per the Forbidden-modifications matrix.

---

## §A. Plan-Phase Artifact Set

| Artifact | Path | Status |
|---|---|---|
| spec.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/spec.md` | emitted |
| plan.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/plan.md` | emitted |
| acceptance.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/acceptance.md` | emitted |
| design.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/design.md` | emitted (Tier L) |
| research.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/research.md` | emitted (Tier L) |
| progress.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/progress.md` | emitted (this file) |

---

## §B. Branch State (at plan-phase emission)

- **Branch**: `feat/factory-bootstrap-guidance`
- **HEAD**: `94025ce0a` (prior-art commit "feat(factory): announce companion session bootstrap from the SessionStart hook")
- **Base**: `chore/revert-kanban-rename` (`24c4674b5`) ← `origin/main`
- **Local ahead by**: 2 (no race)
- **Tree status at emission**: clean working tree, plan-phase artifacts are the only new files

---

## §C. Requirements Budget

- **REQ count**: 18 (REQ-FB-001..018) against the Tier L ceiling of 25.
- **AC count**: 27 against 18 REQ (9-criterion surplus reflects the multi-surface span — env / source / notice / help / docs-site / template-neutrality / sibling-boundary / worktree-isolation / breaking-change-pin). AC-FB-016a is a fail-open sub-clause paired with AC-FB-016, not a separate criterion.
- **MUST criteria**: 25; **SHOULD**: 1 (AC-FB-020); **meta**: 1 (AC-FB-026 worktree isolation).

---

## §D. Pre-Plan-Audit Self-Check

- [x] SPEC ID regex: `SPEC-FACTORY-BOOTSTRAP-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (Bash output: `PASS`).
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + tier L.
- [x] ID uniqueness: no prior `SPEC-FACTORY-BOOTSTRAP-001` directory exists at `.moai/specs/`.
- [x] Requirements in GEARS notation (Where/When/While + shall).
- [x] Out of Scope: §C carries ≥ 1 `### Out of Scope — <topic>` H3 heading with `-` bullets (satisfies `OutOfScopeRule`).
- [x] Artifact set matches Tier L (spec + plan + acceptance + design + research + progress).
- [x] spec.md carries no implementation detail (no function signatures, no API schemas — only observable behaviors, env-var names, file paths as evidence anchors).
- [x] Prior-art commit `94025ce0a` recorded in HISTORY and §A.1 as revised-not-reverted.
- [x] Sibling boundary one-sided (no edits to `.moai/specs/SPEC-KANBAN-*`).
- [x] AC003 preserve-tests named in plan.md §A.5 with package path.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_version: 0.2.0
spec_id: SPEC-FACTORY-BOOTSTRAP-001
tier: L
target_auditor_threshold: 0.85
requirements_count: 18
requirements_ceiling: 25
acceptance_criteria_count: 27
must_criteria: 25
should_criteria: 1
meta_criteria: 1
prior_art_commit: 94025ce0a
prior_art_relation: revised-not-reverted
predecessor_spec: SPEC-FACTORY-MODE-001
predecessor_status: completed
sibling_spec: SPEC-KANBAN-BOOTSTRAP-001
sibling_boundary: one-sided-from-this-side
out_of_scope_sections:
  - "Out of Scope — Topology-config-gated quorum and dispatch"
  - "Out of Scope — Forward reference to the sibling's supersedence"
  - "Out of Scope — Run-phase decisions"
  - "Out of Scope — Single-session factory mode"
preserve_tests:
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal"
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_BlockCapDoctrineClauseSpecific"
baseline_attribution:
  worktree: "/Users/goos/.moai/worktrees/kanban"
  branch: "feat/factory-bootstrap-guidance"
  head: "94025ce0a"
  base: "24c4674b5"
  tree_status: "clean"
audit_revision:
  iteration: 2
  prior_verdict: FAIL
  prior_score: 0.81
  findings_closed: [D1, D2, D3, D4, D5]
  breaking_change_94025ce0a: "companion-shape --name alone reclassified from companion entry to no-op (spec.md §A.2.1, REQ-FB-001 no--f clause, AC-FB-027)"
frontmatter_repo_conventions:
  - "related_specs: [SPEC-FACTORY-MODE-001, SPEC-KANBAN-BOOTSTRAP-001] is a repo convention NOT codified in the canonical 12-field or optional schema; spec-lint passes (moai spec lint → 0 findings) and the sibling SPEC-KANBAN-BOOTSTRAP-001 carries the same field. Schema codification is a separate SPEC."
open_clarifications: 0
blocker_report: none
```

---

## §E.2 Run-phase Evidence

| AC | Status | Milestone | Verification command + result |
|---|---|---|---|
| AC-FB-001 | PASS | M2 | `go test ./internal/cli/ -run TestDispatchOutcome_LeadEnvState` → ok |
| AC-FB-002 | PASS | M2 | `go test ./internal/cli/ -run TestDispatchOutcome_CompanionEnvState` → ok |
| AC-FB-003 | PASS | M2 | `go test ./internal/cli/ -run TestResolveFactoryBranchTruthTable/row3` → ok |
| AC-FB-004 | PASS | M2 | `go test ./internal/cli/ -run TestResolveFactoryBranchTruthTable/row4` → ok |
| AC-FB-005 | PASS | M2 | `grep -nE 'else if .*parseCompanionLabel' cc.go glm.go` → 0 matches; both functions called unconditionally |
| AC-FB-006 | PASS | M2 | `TestDispatchOutcome_CompanionEnvState` asserts MOAI_FACTORY unset |
| AC-FB-007 | PASS | M2 | `TestDispatchOutcome_LeadEnvState` asserts MOAI_FACTORY set |
| AC-FB-008 | PASS | M2 | `go test ./internal/cli/ -run AC003` → both tests PASS |
| AC-FB-009 | PASS | pre-existing | `TestFactoryCompanionRaisesBlockCap` + `TestACFM022a` PASS (block-cap OR-branch untouched) |
| AC-FB-010 | PASS | M1 | `TestPrepareFactorySettingsWritesTransientFile` → ok |
| AC-FB-011 | PASS | M1 | `TestPrepareFactorySettingsHonorsOperatorSupplied` → ok |
| AC-FB-012 | PASS | M1/M3 | `TestFactoryLeadNoticeOperatorSettingsAdvisory` → ok |
| AC-FB-013 | PASS | M3 | `TestFactoryLeadNoticeFullContent` → ok (run id, 4 lines with -f, socket path, auto-accept, SPEC id) |
| AC-FB-014 | PASS | M3 | `TestFactoryLeadNoticeOmitsSPECWhenUnset` → ok |
| AC-FB-015 | PASS | M3 | `TestFactoryLeadNoticeCompanionLinesCarryF` → ok |
| AC-FB-016 | PASS | M3 | `TestFactoryCompanionNoticeRoleless` → ok |
| AC-FB-016a | PASS | M3 | `TestFactoryCompanionNoticeFailOpen` → ok (empty/malformed/non-companion → empty string) |
| AC-FB-017 | PASS | M3 | `TestFactoryCompanionNoticeJoinOnly` → ok |
| AC-FB-018 | PASS | M4 | `TestACFB018_HelpDocumentsLeadEntry` → ok (cc + glm) |
| AC-FB-019 | PASS | M4 | `TestACFB019_HelpDocumentsCompanionEntry` → ok (cc + glm) |
| AC-FB-020 | PASS | M4 | `grep -nE 'cmd\.Flags\(\).*factory' cc.go glm.go` → 0 matches |
| AC-FB-021 | PASS | M5 | `ls docs-site/content/{en,ko,ja,zh}/multi-llm/factory-mode.md` → all 4 exist with title/weight/draft:false |
| AC-FB-022 | PASS | M5 | `grep factory-mode docs-site/data/menu/main.yaml` → in multi-llm sub: with 4-locale name map + ref |
| AC-FB-023 | PASS | M5 | no _meta.yaml / menu.html change (sub:-level entry only) |
| AC-FB-024 | PASS | M6 | `grep -rnE 'SPEC-FACTORY-BOOTSTRAP-001\|REQ-FB-\|94025ce0a' internal/template/templates/` → 0 matches |
| AC-FB-025 | PASS | M6 | `git diff --name-only 24c4674b5..HEAD -- .moai/specs/SPEC-KANBAN-` → 0 matches |
| AC-FB-026 | PASS | M6 | 6 artifacts present in worktree; primary checkout has no such dir |
| AC-FB-027 | PASS | M2 | `TestResolveFactoryBranchTruthTable/row5` → ok (BREAKING: companion-shape --name alone → no-op) |

**Summary**: 27/27 AC PASS (25 MUST + 1 SHOULD + 1 meta). 0 FAIL.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-11
run_commit_sha: a87e6174b
ac_pass_count: 27
ac_fail_count: 0
preserve_list_post_run_count: 2
preserve_list_post_run:
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal"
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_BlockCapDoctrineClauseSpecific"
l44_pre_commit_fetch: n/a
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: PASS
  windows_amd64: PASS
total_run_phase_files: 12
m1_to_mN_commit_strategy: per-milestone Conventional Commits (M1-M5 committed; M6 verification-only)
run_phase_commits:
  - "d3f8ffcda: M1 crossSessionInbound injection via transient --settings"
  - "1b02bcf80: M2 -f redefinition + flat dispatch selection"
  - "884e8e2dc: M3 SessionStart notice revision"
  - "c346682a2: M4 CLI help text for -f lead and companion"
  - "a87e6174b: M5 Factory Mode docs-site 4-locale page"
  - "M6: verification only (make build exit 0; AC-FB-024/025/026 grep PASS)"
coverage_note: "internal/cli 77.0% (package-wide); factory/bootstrap/settings-dispatch files fully covered by new + prior-art tests"
baseline_attribution:
  worktree: "/Users/goos/.moai/worktrees/kanban"
  branch: "feat/factory-bootstrap-guidance"
  head: "a87e6174b"
  base: "24c4674b5"
  pre_existing_failures: "internal/statusline TestBuilderNormalizesMode (rolling-window flake); internal/hook TestBranchGuard_Latency (timing-sensitive under parallel load) — both pass in isolation, unrelated to Factory Mode"
push_state: "NOT pushed (C5: commits only, no push, no PR)"
blocker_report: none
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-11
sync_commit_sha: pending-backfill-sync
run_commit_sha: 89227add0
changelog_entry_position: top of [Unreleased] → Added (English-only CHANGELOG.md)
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (single sync commit)"
  updated_field_refreshed: true
b12_self_test_a:
  pre_emission_grep: "grep -c 'SPEC-FACTORY-BOOTSTRAP-001' CHANGELOG.md → 0 (pre-edit)"
  outcome: pass
b12_self_test_b:
  ac_count_match: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 27"
  changelog_references_ac: true
  outcome: pass
b12_self_test_c:
  file_path_verification: "ls internal/cli/factory.go internal/cli/factory_settings.go internal/cli/cc.go internal/cli/glm.go internal/hook/session_start_factory.go docs-site/content/{en,ko,ja,zh}/multi-llm/factory-mode.md → all 9 paths exist"
  outcome: pass
docs_site_4locale_parity:
  files_exist: true
  frontmatter_complete: true  # title + weight:30 + draft:false in all 4 locales
  menu_entry_present: true    # data/menu/main.yaml multi-llm sub: factory-mode, 4-locale name map + ref
mx_tag_status:
  present_count: 4
  locations:
    - "internal/cli/factory.go:8 @MX:NOTE [AUTO] (process-env signal)"
    - "internal/cli/factory.go:76-77 @MX:ANCHOR [AUTO] + @MX:REASON (deferred restore correctness)"
    - "internal/cli/factory_settings.go:14 @MX:NOTE [AUTO] (session-private transient file)"
  gaps: "resolveFactoryBranch / prepareFactorySettings / operatorSuppliedSettings / revised notice functions carry no @MX tag — all below fan_in>=3 ANCHOR threshold (2,2,1,1 callers respectively); thorough doc comments already carry the context. No autonomous tag addition warranted (would be mechanical inflation)."
canary_compliance_check:
  sibling_off_limits: "git diff --name-only 24c4674b5..HEAD -- .moai/specs/SPEC-KANBAN- → 0 matches (C3 held)"
  template_neutrality: "grep -rnE 'SPEC-FACTORY-BOOTSTRAP-001|REQ-FB-|94025ce0a' internal/template/templates/ → 0 matches (C2 held)"
  worktree_isolation: "all artifacts under .moai/specs/SPEC-FACTORY-BOOTSTRAP-001/ in this worktree (C6 held)"
  ac003_preserve: "TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal + TestAC003_BlockCapDoctrineClauseSpecific green (REQ-FB-018 held)"
baseline_attribution:
  worktree: "/Users/goos/.moai/worktrees/kanban"
  branch: "feat/factory-bootstrap-guidance"
  head_pre_sync: "24b401a3a"
  run_head: "89227add0"
  base: "24c4674b5"
push_state: "NOT pushed (C5: commits only, no push, no PR — per user instruction)"
blocker_report: none
```

### Sync-phase Verification Evidence

- **CHANGELOG entry**: `grep -c 'SPEC-FACTORY-BOOTSTRAP-001' CHANGELOG.md` → `1` (post-edit, the new entry; pre-edit count was 0).
- **docs-site 4-locale parity**: `ls docs-site/content/{en,ko,ja,zh}/multi-llm/factory-mode.md` → all 4 files exist; each carries `title:` + `weight: 30` + `draft: false`.
- **Menu entry**: `grep -nE 'factory-mode' docs-site/data/menu/main.yaml` → `430:        ref: /multi-llm/factory-mode` inside the `multi-llm` section's `sub:` list, with a 4-locale `name:` map (ko/en/ja/zh).
- **Frontmatter transition**: `spec.md` status `in-progress` → `completed` in this sync commit; `updated:` refreshed.
- **Push NOT performed**: `git rev-parse --abbrev-ref @{u}` → fatal (no upstream) — no push, no PR (C5 held).

---

## §F Phase 4 Mode Selection

### Input parameters

- **tier**: L (18 REQ / 27 AC, >15 files affected)
- **scope (file count)**: ~14-15 — Go source (`internal/cli/factory.go`, `cc.go`, `glm.go`, `internal/hook/session_start_factory.go`, new settings-injection helper; `launcher_blockcap_infinite.go` read-only) + Go tests (`factory_bootstrap_test.go`, `session_start_factory_test.go`, `launcher_blockcap_infinite_test.go` extended, +companion-combo test) + docs-site 4-locale `factory-mode.md` + `data/menu/main.yaml`
- **domain count**: 3 (`internal/cli` dispatch + env, `internal/hook` SessionStart notice, docs-site 4-locale publishing)
- **file language mix**: Go source (~6) + Go tests (~4) + markdown docs-site (4) + YAML (1) — mixed, Go-majority
- **concurrency benefit**: LOW — the milestones repeatedly edit the SAME Go files (`factory.go`, `cc.go`, `glm.go` are touched by both M2 and M4; `session_start_factory.go` by M3), so parallel fan-out would race on shared write targets
- **Agent Teams prereqs**: n/a (Mode 3 retired)

### Mode evaluation table

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | Tier L, 6 milestones, multi-file semantic change |
| 2 background | no | not a single async read-only task; multi-milestone write work |
| 3 agent-team | no | RETIRED (tombstone) |
| 4 parallel | no | coding-heavy + shared-file milestones → Anthropic coding-task parallelism caveat; concurrent edits to `cc.go`/`glm.go`/`factory.go` would race |
| 5 sub-agent | **yes** | sequential manager-develop delegation; plan.md provides per-milestone implementation guidance; same-Go-file recurrence demands serialization |
| 6 workflow | no | not a ≥30-file single-uniform-transform; milestones carry distinct semantic changes (dispatch rewrite, notice revision, settings injection, help text, docs pages) |

### Decision

`sub-agent`

### Justification

Mode 5 (sequential sub-agent) is selected per Anthropic's coding-task parallelism caveat: most coding tasks involve fewer truly parallelizable tasks than research, and the six milestones here repeatedly modify the same Go source files (`internal/cli/factory.go`, `cc.go`, `glm.go` by both M2 and M4; `internal/hook/session_start_factory.go` by M3). Parallel fan-out (Mode 4) would race on those shared write targets. plan.md §B already orders the milestones by decision-reversibility (M0 decision → M1 new surface → M2 core semantic → M3 notice → M4 help → M5 docs → M6 verify) and provides concrete implementation guidance per milestone, so a single sequential manager-develop delegation with the Tier L Section A-E template carries the full scope. Implementation Kickoff Approval was obtained in this session; plan-audit review-2 verdict is PASS 1.00 (Tier L threshold 0.85), D1-D5 closed, D6 non-blocking.
