# Progress — SPEC-CLI-UIKIT-KERNEL-001

> Canonical §E lifecycle skeleton. Plan-phase emits §E.1 populated; §E.2/§E.3 are
> populated by manager-develop (run-phase) and §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check: PASS (4 domain segments, last digit-only).
- Frontmatter: 12 canonical fields; `status: in-progress` (draft→in-progress on M1 commit); `tier: L`; `era: V3R6`.
- Artifacts emitted: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md (Tier L 5-file set + progress).
- plan-auditor: iter-2 PASS 0.91 (skip-eligible ≥0.90, 24h, hash-unchanged); Phase 0.5 NOT re-run.
- Mode 5 (sub-agent sequential) selected at Phase 0.95; Implementation Kickoff Approval granted.

## §E.2 Run-phase Evidence

### ANALYZE findings (TRUE blast radius — re-grepped, NOT design.md's figures)

**CheckStatus type references (CRITICAL correction to design.md §D.8):**
- design.md §D.8 claimed "4 refs in doctor.go" — the TRUE recursive grep shows **87 production refs across 4 files**:
  - doctor.go: 77 refs (type def + struct field + fn sigs + case branches)
  - doctor_cache.go: 5 refs (L49/60/70/80/88)
  - doctor_harness.go: 4 refs (L24/96/100/104)
  - render.go: 1 ref (renderStatusLine signature — moves with render.go)
- **13 test files** bearing CheckStatus (design.md §D.8 listed only 5; the other 8 were missed):
  coverage_fixes_test(3), coverage_improvement_test(12), coverage_test(6), doctor_cache_test(4),
  doctor_constitution_test(10), doctor_golden_test(9), doctor_harness_test(12), doctor_new_test(25),
  doctor_test(22), mcp_doctor_coverage_test(4), remaining_coverage_test(2), render_test(1→moves),
  target_coverage_test(8).

**Additional cross-file dependencies design.md did NOT enumerate (resolved this run):**
- **6 style vars** (`cliSuccess/cliWarn/cliError/cliMuted/cliPrimary/cliBorder`, defined update.go:50-56):
  render.go depends on ALL 6. Resolution: moved to uikit as `SuccessStyle/WarnStyle/ErrorStyle/MutedStyle/PrimaryStyle/BorderStyle` (leaf: lipgloss+tui only). Cli consumers (doctor.go, init.go, init_layout.go, update.go) rewrite to `uikit.*Style`.
- **statusIcon** (doctor.go:568): called by render.go:61 renderStatusLine. Resolution: moved to uikit as `StatusIcon` (generic CheckStatus→icon; 0 cli-side rewrite cost — only render.go called it).
- **resolveTheme** (banner.go:41): called by 13 cli files (doctor/help/init_layout/loop×5/status/update×6/update_archive×2/version). Resolution: moved to uikit as `ResolveTheme`; cli retains a LOCAL `resolveTheme()` wrapper (cli/theme.go) so all 13 callers are UNCHANGED (0 additional blast radius).
- **version helpers** (`goVersion/goosArch/gitVersionOverride/claudeVersion/ghVersionOverride`, banner.go): called by doctor.go. Resolution: moved to uikit as exported `GoVersion/GoosArch/GitVersionOverride/ClaudeVersion/GhVersionOverride`; doctor.go rewrites callers.
- **symSuccess/symError/symWarning** (update.go:57-59): called by update.go internally. Resolution: moved to uikit as `SymSuccess/SymError/SymWarning`; update.go rewrites callers.
- **checkStatusToTUI** (doctor.go:152): STAYS in cli (doctor-specific TUI conversion); rewrites CheckStatus→uikit.CheckStatus.

**schema_bridge.go b-ii split resolution:**
- `SchemaKeyToTUIField` + `FieldDefTUILabel` + `TuiLabel` MOVED to uikit (via callback registration pattern: `RegisterSchemaBridge`).
- BOTH `schemaFieldBridge` (L24-58, 19 entries) + `schemaSegmentBridge` (L60-77, 16 entries) STAY in package cli (retyped to `uikit.TuiLabel`).
- cli init() registers the profileSetupText-aware resolver closure; uikit dispatches through the callback.
- uikit has ZERO code references to profileSetupText (5 comment-only mentions in uikit/schema_bridge.go documenting the b-ii split — these are documentation, not coupling).

### PRESERVE: behavior snapshot

