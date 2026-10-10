---
id: SPEC-UPDATE-REPAIR-RC30-001
title: "moai update rc.30 repair: data-safety must-fix items, shared-root profile reach, and update output redesign"
version: "0.1.2"
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
| 0.1.0 | 2026-10-10 | GOOS (manager-spec) | Initial plan-phase draft for card t1619. Source: `.moai/reports/t1619/update-audit.md` and `.moai/reports/t1619/operator-update-output.md`. Lane addendum applied: shared-root model per operator direction D1, landing order t1591 then t1578 then t1619, absorption as run milestone M1. |
| 0.1.1 | 2026-10-10 | GOOS (manager-spec) | Leader revision: U-01 DECISION-PENDING (Q2); queue overlaps recorded with disposition undecided; Q1 resolved from disk; leader intake M3 added as U-31 with REQ-RC30-021 and AC-RC30-26. |
| 0.1.2 | 2026-10-10 | GOOS (manager-spec) | Response to plan-audit verdict FAIL (`.moai/reports/t1619/plan-audit.md`): one REQ form (REQ-RC30-001 to -025) throughout; GEARS rewrite of REQ-RC30-014 to -017; REQ-RC30-002 response pending Q2; Q7 carries no applied default; overlap facts pinned to tip SHAs; AC budget 26 of 25 held as an open leader decision (decision-index.md B1); AC RED cells recorded per MP-8. |

Tree measured: `8fa0dc24b`. Its code paths are identical to main tip `2aab5f797b75983e132af451da68f69e3426557b` (`git diff --name-only 2aab5f797 HEAD -- internal cmd` printed nothing), so RED cells cite `8fa0dc24b`. The SPEC-only commits after it leave those paths unchanged (`git diff --name-only 8fa0dc24b 00ab5bbb1 -- internal cmd` is empty). The installed `moai` binary was built from `4f8aba061`, which descends from this tree. No claim in this SPEC relies on the installed binary.

## A. Background

`moai update --force` on the mo.ai.kr project (binary v3.2.0-rc.29, paste dated 2026-10-10 about 02:34 KST) showed that skills and agents disappeared from the Claude Code session, that settings and config lost fields, and that the output contradicted itself. The audit traced the causes. Its rc.30 must-fix list is the eight items in `update-audit.md` § Fix list for rc.30 (lines 261-271), mapped in § E. The audit's other items are carried as can-follow items in § C.

Operator direction D1 (verbatim, Korean, from the lane addendum): "원래 카드에 ~/.moai ~/.claude ~/.codex 에 공통 스킬과 에이전트를 여기에 moai init, update 시 복사 하도록 하자고 했는데 완료가 되었는지 확인. 그 방식을 따른다." Meaning: common skills and agents are copied into the user-shared roots by `moai init` and `moai update`. That shared-root model is kept.

## B. Requirements (GEARS)

One REQ form is used everywhere: REQ-RC30-NNN. The Tier L requirement ceiling is 25; this SPEC carries 25.

