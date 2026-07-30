---
id: SPEC-CLIFIX-HYGIENE-001
title: "CLI Structure and Hygiene Remediation — update.go decomposition, dead code, hardcoding sweep, i18n policy, rune safety (P4)"
version: "0.2.0"
status: draft
created: 2026-07-10
updated: 2026-07-30
author: manager-spec
priority: Low
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, audit-remediation, refactoring, dead-code, hardcoding, i18n, p4"
era: V3R6
tier: L
depends_on: [SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001, SPEC-CLIFIX-LINTER-STALE-001]
---

# SPEC-CLIFIX-HYGIENE-001 — CLI Structure and Hygiene Remediation (P4)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §3 clusters 1/2/4/5 (structure/hygiene rows) + §4 rows 6-9 + §5 P4 roadmap row |
| 0.2.0 | 2026-07-30 | manager-spec | Stale-anchor rescope after plan-audit iter-1 FAIL (0.58). Re-verified every anchor against current worktree (post P0-P3 merge). Dropped REQ-HYG-001-007 (deepMerge3Way old-key drop — already fixed during extraction to `internal/merge/`; `deepMergeMap` at `strategies.go:360` preserves user-added keys at line 388) and REQ-HYG-001-008 (token 0600 — already satisfied by the F1 security redesign: `update.go:1375-1438` intentionally does NOT persist `github_token`/`gitlab_token`, delegating credentials to `gh`/`glab` CLI). Renumbered old 009→007 (wizard PAT mask; stepper half dropped — `stepperDenominator` at `wizard.go:225` already dynamic) and old 010→008 (worktree YAML parse + tmux quoting; split `team_mode` vs `tmux_preferred` fields). Shrunk REQ-HYG-001-001 (update.go is 1,905 lines, not 3,181; 8 sibling files already extracted — residual is a per-file ceiling on update.go itself) and REQ-HYG-001-002 (init_layout.go absent; excludedDirs/manifest-load symbols gone). Refreshed every line anchor; replaced stale numbers with content-token (symbol) anchors where code still moves. Added `### Out of Scope — Broader i18n sweep` covering ~24 additional Hangul-bearing production files deferred to a follow-up SPEC. |

## §A Context

The P4 roadmap row targets maintenance cost and recurrence risk. As of 2026-07-30 (post P0-P3 merge, worktree HEAD = origin/main): `update.go` is a 1,905-line single file with 8 already-extracted sibling files in the update cluster (`update_archive.go`, `update_clean_install.go`, `update_cleanup.go`, `update_deny_migration.go`, `update_namespace_protect.go`, `update_preserve_inventory.go`, `update_tux.go`, `update_noise.go`) — the bulk decomposition named in the original audit is done, but `update.go` itself still exceeds a reasonable per-file ceiling; env-var names and thresholds are inlined against the §14 hardcoding policy; six audited files carry Korean user-facing strings against the `error_messages: en` policy; four sites truncate strings at byte boundaries corrupting CJK content; the setup wizard echoes PATs in plaintext (the dynamic stepper total is already shipped); and the worktree package "parses" YAML via `strings.Contains` (comments activate CG mode) and injects unquoted paths into tmux commands.

Two originally-scoped defects turned out to be already fixed during the P0-P3 window and are therefore dropped from this SPEC (see HISTORY v0.2.0): the `deepMerge3Way` old-key-drop defect was resolved when the merge machinery was extracted to `internal/merge/` (the `deepMergeMap` function explicitly classifies and preserves user-added keys), and the token-file 0600 requirement was superseded by the stronger F1 security fix that does not persist tokens to `user.yaml` at all.

This is a behavior-preserving DDD SPEC except where the defect IS the behavior (PAT echo, YAML misparse) — those are corrected with reproduction tests. Findings SSOT: audit §3 clusters 1/2/4/5 + §4 cross-cutting rows 6-9, re-verified against the current tree on 2026-07-30.

## §B Requirements (GEARS)

