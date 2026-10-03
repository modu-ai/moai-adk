---
id: SPEC-INIT-SHRINK-001
title: "Shrink moai init to a thin project deploy — skills and commands ride the plugin, update scope follows the deploy mode, existing projects migrate with backup"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-04
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template, internal/cli, internal/config"
lifecycle: spec-anchored
tags: "init, slim, plugin, migration, update, codex-mirror, deploy-mode, design-card-5"
tier: L
depends_on: [SPEC-PLUGIN-MARKETPLACE-001]
---

# SPEC-INIT-SHRINK-001 — moai init shrink: thin deploy, mode-scoped update, backup migration

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-03 | manager-spec | Initial plan-phase draft (card t1438, Class C, Mods design card 5 of 6, operator instruction 2026-10-02). Built on the t1434 verdict (SPEC-PLUGIN-LOAD-SCOPE-001, completed) and the binding OD-1..OD-14 pins of SPEC-PLUGIN-MARKETPLACE-001 (t1435, in flight; operator batch acceptance 2026-10-03 via leader relay). Every mechanism premise was observed in this worktree session at base tree `3f3ebb763`; observations are P-01..P-21 with commands in `research.md`. The card text's 스킬·에이전트·명령 list is reconciled with the t1435 OD-3 pin in §1.2. Every design detail the card text and the inputs do not fix is an Open Decision (§5), never an assumption. |
| 0.1.0 | 2026-10-03 | manager-spec | Amendment (still draft, pre-audit): the OD-1..OD-8 leader verdicts (2026-10-03, codex-informed; `decision-index.md` Q1..Q8) are incorporated into the clauses and criteria they carry (§5 Settlement) — no REQ/AC count change; the RED-now cells whose assertions changed were re-observed at tree `3906f985b`; the t1435 dependency status was refreshed post-absorb. |
| 0.1.0 | 2026-10-03 | manager-spec | Amendment (still draft): cross-model-audit repair (codex receipt `rcpt-c5721bde339669a03956a37e`, which overturned plan-audit iteration 2 to FAIL 0.90 on D-7/D-8/D-9, all source-verified). D-7: the migration's removal executor is scoped to the classified sets (design §3 step 4) — the global `ManagedCleanTargets` walk has no classification gate, so the earlier scoped-FS hand-over is replaced by the classified-list executor and REQ-011/REQ-013/AC-013 name the glob-hit preservation explicitly. D-8: the OD-6 no-verification fallback takes the re-homing form — a kept mirror's entries become real directory copies rendered from the embedded template tree (P-11's copy fallback), never dangling symlinks into the undeployed `.claude/skills/**` (REQ-006/AC-006 bound; AC-016/AC-019 given the path qualifier). D-9: the install-outcome signal is specified — the post-install list-surface probe (design §2.4), a tri-state read of the tools' installed-plugin list surfaces through the REQ-017 runner seam, and REQ-015/AC-015's three arms map onto it. No REQ/AC count change (21/21); AC-006/011/013/015 RED-now cells re-observed at tree `a1f17b038`. |
| 0.1.0 | 2026-10-04 | manager-spec | Amendment (still draft): leader-ruling repair of iteration-3 findings D-11/D-12/D-13 (leader ruling 2026-10-04, mission contract 11c79e1a; the iteration-4 delta audit is authorized by the leader beyond the 3-iteration retry cap). D-12: REQ-010's foreign class is keyed to template carriage — a file the template render does not carry is foreign (every user-created skill or command, `moai-custom` included, preserved byte-for-byte), the removable classes require template carriage, and REQ-013/AC-013/AC-010/the Edge Cases bullet/design §3 step 2/plan M1 are aligned to that predicate. D-11: the install-outcome probe is a pre-execution snapshot vs post-execution diff of the acted tools' installed-plugin list surfaces — only a demonstrated this-install success (ref present post-run AND absent from the pre-snapshot) admits the `plugin` record; the residual arm is renamed `not-demonstrated`, defined only by the observable diff (the causal trigger wording — "the step failed, timed out, found no tool" — is removed), and REQ-005/REQ-015/AC-005/AC-015/design §2.1/§2.3/§2.4/§3/plan M2-M3 carry the form. D-13: design §2.4's arm mapping is scoped per surface — the migration surface (where `confirmed` alone classifies, removes, and records `plugin`) vs the init surface (where the probe governs only the project `moai` entry and the guidance block, and the file set and mode record follow the deploy path per REQ-001/REQ-009). No REQ/AC count change (21/21); the asserted behavior of AC-005 and AC-015 changed, so L-06 and L-16 were re-observed at tree `5c380a251`; AC-010/AC-013's fixture and rationale wording was tightened without an assertion change, so their pins stand. |

## 1. Background and Premise

### 1.1 What this SPEC does and does not do

`moai init` today copies the full project scaffold into every project; the plugin carries the same
skills and commands a second time (t1435 OD-1 accepted that interim double copy and named this card
as the one that ends it). This card changes the **default deploy set**: on the plugin path, init
deploys no `.claude/skills/**` and no `.claude/commands/**` files — the plugin (t1435) is the
carrier — and deploys everything else the scaffold carries today. It defines the `--no-plugin`
counterpart (a full local deploy), re-scopes `moai update` to the same deploy-mode split, and
defines the migration of existing projects (duplicate removal with backup). It **does not** build
the plugin, the marketplace, or the install step (t1435, unmerged at this tree) and does not move
agents, rules, instruction files, hooks, or output styles anywhere (§1.2, §4).

### 1.2 Card-text reconciliations (stated, per the dispatch)

1. **Agents stay.** The card text lists 스킬·에이전트·명령 (skills/agents/commands) as leaving the
   template, but t1435 OD-3 (binding pin, operator batch 2026-10-03) ships **no agents in the v1
   plugin** because plugin subagents ignore `permissionMode`, `hooks`, and `mcpServers` (t1434
   P-13/P-14), which would change the safety posture of the read-only advisors. This card's shrink
   therefore moves **skills and commands only**; `.claude/agents/**` stays in the project scaffold.
2. **Thin-deploy list vs the t1434 boundary.** The card's thin list (.moai/config, rules, hook
   registration, .mcp.json, AGENTS.md) is consistent with R09/R10/R11 (rules and instruction files
   are PROJECT-ONLY). The one contested item is the `.mcp.json` `moai` entry: the plugin carries it
   (t1435 OD-2 pin), a project copy would be inert-but-duplicated (R06), and today's decline
   semantics (`init.go:997-1004`, P-15) gate the project entry by harness. That is Open Decision
   OD-1, not an assumption.