- **REQ-RC30-001** (Ubiquitous): The update flow shall keep the four shared roots `~/.claude/skills`, `~/.claude/agents`, `~/.agents/skills`, and `~/.codex/agents` as the install source for common skills and agents, and shall keep an agent or skill copy that is not refreshed from them visible in the update output. Items: U-01, U-12.
- **REQ-RC30-002** (Event-driven): **When** a project is launched under a profile config directory (`CLAUDE_CONFIG_DIR`), the shared assets shall reach the profile session by the mechanism selected under operator decision Q2 (pending). Items: U-01, U-21. Status: pending operator decision Q2 (spec.md § D.2; decision-index.md Q2). Judged by AC-RC30-21 for profile sync; a criterion for the selected mechanism is added when Q2 is answered.
- **REQ-RC30-003** (Unwanted behavior): The update flow shall not delete a project-side managed skill or agent copy (for example `.claude/skills/moai/**`) until the runnable profile-visibility check AC-RC30-01 has passed on the tree being updated (AC-RC30-01 is moved to SPEC B; within this SPEC, REQ-RC30-003 is judged by AC-RC30-24). Items: U-01, U-25.
- **REQ-RC30-004** (Event-driven): **When** the migration removes a file, the run backup shall contain a copy of that file, the migration step shall run after the backup step, and an emptied skill directory shall be removed with its last file. Items: U-02, U-10. moved to SPEC B (ID kept; criteria moved with it).
- **REQ-RC30-005** (Event-driven): **When** a former plugin-mode project has optional packs that are not installed, or a managed skill has no catalog entry, the update output shall name the exact `moai bundle add <name>` command or the distinct no-catalog message for that skill. Items: U-03, U-09.
- **REQ-RC30-006** (Ubiquitous): `moai bundle list` shall print every catalog bundle name with its skill and agent counts, and `moai bundle add --help` shall name the valid bundles. Items: U-04.
- **REQ-RC30-007** (Ubiquitous): `ConfigManager.Save` shall not drop any `llm.yaml` key present before the save, including `deployment_mode` and keys the typed struct does not model, and the save and the section restore shall keep the comments of the file they write. Items: U-05, U-23, U-24.
- **REQ-RC30-008** (Event-driven): **When** a project has no deployment-mode record and a migration archive exists, the update shall print a visible warning instead of re-stamping the mode silently. Items: U-06.
- **REQ-RC30-009** (Ubiquitous): The user-facing restore does not write the live `.claude/settings.json`. The restore output lists `.claude/settings.json` as a not-restored path and tells the user how to copy it by hand. The restore output shall also list the other paths that the backup does not cover. Items: U-07.
- **REQ-RC30-010** (Ubiquitous): Backup retention shall be counted per run and shall not depend on directory-name shape. Items: U-07.
- **REQ-RC30-011** (Ubiquitous): Every count and status line in the update output shall be a value measured in the same run. No literal count string shall appear in a format string. Items: U-08, U-18.
- **REQ-RC30-012** (Unwanted behavior): The update flow shall not refresh a user root with a binary older than the version recorded there. Items: U-11.
- **REQ-RC30-013** (Ubiquitous): The update output shall print exactly one version banner. Items: U-17.
- **REQ-RC30-014** (State-driven): **While** the update output is a TTY, the update shall show step progress as N of M steps, a progress bar, and one line per step carrying `✓`, `!`, or `✗`, updated in place; **while** the output is not a TTY, the update shall print the same step lines as plain newline-terminated lines. Items: U-26.
- **REQ-RC30-015** (Where-modality): **Where** `--verbose` is set, the update shall print the full file list; **where** `--verbose` is not set, the update shall print per-category counts in place of the file list. Items: U-16.
- **REQ-RC30-016** (Event-driven): **When** the update completes, the closing summary shall fit one terminal screen and state the counts of updated, merged, preserved, moved, and excluded items, the backup location, and one restore command; **when** a user action is required, the summary shall add exactly one user action line. Items: U-27.
- **REQ-RC30-017** (Ubiquitous): Every conflict or warning line shall be printed above the closing summary, and the output shall print one conflict count, equal to the conflict count measured by the reconciliation. Items: U-08, U-28.
- **REQ-RC30-018** (State-driven): **While** `NO_COLOR` is set, the output is not a TTY, `CI` is set, or the terminal is narrower than the width threshold defined in the `internal/tui` defaults (value pending Q7), the update shall print plain lines that carry the same counts and statuses. Items: U-29.
- **REQ-RC30-019** (Ubiquitous): The locale tables of the new output shall define the same keys for each of the four user-facing locales ko, en, ja, and zh, and the rendered summary shall show the same counts in each locale (the four-locale set is resolved from disk in decision-index.md Q1). Items: U-30. Judged by the locale part of AC-RC30-25.
- **REQ-RC30-020** (Ubiquitous): The redesigned output shall reuse the existing `internal/tui` components and shall be covered by the golden snapshot path `make tui-snapshot`. Items: U-26, U-27, U-28, U-29.
- **REQ-RC30-021** (State-driven, leader intake 10-10, t1578 card-review M3): **While** the Claude harness is hidden for a project (codex-only projects, `hideClaude=true`, `internal/template/harness_fs.go:112`), the managed-surface integrity probe shall not treat the absence of `.claude/settings.json` as damage, and shall still check every other entry of its set. Items: U-31. moved to SPEC B (ID kept; criteria moved with it).
- **REQ-RC30-022** (Event-driven): **When** the template tree differs from the recorded tree at the same version, the update shall not skip the template sync. Items: U-13.
- **REQ-RC30-023** (Ubiquitous): The update shall report, as a visible warning, a version stamp that disagrees with the deployed bytes. Items: U-14.
- **REQ-RC30-024** (Unwanted behavior): The settings merge shall not drop a user array entry under permissions.* that the template no longer carries. Hook arrays are out of scope for rc.30. Items: U-15.
- **REQ-RC30-025** (Ubiquitous): Every shipped rule, command, and docs-site reference to a moai skill or agent path shall name a deployed target. Items: U-19, U-20.

## C. Item register

Disposition vocabulary (lane brief): **IN-NOW** = no file-level overlap with either in-flight branch, so the item is scheduled for run now. **BLOCKED-ON <branch>** = the target file is changed by that branch (§ C.1), so the item is implemented after milestone M1 absorbs the branch. **GAP-INVESTIGATE** = the cause is not established; the next investigation step is named. **DECISION-PENDING** = the item waits for an operator decision recorded in decision-index.md; it is not BLOCKED-ON any branch. Citations marked **re-read** were read in this tree; citations marked **audit** are carried from the audit and not yet re-read (research.md § 4).

