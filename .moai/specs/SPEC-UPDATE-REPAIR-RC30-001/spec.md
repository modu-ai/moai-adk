---
id: SPEC-UPDATE-REPAIR-RC30-001
title: "moai update rc.30 repair: data-safety must-fix items, shared-root profile reach, and update output redesign"
version: "0.1.0"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: "GOOS (via manager-spec, card t1619)"
priority: P0
phase: "v3.2.0-rc.30 target"
module: "internal/cli/update"
lifecycle: spec-anchored
tags: "update, repair, profile, backup, migration, output, tux, rc30"
tier: L
---

# SPEC-UPDATE-REPAIR-RC30-001 — moai update rc.30 repair and output redesign

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-10 | GOOS (manager-spec) | Initial plan-phase draft for card t1619. Source: `.moai/reports/t1619/update-audit.md` (audit) and `.moai/reports/t1619/operator-update-output.md` (operator paste). Lane addendum (lane-14) applied: shared-root model per operator direction D1, landing order t1591 then t1578 then t1619, absorption as run milestone M1. |

Tree measured: `2aab5f797b75983e132af451da68f69e3426557b` (main tip at measurement, worktree branch `WT-moai-update-repair`). The installed `moai` binary was built from `4f8aba061`, which descends from this tree (`git merge-base --is-ancestor 2aab5f797 4f8aba061` exited 0). No claim in this SPEC relies on the installed binary.

## A. Background

`moai update --force` on the mo.ai.kr project (binary v3.2.0-rc.29, paste dated 2026-10-10 about 02:34 KST) showed that skills and agents disappeared from the Claude Code session, that settings and config lost fields, and that the output contradicted itself. The audit traced the causes. Its rc.30 must-fix list is the eight items in `update-audit.md` § Fix list for rc.30 (lines 261-271), mapped in § E below. The audit's other items are carried as can-follow items in § C.

Operator direction D1 (verbatim, Korean, from the lane addendum): "원래 카드에 ~/.moai ~/.claude ~/.codex 에 공통 스킬과 에이전트를 여기에 moai init, update 시 복사 하도록 하자고 했는데 완료가 되었는지 확인. 그 방식을 따른다." Meaning: common skills and agents are copied into the user-shared roots by `moai init` and `moai update`. That shared-root model is kept. This SPEC does not adopt "keep project copies" or "warning only" as the steady-state answer.

## B. Requirements (GEARS)

Every requirement is traced to items in § C (U-IDs). Requirement ceiling for Tier L is 25; this SPEC carries 20.