- `moai --help` captured pre-M1 → byte-identical post-M1 (AC-CUK-003 PASS).
- Pre-existing baseline failures documented: TestRunHookEvent_ReadInputError (cli), TestAuthoringDocHasEffortMatrix (agentlint), TestCloseSubjectDoctrineAmendment (spec), TestBuild_WritesContextUsageWithSessionID (statusline) — ALL unrelated to this SPEC.

### IMPROVE: atomic M1 transformation

**uikit package created (6 source + 4 test files):**
- `uikit/types.go` — CheckStatus type + CheckOK/CheckWarn/CheckFail consts (from doctor.go:25-34)
- `uikit/styles.go` — 6 exported style vars + SymSuccess/SymError/SymWarning (from update.go:50-59)
- `uikit/status.go` — StatusIcon (from doctor.go:568-578)
- `uikit/render.go` — RenderCard/RenderKeyValue/RenderKeyValueLines/RenderStatusLine/RenderSuccessCard/RenderInfoCard/RenderSummaryLine/RenderError/CardStyle/KVPair (from render.go, exported)
- `uikit/banner.go` — PrintBanner/PrintWelcomeMessage/ResolveTheme/GoVersion/ClaudeVersion/GitVersionOverride/GhVersionOverride/GoosArch (from banner.go, exported)
- `uikit/schema_bridge.go` — TuiLabel/SchemaKeyToTUIField/FieldDefTUILabel/RegisterSchemaBridge (callback pattern)
- `uikit/render_test.go` — moved from render_test.go (package uikit_test)
- `uikit/banner_test.go` — moved from banner_test.go (package uikit_test, with captureStdout + golden helpers)
- `uikit/welcome_test.go` — PrintWelcomeMessage tests moved from misc_coverage_test.go:673-717
- `uikit/characterization_test.go` — NEW characterization tests (resolver dispatch, version helpers, StatusIcon branches, Sym* helpers)
- `uikit/testdata/*.golden` — golden snapshots copied from cli/testdata/

**cli package changes:**
- DELETED: render.go, banner.go, render_test.go, banner_test.go (moved to uikit)
- NEW: theme.go (local resolveTheme wrapper, keeps 13 callers unchanged)
- REWRITTEN: schema_bridge.go (maps retyped to uikit.TuiLabel, helpers removed, init() resolver registration added)
- REWRITTEN: 14 production files (doctor/doctor_cache/doctor_harness/update/migrate_agency/hook/inventory/research/init/init_layout/root/glm + misc_coverage) — all moved-helper call sites → uikit.*
- REWRITTEN: 13 test files (CheckStatus refs → uikit.CheckStatus etc.)

### AC PASS/FAIL matrix