### C.1 Overlap by branch name

The overlap facts are stated by branch name. Read them at run time with `git diff --name-only main...<branch> -- <path>` (empty output = not changed on that branch); the statements below are dated references, measured 2026-10-09 (research.md § 2 E-OVERLAP):
- WT-user-asset-bundle (t1591): absorbed at run start; the absorption-time tip is recorded in progress.md section G (Absorption Record).
- WT-update-migration-fixes (t1578): absorbed at run start; the absorption-time tip is recorded in progress.md section G (Absorption Record).
- Main = `2aab5f797`. Tree of the RED cells = `8fa0dc24b`, whose code paths equal main's. The SPEC-only commits after `8fa0dc24b` leave the code paths unchanged (`git diff --name-only 8fa0dc24b 00ab5bbb1 -- internal cmd` is empty).

No IN-NOW file is changed on either branch (research.md § 2 E-OVERLAP, verbatim). The BLOCKED-ON facts are re-derived per branch there. M1 re-derives them after absorption.

| U-ID | Source | Priority | Disposition | Root cause (file:line) | Acceptance |
|------|--------|----------|-------------|------------------------|------------|
| U-01 | Must-fix 1; D1; REQ-RC30-001, REQ-RC30-002, REQ-RC30-003 | High | DECISION-PENDING (Q2); no criterion in SPEC A (AC-RC30-01, AC-RC30-02 moved) | `internal/userassets/paths.go:36-43` (re-read): `ResolveRoots` returns `$HOME` roots only. Profile resolution `launcher.go:159` and `profile.go:240` (audit). `paths.go` is changed on WT-user-asset-bundle (research.md E-OVERLAP). | AC-RC30-01, AC-RC30-02 |
| U-02 | Must-fix 2; D2; REQ-RC30-004 | High | moved to SPEC B | Removal `os.Remove(p)` at `internal/cli/migrate_project_assets.go:168` (re-read, E-04). Migration call `internal/cli/update_template_sync.go:1185` (re-read, E-20). `migrate_project_assets.go` is changed on WT-user-asset-bundle. The audit's `update/` path and its `:1183` line are wrong (research.md § 4 item 5). | AC-RC30-03 |
| U-03 | Must-fix 3 (D3); REQ-RC30-005 | High | BLOCKED-ON WT-user-asset-bundle (t1591) | `internal/cli/user_asset_phase.go:105-125` (audit): core-only selection. The file is changed on WT-user-asset-bundle. | AC-RC30-04 |
| U-04 | Must-fix 3 (D15); REQ-RC30-006 | Medium | IN-NOW | `internal/cli/bundle.go` (re-read, E-12): subcommands `bundle`, `add <bundle-name>`, `remove <bundle-name>`; no `list`. Not changed on either branch. | AC-RC30-05 |
| U-05 | Must-fix 4 (D7); REQ-RC30-007 | High | IN-NOW | `internal/config/manager.go:225` (re-read): `Save()`. `internal/config/types.go:389` (re-read, E-15): alias fields are absent, only a comment remains. `manager.go` has no `deployment_mode` reference (re-read, E-14). `deploy_mode.go:23` (re-read). Not changed on either branch. | AC-RC30-06 |
| U-06 | Must-fix 4 (D8); REQ-RC30-008 | High | IN-NOW | `internal/cli/update_template_sync.go:361` and `:1100-1114` (audit). Not changed on either branch. | AC-RC30-07 |
| U-07 | Must-fix 5 (D9); REQ-RC30-009, REQ-RC30-010 | High | IN-NOW | `internal/cli/update_restore.go` has no `settings.json` reference (re-read, E-11). Retention `CleanupOldBackups(…, 5)` at `update_template_sync.go:859` (audit). `update.go:269` (audit) is the entry point and is not edited by this item. Not changed on either branch. | AC-RC30-08 |
| U-08 | Must-fix 6 (D12); REQ-RC30-011, REQ-RC30-017 | High | IN-NOW | `internal/cli/update_archive.go:282` (comment) and `:367` (code) hold the literal `0 user customizations modified` (re-read, E-01). `update_tux.go:119-120` (re-read): `! %d conflict` classification pill. `update_tux.go:237` (re-read): outcome conflict line. Measured-value fix for the no-op `Updated N files` line: `update_tux.go:217-262` (audit). The characterization test `internal/cli/tuxiu_characterization_test.go:201` pins the literal `3 skills archived, 0 user customizations modified` and the expectation `Updated 26 files` (E-22); its expectation is updated in the same M4 change, as part of U-08. | AC-RC30-09 |
| U-09 | Must-fix 7 (D4); REQ-RC30-005 | Medium | BLOCKED-ON WT-user-asset-bundle (t1591) | `internal/cli/migrate_project_assets.go:160` (re-read, E-05): message "opt in via 'moai bundle add <name>'" for a skill with no catalog entry. `update_reconcile.go:64-70` (audit). The file is changed on WT-user-asset-bundle. | AC-RC30-10 |
| U-10 | Must-fix 7 (D5); REQ-RC30-004 | Medium | moved to SPEC B | `internal/cli/migrate_project_assets.go:168` (re-read, E-04): `os.Remove(p)` removes one file per call; no directory-pruning step (E-04). The file is changed on WT-user-asset-bundle. | AC-RC30-11 |
| U-11 | Must-fix 8 (D10); REQ-RC30-012 | High | BLOCKED-ON WT-update-migration-fixes (t1578) and WT-user-asset-bundle (t1591) | `internal/userassets/install.go:612` (re-read, E-16): `case stateManifestMatch:`, with no version comparison. The file is changed on both WT-user-asset-bundle and WT-update-migration-fixes. | AC-RC30-12 |
| U-12 | Can-follow (D6); REQ-RC30-001 | Medium | BLOCKED-ON WT-user-asset-bundle (t1591) | `migrate_project_assets.go:154-161` (audit): kept optional agents and `.agents/skills` copies are never refreshed. The file is changed on WT-user-asset-bundle. | AC-RC30-13 |
| U-13 | Can-follow (D10, skip decision); REQ-RC30-022 | Medium | IN-NOW | `internal/cli/update_template_sync.go:247` and `:1143` (audit): skip compares version strings only. Not changed on either branch. | AC-RC30-14 |
| U-14 | Audit §8; REQ-RC30-023 | Medium | IN-NOW | `internal/cli/update_template_sync.go:456` and `:619` (re-read, E-06): `template.WithVersion(version.GetVersion())`. No call compares the stamp with the deployed bytes (E-06). Audit §8 records `version: "cutover/github-flow"` in the mo.ai.kr HEAD. | AC-RC30-15 |
| U-15 | Can-follow (D11); REQ-RC30-024 | Medium | GAP-INVESTIGATE | Settings merge; `permissions.ask` emptied (audit §4). Cause inferred ("the 3-way merge is the inferred cause"), not established. Starting observation: the RED-now fixture described in the next step, which has not been run. Next step: fixture with a base holding `permissions.ask: [sudo, chmod, chown]`, a template without the key, and a user copy with the key; run the merge and read the result. | AC-RC30-16 |
| U-16 | Can-follow (D13); R2b; REQ-RC30-015 | Medium | IN-NOW | `internal/cli/update/report/outcome.go:115` (re-read, E-02): `result += "\n  preserved: " + p` for each path. `update_tux.go:340-345` (re-read): retained-key advisory already has a `--verbose` hint. `--verbose` keeps the meaning of REQ-UN-005 and REQ-UN-010 of SPEC-V3R6-UPDATE-NOISE-001. | AC-RC30-17 |
| U-17 | Can-follow (D14a); R2c; REQ-RC30-013 | Medium | BLOCKED-ON WT-update-migration-fixes (t1578) | `internal/cli/update.go:200` (re-read, E-03): "Current version" line; `update.go:835` (re-read, E-03): comment names card t1527. The identity band is in `update_tux.go` (audit `:231-232`); REQ-TUXIU-015 of SPEC-CLI-TUX-INIT-UPDATE-001 decides that band. `update.go` is changed on WT-update-migration-fixes. | AC-RC30-18 |
| U-18 | Can-follow (D14b); REQ-RC30-011 | Medium | IN-NOW | Preview deployer `update_template_sync.go:1157` (audit) versus the run deployer, giving 266 against 423 files (audit). Not changed on either branch. | AC-RC30-19 |
| U-19 | Can-follow (D16); REQ-RC30-025 | Low | GAP-INVESTIGATE | Audit: 83 rule references and 17 commands name a project-relative `moai` skill that is not deployed. Re-run (E-07): `internal/template/templates/.claude/rules` lists the files that name `.claude/skills/moai/` (E-07). Next step: classify each hit against the § D.1 outcome. | AC-RC30-20 |
| U-20 | Can-follow (D17); REQ-RC30-025 | Low | GAP-INVESTIGATE | Audit: `docs-site/content/ko/cli-reference/update.md:295` lists `.claude/agents/moai/`. Re-run (E-08): the path `claude/agents/moai` has no match in `update.md` in this tree, so the audit line is not reproduced. Next step: locate the docs page the audit cites and classify the finding. | AC-RC30-20 |
| U-21 | Can-follow (D18); REQ-RC30-002 | Medium | BLOCKED-ON WT-update-migration-fixes (t1578) | `internal/cli/update.go:652` (re-read, E-17): `profileName := profile.GetCurrentName()`; `profile.SyncToProjectConfig(".", prefs)` at `update.go:657` (audit). The profile is read from the environment, so the result depends on where update runs. `update.go` is changed on WT-update-migration-fixes. | AC-RC30-21 |
| U-22 | Can-follow (D19) | Low | BLOCKED-ON WT-update-migration-fixes (t1578) | `normalizeLegacyRootDenySpecifiers` (audit `update.go:428-433`). Owned by card t1578; this card does not implement it. Tracked for traceability. | none |
| U-23 | Audit §4; REQ-RC30-007 | Medium | GAP-INVESTIGATE | `.moai/config/sections/gate.yaml` lost 12 comment lines (audit §4). The audit does not establish which step drops them. SPEC-UPDATE-YAML-PRESERVE-001 (completed, on main) decided comment preservation. Next step: fixture with 12 comment lines; run backup, wipe, deploy, restore, and diff. | AC-RC30-22 |
| U-24 | Audit §4, §E; REQ-RC30-007 | Medium | GAP-INVESTIGATE | Which command re-marshalled `llm.yaml` between 2026-10-08 20:39 and 2026-10-10 02:12, and reset `team_mode` from `glm` to `""` (audit §4 cites `launcher.go:397`). Next step: replay each `ConfigManager.Save` caller against a fixture `llm.yaml` and diff. | AC-RC30-23 |
| U-25 | Audit §E; REQ-RC30-003 | High | GAP-INVESTIGATE | Untraced removal at 2026-10-09 11:26:48 of `.claude/skills/moai/**` (46 files) and six core agents, with no backup directory (audit §A). The audit's log `.moai/logs/config-change-audit.log` is absent from this tree (re-read, E-18). Next step: locate the log in the operator profile, then correlate with shell history. Do not fix blindly. | AC-RC30-24 |
| U-26 | R2a; REQ-RC30-014, REQ-RC30-020 | High | IN-NOW | Step line renderer `update_tux.go:134-139` (re-read, E-10): `"  %s %s %d/%d steps"`. Operator paste shows appended lines (non-TTY). In-place path: `tui.ProgressLine` (SPEC-V3R6-UPDATE-PROGRESS-001). Not changed on either branch. | AC-RC30-25 |
| U-27 | R2d; REQ-RC30-016, REQ-RC30-020 | High | IN-NOW | Closing summary: `update_tux.go:217-299` (audit). Consumes U-08 (measured values, which M4 completes first; plan.md § B.5). The migration tally at `migrate_project_assets.go:200` (re-read, E-05) is read-only input. | AC-RC30-25 |
| U-28 | R2e; REQ-RC30-017 | High | IN-NOW | Pre-merge risk class shown as conflict: `update_tux.go:91-95` (re-read at 91 and 95) and `update_tux.go:119-120` (re-read). Real outcome: `Reconciliation: … 0 conflict(s)` at `update/report/outcome.go:106` (re-read). Fix: show the measured conflict count only, above the summary. | AC-RC30-25 |
| U-29 | R2f; REQ-RC30-018 | Medium | IN-NOW | `NO_COLOR` and non-TTY already decided by REQ-TUXIU-041 and REQ-TUXIU-022 (SPEC-CLI-TUX-INIT-UPDATE-001). The narrow-terminal threshold is new; it lives in `internal/tui` defaults; its value is decision Q7 (pending, no applied default). | AC-RC30-25 |
| U-30 | R2g; REQ-RC30-019 | Medium | GAP-INVESTIGATE | Whether update output uses a locale table is not yet read in this tree. The four-locale set is resolved from disk (decision-index.md Q1; research.md § 2 E-LOCALE). | AC-RC30-25 (locale part) |
| U-31 | Leader intake 10-10, t1578 card-review M3; REQ-RC30-021 | Medium | moved to SPEC B | Branch WT-update-migration-fixes, `internal/cli/update_integrity_probe.go:52` (line recorded at absorption, progress.md section G (Absorption Record); re-read, E-19): entry `{rel: ".claude/settings.json", check: probeJSON}`; `damageReason` returns `"missing"` when the file is absent (line recorded at absorption). Absent on main (`git cat-file -e main:internal/cli/update_integrity_probe.go` exits 128). On the tree, `internal/template/harness_fs.go:112` (re-read) hides `.claude` paths when `hideClaude` is set. | AC-RC30-26 |