3. **`--no-plugin` is a first-class path.** t1435 OD-5 (binding pin) defines the opt-out surface;
   users who take it get no plugin, so the shrink must keep a full local deploy available — REQ-003
   defines it; the template sources stay embedded (REQ-002).
4. **Codex mirror parity (Codex 미러 정합).** Skills move for both tools (R01 PLUGIN-OK on both);
   Codex receives plugin commands as generated skills (R03-codex). The `.agents/skills` mirror and
   its deploy-time symlink machinery (P-11) follow the deploy mode — OD-6 — and the bare-name
   reference question is measured, not assumed (REQ-008, OD-2).
5. **Existing projects never ran the t1435 install step.** The install step attaches to init and
   the three install scripts (t1435 REQ-010, REQ-018), so an old project that only runs
   `moai update` holds local copies and no plugin. The migration surface for those projects is
   Open Decision OD-4; the default makes update itself the migration (fail-open install, then
   dedupe with backup).

### 1.3 Observed premises

All rows were observed by the author in this worktree session at base tree `3f3ebb763`
(branch `WT-moai-init-slim`); the command and its deciding output are quoted; full outputs are in
`research.md` as R-nn.

| # | Observation | Command or read | Deciding output (verbatim) | Source |
|---|-------------|-----------------|----------------------------|--------|
| P-01 | Tree pin | `git rev-parse --short HEAD`; `git branch --show-current` | `3f3ebb763`; `WT-moai-init-slim` | author |
| P-02 | Template skill set and catalog tiers | `ls internal/template/templates/.claude/skills/ \| grep -c .`; `grep -c 'tier: core' internal/template/catalog.yaml`; `grep -c 'tier: optional-pack' …`; `grep -c 'tier: harness-generated' …` | `41`; `36`; `13`; `1` | author |
| P-03 | Command template set | `ls internal/template/templates/.claude/commands/moai/` | 17 entries — 15 `*.md.tmpl` plus `gtd.md` and `todo.md` | author |
| P-04 | Codex command-skill mirror template | `find internal/template/templates/.agents -type f \| wc -l` | `17` (one `SKILL.md` per `/moai` command) | author |
| P-05 | Deployer selection per harness and slim mode | read of `internal/cli/init.go:738-767` | `agentWiringGPT` → `NewCodexOnlyDeployer…`; `both` → Dual (slim unless `shouldDistributeAll`); default → Claude (slim unless `shouldDistributeAll`) | author |
| P-06 | Update runs Clean Managed Paths on every update, before Deploy | read of `internal/cli/update_template_sync.go:388-432` | step order `Backup → Validate Templates → Clean Managed Paths → Deploy Templates → Restore Settings`; the Clean step runs `archiveLegacySkills`, snapshots `InventoryManagedPaths`, then `guardFirstDestructiveStep` wraps `deploy.CleanMoaiManagedPaths(projectRoot, out, tmplFS)` | author |
| P-07 | The managed clean roots include the components this card drops | read of `internal/cli/update/deploy/deploy.go:40-87` | `ManagedCleanTargets` lists `.claude/settings.json`, `.claude/commands/moai`, `.claude/agents/moai`, `.claude/skills/moai*` (glob), `.claude/rules/moai`, `.claude/output-styles/moai`, `.claude/hooks/moai`, and `.moai/config` | author |
| P-08 | Pre-clean backup covers only files the **raw template** does not carry | read of `internal/cli/update/deploy/deploy.go:95-106` | "BEFORE each root is removed, every regular file tmplFS does not carry at the same relative path is copied into the run's pre-clean backup … Template-managed files are NOT backed up: deployment rewrites them moments later" | author |
| P-09 | The archive contract this card reuses | read of `internal/cli/update_archive.go:69-114, 267-367` | source absent → nil; archive matches → nil; drift → `ARCHIVE_DRIFT`; `--force` routes drift to `.moai/archive/skills/v2.16-drift-<UTC-ISO8601>/<id>/`; symlink entries skipped (REQ-SEC-003) | author |
| P-10 | The legacy-list guard precedent | read of `internal/cli/update_archive.go:41-44` | "TestLegacySkillIDsNotEmbedded (update_archive_guard_test.go) asserts this list stays disjoint from the embedded template skill set" | author |
| P-11 | Skill-mirror mechanism and its clean-path ownership | read of `internal/template/skill_mirror.go:1-33` | ".agents/skills/<name> as a relative symlink to ../../.claude/skills/<name>, falling back to a real directory copy"; "Lifecycle management of the mirror … belongs to the clean path" | author |
| P-12 | Codex-only deploy re-homes skills, mirror off | read of `internal/cli/init.go:739-746` | "a codex-only selection reroutes the WHOLE deployment through the harnessFS wrapper — claude-only surfaces hidden, the skill catalog re-homed to .agents/skills as real directories, skill mirror off" | author |
| P-13 | The manifest is the per-file record | read of `internal/manifest/types.go:92`, `internal/manifest/manifest.go:14-18` | "Manifest represents the file tracking manifest stored at .moai/manifest.json"; `Load` reads `{projectRoot}/.moai/manifest.json` | author |
| P-14 | Re-init redirects to update | read of `internal/cli/init.go:838-843` | "did you mean 'moai update' (refresh templates in place)? Re-run with --force only to reinitialize from scratch" | author |
| P-15 | MCP decline semantics | read of `internal/cli/init.go:997-1004` | `mcpDeclined := !opts.MCPProvision`; `case agentWiringGPT: mcpDeclined = true`; `case agentWiringBoth: mcpDeclined = false`; non-interactive leaves `MCPProvision` false | author |
| P-16 | The project `.mcp.json` template carries two servers plus startup shaping | `cat internal/template/templates/.mcp.json` | `mcpServers` holds `moai` (`command: moai, args: [mcp-server]`) and `context7` (`npx -y @upstash/context7-mcp@latest`), plus `staggeredStartup` | author |
| P-17 | Bare-name reference surface in scaffold instructions | `grep -rn 'Skill("' internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl \| wc -l`; `grep -rln 'Skill("moai' internal/template/templates/.claude/rules/ \| wc -l` | `5`; `7` rule files | author |
| P-18 | The t1435 install step is absent at this tree | `grep -c 'plugin install' internal/cli/init.go`; t1435 SPEC frontmatter | `0`; SPEC-PLUGIN-MARKETPLACE-001 `status: in-progress` | author |
| P-19 | User-owned namespace protection | read of `internal/cli/update/plan/plan.go:200-235` | user direct-added skills (any name except `moai` / `moai-*` prefix) classify user-owned; `.claude/agents/harness/` is user-owned | author |
| P-20 | Snapshots are the next update's merge BASE | read of `internal/cli/update_template_sync.go:462-480` | `StageDeployedSettingsSnapshot`, `StageDeployedMCPSnapshot`, `writeTemplateSnapshotBestEffort` run inside Deploy Templates, before Restore Settings | author |
| P-21 | Managed-path classification rule | read of `internal/cli/update/plan/plan.go:236-285` | `IsMoaiManaged`: `.moai/config/**` and `.moai/evolution/**` managed; `skills/rules/commands/output-styles/hooks` + `moai-` prefix managed; agents in `core/expert/meta` or `moai-` prefix managed | author |

