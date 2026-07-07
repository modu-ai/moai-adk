---
id: SPEC-HOOK-CONFIG-SAFETY-001
title: "moai update Configuration Safety /hooks Review Guidance"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec (P5 of hooks-hardening Epic)
priority: P3
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "hooks, cli, configuration-safety, ux, discoverability"
---

# progress.md — SPEC-HOOK-CONFIG-SAFETY-001

> Plan-phase artifact. §E.1 populated by manager-spec. §E.2 / §E.3 / §E.4 are placeholder headings emitted per the system-prompt HARD rule — era.go greps for the literal `§E.2`/`§E.3`/`§E.4` substrings for era classification, so omitting them risks H-2 era misclassification. The placeholders are NOT populated at plan-phase; §E.2/§E.3 belong to manager-develop (run-phase) and §E.4 belongs to manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- **plan_status**: `audit-ready`
- **plan_complete_at**: 2026-07-07
- **plan_author**: manager-spec (P5 of moai-adk hooks-hardening Epic)
- **tier**: S — single domain (CLI completion UX), 2-3 files, no architectural change, no new dependencies
- **design_decision**: Option B (unconditional on "Template sync complete" branch) — see plan.md §A.2
- **frontmatter_schema_check**: PASS — 12 canonical fields present; no snake_case aliases (`created_at` / `updated_at` / `labels` / `spec_id` all avoided)
- **spec_id_regex_check**: PASS — `SPEC-HOOK-CONFIG-SAFETY-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`
  - decomposition: `SPEC ✓ | HOOK ✓ | CONFIG ✓ | SAFETY ✓ | 001 ✓ → PASS`
- **outofscope_rule_check**: PASS — 5 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets present in spec.md §D
- **gears_notation_check**: PASS — 5 requirements in spec.md §B use canonical GEARS keywords (`When`/`Where` + `shall`/`shall not`); no deprecated `IF/THEN` modality
- **verify_before_author_complete**: YES — Greps 1-3 in spec.md §A.2 re-run independently during plan-phase; 0-hit result confirmed (not memory-derived)
- **artifact_set**: spec.md + plan.md + acceptance.md + progress.md (4 artifacts; Tier S scope per frontmatter `tier: S` — acceptance.md authored separately for AC testability detail; plan-auditor D1 SHOULD-FIX noted the count exceeds the Tier S 2-file lean set, accepted as quality-positive)
- **insertion_point_anchor**: content token `Label: "Template sync complete"` in `internal/cli/update.go` (grep-date line ~874; run-phase must re-locate by content token per `feedback_line_number_drift_asymmetry`)
- **no_op_branches_enumerated**: 3 — `Label: "Already up to date"`, `Label: "Template version up-to-date · Skipping sync"`, `Label: "Binary updated (template sync skipped)"` (all in `internal/cli/update.go`)
- **epic_context**: P5 of moai-adk hooks-hardening Epic. P1 = SPEC-HOOK-SESSIONSTART-PROBE-001 (completed, sync commit 10cd484a0). P2/P3/P4 absorbed (3x stale-memory non-defects per `feedback_hypothesis_as_defect`). P5 = this SPEC (real gap confirmed by mechanical grep).
- **plan_auditor_verdict**: iter-1 PASS-WITH-DEBT 0.904 (D1 SHOULD-FIX: tier 정합성 — frontmatter `tier:` absent + artifact_set mislabel) → orchestrator D1 fix (spec.md frontmatter `tier: S` line 10 + progress.md `artifact_set` 정정) → iter-2 PASS-WITH-DEBT **0.96** (D1 RESOLVED across 5 surfaces, D2 MINOR §B.1 "conveys" testability persists, skip-eligible ≥0.90 — 4/4 compound predicate hold). Report: `.moai/reports/plan-audit/SPEC-HOOK-CONFIG-SAFETY-001-2026-07-07.md`.
- **run_phase_status**: DEFERRED at Implementation Kickoff — pre-spawn check detected multi-session race (origin ahead 1): 병렬 세션 commit `5326b7632` SPEC-CLI-UIKIT-KERNEL-001 M1이 `internal/cli/update.go` 44 lines 변경(uikit kernel 추출, 51 files/1100 insertions), P5 수정 대상("Template sync complete" 분기)과 직접 겹침 + 이 세션 uncommitted(glm.go 등)도 겹침. 사용자 결정: P5 보류 + 병렬 세션 quiesce 대기 (memory Phase 3 지침 준수). 재개 = `/moai run SPEC-HOOK-CONFIG-SAFETY-001` after quiesce.