Item counts (research.md § 2 E-R5 is the check): 31 rows; 3 moved to SPEC B (U-02, U-10, U-31); the 28 rows in SPEC A carry IN-NOW 13 (U-04, U-05, U-06, U-07, U-08, U-13, U-14, U-16, U-18, U-26, U-27, U-28, U-29). DECISION-PENDING 1 (U-01). BLOCKED-ON 7 (U-03, U-09, U-11, U-12, U-17, U-21, U-22). GAP-INVESTIGATE 7 (U-15, U-19, U-20, U-23, U-24, U-25, U-30).

Note on U-22: owned by card t1578 and tracked here for traceability only. Note on U-31: not one of the eight must-fix items; it comes from leader intake (§ E).

## D. Shared-root status, profile reach, and init path

### D.1 Status of the D1 shared-root work, one row per location (disposition: IN-NOW for the non-profile part)

Intended content is taken from D1 and from `SPEC-USER-ASSET-INSTALL-001` (completed, on main). Code citations are from `internal/userassets/paths.go:36-43` (re-read in this run). Runtime cells are marked Gap where the command was not run in this run. This section is not a U-row. Its completion is documentary evidence in research.md § 2, not a criterion.

| Location | Intended content (D1) | What the code does (this tree) | Observed in this run | Gap |
|----------|-----------------------|--------------------------------|----------------------|-----|
| `~/.claude/skills` | Common `moai*` skills copied by `moai init` and `moai update` | Root declared at `paths.go:38`: `filepath.Join(home, ".claude", "skills")`, `home` from `userHomeDirFn()` (audit `update.go:558`) | Gap: no `ls` run on the operator home in this run. Audit §2 records 19 `moai*` skills, byte-identical to the template, but that is audit evidence. | Runtime state not measured. Profile sessions do not read it (audit §3). |
| `~/.claude/agents` | Six core agents | Root at `paths.go:39` | Gap: not listed in this run. Audit §A records the six core agents removed on 2026-10-09 with no backup. | Restoration path and the untraced removal (U-25). |
| `~/.agents/skills` | Codex mirror of common skills | Root at `paths.go:40` | Gap: not listed in this run. | Whether `moai init` and `moai update` populate this root is not verified. |
| `~/.codex/agents` | Codex agent definitions | Root at `paths.go:41` | Gap: not listed in this run. | Whether the user root is populated is not verified. |
| `~/.moai` | Manifest and backups only; no skills or agents directories (leader measurement on main) | Not a `ResolveRoots` entry. The manifest and backup roots are `.moai-backups/` and `.moai/backups/` in the project (audit §6). | Gap: `ls ~/.moai` not run in this run. | Leader's claim is unverified in this tree. |