### 1.4 Interactions detected before any code exists

1. **Silent-loss hazard (the load-bearing one).** After the shrink, the deployed copies of dropped
   components are removed by the Clean step (P-06/P-07), but `backupThenRemove` backs up only files
   the **raw template** does not carry (P-08) — and the raw template keeps skills/commands as the
   plugin derivation source (REQ-002). A user-modified deployed skill would therefore be deleted
   with **no backup**. REQ-012 and the milestone that lands it exist to close exactly this hole;
   the acceptance carries a negative control that shows the pre-fix behavior losing the file. The
   same machinery has the mirror-image hazard: the global walk has **no classification gate** —
   `CleanMoaiManagedPaths` removes every `ManagedCleanTargets` match (the `.claude/skills/moai*`
   glob included) with no user-owned skip, and its `tmplFS` argument scopes only the backup — so a
   foreign `moai-custom` skill and symlink entries would be removed too; REQ-010/011/013 and the
   classified-set removal executor (design §3 step 4) close that side.
2. **Existing projects have no plugin.** The install step never ran for them (§1.2-5); a naive
   dedupe on update would leave a plugin-mode project with no skills at all. OD-4 decides the
   surface; the default runs the install step fail-open before the dedupe.
3. **Bare-name resolution is unmeasured.** Command and skill bodies and scaffold instructions call
   `Skill("moai…")` bare (P-17; t1435 §1.5-3 named this load-bearing at this card). A plugin skill
   is namespaced (`moai:<name>`). Whether a bare name resolves to the namespaced component when no
   scaffold copy exists is UNMEASURED; REQ-008 makes the measurement a gate before the default-path
   flip merges, and OD-2 fixes the fallback if it does not resolve.
4. **MCP behavior changes shape on the default path.** Today non-interactive init provisions no
   project moai entry (P-15) and Codex MCP comes from `config.toml` wiring. With the plugin as the
   carrier the default path yields plugin-borne MCP; the explicit-decline reading is re-stated by
   OD-1 rather than silently changed.
5. **Version-compare skip.** Update skips the whole sync when the project's template version equals
   the package version (`update_template_sync.go:180-189`). A post-shrink binary bumps the version,
   so the migration runs; the skip is named here so the migration is not accidentally scheduled
   behind a gate that can hold it.

## 2. Requirements (GEARS)

Each requirement has exactly one acceptance criterion with the same number in `acceptance.md`. At
authoring, clauses carried a `default pending OD-n` marker for the behavior taken when no operator
verdict existed for that decision in §5. On 2026-10-03 the leader settled all eight decisions
(codex-informed; the rulings are verbatim on the `Operator verdict:` lines of `decision-index.md`
Q1..Q8), and every marked clause below now states its settled branch — see §5 Settlement. No
requirement or criterion count changed: the verdicts' conditions folded into the existing rows.