- REQ-HYG-001-001: The update command implementation shall establish a per-file line ceiling for the update cluster (no file in `internal/cli/update.go` or its `update_*.go` siblings exceeds 1,200 lines), with characterization tests establishing behavior before any further split and passing unchanged after it. (Current state: `update.go` = 1,905 lines; 8 siblings already extracted. Residual decomposition work, if any, targets only the residual over-ceiling portion of `update.go` itself.)
- REQ-HYG-001-002: The CLI shall remove the verified dead code from the update cluster and adjacent packages — `buildGLMEnvVars` (test-only callers), `ttyConfirmer` (after its deferred SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 pairing is resolved), and the `worktree_validation.go` whole-file candidate — after per-symbol caller-graph re-verification at run time. Symbols confirmed LIVE (`scanDeprecatedPaths`, `cleanup_old_backups`) and P0-wired (`acquireUpdateLock`, `cleanStaleLock`) are EXCLUDE-KEEP. (Original-audit items `init_layout.go`, `excludedDirs`, and the unused manifest load are gone from the current tree and are no longer actionable.)
- REQ-HYG-001-003: The CLI shall replace inline GLM env-var name literals with `envkeys.go` constants and one shared `glmEnvVarSet()` helper consumed by every inject and clear site, so the inject/clear sets cannot drift.
- REQ-HYG-001-004: The CLI shall extract inline thresholds and limits (the `[]int{1, 3, 5, 10}` tier-threshold literal currently duplicated at THREE sites — `harness.go:150` default, `harness.go:480` struct default, `hook.go:1013` `defaultTierThresholds`; dispatcher timeouts at `hook.go:237` and `hook.go:361`, both `30*time.Second`; size caps; retry/circuit literals) into `defaults.go` single-source constants referenced by all users.
- REQ-HYG-001-005: The CLI shall present English user-facing strings in `doctor.go`, `migration.go`, `clean.go`, and `web_port*.go` per the `error_messages: en` language policy, removing the Korean hardcoded strings. (Broader Hangul sweep across ~24 additional production files is deferred — see §C `Out of Scope — Broader i18n sweep`.)
- REQ-HYG-001-006: The CLI shall provide a single rune-boundary truncate helper and use it at the four byte-slicing truncation sites (constitution, tool_policy, github.go, and the remaining audited site), so multi-byte content is never split mid-rune.
- REQ-HYG-001-007: The setup wizard shall mask PAT input using password echo mode for the `github_token` and `gitlab_token` questions. The input is currently rendered via `huh.NewInput()` at `internal/cli/wizard/wizard.go:329` (the `QuestionTypeInput` render dispatch) with no echo masking; the fix adds `.EchoMode(tty.EchoPassword)` (or equivalent) gated on the token-question IDs. (The dynamic stepper half of the original audit finding is already satisfied by `stepperDenominator` at `wizard.go:225` and is dropped from this REQ.)
- REQ-HYG-001-008: The worktree package shall parse workflow configuration via `yaml.v3` instead of `strings.Contains` line matching. Two distinct parse sites are in scope, parsing two distinct fields: (a) `team_mode` field — `internal/cli/worktree/tmux_integration.go:175` currently uses `strings.Contains(trimmed, "team_mode:")` so commented-out lines activate CG mode; (b) `tmux_preferred` field — `internal/cli/worktree/new.go:481` currently uses `strings.Contains(trimmed, "tmux_preferred:")` with the same comment-activation hazard. The worktree package shall additionally quote worktree paths inserted into tmux initial commands (`tmux_integration.go:114,124`) so paths with spaces or metacharacters survive.

## §C Scope

### In Scope

- Per-file line ceiling for the update cluster with characterization safety net; the verified dead-code inventory (buildGLMEnvVars, ttyConfirmer post-pairing, worktree_validation.go); §14 hardcoding sweep (env names + thresholds, with the 3-site tier-threshold duplication); Korean→English UI strings in the six audited files; rune-truncate helper adoption at 4 sites; wizard PAT masking; worktree `yaml.v3` parse for BOTH `team_mode` and `tmux_preferred` + tmux path quoting.

### Out of Scope — Earlier-priority remediation

- Everything owned by the P0-P3 SPECs (critical data-loss fixes, exit-code contracts, locked writers, linter staleness). This SPEC runs last (P4) and rebases on all of them; in particular the dead-code removal must not delete the lock path newly wired by SPEC-CLIFIX-CRITICAL-001.

### Out of Scope — Full i18n framework

- No message-catalog/i18n framework is introduced; strings are corrected to English literals per policy. Localizing CLI output is future work.

### Out of Scope — YAML 3-way merge package extraction

- The audit's original suggestion to extract the YAML 3-way merge into an independent package has been overtaken by events: the merge machinery now lives in `internal/merge/` (extracted during the P0-P3 window), and the old-key-drop defect named in REQ-HYG-001-007 (v0.1.0) is already fixed there. No further merge-package work is in scope.