## §E.2 Run-phase Evidence

**Implementation** (2 files, Tier S):
- `internal/cli/update.go` — added `emitHooksReviewGuidance(w io.Writer)` helper (writes the `/hooks` review guidance line) + wired it as a single call site immediately after the `Label: "Template sync complete"` pill (content-token-anchored, AC-CS-005). `@MX:NOTE` + `@MX:SPEC` tags annotate the helper.
- `internal/cli/update_hooks_guidance_test.go` (NEW) — 5 characterization/static-guard tests + 2 test helpers (`seedSystemYAMLVersion`, `bindUpdateCmd`).

**E1 — AC PASS/FAIL matrix** (5-section format: Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

| AC | Status | Verification command | Observed output |
|----|--------|----------------------|-----------------|
| AC-CS-001 (guidance emitted on Template-sync-complete) | PASS | `go test -run TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance -v ./internal/cli/` | `--- PASS: TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance (0.00s)` — calls `emitHooksReviewGuidance(&buf)` directly; asserts review-or-restart remedy present |
| AC-CS-002a (no guidance on Already-up-to-date / --binary path) | PASS | `go test -run TestUpdate_Characterize_AlreadyUpToDate_NoHooksGuidance -v ./internal/cli/` | `--- PASS` — `runUpdate` with `--binary --yes` + `MOAI_SKIP_BINARY_UPDATE=1`; asserts `/hooks` absent |
| AC-CS-002b (no guidance on Skipping-sync) | PASS | `go test -run TestUpdate_Characterize_SkippingSync_NoHooksGuidance -v ./internal/cli/` | `--- PASS` — `runTemplateSyncWithProgress` with version-match `system.yaml`; asserts `/hooks` absent; verifies "Skipping sync" branch fired |
| AC-CS-002c (no guidance on Binary-only) | PASS | `go test -run TestUpdate_Characterize_BinaryOnly_NoHooksGuidance -v ./internal/cli/` | `--- PASS` — `runUpdate` with `--binary --yes` + `MOAI_SKIP_BINARY_UPDATE=1`; asserts `/hooks` absent |
| AC-CS-003 (literal `/hooks` token) | PASS | (covered by AC-CS-001 test) | `strings.Contains(got, "/hooks")` — case-sensitive literal substring assertion (NOT regex, NOT case-insensitive) |
| AC-CS-004 (no AskUserQuestion / mcp__askuser) | PASS | `go test -run TestUpdate_NoAskUserQuestion -v ./internal/cli/` | `--- PASS` — reads `update.go`, asserts 0 occurrences of `AskUserQuestion` and `mcp__askuser` (mirrors `TestNew_NoAskUserQuestion`) |
| AC-CS-005 (content-token anchored, SHOULD) | PASS | `grep -n 'Label: "Template sync complete"' internal/cli/update.go` | `875: ...Label: "Template sync complete"...` followed immediately by `emitHooksReviewGuidance(out)` — insertion located by content token, NOT line number |
| AC-CS-006 (full suite 0 failures) | PASS-WITH-DEBT | `go test -count=1 ./...` | 3 pre-existing failures, ALL verified pre-existing via `git stash --include-untracked` + re-run (see Gaps below). This SPEC introduced 0 new failures. |
| AC-CS-007 (lint clean internal/cli) | PASS | `golangci-lint run ./internal/cli/...` + `go vet ./internal/cli/...` | `0 issues.` / `VET_EXIT=0` |

**Baseline-attribution**: all commands run against worktree HEAD `d9798065459094815ce2860c1b56a70ba42b1e7f` (fast-forwarded from `5326b7632` after detecting origin/main advanced 1 commit — the parallel commit `feat(spec-lint)` touched `internal/spec/` + template README, zero scope overlap with this SPEC). The 5 new characterization tests were observed PASS in this run, this tree.

**E2 — Cross-platform build**: `go build ./...` → `EXIT=0` (darwin/arm64); `GOOS=windows GOARCH=amd64 go build ./...` → `EXIT=0`.

**E3 — Coverage**: focused run `go test -coverprofile=/tmp/cs_cover.out -run 'TestUpdate_Characterize|TestUpdate_NoAskUserQuestion' ./internal/cli/` → `emitHooksReviewGuidance 100.0%` (the new code is fully covered). Package-wide `internal/cli` coverage number is NOT separately measurable here because a PRE-EXISTING test (`TestRunHookEvent_ReadInputError`) panics with a nil-pointer dereference and aborts the full-package coverage report — see Gaps.

**E4 — Subagent boundary grep (AC-CS-004 mechanical)**: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/update.go | grep -v "_test.go" | grep -v "// "` → exit 1 (0 matches). PASS.

**E5 — Lint (NEW vs pre-existing)**: `golangci-lint run --timeout=5m ./internal/cli/...` → `0 issues` (no NEW and no pre-existing findings on internal/cli). `go vet ./internal/cli/...` → `VET_EXIT=0`.

**E6 — Full suite**: `go test -count=1 ./...` → 3 failures, ALL pre-existing and unrelated to this SPEC:
1. `internal/cli` — `TestRunHookEvent_ReadInputError` (nil-pointer panic in hook-event reading code). Verified pre-existing: stashed ALL my changes (`git stash push --include-untracked`) → test STILL fails identically → restored.
2. `internal/spec` — `TestCloseSubjectDoctrineAmendment/lifecycle-sync-gate.md` (AC-DLC-011 close-subject full-ID mandate in a doctrine rule file). Verified pre-existing: fails with `update.go` stashed.
3. `internal/statusline` — `TestBuild_WritesContextUsageWithSessionID` (1M vs 256K context-window-size assertion). Verified pre-existing: fails with `update.go` stashed.

This SPEC's 5 new tests PASS; 0 new failures introduced.

**Gaps (NOT observed)**:
- Package-wide `internal/cli` coverage % not measured (pre-existing `TestRunHookEvent_ReadInputError` panic aborts the coverage report before it is written). Only the focused coverage of the new helper was measured (`emitHooksReviewGuidance 100.0%`).
- AC-CS-002a/c target the production pills "Already up to date" / "Binary updated (template sync skipped)"; in the test environment (dev build + `MOAI_SKIP_BINARY_UPDATE=1`) the `--binary` path emits "Skipped (dev build, --binary)" instead. All three pills share the same return-before-line-875 property, so the `/hooks`-absent invariant holds on the entire `--binary` path — but the exact production pills were not exercised (require a release build + real binary update).
- A full `runUpdate` integration test reaching the line-875 "Template sync complete" pill was evaluated and rejected: full template deployment into a bare `t.TempDir()` is unreliable (precedent: `update_skip_sync_test.go` notes "template deployment may fail"), and `runUpdate`'s v2-fingerprint block diverts a bare tmpDir to `runCleanReinstall`. AC-CS-001 is therefore evidenced via the direct helper test (`emitHooksReviewGuidance`) + the structural fact that the helper is the single emission site (AC-CS-005 content-token-anchored call).

**Residual-risk**:
- The guidance is unconditional on the Template-sync-complete branch (Option B per plan.md §A.2). If template sync ran but the `.claude/settings.json` hooks section did not actually change, the guidance fires anyway — an accepted Tier S trade-off (spec.md §E.2, acceptance.md §D.1). Cost: one advisory terminal line.
- `--json` is NOT a registered flag on `moai update` (flags: check, shell-env, config, force, yes, templates-only, binary, dry-run, no-hooks, verbose), so the guidance cannot corrupt JSON output. Characterized at run-phase (plan.md §B.4 resolution): §D.3 / acceptance.md §D.3 edge case is moot.

## §E.3 Run-phase Audit-Ready Signal

- **run_complete_at**: 2026-07-07
- **run_commit_sha**: `e87bc205c`
- **run_status**: `audit-ready` (all 7 ACs at PASS or PASS-WITH-DEBT; PASS-WITH-DEBT is AC-CS-006 only, debt = 3 pre-existing unrelated failures)
- **ac_pass_count**: 7 (AC-CS-001, 002a, 002b, 002c, 003, 004, 005, 007 — PASS; AC-CS-006 PASS-WITH-DEBT)
- **ac_fail_count**: 0
- **preserve_list_post_run_count**: 0 (every file outside internal/cli/update.go + internal/cli/update_hooks_guidance_test.go + .moai/specs/SPEC-HOOK-CONFIG-SAFETY-001/progress.md PRESERVED; verified via `git status --porcelain` showing only the 3 in-scope paths + the copied SPEC dir)
- **l44_pre_commit_fetch**: YES — pre-flight `git fetch origin main` detected origin advanced to `d9798065` (parallel `feat(spec-lint)` commit, zero scope overlap); worktree fast-forwarded to `d9798065` before committing (clean `0 0` divergence post-ff)
- **l44_post_push_fetch**: `<pending push>`
- **new_warnings_or_lints_introduced**: 0 (`golangci-lint run ./internal/cli/...` → 0 issues; `go vet` → 0)
- **cross_platform_build.darwin**: exit 0
- **cross_platform_build.windows**: exit 0 (GOOS=windows GOARCH=amd64)
- **total_run_phase_files**: 3 (internal/cli/update.go [modified], internal/cli/update_hooks_guidance_test.go [new], .moai/specs/SPEC-HOOK-CONFIG-SAFETY-001/progress.md [modified — §E.2/§E.3 + frontmatter]) + 3 SPEC frontmatter-only edits (spec.md, plan.md, acceptance.md — status draft→in-progress)
- **m1_to_mN_commit_strategy**: single M1 commit (Tier S, one milestone) carrying implementation + characterization tests + frontmatter `draft → in-progress` transition across all 4 plan-phase artifacts + progress.md §E.2/§E.3 population

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-07
sync_commit_sha: "82cc1ff95"  # single sync commit carrying in-progress → completed transition + 3-phase close; backfilled in chore commit per feedback_era_commit_sha_field_format
sync_status: pass
changelog_entry_position: "[Unreleased] > Added"
frontmatter_status_transitions:
  spec_md: in-progress → completed       # single sync commit (manager-docs owned; merged 3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 REQ-LR-008/009)
  plan_md: in-progress → completed
  acceptance_md: in-progress → completed
  progress_md: in-progress → completed
canary_compliance_check:
  spec_lint: pending                      # moai spec lint not run in this sync-phase; SPEC frontmatter 12-field schema manually verified (status + updated + 10 other canonical fields present)
  era_classification: V3R6                # H-4: §E.2 + §E.4 + sync_commit_sha present (once backfilled); created 2026-07-07 ≥ 2026-04-01 modernEraThreshold
sync_phase_artifacts:
  changelog_updated: true                 # CHANGELOG.md [Unreleased] > Added — SPEC-HOOK-CONFIG-SAFETY-001 entry
  readme_updated: false                   # moai update /hooks guidance is internal CLI UX; README/README.ko/docs-site do not document it
  frontmatter_transitioned: true          # spec/plan/acceptance/progress frontmatter status → completed (all 4 artifacts atomic on single sync commit)
  progress_e4_backfilled: true            # §E.4 Sync-phase Audit-Ready Signal populated (this block)
mx_tag_validation: pass                   # @MX:NOTE + @MX:SPEC annotations on emitHooksReviewGuidance helper verified present at internal/cli/update.go:353-365 (run-phase emission; sync-phase validation-only, no new MX tags added — sync is cross-cutting concern)
b12_self_test:
  a_pre_emission_grep: pass               # grep -c 'SPEC-HOOK-CONFIG-SAFETY-001' CHANGELOG.md = 0 pre-emission (no duplicate, HALT avoided)
  b_ac_count_match: pass                  # acceptance.md SSOT AC count = 7 (AC-CS-001..007); CHANGELOG entry references "7/7 AC"
  c_file_path_verification: pass          # all paths claimed in CHANGELOG entry (internal/cli/update.go, internal/cli/update_hooks_guidance_test.go, .moai/specs/SPEC-HOOK-CONFIG-SAFETY-001/) verified via ls
```

### Sync-phase Self-Verification (manager-docs variant)

- **E1 Frontmatter status transitions**: all 4 SPEC artifacts (spec.md / plan.md / acceptance.md / progress.md) `status: in-progress → completed` applied atomically on this single sync commit per Status Transition Ownership Matrix (spec-frontmatter-schema.md L72+L106). Body content untouched — only frontmatter `status:` field edited; `updated:` field already at sync commit date (2026-07-07).
- **E2 SHA backfill**: `sync_commit_sha` field will be backfilled in a separate chore commit per `feedback_era_commit_sha_field_format` (sync_commit_sha is a literal YAML field; backfill chore commit pattern is the established SPEC discipline). Run-phase SHA `e87bc205c` already in §E.3.
- **E3 CHANGELOG entry**: `[Unreleased] > Added` section — SPEC-HOOK-CONFIG-SAFETY-001 entry appended (B12 pre-emission grep count 0 confirmed; 7 AC count matches acceptance.md SSOT AC-CS-001..007; file paths verified via ls).
- **E4 Scope-boundary grep**: only spec.md / plan.md / acceptance.md / progress.md (frontmatter status only — body untouched) + CHANGELOG.md in this commit's stat (no implementation source files in scope per Forbidden modifications).
- **E5 Pathspec discipline**: `git add` invoked with specific paths only — no `git add -A`.
- **E6 Commit subject**: `docs(SPEC-HOOK-CONFIG-SAFETY-001): sync-phase artifacts — Tier S 3-phase close` follows the canonical sync-commit subject convention per Status Transition Ownership Matrix.

### 3-Phase Close Record

- **Plan → Run → Sync**: complete on 2026-07-07 (single-day close, Tier S scope).
- **Plan-phase**: manager-spec, iter-2 PASS-WITH-DEBT 0.96 (skip-eligible ≥ 0.90). Verify-before-author pipeline complete (Greps 1-3 in spec.md §A.2 re-run independently, 0-hit confirmed — not memory-derived).
- **Run-phase**: manager-develop, single M1 commit `e87bc205c` carrying implementation + characterization tests + progress.md §E.2/§E.3 population + 4-artifact frontmatter `draft → in-progress` transition + run_commit_sha backfill `f7107f76b`. All 7 AC at PASS or PASS-WITH-DEBT (only PASS-WITH-DEBT is AC-CS-006 — 3 pre-existing unrelated test failures verified via stash-and-retest: `TestRunHookEvent_ReadInputError` nil-pointer panic, `TestCloseSubjectDoctrineAmendment` lifecycle-sync-gate.md drift, `TestBuild_WritesContextUsageWithSessionID` 1M-vs-256K context-window assertion).
- **Sync-phase** (this commit): manager-docs, single sync commit carrying 4-artifact frontmatter `in-progress → completed` transition + §E.4 audit-ready signal + CHANGELOG entry. MX Tag validation executed as sync sub-step (no new tags needed — run-phase `@MX:NOTE`/`@MX:SPEC` already present on `emitHooksReviewGuidance` helper at update.go:353-365).
- **Epic context**: P5 of moai-adk hooks-hardening Epic. P1 = SPEC-HOOK-SESSIONSTART-PROBE-001 (completed 2026-07-07, sync recovery `a503af8c5`). P2/P3/P4 absorbed as stale-memory non-defects (`feedback_hypothesis_as_defect`). **P5 = this SPEC (real UX gap confirmed by mechanical grep — the opposite of P2/P3/P4 stale-memory phantoms)**. P5 closes the hooks-hardening Epic.

## §F Phase 0.95 Mode Selection

**Decision**: Mode 5 (sub-agent) — sequential single Agent() per milestone (Tier S coding-heavy default).

**Input parameters**:
- tier: S
- scope (file count): 2 (internal/cli/update.go edit + internal/cli/update_test.go characterization test)
- domain count: 1 (CLI completion UX)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: N/A (Tier S, single domain)

**Mode evaluation**:
- Mode 1 (trivial): not selected — semantic change (new guidance line) + characterization test authoring required
- Mode 2 (background): not selected — write task (update.go edit)
- Mode 3 (agent-team): not selected — Tier S, single domain, team prereqs unmet
- Mode 4 (parallel): not selected — coding-heavy + single domain (Anthropic coding-task parallelism caveat)
- Mode 5 (sub-agent): **selected** — Tier S coding-heavy default, sequential single Agent() per milestone
- Mode 6 (workflow): not selected — scope < 30 files, not high-volume mechanical

**Decision**: Mode 5 (sub-agent)

**Justification**: Tier S coding-heavy work (update.go edit + characterization tests). Per Anthropic's coding-task parallelism caveat, sequential sub-agent is the safe default. Single domain (CLI completion UX), 2 files — no parallelism benefit. Manager-develop delegated with cycle_type=tdd (characterization test-first per acceptance.md).

## §G IGGDA Kickoff Predicate

**Conditions** (4-condition compound predicate per orchestration-mode-selection.md §H.1):
- (a) intent clarity 100%: PASS — resume message drained intent (WHAT = /hooks guidance on Template sync complete branch; HOW = Option B unconditional trigger; AC = AC-CS-001..007; scope = update.go + update_test.go)
- (b) plan-auditor PASS: PASS — iter-2 PASS-WITH-DEBT 0.96 (skip-eligible ≥0.90; Phase 0.5 re-execution skipped per 4-condition compound predicate: PASS / ≥0.90 / artifact-hash unchanged / within 24h)
- (c) Tier S or M: PASS — Tier S
- (d) no dangerous keywords / destructive scope: PASS — scope = CLI completion UX guidance; no security/payment/critical-infrastructure keywords; no --pr flag; not destructive

**Verdict**: auto-proceed (all 4 conditions hold)

**Implementation Kickoff Approval**: user-approved via AskUserQuestion (2026-07-07, "run-phase 진입" selected) — the AskUserQuestion round was ISSUED and BLOCKED until user response per §H.2 (auto-proceed reduces framing, not blocking nature).

**Timestamp**: 2026-07-07