### D.2 Profile reach: shared assets for `CLAUDE_CONFIG_DIR` sessions (U-01, pending operator decision Q2)

Scope: U-01 only. The status table (§ D.1) and the init path check (§ D.3) are IN-NOW and are not part of this decision.

Facts, with their source:
- The shared roots are the four `$HOME` roots at `internal/userassets/paths.go:36-43` (re-read in this run).
- Audit §3 states that profile sessions do not see the shared user assets in this build (audit claim).
- SPEC-USER-ASSET-INSTALL-001 (completed, on main) declares P6 at spec.md lines 237-238 and REQ-002 at spec.md line 255 (both re-read on main in this run). Line 255 begins "The system shall keep per-profile settings folders per-profile". The rest of the REQ-002 clause, after line 255, was not re-read in this run (Gap).
- Audit §3 cites the profile mapping in `~/.moai/claude-profiles/launch.yaml` and the resolution at `launcher.go:159` (audit claims, not re-read).

Options (factual; the operator decides; no recommendation; full text in decision-index.md Q2):
- **(a)** A reach mechanism that writes nothing into a profile directory. None is established by evidence in this card.
- **(b)** A link or copy of the common assets inside the profile directory. It requires amending REQ-002 (spec.md line 255) to permit a link or copy in a profile directory. P6 (spec.md lines 237-238) would be restated to match. The amendment scope is the line-255 clause only.
- **(c)** Keep project-side copies when a profile session is detected. Its consequence against D1 is stated in decision-index.md Q2.