### Out of Scope — Token persistence policy

- The v0.1.0 token-file-0600 requirement is overtaken by the stronger F1 security redesign: the update wizard intentionally does NOT persist `github_token`/`gitlab_token` to `user.yaml` at all (delegating credentials to the `gh`/`glab` CLI store). See `internal/cli/update.go:1375-1438` for the in-tree characterization of this behavior. No file-permission work is in scope.

### Out of Scope — Broader i18n sweep

- A full-tree Hangul sweep (`grep -rl '[가-힣]' internal/cli/` returned 30 production `.go` files on 2026-07-30) materially exceeds the six audited files in REQ-HYG-001-005. The additional ~24 files — including `fang.go`, `glm_tools.go`, `harness_clusters.go`, `loop.go`, `preference/{correction,decay,freshness,gate,proficiency,toggle}.go`, `schema_bridge.go`, `specid/specid.go`, `state.go`, `web.go`, `wizard/questions.go`, `wizard/translations.go`, `worktree/new.go`, `harness/execute.go`, `pr/watch.go`, `agentlint/agent_lint.go`, `profile_setup*.go`, `glamour_style.go` — are deferred to a follow-up SPEC pending a per-file classification (user-facing CLI output vs internal diagnostic vs test-fixture Hangul). Scope-widening is a user decision; this SPEC does not guess policy scope for those files.

### Out of Scope — Behavioral redesign of update flow

- Step ordering, prompts, and update semantics stay identical; decomposition is file/mechanical structure only.

## §D Acceptance Criteria

- AC-HYG-001-001: Given the characterization suite captured pre-split, When the residual over-ceiling portion of `update.go` (1,905 lines today) is decomposed to satisfy the 1,200-line ceiling, Then the suite passes unchanged and no file in the update cluster (`internal/cli/update.go` + `update_*.go` siblings) exceeds 1,200 lines (maps REQ-HYG-001-001)
- AC-HYG-001-002: Given the verified dead-code inventory, When the cleanup lands, Then the listed symbols (`buildGLMEnvVars`, `ttyConfirmer` post-pairing, `worktree_validation.go` whole-file candidate) are gone, the package still builds, and no production caller referenced them (maps REQ-HYG-001-002)
- AC-HYG-001-003: Given the GLM env inject and clear sites, When their key sets are compared, Then both derive from the single `glmEnvVarSet()`/`envkeys` constants and are identical (maps REQ-HYG-001-003)
- AC-HYG-001-004: Given the audited threshold literals, When grepping the CLI tree, Then the `[]int{1, 3, 5, 10}` tier-threshold literal, dispatcher timeouts, and size caps exist once in `defaults.go` and the THREE inline tier-threshold sites are reduced to one (maps REQ-HYG-001-004)
- AC-HYG-001-005: Given the six audited files, When scanning for Hangul characters in user-facing strings, Then zero remain (maps REQ-HYG-001-005)
- AC-HYG-001-006: Given a CJK string longer than each truncation limit, When each of the four sites truncates it, Then output is valid UTF-8 with no broken rune (maps REQ-HYG-001-006)
- AC-HYG-001-007: Given the `github_token` / `gitlab_token` wizard questions, When the wizard renders them, Then PAT input is masked (password echo mode) — verified at the `huh.NewInput()` render site at `internal/cli/wizard/wizard.go:329` (maps REQ-HYG-001-007)
- AC-HYG-001-008: Given a `workflow.yaml` where `team_mode` lines exist only as comments AND a worktree path containing spaces, When worktree CG detection (`tmux_integration.go:175`) and tmux spawn run, Then CG mode is not activated by the comment AND the `tmux_preferred` parse (`new.go:481`) is also YAML-driven AND the tmux command receives the quoted path intact (maps REQ-HYG-001-008)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001, SPEC-CLIFIX-LINTER-STALE-001 — strict P0→P1→P2→P3→P4 execution order; this SPEC touches nearly every file the earlier SPECs modified and must rebase on their results.
- Non-goal: performance work, new CLI features, or dependency additions.
- Non-goal: ast-grep recurrence-prevention rules for hardcoding (audit §4 processual recommendation) — tracked separately from this code SPEC.
- Non-goal: re-introducing the dropped v0.1.0 REQ-HYG-001-007 (deepMerge3Way) or REQ-HYG-001-008 (token 0600) — both are documented as already-satisfied in HISTORY v0.2.0 and §C Out-of-Scope subsections.