- **REQ-RC30-001** (Ubiquitous): The update flow shall keep the four shared roots `~/.claude/skills`, `~/.claude/agents`, `~/.agents/skills`, and `~/.codex/agents` as the install source for common skills and agents, per operator direction D1. Items: U-01.
- **REQ-RC30-002** (Event-driven): **When** a project is launched under a profile config directory (`CLAUDE_CONFIG_DIR`), the shared assets shall reach the profile session through a mechanism that does not violate REQ-002 of SPEC-USER-ASSET-INSTALL-001 (spec.md:255). Items: U-01. Status: BLOCKED-ON t1591 (see § D.2).
- **REQ-RC30-003** (Unwanted behavior): The update flow shall not delete a project-side managed skill or agent copy (for example `.claude/skills/moai/**`) until the runnable profile-visibility check AC-RC30-01 has passed on the tree being updated. Items: U-01, U-02.
- **REQ-RC30-004** (Ubiquitous): Before any migration removal, the run backup shall contain a copy of every file the removal deletes, and the migration step shall run after the backup step. Items: U-02.
- **REQ-RC30-005** (Event-driven): **When** a former plugin-mode project has optional packs that are not installed, the update output shall name the exact `moai bundle add <name>` commands for those packs. Items: U-03.
- **REQ-RC30-006** (Ubiquitous): `moai bundle list` shall print every catalog bundle name with its skill and agent counts, and `moai bundle add --help` shall name the valid bundles. Items: U-04.
- **REQ-RC30-007** (Ubiquitous): `ConfigManager.Save` shall not drop any `llm.yaml` key present before the save, including `deployment_mode` and keys the typed struct does not model, and shall keep the file's comments. Items: U-05.
- **REQ-RC30-008** (Event-driven): **When** a project has no deployment-mode record and a migration archive exists, the update shall print a visible warning instead of re-stamping the mode silently. Items: U-06.
- **REQ-RC30-009** (Ubiquitous): `moai update --restore <dir>` shall restore the backed-up `.claude/settings.json`, and the restore output shall list the paths that the backup does not cover. Items: U-07.
- **REQ-RC30-010** (Ubiquitous): Backup retention shall be counted per run and shall not depend on directory-name shape. Items: U-07.
- **REQ-RC30-011** (Ubiquitous): Every count and status line in the update output shall be a measured value. No literal count string shall appear in a format string. Items: U-08, U-18.
- **REQ-RC30-012** (Unwanted behavior): The update flow shall refuse to refresh a user root with a binary older than the version recorded there. Items: U-11.
- **REQ-RC30-013** (Ubiquitous): The update output shall print one version banner. Items: U-17.
- **REQ-RC30-014** (Ubiquitous, R2a): Step progress shall show N of M steps, a progress bar, and one line per step with `✓`, `!`, or `✗`, updated in place on a TTY and printed as plain lines otherwise. The glyph set is the one already whitelisted by REQ-TUXIU-004 of SPEC-CLI-TUX-INIT-UPDATE-001. Items: U-26.
- **REQ-RC30-015** (Ubiquitous, R2b): File lists shall collapse to per-category counts by default. The full list shall appear only behind `--verbose` or in a log file whose path the summary prints. Items: U-16.
- **REQ-RC30-016** (Ubiquitous, R2d): The closing summary shall fit one screen and show counts of updated, merged, preserved, moved, and excluded items, a user action line only when one is needed, the backup location, and one restore command. Items: U-27.
- **REQ-RC30-017** (Ubiquitous, R2e): Conflicts and warnings shall appear above the summary with prominence, and no two sentences of the output shall contradict each other. Items: U-08, U-28.
- **REQ-RC30-018** (State-driven, R2f): **While** `NO_COLOR` is set, output is not a TTY, `CI` is set, or the terminal is narrower than the threshold defined in `internal/tui` defaults, the output shall fall back to plain lines that carry the same information. Items: U-29.
- **REQ-RC30-019** (Ubiquitous, R2g): The locale strings of the new output shall be consistent across the locale set fixed by decision Q1 (§ J). Items: U-30.
- **REQ-RC30-020** (Ubiquitous, R2 reuse): The redesigned output shall reuse the existing `internal/tui` components and shall be covered by the golden snapshot path `make tui-snapshot`. Items: U-26, U-27, U-28, U-29.

## C. Item register

Disposition vocabulary (lane brief): **IN-NOW** = no overlap with either in-flight branch by file, so the item is scheduled for run now. **BLOCKED-ON <branch>** = the target file is changed by that branch (file-level overlap per § G), so the item is implemented after milestone M1 absorbs the branch. **GAP-INVESTIGATE** = the cause is not established; the next investigation step is named. Overlap is measured with `git diff --name-only main...<branch>` at tree `2aab5f797`. Root-cause citations marked **re-read** were read in this tree; citations marked **audit** are carried from the audit and are not yet re-read (see research.md Gaps).

