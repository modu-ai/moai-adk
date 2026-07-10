---
id: SPEC-CLIFIX-HYGIENE-001
title: "CLI Structure and Hygiene Remediation — update.go decomposition, dead code, hardcoding sweep, i18n policy, rune safety (P4)"
version: "0.1.0"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: Low
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, audit-remediation, refactoring, dead-code, hardcoding, i18n, p4"
era: V3R6
tier: L
dependencies: [SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001, SPEC-CLIFIX-LINTER-STALE-001]
---

# SPEC-CLIFIX-HYGIENE-001 — CLI Structure and Hygiene Remediation (P4)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §3 clusters 1/2/4/5 (structure/hygiene rows) + §4 rows 6-9 + §5 P4 roadmap row |

## §A Context

The P4 roadmap row targets maintenance cost and recurrence risk: update.go is a 3,181-line single file; roughly 500 lines of dead code remain across the CLI (unwired safety mechanisms, unreachable step closures, unused helpers); env-var names and thresholds are inlined against the §14 hardcoding policy (with inject/clear drift already observed for GLM env names); six files carry Korean user-facing strings against the `error_messages: en` policy; four sites truncate strings at byte boundaries corrupting CJK content; deepMerge3Way silently drops user-added keys; tokens are written 0644; the wizard echoes PATs and hardcodes its stepper total; and the worktree package "parses" YAML via strings.Contains (comments activate CG mode) and injects unquoted paths into tmux commands.

This is a behavior-preserving DDD SPEC except where the defect IS the behavior (merge key drop, permissions, PAT echo, YAML misparse) — those are corrected with reproduction tests. Findings SSOT: audit §3 clusters 1/2/4/5 + §4 cross-cutting rows 6-9. Re-verify all anchors at run time.

## §B Requirements (GEARS)

- REQ-HYG-001-001: The update command implementation shall be decomposed from the single 3,181-line update.go into focused files (update_sync.go, update_merge.go, update_wizard.go, update_settings.go or equivalent seams), with characterization tests establishing behavior before the split and passing unchanged after it.
- REQ-HYG-001-002: The CLI shall remove the dead code identified by the audit (~500 lines), including update_cleanup.go functions left unwired after the P0 lock wiring, unreachable Backup/Restore step closures (update.go:618-717), buildGLMEnvVars, ttyConfirmer, worktree_validation/init_layout dead paths, the always-empty excludedDirs block, and the unused manifest load.
- REQ-HYG-001-003: The CLI shall replace inline GLM env-var name literals with envkeys.go constants and one shared glmEnvVarSet() helper consumed by every inject and clear site, so the inject/clear sets cannot drift.
- REQ-HYG-001-004: The CLI shall extract inline thresholds and limits (tier thresholds duplicated twice, dispatcher timeouts, size caps, retry/circuit literals) into defaults.go single-source constants referenced by all users.
- REQ-HYG-001-005: The CLI shall present English user-facing strings in doctor.go, migration.go, clean.go, and web_port* per the `error_messages: en` language policy, removing the Korean hardcoded strings.
- REQ-HYG-001-006: The CLI shall provide a single rune-boundary truncate helper and use it at the four byte-slicing truncation sites (constitution, tool_policy, github.go, and the remaining audited site), so multi-byte content is never split mid-rune.
- REQ-HYG-001-007: When deepMerge3Way merges user configuration with a new template (update.go:2396-2452), keys present only in the old user configuration shall be preserved in the merge result instead of being silently dropped.
- REQ-HYG-001-008: When update persists github_token or gitlab_token into user.yaml (update.go:2641-2651), the CLI shall write the file with 0600 permissions.
- REQ-HYG-001-009: The setup wizard shall mask PAT input using password echo mode (wizard/questions.go:119-154) and shall compute the stepper total dynamically from the actually-displayed questions (wizard/wizard.go:99,126).
- REQ-HYG-001-010: The worktree package shall parse workflow configuration via yaml.v3 instead of strings.Contains line matching (worktree/tmux_integration.go:160, worktree/new.go:481) so commented-out lines cannot activate CG mode, and shall quote worktree paths inserted into tmux initial commands (tmux_integration.go:114,124) so paths with spaces or metacharacters survive.