Follow-on if option (b) is chosen: manager-spec performs the REQ-002 amendment in place in this card, after WT-user-asset-bundle (t1591) is absorbed at run start (M1). The leader coordinates that with the other lane.

Gate: REQ-RC30-003 applies. No project-side copy is deleted; the absorbed removal call is merged inert behind `removalEnabled` (default false); no removal path is enabled until AC-RC30-01 is adopted (AC-RC30-01 is moved to SPEC B). The absorbed removal call (line recorded at absorption, progress.md section G (Absorption Record)) stays inert in SPEC A: switch `removalEnabled` defaults to false (plan.md § B.2, step 1; its judging test AC-RC30-02 moved to SPEC B).

### D.3 Init path check: which roots `moai init` writes (disposition: IN-NOW, not a U-row)

Status: **Gap**. This run did not execute `moai init` from a build of this tree. Required evidence, produced in the first run milestone: build this tree (`make build`, binary invoked by path), run `HOME=<scratch> <built-binary> init <scratch-project>`, then list the four roots and `~/.moai` under `<scratch>`.

## E. rc.30 must-fix items mapped to U-IDs

The eight must-fix items are the entries of `update-audit.md` § Fix list for rc.30, "Must fix", items 1-8 (lines 264-271). No other item is treated as must-fix. U-31 is not a must-fix item.