| AC | Status | Verification | Result |
|----|--------|-------------|--------|
| AC-CUK-001 | PASS | `go test ./...` | Same 4 pre-existing failures, zero NEW failures |
| AC-CUK-002 | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | Both exit 0 |
| AC-CUK-003 | PASS | `diff /tmp/uikit-help-before.txt /tmp/uikit-help-after.txt` | IDENTICAL |
| AC-CUK-004 | PASS | `grep 'cli.Execute' cmd/moai/main.go` + `go build ./cmd/moai` | Present, exit 0 |
| AC-CUK-005 | PASS | grep uikit/*.go (non-test) for `internal/cli` parent/sibling | No match |
| AC-CUK-006 | PASS | `go build ./...` exit 0 + `go list -deps` | No cycle |
| AC-CUK-007 | PASS | type CheckStatus in uikit only; profileSetupText maps stay in cli; settings.go STAYS | See §E.2 above |
| AC-CUK-008 | PASS | grep for residual unqualified moved symbols in cli/*.go | No residual |
| AC-CUK-009 | PASS | grep AskUserQuestion in uikit | 0 matches |
| AC-CUK-010 | PASS | `go doc ./internal/cli/uikit` | All needed helpers exported |
| AC-CUK-011 | PASS | Single atomic commit | M1 commit |
| AC-CUK-012 | PASS | No test deleted/skipped; test count preserved | Moved + characterization only |
| AC-CUK-013 | DEFERRED | Checkpoint decision | Post-M1 (this SPEC ships M1; follow-up cluster SPECs are independent) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-07
run_commit_sha: "5326b7632"
run_status: audit-ready
ac_pass_count: 12
ac_fail_count: 0
ac_deferred_count: 1
ac_deferred_list:
  - AC-CUK-013 (checkpoint — deferred to sync-phase / post-M1 decision)
preserve_list_post_run_count: 7  # 7 existing subpackages unchanged (worktree/harness/preference/wizard/specid/pr/agentlint)
l44_pre_commit_fetch: "0 0"  # synced
l44_post_push_fetch: "0 0"  # synced post-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_darwin: PASS
  windows: PASS
total_run_phase_files: 39  # 4 deleted + 2 new + 33 modified
m1_to_mN_commit_strategy: single-atomic-M1
```

### Notes
- **Characterization tests added** (uikit/characterization_test.go): 4 test functions covering resolver dispatch (RegisterSchemaBridge + SchemaKeyToTUIField + FieldDefTUILabel), StatusIcon all branches, version helpers (GitVersionOverride/GhVersionOverride/GoosArch), Sym* helpers. These are DDD PRESERVE-phase characterization tests capturing EXISTING behavior — permitted under design.md §E AP-5 ("Characterization only").
- **Coverage**: uikit 98.9% (≥85% target).
- **b-ii split characterization gap**: design.md anticipated the helpers move cleanly; ANALYZE found they're inseparable from the profileSetupText-coupled maps without a callback-registration mechanism. The callback pattern (RegisterSchemaBridge) is the faithful b-ii realization — standard Go DI (precedent: worktree.WorktreeProvider).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-08
sync_commit_sha: "8303eaf6d"  # backfilled in separate chore commit per feedback_era_commit_sha_field_format; run_commit_sha 5326b7632 already on origin/main (code landed prior session)
sync_status: audit-ready
close_decision: "3-phase close (plan→run→sync) — M1 uikit kernel extraction complete; AC-CUK-013 checkpoint deferred to post-M1 per §F (future migrate/doctor/update clusters gate on this uikit leaf and are separate SPECs)"
run_commit_sha: "5326b7632"  # ancestor of origin/main, verified pre-sync
l44_pre_sync_fetch: "0 10 (local ahead by 10, clean-ahead — NOT a race; origin/main not ahead)"
ac_close_subset: "12/13 AC PASS (AC-CUK-001..012); 1 deferred (AC-CUK-013 checkpoint — post-M1 follow-up cluster SPECs independent)"
changelog_entry_position: "[Unreleased] > Added"
frontmatter_status_transitions:
  spec_md: in-progress → completed       # single sync commit (manager-docs owned; merged 3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-008/009)
  plan_md: n/a                            # plan.md has no YAML frontmatter status field (pure markdown)
  acceptance_md: n/a                      # acceptance.md has no YAML frontmatter status field (pure markdown)
  progress_md: n/a                        # progress.md uses §E.* sections for lifecycle signaling, not a frontmatter status field
canary_compliance_check:
  spec_lint: pending                      # moai spec lint not run in this sync-phase; SPEC frontmatter 12-field schema manually verified (status + updated + 10 other canonical fields present)
  era_classification: V3R6                # H-4: §E.2 + §E.4 + sync_commit_sha present (once backfilled); created 2026-07-07 ≥ 2026-04-01 modernEraThreshold; explicit era: V3R6 in frontmatter
sync_phase_artifacts:
  changelog_updated: true                 # CHANGELOG.md [Unreleased] > Added — SPEC-CLI-UIKIT-KERNEL-001 entry
  readme_updated: false                   # internal refactor (behavior-preserving, zero user-visible change); README/README.ko/docs-site do not document internal package structure per Tier L internal-refactor convention (cf. SPEC-CLI-SUBPKG-SPLIT-001 also readme_updated: false)
  frontmatter_transitioned: true          # spec.md frontmatter status → completed (only spec.md carries the YAML status field)
  progress_e4_backfilled: true            # §E.4 Sync-phase Audit-Ready Signal populated (this block)
mx_tag_validation: n/a                    # internal refactor — no new @MX tags added; sync is cross-cutting concern, validation-only
b12_self_test:
  a_pre_emission_grep: pass               # grep -c 'SPEC-CLI-UIKIT-KERNEL-001' CHANGELOG.md = 0 pre-emission (no duplicate, HALT avoided)
  b_ac_count_match: pass                  # acceptance.md SSOT AC count = 13 (AC-CUK-001..013); CHANGELOG entry references "12/13 AC PASS + 1 deferred (AC-CUK-013)"
  c_file_path_verification: pass          # all paths claimed in CHANGELOG entry (.moai/specs/SPEC-CLI-UIKIT-KERNEL-001/, internal/cli/uikit/*.go 6 source + 4 test files) verified via ls
```

### Sync-phase Self-Verification (manager-docs variant)

- **E1 Frontmatter status transitions**: spec.md `status: in-progress → completed` + `updated: 2026-07-07 → 2026-07-08` applied on this single sync commit per Status Transition Ownership Matrix (spec-frontmatter-schema.md § Status Transition Ownership Matrix). plan.md / acceptance.md / progress.md have NO YAML frontmatter status field (pure markdown bodies + §E.* section-based lifecycle signaling), so no frontmatter edit applies to them; the transition is atomic on spec.md (the SSOT carrier of the `status:` field). Body content untouched.
- **E2 SHA backfill**: `sync_commit_sha` field will be backfilled in a separate chore commit per `feedback_era_commit_sha_field_format` (sync_commit_sha is a literal YAML field; backfill chore commit pattern is the established SPEC discipline). Run-phase SHA `5326b7632` already in §E.3 and verified ancestor of origin/main pre-sync.
- **E3 CHANGELOG entry**: `[Unreleased] > Added` section — SPEC-CLI-UIKIT-KERNEL-001 entry appended (B12 pre-emission grep count 0 confirmed; 13 AC count matches acceptance.md SSOT AC-CUK-001..013 with 12 PASS + 1 deferred AC-CUK-013; file paths verified via ls).
- **E4 Scope-boundary grep**: only spec.md (frontmatter status only — body untouched) + progress.md §E.4 + CHANGELOG.md in this commit's stat (no implementation source files in scope per Forbidden modifications).
- **E5 Pathspec discipline**: `git add` invoked with specific paths only — no `git add -A`.
- **E6 Commit subject**: `docs(SPEC-CLI-UIKIT-KERNEL-001): sync-phase artifacts — Tier L 3-phase close` follows the canonical sync-commit subject convention per Status Transition Ownership Matrix.

### 3-Phase Close Record

- **Plan → Run → Sync**: complete on 2026-07-08 (Tier L 3-phase close).
- **Plan-phase**: manager-spec, 5-file Tier L artifact set (spec/plan/acceptance/design/research/progress). plan-auditor iter-2 PASS 0.91 (skip-eligible ≥0.90, 24h, hash-unchanged); Phase 0.5 NOT re-run.
- **Run-phase**: manager-develop ddd (single-atomic M1 commit `5326b7632` — refactor extracting TUI kernel helpers render.go/banner.go/schema_bridge.go via RegisterSchemaBridge callback into new leaf package `internal/cli/uikit`). All 12 binding AC PASS (AC-CUK-001..012); 1 AC deferred (AC-CUK-013 checkpoint — future migrate/doctor/update cluster SPECs are independent). 7 existing subpackages preserved unchanged (worktree/harness/preference/wizard/specid/pr/agentlint). Run commit already on origin/main from prior session.
- **Sync-phase** (this commit): manager-docs, single sync commit carrying spec.md frontmatter `in-progress → completed` transition + §E.4 audit-ready signal + CHANGELOG entry. MX Tag validation: n/a (internal refactor, no new tags needed). README unchanged (Tier L internal refactor, zero user-visible change).
- **Epic context**: First follow-up to SPEC-CLI-SUBPKG-SPLIT-001 (M1-only close at sync_commit_sha `d0d9b49d7` — agentlint cluster shipped, §F CHECKPOINT STOP). This SPEC extracts the uikit kernel (SPLIT-001 §F M5) that unblocks future kernel-dependent clusters (migrate/doctor/update) from their axis-(i) blocker. Per progress.md §E.2 ANALYZE, this is the foundational unblock for the cluster ladder SPLIT-001 envisioned.

## §F Phase 0.95 Mode Selection

- **Decision**: Mode 5 (sub-agent sequential)
- **Justification**: Tier L coding-heavy refactor with tightly-coupled caller rewrite. Single sequential manager-develop ddd cycle.
- **Pre-Spawn Sync Check**: `0 0` (synced); no active sessions. HEAD `fdc411807`.

## §G IGGDA Kickoff Predicate

- Verdict: **explicit-gate** (condition (c) Tier L FAIL → mandatory blocking AskUserQuestion).
- Resolution: Implementation Kickoff Approval **granted by user** (Option A — run-phase 진입 승인) at 2026-07-07.
- source_session_id: 9b68bae7-b26f-4d1c-bad0-fb800ff626f6