## §C Scope

### In Scope

- update.go decomposition with characterization safety net; the audited dead-code inventory; §14 hardcoding sweep (env names + thresholds); Korean→English UI strings in the six audited files; rune-truncate helper adoption at 4 sites; deepMerge3Way old-key preservation; token file 0600; wizard PAT masking + dynamic stepper; worktree yaml.v3 + tmux quoting.

### Out of Scope — Earlier-priority remediation

- Everything owned by the P0-P3 SPECs (critical data-loss fixes, exit-code contracts, locked writers, linter staleness). This SPEC runs last (P4) and rebases on all of them; in particular the dead-code removal must not delete the lock path newly wired by SPEC-CLIFIX-CRITICAL-001.

### Out of Scope — Full i18n framework

- No message-catalog/i18n framework is introduced; strings are corrected to English literals per policy. Localizing CLI output is future work.

### Out of Scope — YAML 3-way merge package extraction

- The audit's suggestion to extract the YAML 3-way merge into an independent package is noted but deferred; only the old-key preservation defect is fixed in place.

### Out of Scope — Behavioral redesign of update flow

- Step ordering, prompts, and update semantics stay identical; decomposition is file/mechanical structure only.

## §D Acceptance Criteria

- AC-HYG-001-001: Given the characterization suite captured pre-split, When update.go is decomposed into the focused files, Then the suite passes unchanged and no file in the update cluster exceeds 1,200 lines (maps REQ-HYG-001-001)
- AC-HYG-001-002: Given the audited dead-code inventory, When the cleanup lands, Then the listed symbols are gone, the package still builds, and no production caller referenced them (maps REQ-HYG-001-002)
- AC-HYG-001-003: Given the GLM env inject and clear sites, When their key sets are compared, Then both derive from the single glmEnvVarSet()/envkeys constants and are identical (maps REQ-HYG-001-003)
- AC-HYG-001-004: Given the audited threshold literals, When grepping the CLI tree, Then tier thresholds, dispatcher timeouts, and size caps exist once in defaults.go and inline duplicates are gone (maps REQ-HYG-001-004)
- AC-HYG-001-005: Given the six audited files, When scanning for Hangul characters in user-facing strings, Then zero remain (maps REQ-HYG-001-005)
- AC-HYG-001-006: Given a CJK string longer than each truncation limit, When each of the four sites truncates it, Then output is valid UTF-8 with no broken rune (maps REQ-HYG-001-006)
- AC-HYG-001-007: Given a user config with keys absent from the new template, When deepMerge3Way merges, Then the old-only keys are present in the result (maps REQ-HYG-001-007)
- AC-HYG-001-008: Given an update that writes tokens to user.yaml, When file permissions are inspected, Then the mode is 0600 (maps REQ-HYG-001-008)
- AC-HYG-001-009: Given the wizard PAT question and a variable question count, When the wizard renders, Then PAT input is masked and the stepper shows N/N with dynamic totals never exceeding 100% (maps REQ-HYG-001-009)
- AC-HYG-001-010: Given a workflow.yaml where team_mode lines exist only as comments and a worktree path containing spaces, When worktree CG detection and tmux spawn run, Then CG mode is not activated by the comment and the tmux command receives the quoted path intact (maps REQ-HYG-001-010)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001, SPEC-CLIFIX-CONTRACT-001, SPEC-CLIFIX-CONCURRENCY-001, SPEC-CLIFIX-LINTER-STALE-001 — strict P0→P1→P2→P3→P4 execution order; this SPEC touches nearly every file the earlier SPECs modified and must rebase on their results.
- Non-goal: performance work, new CLI features, or dependency additions.
- Non-goal: ast-grep recurrence-prevention rules for hardcoding (audit §4 processual recommendation) — tracked separately from this code SPEC.