| # | Audit must-fix text (abridged) | U-IDs | Disposition |
|---|-------------------------------|-------|-------------|
| 1 | D1: do not remove project-side skills or agents under a profile | U-01 | DECISION-PENDING (operator decision Q2). The "blocking warning" option is not adopted (lane addendum). |
| 2 | D2: back up every migration-removed file; move migration after backup | U-02 | moved to SPEC B (split, d-20261009T235135Z-5e35) |
| 3 | D3 and D15: seed or announce optional bundles; add `moai bundle list` | U-03 (D3), U-04 (D15) | U-03 BLOCKED-ON WT-user-asset-bundle (t1591); U-04 IN-NOW |
| 4 | D7 and D8: stop `ConfigManager.Save` dropping `deployment_mode` and unknown keys | U-05 (D7), U-06 (D8) | IN-NOW, IN-NOW |
| 5 | D9: the user-facing restore does not write the live `.claude/settings.json` (verdict d-20261010T053033Z-7012); per-run retention; plain no-backup statement | U-07 | IN-NOW |
| 6 | D12: remove hardcoded "0 user customizations modified", pre-merge "! conflict", "Updated N files" on no-op | U-08 | IN-NOW |
| 7 | D4 and D5: correct non-catalog skill message; remove emptied dirs | U-09 (D4), U-10 (D5) | U-09 BLOCKED-ON WT-user-asset-bundle (t1591); U-10 moved to SPEC B |
| 8 | D10: refuse a user-root downgrade | U-11 | BLOCKED-ON WT-update-migration-fixes (t1578) and WT-user-asset-bundle (t1591) |

Item 1 (U-01) is DECISION-PENDING with no criterion in SPEC A. Item 2 (U-02) moved to SPEC B. Item 7 (U-09) stays BLOCKED-ON; U-10 is moved to SPEC B. Items 3 (U-03) and 8 (U-11) are BLOCKED-ON an in-flight branch.

## F. Queue overlaps (disposition undecided)

Each card below is recorded only as an overlap with this SPEC. No disposition is recorded for any of the three cards. Card text for t1527, t1594, and t1612 was not read in this run (research.md § 4 item 9).

| Card | Overlap facts (basis) | Status |
|------|------------------------|--------|
| t1527 | Owner of banners, progress rendering, and severity prefixes per `SPEC-UPDATE-MIGRATION-001` § H (spec.md:228-230). Overlaps this SPEC's update-side items U-16, U-17, U-18, and U-26 to U-29. The init-side presentation is decided by SPEC-CLI-TUX-INIT-UPDATE-001 Group D and touches `internal/cli/init*.go`, which WT-user-asset-bundle also changes (file-level overlap on WT-user-asset-bundle, research.md § 2 E-OVERLAP). | overlap, disposition undecided |
| t1594 | User-asset residue. The audit maps D4 and D5 to t1594 by inference. Overlaps U-09 and U-10, which sit in `internal/cli/migrate_project_assets.go`, a file changed on WT-user-asset-bundle. | overlap, disposition undecided |
| t1612 | CLI output redesign. The audit maps D12 and D13 to t1527 and t1612, and D17 as adjacent. Overlaps U-08 and U-16. Non-update CLI output has no decided scope in this tree (Gap). | overlap, disposition undecided |

## G. Landing and absorption

1. **Landing order.** t1591 lands first, then t1578, then t1619. This card ends at merge-ready. No card merges anything until card t1616 lands (it builds the local-main integration tooling). The landing face after t1616 is local main. This SPEC names no remote push, no develop command, and no merge into main as a landing path. The AGENTS.md wording for that path belongs to t1616.
2. **Absorption as run milestone M1.** After plan-audit PASS, at M1, this tree absorbs `WT-user-asset-bundle` (t1591) and `WT-update-migration-fixes` (t1578) with plain `git merge`, absorbed at run start; the absorption-time tip is recorded in progress.md section G (Absorption Record). Every range is measured from the merge-base with the absorbed ref, not from a literal SHA. If a branch moves later, the lane absorbs again.
3. **Non-overlapping items run now.** IN-NOW items (output TUX, report wording, config-save loss, and the rest of § C marked IN-NOW) are scheduled for run now. The plan does not wait for the overlap decision.

Note on the branch base: the worktree was cut from main tip `2aab5f797`. `.claude/rules/local/gitflow-lane-protocol.md` § 1 names `develop` as the card base, and AGENTS.md § 3 names main. The lane must confirm which base applies before absorption (research.md § 4 item 14).

## H. Dependencies