- REQ-001: (Ubiquitous) The default init deploy — the deploy set used when the plugin opt-out is not set — shall carry no `.claude/skills/**` and no `.claude/commands/**` files, and shall otherwise carry the same components the default deploy carries today (`.claude/agents/**`, `.claude/rules/**`, the hooks registration in `settings.json.tmpl`, `.claude/output-styles/**`, `.claude/workflows/**`, `.claude/loop.md`, `CLAUDE.md`, `AGENTS.md`, `.moai/**`, `.gitignore`, `.claudeignore`, `.worktreeinclude`, `.git_hooks/**`, `.github/**`, `.codex/**`), with the Codex mirror of REQ-006 and the `.mcp.json` moai entry of REQ-005 decided by their own clauses (OD-1 settled (c) and OD-6 settled (a) with its Codex-execution condition — see those clauses).
- REQ-002: (Ubiquitous) The embedded template tree shall retain every skill and command source it carries today — they remain the derivation source of the plugin payload (t1435 REQ-004/005) and the deploy source of the `--no-plugin` and `--all` paths — so that `internal/template/catalog.yaml`, the slim-mode catalog routing (P-05), and the `commands-emit`/`agents-emit` golden machinery keep passing unchanged as data.
- REQ-003: (Event-driven) When the plugin opt-out is set (t1435 OD-5 pin: the `--no-plugin` flag of `moai init`, or `MOAI_SKIP_PLUGIN_INSTALL=1`/`true`), init shall deploy the full local payload — the same file set today's default deploys, including `.claude/skills/**`, `.claude/commands/**`, the project `.mcp.json` moai entry, and the Codex mirror — so that a user who declines the plugin loses no capability.
- REQ-004: (Event-driven) When the default init deploy completes and the plugin install step was skipped, or ran and did not exit 0, init shall print one guidance block naming both recourses (re-running with `--no-plugin` for a full local deploy; the t1435 manual-install commands) and shall leave the exit status of the init unchanged.
- REQ-005: (Ubiquitous) On the default init deploy the project `.mcp.json` shall keep the `context7` server and the `staggeredStartup` block of the template (P-16) and the `moai` entry shall follow OD-1, settled (c) 2026-10-03, read through the probe of design §2.4: the default deploy writes the project `moai` entry only when the probe's pre/post diff does not demonstrate this-install success (`not-demonstrated`, or `opted-out` on the local path) — the entry is the fallback carrier for a plugin-less Claude user — and when the probe reads `confirmed` no project `moai` entry is written (the plugin is the sole carrier); an explicit decline (`--llm gpt` project-entry decline, non-interactive absence) shall keep its present meaning for the project file, and on the `--no-plugin` path the moai entry shall be written as today.
- REQ-006: (Ubiquitous) The Codex command-skill mirror shall follow the deploy mode (OD-6 settled (a) 2026-10-03, with its condition: plugin mode deploys no `.agents/skills` mirror and deploys no symlink or copy under it — but only after Codex is verified to actually execute plugin-borne skills, and where that verification cannot be produced the local mirror stays deployed on every path with its entries re-homed to real directory copies rendered from the embedded template tree (P-11's existing copy fallback) — never left as the symlinks into `../../.claude/skills/<name>` that would dangle once the plugin path stops deploying `.claude/skills/**`; `--no-plugin` and every local deploy, codex-only re-homing included (P-12), deploy the mirror exactly as today); a mirror entry left behind by a mode transition shall be removed by the migration of REQ-010..014, never left dangling, and that removal is under the same verification condition.
- REQ-007: (Ubiquitous) The `--all` flag shall keep naming a local full deploy (OD-7 settled (a) 2026-10-03: `--all` deploys every catalog tier locally and carries the same skills-and-commands payload as `--no-plugin` plus the wider tier, superseding slim-mode hiding for that run), so that optional-pack catalog entries (P-02: 13 `optional-pack` entries), which the core-only plugin does not carry (t1435 OD-8 pin), remain reachable.
- REQ-008: (Event-driven) When the default init deploy or the migration of an existing project is prepared for release, the bare-name resolution measurement shall have run and its verdict shall have been recorded in `progress.md` §E.2 before the flip merges: the measurement drives the real tool runtimes under scratch config homes (the t1434 `--plugin-dir` route) with no scaffold copies present and answers three questions — does `Skill("<bare-skill-name>")` resolve to the namespaced plugin skill, does a plugin command body's bare `Skill("moai")` resolve, and what names Codex lists for plugin-borne components — and (OD-2 settled (a) 2026-10-03, with its condition) BOTH the scaffold-shipped instruction files (`CLAUDE.md`, `AGENTS.md.tmpl`, the seven referencing rule files of P-17) AND the plugin command bodies shall pass an actual name-resolution verification test, and after the flip shall reference only the names that test proves resolvable in each deployment mode — the mode-aware reference rewrite ships only what the test proves resolvable, per mode.
- REQ-009: (Ubiquitous) Init shall persist the resolved deploy mode of the project — `plugin` when the default deploy ran, `local` when the opt-out or `--all` path ran — in the config key OD-5 names (settled (a) 2026-10-03: `deployment_mode` in `.moai/config/sections/llm.yaml`, read and written through the same config seam as `llm.harness`), on every init run including re-init, so that update never infers the mode; and the recorded value shall survive update's redeploy/restore process unchanged (OD-5 settled condition: the Clean step's `.moai/config` root wipe (P-07) must not cost the key — a re-read after the full update cycle returns the recorded value byte-identical).
- REQ-010: (Event-driven) When update runs the migration, it shall classify every deployed file under the dropped component roots (`.claude/skills/**`, `.claude/commands/**`, and the mirror of REQ-006) into exactly three classes before removing anything — template-identical: the file is carried by the template render for this project's context (a template source exists at its relative path), its content equals that render, and its manifest record (P-13) shows it managed and not user-modified; modified: template-carried, and its content differs from that render, or its manifest record is absent or stale for it; foreign: NOT carried by the template render for this project's context — every user-created skill or command is foreign on this ground alone, a `moai-custom` name included, whatever the managed-name rules (P-19/P-21) match — and shall print the three counts; a symlink is never classified for removal — foreign files and symlinks are preserved untouched byte-for-byte (REQ-013 carries the prohibition; OD-3 settled condition, 2026-10-03). The removable classes require template carriage; managed-name matching alone never routes a file into a removal class.
- REQ-011: (Event-driven) When the migration holds template-identical classified files, it shall remove them from the project tree and report the removed count — through the classified-set-scoped removal executor: the executor's removal list is exactly the classified sets (the identical set plus the modified set once REQ-012's batch archive has fully succeeded), and the Clean step's global managed-roots walk (P-06/P-07) does not run over the dropped roots in a migration run — without archiving copies identical to the template render (OD-3 settled (a) 2026-10-03; the REQ-010 routing sends absent-or-stale-manifest files to the modified class — archive-then-remove — never into this removal set).
- REQ-012: (Event-driven) When the migration holds modified classified files, it shall archive each one — skill directories through the `archiveSkill` contract (P-09: idempotent, drift-checked, symlink-refusing) and standalone command or mirror files into an archive directory of the same layout — before any removal; removal shall run only after EVERY archive in the batch has succeeded (OD-3 settled condition, 2026-10-03) — any single archive-write failure aborts the migration before anything is removed — so that a modified deployed component is never deleted without its archived copy; the raw template's continued carriage of the source (REQ-002) shall not exempt these files from backup the way P-08 exempts template-carried files today.
- REQ-013: (Unwanted) The migration shall not remove, modify, or archive any file outside the classified sets of REQ-010 — including any file a managed-root glob would match: a user-created skill at a name such as `moai-custom` under `.claude/skills/` is foreign under REQ-010 (the template render does not carry it) and is preserved byte-unchanged even though the Clean step's global walk (P-06/P-07) would remove it — the migration's removal list never contains it — and shall not follow or dereference a symlink while classifying, archiving, or removing (the REQ-SEC-003 rule of P-09; mirror-entry handling is REQ-006's own clause — removed as link entries or re-homed per its fallback form, never dereferenced).
- REQ-014: (Event-driven) When update runs the migration a second time on a project it already migrated, it shall remove nothing further, archive nothing further, and report zero counts.
- REQ-015: (Event-driven) When update runs on a project with no deploy-mode record — every project initialized before this SPEC's init — it shall run the migration path (OD-4 settled (a, amended) 2026-10-03: update performs the migration — run the t1435 plugin install step first, fail-open and under the same opt-out, and read its outcome through the post-install list-surface probe of design §2.4, which snapshots the installed-plugin list surface of every tool the step will act on BEFORE the step runs and diffs that pre-execution snapshot against the post-execution state; probe `confirmed` — the diff demonstrates this-install success: the plugin ref present in the post-execution surface AND absent from the pre-execution snapshot for every acted tool — then classify, then remove and archive per REQ-011/012, then write the mode record `plugin`; probe `not-demonstrated` — every other outcome, decided only by the observable diff and never by a cause inside the step — the migration MUST NOT dedupe or remove anything, MUST keep `deployment_mode: local`, and no path records `plugin`; probe `opted-out` — the opt-out was set, decided before the step runs — the full local payload deploys through update's normal local path, the record is written `local`, and the migration removes nothing).
- REQ-016: (Ubiquitous) The update template-sync deployer shall be selected by the project's mode record (P-05 construction, OD-5 key): a `plugin`-mode project deploys the thin set of REQ-001, a `local`-mode project deploys the full set of REQ-003, and the outcome accounting (`managedRedeployCount`, `restoredSet`, `preCleanFiles`, P-06) shall report what this run actually deployed and removed — a run that re-deployed no dropped component shall not count one; and the mode record the selection reads shall survive the run's Clean/redeploy/restore cycle unchanged (OD-5 settled condition, 2026-10-03: REQ-018's no-flip rule holds through the mechanics — the `.moai/config` Clean root (P-07) does not cost the recorded key).
- REQ-017: (Event-driven) When update runs on a `local`-mode project, it shall keep today's full merge scope (deploy, 3-way merge, snapshot semantics of P-20 unchanged), so that the opt-out and `--all` populations see no update-scope regression.
- REQ-018: (Event-driven) When a user wants to switch a project's deploy mode after initialization, the only supported surface shall be a re-run of `moai init` (`--force` re-init from scratch, or the documented init re-entry), and update shall print that guidance and change no mode record by itself.
- REQ-019: (Unwanted) Update shall not resurrect a dropped component on a `plugin`-mode project: no run of update, with or without `--force`, shall re-create `.claude/skills/**`, `.claude/commands/**`, or a mirror entry the thin deploy does not carry.
- REQ-020: (State-driven) While the verification of this SPEC runs — its acceptance commands and its harness scripts — it shall not reach a real Claude or Codex profile, a real `$HOME/.claude/settings.json`, or the network, and shall report a failure when it does: isolation uses scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `MOAI_HOME` (never a `HOME` assignment, which the worktree guard refuses and no script may set), the default external-command runner refuses to exec under a test binary (the t1435 REQ-017 seam, reused), the one script that runs the real tool runtimes for the REQ-008 measurement applies a live-enumerated environment scrub and a before/after protected-set hash with directory entries, and no harness case runs the real binary's `moai init` (the t1435 OD-14 verdict, inherited).
- REQ-021: (Ubiquitous) The user-facing surfaces that describe the deploy shall describe the shrink: the `init` flag help and success card, the README and docs-site init pages, and every guarded test that asserts a deployed file set (the settings snapshot, template-count, update dry-run preview, and e2e journey assertions) shall be updated in the same change set that flips the default, so that no guard asserts the pre-shrink file set after the flip lands; and per OD-8 settled (a) 2026-10-03 this card's report deliverable states the t1466 handover scope (instruction-file consistency beyond REQ-008 name resolution defers to t1466 — no broader sweep joins this SPEC).

## 3. Constraints

- Tier L: 21 requirements and 21 acceptance criteria (Tier L ceilings 25/25), with `design.md` and
  `research.md`. The hand-authored change set is on the order of 40 files across
  `internal/template`, `internal/cli` (init + update + a new migration unit and the install-outcome
  probe of design §2.4), `internal/config`, one new harness script, docs, and the guarded test
  surfaces — see `plan.md` §3 for the per-milestone list. Generated files are not counted.
- Template-First: every behavior change lands in `internal/template/templates/**` or the Go code
  that deploys it; no file is hand-edited in a deployed tree. The plugin payload itself is
  t1435's generated tree and is never edited here.
- Run-phase dependency (leader-designated): this card's run phase is sequenced after
  SPEC-PLUGIN-MARKETPLACE-001 lands on `develop`. Its REQ-015/REQ-016 (install step attached to
  init), REQ-017 (test-binary runner refusal), REQ-018 (installer verb and opt-out) and its
  `--no-plugin` flag are consumed here by reference; at the authoring pin 3f3ebb763 they did not
  exist (P-18). t1435 is `completed` in this tree since the 2026-10-03 absorb of develop 6770c714f;
  the consumed surfaces are verified by name at run-phase pre-flight, and any contradiction found
  there is a blocker report, not a local improvisation.
- The worktree guard refuses a `HOME=` prefix and refuses `pwsh` (t1435 P-36), and repository
  doctrine says moving a command into a script file is not a way round it. Verification isolation
  uses the explicit seams only: scratch `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, `MOAI_HOME`, working
  directory, and the protected-set hash of REQ-020.
- No acceptance command makes a network call except the REQ-008 measurement script's real-runtime
  plugin load (local fixture via `--plugin-dir`, no marketplace network), and no acceptance command
  writes outside the scratch homes and the tree it is verifying.
- Acceptance commands use the plain single-invocation forms measured in this worktree session
  (anchored `-run '^Name$'`, no `HOME=`, no `pwsh`, no pipes-and-redirect bundles); exit codes are
  read from the tool result.
- No time estimates; ordering is by priority and milestone sequence.

## 4. Non-goals and Out of Scope

### Out of Scope — Agents, rules, instruction files, hooks, output styles

- No agent ships in or leaves the scaffold this card (t1435 OD-3 pin; §1.2-1). Always-loaded and
  path-scoped rules, `CLAUDE.md` and `AGENTS.md` stay in the project scaffold (R09/R10/R11). Hook
  registrations and output styles stay in the scaffold (t1435 OD-2 pin; R05/R12 unobserved for
  plugin activation). Other carriers for the rules payload (a skill, a SessionStart hook returning
  context) are UNMEASURED and are not used here.

### Out of Scope — The plugin, marketplace, and install step themselves

- The marketplace manifests, the plugin payload derivation, the install verb, and the doctor
  check belong to SPEC-PLUGIN-MARKETPLACE-001. This card consumes them by reference and builds
  none of them.

### Out of Scope — User-global install mode removal

- The t1434 non-goal stands: nothing here touches user-global install mode history.

### Out of Scope — moai update refreshing the plugin

- Update does not refresh, reinstall, or version-check the plugin (t1435's doctor check reports
  and names the remedy). Update's plugin contact is limited to the OD-4 migration install step,
  fail-open.

### Out of Scope — Deployed-template instruction drift beyond name resolution

- The broader instruction-drift cleanup of deployed templates is card t1466 (leader-issued
  2026-10-03, per t1418 verdict H1). This card ensures names referenced by the shipped instruction
  files resolve in each deployment mode (REQ-008) and defers the rest to t1466's approach (OD-8,
  settled (a) 2026-10-03).

### Out of Scope — Branch-flow narratives

- The github-flow switch is card t1453; branch-flow narratives in docs are out of scope here.

### Out of Scope — Real-profile writes and release mechanics

- No tag, push, release, pull request, GitHub issue, or write to any real Claude or Codex profile
  is performed by this card. PowerShell-specific behavior is not executed locally (no Windows
  host; the guard refuses `pwsh`), matching the t1435 gap treatment.

## 5. Open Decisions

`interview.decision_gate: on` and `interview.recommendation_mode: pull`. Every row below is a
decision surfaced while assembling card t1438 that no committed artifact in the authority register
settles as written; none carries a recommendation. Rows are mirrored in `decision-index.md`
(Q-n = OD-n), whose verdict lines the Kickoff gate reads. Per the leader's dispatch of
2026-10-03, operator-gate-level decisions for this card are set by the leader on audit evidence;
no verdict line is invented here. All eight rows were settled by the leader on 2026-10-03 — see
Settlement below.

| OD | Question | Options | Default if unanswered | Who decides |
|----|----------|---------|-----------------------|-------------|
| OD-1 | On the default (plugin) path, what does init write into the project `.mcp.json` for the `moai` entry? The plugin carries the entry (t1435 OD-2 pin); a project copy is inert-but-duplicated (R06: the project copy wins, the plugin copy is suppressed); today non-interactive init writes no project moai entry and `--llm gpt` declines it (P-15). The Codex side keeps its `config.toml` wiring on every path, so `gpt` plus plugin holds two registrations whose combined effect is UNMEASURED. | (a) default path writes no project `moai` entry — plugin is the sole carrier; context7 and staggeredStartup always written; `--no-plugin` writes it as today; (b) keep writing the project `moai` entry on every path (inert duplication accepted, R06 suppression applies); (c) default path writes the entry only when the plugin install step failed or was skipped (conditional fallback) | (a) | operator |
| OD-2 | If the REQ-008 measurement shows bare names do NOT resolve to namespaced plugin components, what ships? The measurement itself is a fact (REQ-008); this row fixes the fallback. | (a) mode-aware reference rewrite: template context grows a deploy-mode variable; scaffold instructions render bare names in `local` mode and namespaced names in `plugin` mode; plugin command bodies are rewritten to the resolvable form; (b) partial shrink: the `moai` router skill and the command set stay in the scaffold, only the remaining skills shrink; (c) hold: the default-path flip and the migration wait; this card ships the mode record, `--no-plugin`, and migration machinery only | (a) | operator, informed by the REQ-008 verdict |
| OD-3 | What does the migration remove and what does it back up? P-08 shows today's pre-clean backup exempts template-carried files — exactly the copies this card removes. | (a) remove template-identical without archive (recoverable from the plugin/template), archive-then-remove modified, leave foreign untouched (REQ-010..012 as written); (b) archive-then-remove everything classified for removal, identical included (uniform backup, larger archives); (c) two-step: first post-shrink update reports the classification only, removal lands in the next release | (a) | operator |
| OD-4 | What is the migration surface for a project that has no mode record (every pre-shrink project), given that the t1435 install step has never run for it (§1.2-5)? | (a) update performs the migration: run the install step fail-open under the opt-out, then classify and dedupe per OD-3, then write the record; (b) update stays offline: print migration guidance, keep the full local deploy, write the record `local`; the plugin switch is a later init re-run; (c) a dedicated `moai migrate` verb owns install-plus-dedupe, update only prints its name | (a) | operator |
| OD-5 | Where does the deploy-mode record live, so update never infers it? | (a) `deployment_mode: plugin\|local` in `.moai/config/sections/llm.yaml`, through the same config seam as `llm.harness` (the harness persistence precedent, `init.go:947`); (b) a dedicated `.moai/config/sections/deployment.yaml` section file; (c) no config key — infer from `.moai/manifest.json` contents each run | (a) | operator |
| OD-6 | What happens to the Codex command-skill mirror (`.agents/skills`, P-04/P-11/P-12) after the shrink? Codex lists plugin skills (R01-codex PLUGIN-OK) and receives plugin commands as generated skills (R03-codex), but activation observations are render-level only (t1434 G-f). | (a) mirror follows the mode: plugin mode deploys no mirror and the migration removes mirror entries it left behind; `local` mode deploys it exactly as today, codex-only re-homing included; (b) mirror always deployed on every path (it is small and Codex-local); (c) mirror retired entirely on every path, including `local` | (a) | operator |
| OD-7 | What does `--all` mean after the shrink? Today it bypasses slim mode (P-05); the plugin carries the core tier only (t1435 OD-8 pin), so 13 optional-pack entries have no plugin home. | (a) `--all` = a local full deploy: every catalog tier locally, skills and commands included, the `--no-plugin` payload plus the wider tier; (b) `--all` widens only the tier of a thin deploy (plugin remains the carrier; optional packs land locally, core rides the plugin); (c) deprecate `--all` with guidance to `--no-plugin` | (a) | operator |
| OD-8 | Who owns instruction-file consistency beyond name resolution? This card's REQ-008 fixes resolvability of the names the shipped instructions reference; the broader deployed-template instruction drift is card t1466 (leader-issued, queue text — not a committed artifact). | (a) defer the broader sweep to t1466's approach; this card owns only REQ-008 resolvability and states the interface in its report; (b) this card performs the full reference sweep of CLAUDE.md/AGENTS.md/rules now, superseding t1466's scope for the moved components; (c) no consistency work beyond what the flip mechanically forces | (a) | operator |

### Marker table — every default-bound clause, and what the other verdict changes

Every requirement named here carries the marker in its own text, and its criterion (the `AC-nnn`
with the same number) carries the marker and an `Alternate` line.

| OD | Clauses carrying the marker | If the verdict differs |
|----|-----------------------------|------------------------|
| OD-1 | REQ-001, REQ-005 | (b) REQ-005 inverts (the project entry is always written) and AC-005's plugin-mode expectation flips; (c) REQ-005 gains the conditional write and AC-005 a third arm |
| OD-2 | REQ-008 | (b) the shrink keeps the router skill and commands in the scaffold — REQ-001's command exclusion narrows to the non-command skills, and REQ-021's docs describe the partial shape; (c) REQ-001's flip is void for this card and REQ-015's migration becomes report-only (OD-3 (c) semantics) |
| OD-3 | REQ-011, REQ-012 | (b) REQ-011 archives identical copies too and AC-011's no-archive expectation flips; (c) REQ-011 and REQ-012's removal halves are deferred one release and their criteria assert report-only output |
| OD-4 | REQ-015 | (b) the install-step call leaves REQ-015, the record is written `local`, and AC-015's plugin arm becomes the guidance arm; (c) REQ-015 names `moai migrate` and update's part is guidance only |
| OD-5 | REQ-009, REQ-016 | (b) the key moves files; (c) REQ-009 is void and REQ-016's selection reads the inference; every criterion naming the key retargets |
| OD-6 | REQ-001, REQ-006 | (b) REQ-006 inverts (mirror always deployed); (c) REQ-006's local arm is void and the migration removes mirrors on local projects too |
| OD-7 | REQ-007 | (b) REQ-007's payload narrows to the optional-pack entries; (c) REQ-007 becomes a deprecation notice and the `--all` criteria assert the notice |
| OD-8 | REQ-008 (scope sentence), REQ-021 | (b) the full sweep joins this card's milestones and t1466's moved-component scope is absorbed; (c) REQ-008's instruction-file sentence is void, resolvability applies to component bodies only |

### Settlement (leader rulings, 2026-10-03, codex-informed)

All eight rows were settled on 2026-10-03; the rulings are recorded verbatim on the `Operator
verdict:` line of each `decision-index.md` row (Q-n = OD-n) and are the authoritative content.
Branches taken and their conditions, now folded into the clauses of §2:

| OD | Verdict | Where it landed |
|----|---------|-----------------|
| OD-1 | (c) | REQ-005's conditional write: the project `moai` entry is written only when the probe of design §2.4 does not demonstrate this-install success (the diff-keyed `not-demonstrated` arm); probe `confirmed` install → no project entry (the plugin is the sole carrier); `--no-plugin` unchanged |
| OD-2 | (a) + condition | REQ-008: BOTH the scaffold references AND the plugin command bodies pass an actual name-resolution verification test; the mode-aware rewrite ships only what the test proves resolvable, per mode |
| OD-3 | (a) + three conditions | REQ-010..012: an absent-or-stale manifest entry classifies modified (archive-then-remove); removal runs only after EVERY archive in the batch succeeded; foreign files AND symlinks preserved untouched (REQ-013) |
| OD-4 | (a, amended) | REQ-015: update performs the migration; when the install step fails — no dedupe, no removal, `deployment_mode: local` kept, and no path records `plugin` (the "failed" arm is read through the post-install list-surface probe, design §2.4 — its diff-keyed `not-demonstrated` arm, defined only by the pre/post snapshot diff) |
| OD-5 | (a) + condition | REQ-009/REQ-016: the recorded `deployment_mode` value survives update's redeploy/restore process |
| OD-6 | (a) + condition | REQ-006: plugin-mode mirror removal only after Codex is verified to actually execute skills; where that verification cannot be produced, the local mirror stays — kept coherent per design §2.1 by re-homing its entries to real copies (never dangling symlinks into the undeployed `.claude/skills/**`) |
| OD-7 | (a) | REQ-007 stands as written — `--all` is the local full deploy |
| OD-8 | (a) | REQ-021: the report deliverable states the t1466 handover scope; no broader sweep joins this SPEC |

The marker table above remains the record of what each non-taken branch would have changed.

## 6. Risks

| # | Risk | Observed or inferred | Handling |
|---|------|----------------------|----------|
| RK-1 | Silent loss of a user-modified deployed skill or command: post-shrink removal runs through machinery whose backup rule exempts template-carried files (P-08), and the raw template keeps carrying the sources (REQ-002) | measured mechanism (P-06..P-08) | REQ-010/012 classify against the deploy-mode render and archive before removal; AC-012 carries a negative control that shows the unguarded path losing the file |
| RK-2 | Bare-name references break after the flip: scaffold instructions and component bodies call `Skill("moai…")` bare (P-17; t1435 §1.5-3); plugin components are namespaced | UNMEASURED by design — this card measures it | REQ-008 gate before the flip merges; OD-2 fallback; the measurement harness is hermetic (REQ-020) |
| RK-3 | A `plugin`-record project whose install the probe does not demonstrate has no skills at all | §1.4-2 | REQ-004 guidance names both recourses; OD-1 option (c) is the same posture for MCP |
| RK-4 | Existing projects have no plugin when the migration reaches them | §1.2-5, P-18 | OD-4; the default runs the install step fail-open before dedupe, under the opt-out |
| RK-5 | MCP behavior change on the default path for non-interactive and `gpt` users; Codex dual registration effect UNMEASURED | P-15, R06 | OD-1 decides and records; the `gpt` + plugin double registration is named in OD-1 and left to the doctor/report surface of t1435, not silently accepted here |
| RK-6 | Codex parity rests on render-level observations; plugin-borne Codex skills are listed but not activation-proven (t1434 G-f) | t1434 verdict conditions | OD-6 keeps the local mirror available; the REQ-008 measurement includes the Codex naming question |
| RK-7 | Pre-shrink update skips the sync on version equality, so a migration could be skipped | P read (`update_template_sync.go:180-189`) | the post-shrink binary bumps the version; REQ-015's trigger names the version-bump moment and AC-014 pins idempotence so a repeat run is a no-op |
| RK-8 | Locale: plugin commands carry English description text only (t1435 OD-11 pin); `local`-mode users keep localized scaffold copies; the two populations see different text | t1435 P-32 | stated in docs (REQ-021); no code remedy in this card |
| RK-9 | Classification against a stale or absent manifest misreads modified files as identical | P-13; the t1275 note that update did not always persist the manifest | REQ-010 routes absent/stale records to the modified class for template-carried files (conservative: archive-then-remove); a file the template does not carry is foreign regardless of its manifest state |
| RK-10 | Mode flipping through update by accident (a run re-deploying the other mode's payload) | inferred from the deployer selection seam (P-05) | REQ-016 selects by the record; REQ-018 makes init the only switch; REQ-019 forbids resurrection |
| RK-11 | Update output/golden drift: dry-run preview, outcome accounting, and tui surfaces assert shapes this card changes | P-06 accounting names | REQ-016's accounting clause + REQ-021's guarded-surface sweep in the same change set |
| RK-12 | The migration reaches a real profile through moai's own re-entry or a test | t1435 P-20 class (32 `runInit(` test sites) | REQ-020 reuses the t1435 runner-refusal seam; automated callers keep the opt-out (t1435 REQ-016 pin) |

## 7. Dependencies and Prior Art

- `depends_on`: SPEC-PLUGIN-MARKETPLACE-001 (status `completed` in this tree since the 2026-10-03
  absorb of develop 6770c714f; it read `in-progress` at the authoring pin 3f3ebb763 — P-18.
  Run-phase sequencing after its landing is leader-designated). Its OD-1..OD-14 verdicts (operator batch 2026-10-03, via leader
  relay) are binding pins for this card: OD-1 (install on by default until this card ends it),
  OD-2 (MCP entry only), OD-3 (no agents), OD-5 (opt-out surface), OD-8 (core tier), OD-9
  (harness gating), OD-11 (English-only payload), OD-14 (no real-binary init harness case).
- Prior art, read only: SPEC-PLUGIN-LOAD-SCOPE-001 (completed; its `progress.md §E.2` verdict
  table is the placement evidence R01-R14); the legacy-skill archive contract
  (`internal/cli/update_archive.go`, P-09/P-10); the managed-path clean machinery
  (`internal/cli/update/deploy/deploy.go`, P-06..P-08); the skill mirror
  (`internal/template/skill_mirror.go`, P-11); the harness deployer family (P-05, P-12).
- Related cards, referenced not absorbed: t1466 (deployed-template instruction drift; OD-8
  interface), t1453 (github-flow switch; narratives out of scope), t1436/t1437 (sibling Mods
  design cards).