| U-ID | Source | Priority | Disposition | Root cause (file:line) | Acceptance |
|------|--------|----------|-------------|------------------------|------------|
| U-01 | Must-fix 1; D1; REQ-001/002/003 | High | BLOCKED-ON WT-user-asset-bundle (t1591) | `internal/userassets/paths.go:36-43` (re-read): `ResolveRoots` returns `$HOME` roots only. `internal/cli/update.go:200` (re-read) prints "Current version"; the install and prune phase is `user_asset_phase.go:105` (audit). Branch touches `paths.go`, `user_asset_phase.go`, `migrate_project_assets.go`. | AC-RC30-01, AC-RC30-02 |
| U-02 | Must-fix 2; D2; REQ-004 | High | BLOCKED-ON WT-user-asset-bundle (t1591) | Removal `os.Remove` at `internal/cli/migrate_project_assets.go:168` (audit line; the audit's `update/` path is wrong, see research.md Gaps). The migration call is `update_template_sync.go:1183` (audit). Branch touches `internal/cli/migrate_project_assets.go` (the file the audit cites under the `update/` path; both paths are checked in research.md). | AC-RC30-03 |
| U-03 | Must-fix 3 (D3); REQ-005 | High | BLOCKED-ON WT-user-asset-bundle (t1591) | `internal/cli/user_asset_phase.go:105-125` (audit): core-only selection for former plugin-mode projects. Branch touches `internal/cli/user_asset_phase.go`. | AC-RC30-04 |
| U-04 | Must-fix 3 (D15); REQ-006 | Medium | IN-NOW | `internal/cli/bundle.go` (audit): no listing verb. Neither branch touches `bundle.go`. | AC-RC30-05 |
| U-05 | Must-fix 4 (D7); REQ-007 | High | IN-NOW | `internal/config/manager.go:225` (re-read): `func (m *ConfigManager) Save() error`. `internal/config/types.go:284` (re-read): `TeamMode` field. `types.go:383` and `:398` (re-read): `Fable` fields. `types.go:389` (re-read): comment says legacy alias fields Opus/Sonnet/Haiku were removed. `deploy_mode.go:23` (re-read): `DeploymentMode` read in a separate wrapper. Neither branch touches `internal/config/manager.go` or `types.go`. | AC-RC30-06 |
| U-06 | Must-fix 4 (D8); REQ-008 | High | IN-NOW | `internal/cli/update_template_sync.go:361` and `:1100-1114` (audit). Neither branch touches this file. | AC-RC30-07 |
| U-07 | Must-fix 5 (D9); REQ-009, REQ-010 | High | IN-NOW | `internal/cli/update_restore.go:43` (audit): restore covers `.moai/config` only. Retention `CleanupOldBackups(…, 5)` at `update_template_sync.go:859` (audit), backup shape at `internal/backup/backup.go:209` (audit). The entry point `update.go:269` is not edited by this item (it stays in the migration-fixes file set). | AC-RC30-08 |
| U-08 | Must-fix 6 (D12); REQ-011, REQ-017 | High | IN-NOW | `internal/cli/update_archive.go:367` (re-read): literal `"total: %d skills archived, 0 user customizations modified"`. `internal/cli/update_tux.go:119-120` (re-read): `! %d conflict` classification pill. `update_tux.go:237` (re-read): outcome `! %d conflict`. "Updated N files" on a no-op: `update_tux.go:217-262` (audit). | AC-RC30-09 |
| U-09 | Must-fix 7 (D4); REQ-001 scope note | Medium | BLOCKED-ON WT-user-asset-bundle (t1591) | `internal/cli/migrate_project_assets.go:160` and `:238-242` (audit): message "opt in via 'moai bundle add'" for a skill with no catalog entry; `update_reconcile.go:64-70` (audit): preserve-only roots. Branch touches `migrate_project_assets.go`. | AC-RC30-10 |
| U-10 | Must-fix 7 (D5) | Medium | BLOCKED-ON WT-user-asset-bundle (t1591) | `internal/cli/migrate_project_assets.go:168` (audit): file removed, emptied parent directory left. Branch touches `migrate_project_assets.go`. | AC-RC30-11 |
| U-11 | Must-fix 8 (D10); REQ-012 | High | BLOCKED-ON WT-update-migration-fixes (t1578) and WT-user-asset-bundle (t1591) | `internal/userassets/install.go:612-624` (audit): `stateManifestMatch` rewrites a pristine file to the binary's template with no version comparison. Both branches touch `internal/userassets/install.go`. | AC-RC30-12 |
| U-12 | Can-follow (D6) | Medium | BLOCKED-ON WT-user-asset-bundle (t1591) | `migrate_project_assets.go:154-161` (audit): kept optional agents and `.agents/skills` copies are never refreshed. | AC-RC30-13 |
| U-13 | Can-follow (D10, skip decision) | Medium | IN-NOW | `internal/cli/update_template_sync.go:247` and `:1143` (audit): skip compares version strings only. Neither branch touches this file. | AC-RC30-14 |
| U-14 | Lane brief IN-NOW example; audit §8 | Medium | IN-NOW | `internal/cli/update_template_sync.go:619` (audit): `template.WithVersion(version.GetVersion())`. Nothing compares the stamp with the deployed bytes. Audit §8 records `version: "cutover/github-flow"` in the mo.ai.kr HEAD. | AC-RC30-15 |
| U-15 | Can-follow (D11) | Medium | GAP-INVESTIGATE | Settings merge; `permissions.ask` emptied (audit §4). Cause inferred ("the 3-way merge is the inferred cause"), not reproduced. Next step: a fixture with a base containing `permissions.ask: [sudo, chmod, chown]`, a template without the key, and a user copy with the key; run the merge and read the result. | AC-RC30-16 |
| U-16 | Can-follow (D13); R2b; REQ-015 | Medium | IN-NOW | `internal/cli/update/report/outcome.go:101-115` (re-read): `RenderReconciliation` appends one `preserved:` line per path. `update_tux.go:340-345` (re-read): retained-key advisory already has a `--verbose` hint. `--verbose` is bound by REQ-UN-005 and REQ-UN-010 of SPEC-V3R6-UPDATE-NOISE-001 and must keep that meaning. | AC-RC30-17 |
| U-17 | Can-follow (D14a); R2c; REQ-013 | Medium | BLOCKED-ON WT-update-migration-fixes (t1578) | `internal/cli/update.go:200` (re-read): "Current version" line. The identity band is `update_tux.go` (audit `:231-232`), and REQ-TUXIU-015 of SPEC-CLI-TUX-INIT-UPDATE-001 decides that band. `update.go:835` (re-read) carries a comment naming card t1527 D1 for the re-exec pass. The removal of the `update.go:200` line is inside the t1578 file. | AC-RC30-18 |
| U-18 | Can-follow (D14b); REQ-011 | Medium | IN-NOW | Preview deployer `update_template_sync.go:1157` (audit) versus the run deployer, giving 266 against 423 files. | AC-RC30-19 |
| U-19 | Can-follow (D16) | Low | GAP-INVESTIGATE | Audit: 83 rule references and 17 commands name a project-relative `moai` skill that is not deployed. Count not reproduced. Next step: grep the templates for the path and classify each hit against the § D.1 outcome. | AC-RC30-20 |
| U-20 | Can-follow (D17) | Low | GAP-INVESTIGATE | Audit: `docs-site/content/ko/cli-reference/update.md:295` lists `.claude/agents/moai/`. Path and 4-locale parity not verified in this tree. | AC-RC30-20 |
| U-21 | Can-follow (D18) | Medium | BLOCKED-ON WT-update-migration-fixes (t1578) | `internal/cli/update.go:652-657` (re-read): `profile.GetCurrentName()` and `profile.SyncToProjectConfig(".", prefs)`. The profile is read from the environment, so the result depends on where update runs. Relevant to R1's profile requirement. | AC-RC30-21 |
| U-22 | Can-follow (D19) | Low | BLOCKED-ON WT-update-migration-fixes (t1578) | `normalizeLegacyRootDenySpecifiers` (audit `update.go:428-433`). Owned by card t1578; this card does not implement it. Tracked for traceability. | (owned by t1578) |
| U-23 | Audit §4 | Medium | GAP-INVESTIGATE | `.moai/config/sections/gate.yaml` lost 12 comment lines (audit §4). The audit does not establish which step drops them (backup, wipe, deploy, or YAML-node merge-back). SPEC-UPDATE-YAML-PRESERVE-001 (completed, on main) decided comment preservation. Next step: a fixture with 12 comment lines, run backup, wipe, deploy, restore, and diff. | AC-RC30-22 |
| U-24 | Audit §4, §E | Medium | GAP-INVESTIGATE | Which command re-marshalled `llm.yaml` between 2026-10-08 20:39 and 2026-10-10 02:12, and reset `team_mode` from `glm` to `""` (audit §4 cites `launcher.go:397`, `resetTeamModeForCC`). Next step: replay each `ConfigManager.Save` caller against a fixture `llm.yaml` and diff. | AC-RC30-23 |
| U-25 | Audit §E | High | GAP-INVESTIGATE | Untraced removal at 2026-10-09 11:26:48 of `.claude/skills/moai/**` (46 files) and six core agents, with no backup directory (audit §A). Next step: read `config-change-audit.log` at `2026-10-09T02:26Z` and correlate with the shell history of the operator. Do not fix blindly. | AC-RC30-24 |
| U-26 | R2a; REQ-014, REQ-020 | High | IN-NOW | Step lines and bar: `update_tux.go:134-139` (re-read): `"  %s %s %d/%d steps"` bar renderer. Operator paste shows appended lines (non-TTY). The in-place path is `tui.ProgressLine` (SPEC-V3R6-UPDATE-PROGRESS-001, REQ-UPR-001..007). Neither branch touches `update_tux.go` or `internal/tui`. | AC-RC30-25 |
| U-27 | R2d; REQ-016, REQ-020 | High | IN-NOW | Closing summary: `update_tux.go:217-299` (audit). Consumes U-07 (restore wording) and U-08 (measured values). The migration tally at `migrate_project_assets.go:200` (audit) is read-only input; no edit to that file is required if the tally is returned (research.md Gaps). | AC-RC30-25 |
| U-28 | R2e; REQ-017 | High | IN-NOW | Pre-merge risk class shown as conflict: `update_tux.go:91-95` (audit) and `update_tux.go:119-120` (re-read). Real outcome: `Reconciliation: … 0 conflict(s)` at `update/report/outcome.go:106` (re-read). The disagreement is between two measured values; the fix is to show the measured outcome above the summary. | AC-RC30-25 |
| U-29 | R2f; REQ-018 | Medium | IN-NOW | `NO_COLOR` and non-TTY already decided by REQ-TUXIU-041 and REQ-TUXIU-022 (SPEC-CLI-TUX-INIT-UPDATE-001). The narrow-terminal threshold is new. Threshold lives in `internal/tui` defaults (hardcoding rule). | AC-RC30-25 |
| U-30 | R2g; REQ-019 | Medium | GAP-INVESTIGATE | Locale mechanism for update output not yet read in this tree. Brief says "four locale strings" (R2g) and "16 locales" (R1). Blocked on decision Q1 (§ J). | none (see acceptance.md Budget) |

Item counts: 30 items. IN-NOW 13 (U-04, U-05, U-06, U-07, U-08, U-13, U-14, U-16, U-18, U-26, U-27, U-28, U-29). BLOCKED-ON 10 (U-01, U-02, U-03, U-09, U-10, U-11, U-12, U-17, U-21, U-22). GAP-INVESTIGATE 7 (U-15, U-19, U-20, U-23, U-24, U-25, U-30).

Note on U-22: it is owned by card t1578 and is tracked here for traceability only. It is not implemented by this card.

## D. Shared-root status, profile reach, and init path

### D.1 Status of the D1 shared-root work, one row per location

Intended content is taken from D1 and from `SPEC-USER-ASSET-INSTALL-001` (completed, on main). Code citations are from `internal/userassets/paths.go:36-43` (re-read in this run). Runtime cells are marked Gap where the command was not run in this run.

| Location | Intended content (D1) | What the code does (this tree) | Observed in this run | Gap |
|----------|-----------------------|--------------------------------|----------------------|-----|
| `~/.claude/skills` | Common `moai*` skills copied by `moai init` and `moai update` | Root declared at `paths.go:38`: `filepath.Join(home, ".claude", "skills")`, `home` from `userHomeDirFn()` (audit `update.go:558`) | Gap: no `ls` run on the operator home in this run. Audit §2 records 19 `moai*` skills, byte-identical to the template (`diff -rq`), but that is audit evidence. | Runtime state of the root in this run not measured. Profile sessions do not read it (audit §3, empirical check). |
| `~/.claude/agents` | Six core agents | Root at `paths.go:39` | Gap: not listed in this run. Audit §A records the six core agents removed on 2026-10-09 with no backup. | Restoration path and the untraced removal (U-25). |
| `~/.agents/skills` | Codex mirror of common skills | Root at `paths.go:40` | Gap: not listed in this run. Audit §2 records stale `.agents/skills` copies in the gitignored project tree, which are a different location. | Whether `moai init` and `moai update` populate this root is not verified in this run. |
| `~/.codex/agents` | Codex agent definitions | Root at `paths.go:41` | Gap: not listed in this run. Audit §2 records `.codex/agents/moai/*.toml` as project-side. | Whether the user root is populated is not verified. |
| `~/.moai` | Manifest and backups only; no skills or agents directories (leader measurement on main) | Not a `ResolveRoots` entry. The manifest and backup roots are `.moai-backups/` and `.moai/backups/` in the project (audit §6). | Gap: `ls ~/.moai` not run in this run. | Leader's claim is unverified in this tree. |

### D.2 Profile reach: shared assets for `CLAUDE_CONFIG_DIR` sessions

Facts established: profile sessions launch with `CLAUDE_CONFIG_DIR` set to a profile directory, and the shared roots in § D.1 are under `$HOME`, not under the profile (audit §3; `paths.go:36-43` re-read). Audit §3 records an empirical check: a profile session's skill listing lacks the `moai-plan`, `moai-run`, and `moai-sync` skills that exist only in `~/.claude/skills`.

Hard rule (binding on this SPEC): project-side copies, for example `.claude/skills/moai/**` and `.claude/agents/moai/**`, are **not deleted** until AC-RC30-01 has passed on the tree being updated. AC-RC30-01 is a runnable profile-visibility check. Until it passes, the migration keeps those copies (REQ-RC30-003).

Reach mechanism. Two candidates, one decision:
1. **Write into the profile directory.** Install the common assets into `<profile>/skills` and `<profile>/agents` as well. This writes into a profile directory. SPEC-USER-ASSET-INSTALL-001 REQ-002 (spec.md:255, "The system shall keep per-profile settings folders per-profile …") and P6 (spec.md:237-238, "profile sessions do not see the shared user assets in v1 (declared limitation, doc-visible). Forced by REQ-002 and C2: every install-into-profile option violated REQ-002") forbid this. The lane addendum reads REQ-002 as making a profile directory never an install target. The full REQ-002 clause has not been read beyond line 255 (Gap). Status: **BLOCKED-ON t1591** (WT-user-asset-bundle carries SPEC-USERASSET-DEPLOY-GUARD-001, which changes the same `userassets` code).
2. **Shared roots stay the only source, and the profile session reads them.** Not achievable from this repository: `CLAUDE_CONFIG_DIR` governs where Claude Code reads user-level assets, and this card does not control that path. Recorded as a Gap, not a design.

Decision point (for the operator, § J Q2): either (a) amend REQ-002 by reopening SPEC-USER-ASSET-INSTALL-001, which allows an install into a profile directory, or (b) keep the project-side copies as the profile-reach mechanism permanently. Option (b) conflicts with the lane addendum's instruction not to pick "keep project copies", so it is listed only as the alternative. Until the operator decides, U-01 stays BLOCKED-ON t1591 and the removal gate of REQ-RC30-003 holds.

AC-RC30-01 (runnable profile-visibility check): the command form is a Gap. The form must be measured before it is written into the acceptance matrix. The check must show that a non-interactive Claude Code session launched with `CLAUDE_CONFIG_DIR=<profile>` lists one known shared skill. Its verdict is not a pass until the measured command and output are recorded in research.md.

### D.3 Init path check (which roots `moai init` writes)

Status: **Gap**. This run did not execute `moai init` from a build of this tree, and the static path through `init.go` was not traced in time. Required evidence, to be produced in the first run milestone: build this tree (`make build`, binary placed by path), run `HOME=<scratch> <built-binary> init <scratch-project>`, then list the four roots and `~/.moai` under `<scratch>`. The gap against D1 is the difference between the roots listed and § D.1.

## E. rc.30 must-fix items mapped to U-IDs

The eight must-fix items are the entries of `update-audit.md` § Fix list for rc.30, "Must fix", items 1-8 (lines 264-271). No other item is treated as must-fix.

| # | Audit must-fix text (abridged) | U-IDs | Disposition |
|---|-------------------------------|-------|-------------|
| 1 | D1: do not remove project-side skills or agents under a profile | U-01 | BLOCKED-ON WT-user-asset-bundle (t1591). The "blocking warning" option is not adopted (lane addendum). |
| 2 | D2: back up every migration-removed file; move migration after backup | U-02 | BLOCKED-ON WT-user-asset-bundle (t1591) |
| 3 | D3 and D15: seed or announce optional bundles; add `moai bundle list` | U-03 (D3), U-04 (D15) | U-03 BLOCKED-ON WT-user-asset-bundle (t1591); U-04 IN-NOW |
| 4 | D7 and D8: stop `ConfigManager.Save` dropping `deployment_mode` and unknown keys | U-05 (D7), U-06 (D8) | IN-NOW, IN-NOW |
| 5 | D9: `--restore` applies backed-up `settings.json`; per-run retention; plain no-backup statement | U-07 | IN-NOW |
| 6 | D12: remove hardcoded "0 user customizations modified", pre-merge "! conflict", "Updated N files" on no-op | U-08 | IN-NOW |
| 7 | D4 and D5: correct non-catalog skill message; remove emptied dirs | U-09 (D4), U-10 (D5) | BLOCKED-ON WT-user-asset-bundle (t1591), both |
| 8 | D10: refuse a user-root downgrade | U-11 | BLOCKED-ON WT-update-migration-fixes (t1578) and WT-user-asset-bundle (t1591) |

Eight of eight must-fix items are mapped. Four of them (items 1, 2, 7, 8) and the D3 half of item 3 are BLOCKED-ON an in-flight branch.

## F. Queue dispositions (QUEUE-DISPOSITION)

Card text for t1527, t1594, and t1612 was not read in this run (Gap). The recommendations use the scopes the audit mapped and the committed SPEC text. Each row must be confirmed by the operator.

| Card | Scope (basis) | Recommendation | Reason |
|------|---------------|----------------|--------|
| t1527 | Init and update output redesign. `SPEC-UPDATE-MIGRATION-001` § H (spec.md:228-230) names "update/init UX overhaul (t1527)" as the owner of banners, progress, and severity prefixes. | **Absorb** the update-side presentation into this SPEC (R2 items U-16 to U-18, U-26 to U-29). **Defer** the init-side presentation. | The update output is this SPEC's R2 scope. Init presentation is already decided by SPEC-CLI-TUX-INIT-UPDATE-001 Group D and touches `init*.go`, which WT-user-asset-bundle also changes. |
| t1594 | User-asset residue. Audit maps D4 and D5 to t1594 by inference. | **Absorb** U-09 and U-10 into this SPEC, then close t1594 once they land. | D4 and D5 sit in `migrate_project_assets.go`, the file the BLOCKED-ON items already cover. Two cards on one file would conflict. |
| t1612 | CLI output redesign. Audit maps D12 and D13 to t1527 and t1612 and D17 as adjacent. | **Overlap**: update-side items are absorbed here; any CLI output outside `update` is deferred to a follow-up card. | R2 covers the update surface. Non-update CLI output has no decided scope in this tree. |

## G. Landing and absorption

1. **Landing order.** t1591 lands first, then t1578, then t1619. This card ends at merge-ready. No card merges anything until card t1616 lands (it builds the local-main integration tooling). The landing face after t1616 is local main. This SPEC names no remote push, no develop command, and no merge into main as a landing path. The AGENTS.md wording for that path belongs to t1616.
2. **Absorption as run milestone M1.** After plan-audit PASS, and at run start, this tree absorbs `WT-user-asset-bundle` (t1591; cae7e28fe at measurement) and `WT-update-migration-fixes` (t1578; 317ed31b6 at measurement) with plain `git merge`. Both tip SHAs are recorded in `progress.md` § G Absorption Record. Every range is measured from the merge-base with the absorbed ref, not from a literal SHA. If a branch moves later, the lane absorbs again.
3. **Non-overlapping items run now.** IN-NOW items (output TUX, report wording, config-save loss, and the rest of § C marked IN-NOW) are scheduled for run now. The plan does not wait for the overlap decision.

Note on the branch base: the worktree was cut from main tip `2aab5f797`. `.claude/rules/local/gitflow-lane-protocol.md` § 1 names `develop` as the card base, and AGENTS.md § 3 names main. The lane must confirm which base applies before absorption (Gap).

## H. Dependencies

- Card **t1621**: the slug rule in `factoryWorktreeSlug`. Out of scope for this SPEC. Named here, not as an item.
- Card **t1622**: the factory inbox error in `factory_msg_list` (`sql: no rows in result set`). Out of scope for this SPEC. Named here, not as an item.
- Card **t1591** (WT-user-asset-bundle, SPEC-USERASSET-DEPLOY-GUARD-001 on that branch; PR pending in the audit's record): BLOCKED-ON source for U-01, U-02, U-03, U-09, U-10, U-11, U-12.
- Card **t1578** (WT-update-migration-fixes, SPEC-UPDATE-MIGRATION-FIX-001 on that branch): BLOCKED-ON source for U-11, U-17, U-21, U-22.
- Card **t1616**: builds the local-main integration tooling. Landing of this card waits for it.
- SPEC-USER-ASSET-INSTALL-001 (completed, on main): P6 and REQ-002 (spec.md:237-238, :255).
- SPEC-CLI-TUX-INIT-UPDATE-001 (completed, on main): REQ-TUXIU-004, -012 to -016, -020 to -022, -041, -054.
- SPEC-V3R6-UPDATE-NOISE-001 (implemented, on main): `--verbose` semantics, REQ-UN-005 and REQ-UN-010.
- SPEC-V3R6-UPDATE-PROGRESS-001 (implemented, on main): `tui.ProgressLine`, REQ-UPR-001 to -007.
- SPEC-UPDATE-MIGRATION-001 (completed, on main): REQ-UPM-030 to -032 (summary reporting, plain counts).

## I. Exclusions

### Out of Scope — factory slug and inbox defects

- The slug rule in `factoryWorktreeSlug` (card t1621). See § H.
- The factory inbox error in `factory_msg_list` (card t1622). See § H.

### Out of Scope — init-side presentation

- Init banner, init success card, and init-side output (decided by SPEC-CLI-TUX-INIT-UPDATE-001 Group D). Deferred per § F.

### Out of Scope — mo.ai.kr operator recovery

- The recovery steps in `update-audit.md` § C are operator actions in the mo.ai.kr project. They are not code items of this SPEC.

### Out of Scope — plugin distribution and new profile semantics

- No plugin distribution path (voided by operator decision 2026-10-06, SPEC-UPDATE-MIGRATION-001 § H).
- No change to REQ-002 of SPEC-USER-ASSET-INSTALL-001 in this SPEC. Amending it is decision Q2 (§ J).

## J. Decision points

Recorded in `decision-index.md`. Summary:

- **Q1 (FOUNDER, product-level).** "16 locales" in R1 and "four locale strings" in R2g: which set is meant? The repository's neutrality contract (`.claude/rules/moai/development/coding-standards.md`) lists 16 programming languages, and the user-facing locales are ko, en, ja, and zh. Blocks U-30 and REQ-RC30-019.
- **Q2 (FOUNDER, product-level).** Profile reach: amend REQ-002 of SPEC-USER-ASSET-INSTALL-001 to permit a profile install, or keep project copies as the permanent mechanism. Blocks U-01 and AC-RC30-01.
- **Q3 (FOUNDER, implementation-level, Default applied by the published rule if no operator verdict).** Queue disposition of t1527's init side. Blocks no item of this SPEC.

## K. Requirement and AC budget

Tier L: requirements 20 of 25 (§ B). Acceptance criteria 25 of 25 (`acceptance.md`, AC-RC30-01 to AC-RC30-25). U-22 (owned by card t1578) and U-30 (blocked on decision Q1) carry no criterion. Several U-items share a criterion (for example U-26 to U-29 share AC-RC30-25). If the operator splits the SPEC, the U-30 criterion moves with it.