- Card **t1621**: the slug rule in `factoryWorktreeSlug`. Out of scope for this SPEC. Named here, not as an item.
- Card **t1622**: the factory inbox error in `factory_msg_list` (`sql: no rows in result set`). Out of scope for this SPEC. Named here, not as an item.
- Card **t1591** (WT-user-asset-bundle, which carries SPEC-USERASSET-DEPLOY-GUARD-001): BLOCKED-ON source for U-03, U-09, U-11, and U-12 (U-02 and U-10 moved to SPEC B). U-01 is not BLOCKED-ON t1591; it is pending Q2.
- Card **t1578** (WT-update-migration-fixes, which carries SPEC-UPDATE-MIGRATION-FIX-001): BLOCKED-ON source for U-11, U-17, U-21, and U-22 (U-31 moved to SPEC B).
- Card **t1616**: builds the local-main integration tooling. Landing of this card waits for it. Merge-ready is this card's terminal state until t1616 lands.
- SPEC-USER-ASSET-INSTALL-001 (completed, on main): P6 at spec.md lines 237-238 and REQ-002 at spec.md line 255. Amending REQ-002 is option (b) of decision Q2.
- SPEC-CLI-TUX-INIT-UPDATE-001 (completed, on main): REQ-TUXIU-004, -012 to -016, -020 to -022, -041, -054.
- SPEC-V3R6-UPDATE-NOISE-001 (implemented, on main): `--verbose` semantics, REQ-UN-005 and REQ-UN-010.
- SPEC-V3R6-UPDATE-PROGRESS-001 (implemented, on main): `tui.ProgressLine`.
- SPEC-UPDATE-MIGRATION-001 (completed, on main): REQ-UPM-030 to -032 (summary reporting, plain counts).
- SPEC-UPDATE-YAML-PRESERVE-001 (completed, on main): comment preservation (U-23).

## I. Exclusions

### Out of Scope — factory slug and inbox defects

- The slug rule in `factoryWorktreeSlug` (card t1621). See § H.
- The factory inbox error in `factory_msg_list` (card t1622). See § H.

### Out of Scope — init-side presentation

- Init banner, init success card, and init-side output are not in this SPEC's item register. Their queue overlap with t1527 is recorded in § F with disposition undecided. The decided init presentation is SPEC-CLI-TUX-INIT-UPDATE-001 Group D.

### Out of Scope — mo.ai.kr operator recovery

- The recovery steps in `update-audit.md` § C are operator actions in the mo.ai.kr project. They are not code items of this SPEC.

### Out of Scope — plugin distribution, new profile semantics, and other locales

- No plugin distribution path (voided by operator decision 2026-10-06, SPEC-UPDATE-MIGRATION-001 § H).
- No change to REQ-002 of SPEC-USER-ASSET-INSTALL-001 is made in this SPEC. An amendment is possible only under decision Q2 option (b).
- Locales es, fr, and de (the other `language.yaml` conversation-language options) are not in REQ-RC30-019 and are out of scope.

## J. Decision points

Recorded in `decision-index.md`. Summary:

- **Q1 (resolved from disk; not an operator question).** "16" in R1 names the 16 programming languages (a neutrality concern). "Four locale strings" in R2(g) names the four user-facing locales ko, en, ja, zh (a translation concern). The sets are disjoint. REQ-RC30-019 uses the four locales.
- **Q2 (pending operator decision; FOUNDER, product-level).** Profile reach: options (a), (b), and (c) are stated factually in decision-index.md. No option is selected. Blocks U-01 (AC-RC30-01 moved to SPEC B).
- **Q3 (queue overlaps; disposition undecided).** t1527, t1594, and t1612 are recorded as overlaps in § F. Blocks no item of this SPEC.
- **Q7 (recorded as the leader's choice of 60 columns, open to operator override (decision-index.md Q7)).** The narrow-terminal width threshold. Options are stated factually in decision-index.md. REQ-RC30-018 carries the value as pending.
- **Q8 (pending; decision-index.md Q8).** U-02 and U-10 moved to SPEC B with the removal family (d-20261009T235135Z-5e35); Q8 stays pending and no longer gates SPEC A.
- **B1 (CLOSED).** Decided by the leader's split (d-20261009T235135Z-5e35, option 1): SPEC A carries 21 criteria against the budget of 25 and 23 requirements. Options (a) and (b) remain in decision-index.md as the record.

## K. Requirement and AC budget

Tier L requirements: **23 of 25** in SPEC A (§ B; REQ-RC30-004 and REQ-RC30-021 moved to SPEC B with their criteria). Acceptance criteria: **21 of 25** in SPEC A (`acceptance.md`; AC-RC30-01, -02, -03, -11, -26 moved to SPEC B). Both are within the Tier L ceiling. B1 is closed by the leader's split (decision-index.md; d-20261009T235135Z-5e35).
