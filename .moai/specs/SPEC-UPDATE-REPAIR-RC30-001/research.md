# SPEC-UPDATE-REPAIR-RC30-001 — Research and verification appendix

Five sections, per `verification-claim-integrity.md` § 3: Claim, Evidence, Baseline-attribution, Gaps, Residual-risk. Tree for every RED cell in this appendix: `8fa0dc24b` (code paths equal main `2aab5f797b75983e132af451da68f69e3426557b`), unless a row names another ref. Revision 0.1.4: the evidence blocks below are the revision 0.1.3 measurements at `8fa0dc24b`, kept as the historical record; the RED ledger is re-pinned to `6f95917a1` in acceptance.md, and its drift is recorded there.

## 1. Claim

| # | Claim | Status |
|---|-------|--------|
| C1 | The RED cells are measured on tree `8fa0dc24b`, whose code paths equal main tip `2aab5f797b75983e132af451da68f69e3426557b`. | Observed |
| C2 | The installed binary (`4f8aba061`) descends from the tree, so no claim here relies on it. | Observed |
| C3 | `WT-user-asset-bundle` (t1591, tip `df559b729`) changes 81 files and `WT-update-migration-fixes` (t1578, tip `40ab940f4`) changes 13 files, relative to main (re-measured 2026-10-09 19:15Z; E-OVERLAP). | Observed |
| C4 | Branch tips, re-measured 2026-10-09 19:15Z with `git rev-parse --short`: t1591 `df559b729`, t1578 `40ab940f4`. The earlier tips `01b519c21` (19:06Z), `cae7e28fe`, and `317ed31b6` are superseded. | Observed |
| C5 | SPEC-USER-ASSET-INSTALL-001 (completed, on main) declares P6 at spec.md:237-238 and REQ-002 at spec.md:255. The branch copy is identical. | Observed |
| C6 | The branch carries SPEC-USERASSET-DEPLOY-GUARD-001, not SPEC-USER-ASSET-INSTALL-001. The lane brief's attribution is wrong. | Observed |
| C7 | The migration function is `migrateProjectCommonAssets` in `internal/cli/migrate_project_assets.go`. The audit's `internal/cli/update/...` path is wrong. | Observed |
| C8 | Install roots are the four `$HOME` roots at `internal/userassets/paths.go:36-43`. | Observed |
| C9 | The "Current version" line is `internal/cli/update.go:200`. The literal `0 user customizations modified` is `internal/cli/update_archive.go:367`. | Observed |
| C10 | `ConfigManager.Save` is `internal/config/manager.go:225`. `LLMConfig` carries `TeamMode` (`types.go:284`) and `Fable` (`:383`, `:398`). Comment at `:389` says alias fields were removed. | Observed |
| C11 | Profile sync reads the profile from the environment (`update.go:652-657`). | Observed |
| C12 | The eight rc.30 must-fix items are at `update-audit.md` lines 264-271. | Observed (Read, lines 261-279) |
| C13 | The SPEC-ID `SPEC-UPDATE-REPAIR-RC30-001` passes the protocol regex. | Observed (see Evidence E-SPEC) |
| C14 | Migration removal at `migrate_project_assets.go:168`, messages at `:160` and `:238-242`, backup ordering at `update_template_sync.go:1183`. | Audit claim, not re-read |
| C15 | `moai init` writes which roots, and what `~/.moai` and the four roots hold on the operator machine. | Gap (not measured) |
| C16 | Rule-file hits (E-07-DISPO): 45 files and 85 occurrences of the skill-path token. 81 occurrences resolve to template targets. Four occurrences in two files name targets the template does not ship: `moai/design/constitution.md` (the `brand*.md` glob) and `moai/development/skill-authoring.md` (three retired rows). The audit's "83 rule references" is not reproduced. | Observed |
| C17 | Docs line (E-08): `docs-site/content/ko/cli-reference/update.md:295` is present as a tree-diagram entry (`moai/` under `.claude/` › `agents/`). The literal `claude/agents/moai` has no match. The named path is a template target. Value recorded: `found-at-real-path`. | Observed |
| C18 | The config-change audit log is absent from the card tree and from the operator profile. It exists in the mo.ai.kr project at `/Users/goos/MoAI/mo.ai.kr/.moai/logs/config-change-audit.log`. Its entries (E-18) bracket the audit's removal time (02:26:48Z) between observations at 02:26:46Z and 02:27:12Z. No command that removes the files is found. | Observed (bracket); Gap (attribution, §4 item 18) |

## 2. Evidence

Every command below ran in this run from the worktree root. Output is verbatim; long outputs are truncated and marked `…`.

**E-TREE (C1)**
```text
$ git rev-parse HEAD
2aab5f797b75983e132af451da68f69e3426557b
```

**E-BUILD (C2)**
```text
$ git merge-base --is-ancestor 2aab5f797 4f8aba061; echo "exit=$?"
exit=0
```

**E-BRANCH (C3, C4; historical, measured before the 19:06Z re-measurement; superseded by E-OVERLAP)**
```text
$ git diff --stat main...WT-user-asset-bundle | tail -3
 internal/web/agentfm.go                            |  16 +
 internal/web/m0_agentfm_red_test.go                | 117 +++
 75 files changed, 8471 insertions(+), 142 deletions(-)
$ git diff --stat main...WT-update-migration-fixes | tail -3
 internal/userassets/install.go                     |  33 +-
 internal/userassets/installer_empty_target_test.go |  58 +++
 13 files changed, 2066 insertions(+), 7 deletions(-)
$ git rev-parse --short WT-user-asset-bundle
cae7e28fe
$ git rev-parse --short WT-update-migration-fixes
317ed31b6
```

**E-MERGEBASE (C3)**
```text
$ git merge-base main WT-user-asset-bundle
db0c514d3b0ce16befa5a6d35126a35b6b0f73d0
$ git merge-base main WT-update-migration-fixes
2aab5f797b75983e132af451da68f69e3426557b
```

**E-OVERLAP-OLD (superseded; branch-name form, not pinned to a SHA; kept for history)**
```text
$ git diff --name-only main...WT-user-asset-bundle   (excerpt of relevant paths)
internal/cli/migrate_project_assets.go
internal/cli/user_asset_phase.go
internal/cli/update_migrate_test.go
internal/userassets/install.go
internal/userassets/paths.go
internal/userassets/remove.go
internal/web/agentfm.go
.moai/specs/SPEC-USERASSET-DEPLOY-GUARD-001/spec.md   (and sibling files)
$ git diff --name-only main...WT-update-migration-fixes
.moai/specs/SPEC-UPDATE-MIGRATION-FIX-001/spec.md   (and sibling files)
internal/cli/update.go
internal/cli/update_integrity_probe.go
internal/cli/update_integrity_probe_test.go
internal/cli/update_managed_dir_guard_test.go
internal/userassets/install.go
internal/userassets/installer_empty_target_test.go
CHANGELOG.md
```
Neither branch touches `internal/config/manager.go`, `internal/config/types.go`, `internal/cli/update_tux.go`, `internal/cli/update_archive.go`, `internal/cli/update/report/outcome.go`, `internal/cli/update_restore.go`, `internal/cli/bundle.go`, or `internal/cli/update_template_sync.go`. Those files are the basis for the IN-NOW dispositions.

**E-SPECS (C5, C6)**
```text
$ grep -n "REQ-002\|P6" .moai/specs/SPEC-USER-ASSET-INSTALL-001/spec.md | head -12
237:- P6: profile sessions do not see the shared user assets in v1 (declared
238:  limitation, doc-visible). Forced by REQ-002 and C2: every
255:- REQ-002: The system shall keep per-profile settings folders per-profile: a
462:REQ-002 → AC-019; REQ-003 → AC-017; REQ-004 → AC-018; REQ-005 → AC-011/012;
526:P5/P6 and closed in the decision register):
528:  (declared limitation) — every install-into-profile option violated REQ-002
$ git show WT-user-asset-bundle:.moai/specs/SPEC-USER-ASSET-INSTALL-001/spec.md | grep -n "REQ-002\|P6" | head -12
(same seven lines as above; branch copy identical)
$ git diff --stat main...WT-user-asset-bundle -- .moai/specs/SPEC-USER-ASSET-INSTALL-001 | tail -1
(empty: the branch does not change this SPEC directory)
$ git show WT-user-asset-bundle:.moai/specs/SPEC-USERASSET-DEPLOY-GUARD-001/spec.md | grep -n "^id:\|^title:\|^status:" | head -10
2:id: SPEC-USERASSET-DEPLOY-GUARD-001
3:title: "User-asset install/deploy defect-repair bundle: lock-guard recovery, journal durability, Codex reference normalization, collision safety, frozen-guard user-path protection, deploy-surface and doctor accuracy (card t1591)"
5:status: completed
```

**E-FIXLIST (§ E of spec.md)**
The Read of `update-audit.md` lines 261-279 is the source. Must-fix items 1-8 are at lines 264-271, and "Can follow" items are at lines 274-279.

**E-CODE (C8-C11, re-read in this run)**
```text
$ grep -n "func runUpdate(\|Current version\|func checkProjectMarker\|CLAUDE_CONFIG_DIR\|LookupEnv\|Getenv" internal/cli/update.go | head -30
140:func runUpdate(cmd *cobra.Command, _ []string) error {
200:		_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
765:	if os.Getenv(config.EnvSkipBinaryUpdate) == "1" {
835:	// Card t1527 D1: the "Current version" line is dropped here — the top of
$ grep -n "func ResolveRoots" -A 12 internal/userassets/paths.go | head -24
36:func ResolveRoots(home string) []Root {
37-	return []Root{
38-		{Slug: RootClaudeSkills, Dir: filepath.Join(home, ".claude", "skills")},
39-		{Slug: RootClaudeAgents, Dir: filepath.Join(home, ".claude", "agents")},
40-		{Slug: RootAgentsSkills, Dir: filepath.Join(home, ".agents", "skills")},
41-		{Slug: RootCodexAgents, Dir: filepath.Join(home, ".codex", "agents")},
42-	}
43-}
$ grep -n "func (m \*ConfigManager) Save\|…" internal/config/manager.go   (Save at line 225, from grep "func (.*) Save")
225:func (m *ConfigManager) Save() error {
$ grep -n "DeploymentMode\|deployment_mode\|Alias\|Fable\|Opus\|Sonnet\|Haiku\|TeamMode" internal/config/types.go | head -20
284:	TeamMode string `yaml:"team_mode"`
383:	Fable  string `yaml:"fable,omitempty"`
389:// legacy backward-compat alias fields Opus/Sonnet/Haiku are removed. An
398:	Fable  string `yaml:"fable"`  // Fable tier (Claude Code ANTHROPIC_DEFAULT_FABLE_MODEL)
$ grep -n "deployment_mode" internal/config/deploy_mode.go | head -5
23:		DeploymentMode string `yaml:"deployment_mode"`
$ grep -n "customizations modified" internal/cli/update_archive.go | head -5
282:…
367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
$ grep -n "conflict\|Updated %d\|user settings key\|copy back manually\|steps" internal/cli/update_tux.go | head -40
(outputs used for U-08, U-26, U-27, U-28, U-16: lines 91, 95, 107, 111, 119-120, 134-139, 237, 286, 340-345)
$ grep -n "preserved" internal/cli/update/report/outcome.go | head -10
101 (RenderReconciliation), 106 ("Reconciliation: %d refreshed, %d merged, %d conflict(s), %d preserved, %d archived-removed"), 114-115 ("  preserved: " + p)
$ grep -n "SyncToProjectConfig\|profile\." internal/cli/update.go | head -12
652:	profileName := profile.GetCurrentName()
653:	prefs, err := profile.ReadPreferences(profileName)
657:		if err := profile.SyncToProjectConfig(".", prefs); err != nil {
```

**E-SPECID (C13)**
```text
$ bash -c 'ID="SPEC-UPDATE-REPAIR-RC30-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL'
PASS
```
This check ran after spec.md was written. The protocol orders the check before the write. The check passed, and the order is recorded as a deviation under Gaps.

**E-LOCALE (R4; decision-index.md Q1)**
```text
$ grep -n "Neutrality\|16 \|user-facing\|locales" .claude/rules/moai/development/coding-standards.md
23:### Programming-Language Neutrality Contract (16 languages)
25:Template content treats the 16 supported programming languages equally:
35:  the 16 programming languages above are a NEUTRALITY concern (templates must
36:  treat them equally), while the 4 user-facing locales (ko, en, ja, zh) are a
37:  TRANSLATION concern (user-facing prose is rendered per locale). The two
$ grep -n "go, python, typescript" .claude/rules/moai/development/coding-standards.md
26:go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php,
(line 27 continues the list: elixir, cpp, scala, r, flutter, swift; the list was displayed by the Read of this file in this run)
$ grep -n "Options" .moai/config/sections/language.yaml
5:  # Options: ko (Korean), en (English), ja (Japanese), zh (Chinese), es (Spanish), fr (French), de (German)
```
Result: the "16" of R1 is the 16 programming languages (neutrality). The "four" of R2(g) is the four locales ko, en, ja, zh (translation). The sets are disjoint and differ. language.yaml lists seven conversation-language options.

**E-OVERLAP (pinned to tip SHAs; measured at 18:4x UTC on 2026-10-09; supersedes E-OVERLAP-OLD)**
```text
$ git rev-parse --short WT-user-asset-bundle        -> df559b729
$ git rev-parse --short WT-update-migration-fixes   -> 40ab940f4
$ git rev-parse --short main                        -> 2aab5f797
$ git diff --stat main...df559b729 | tail -1        -> 81 files changed, 9149 insertions(+), 157 deletions(-)
$ git diff --stat main...40ab940f4 | tail -1        -> 13 files changed, 2068 insertions(+), 7 deletions(-)
$ git diff --name-only main...df559b729 -- internal/cli/update_tux.go internal/cli/update_archive.go internal/cli/update/report/outcome.go internal/cli/update_restore.go internal/cli/bundle.go internal/cli/update_template_sync.go internal/config/manager.go internal/config/types.go
(empty: no IN-NOW file is changed at df559b729)
$ git diff --name-only main...40ab940f4 -- (the same eight IN-NOW paths)
(empty: no IN-NOW file is changed at 40ab940f4)
$ git diff --name-only main...df559b729 -- internal/cli/migrate_project_assets.go internal/cli/user_asset_phase.go internal/userassets/install.go internal/userassets/paths.go internal/cli/update.go internal/template/harness_fs.go
internal/cli/migrate_project_assets.go
internal/cli/user_asset_phase.go
internal/template/harness_fs.go
internal/userassets/install.go
internal/userassets/paths.go
$ git diff --name-only main...40ab940f4 -- (the same six paths)
internal/cli/update.go
internal/userassets/install.go
```
Re-derivation at M1 is required after absorption (spec.md § G.2). The audit's tip `2022e4633` for t1591 is superseded by `df559b729` (`01b519c21` was the tip at 19:06Z).

**E-07-COUNT (rule files naming `.claude/skills/moai/`)**
```text
$ grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules | wc -l
45
```

**E-PROBE (R3; U-31)**
```text
$ git show WT-update-migration-fixes:internal/cli/update_integrity_probe.go | head -80   (excerpt of the 80-line output)
 47://   - .claude/settings.json: the Claude Code settings merge target; must parse as JSON.
 52:	{rel: ".claude/settings.json", check: probeJSON},
 64:		return "missing"
$ git cat-file -e main:internal/cli/update_integrity_probe.go; echo "main-exit=$?"
fatal: path 'internal/cli/update_integrity_probe.go' does not exist in 'main'
main-exit=128
$ grep -n "hideClaude" internal/template/harness_fs.go   (excerpt)
112:	if h.hideClaude && (name == ".claude" || strings.HasPrefix(name, ".claude/")) {
```

**E-R5 (item-register count, re-run in revision 0.1.4 on the current spec.md)**
```text
$ grep -Ec "^\| U-[0-9]{2} \|" .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md
31
$ grep -E "^\| U-[0-9]{2} \|" .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md | grep -Eo "IN-NOW|BLOCKED-ON|GAP-INVESTIGATE|DECISION-PENDING" | wc -l
      23
$ grep -E "^\| U-[0-9]{2} \|" .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md | grep -Evc "IN-NOW|BLOCKED-ON|GAP-INVESTIGATE|DECISION-PENDING"
8
(per-token counts, the same pipeline with `grep -Eo` on one token at a time: IN-NOW 9, BLOCKED-ON 8, GAP-INVESTIGATE 6, DECISION-PENDING 0; the 8 rows without a token are U-02, U-10 and U-31 (card t1668) and U-26 to U-30 (card t1669))
```
Revision 0.1.3 recorded 28 tokens, 3 rows without a token, and the counts 13, 7, 7 and 1. That transcript did not reproduce: on the revision 0.1.3 text, 8 rows carry no token, not 3, and the per-token counts differ. Revision 0.1.4 re-ran the check above on the current spec.md; the row count is unchanged at 31.

**E-MIGRATE-PATH (C7)**
```text
$ grep -rln "func migrateProjectCommonAssets" internal/cli
internal/cli/migrate_project_assets.go
$ ls internal/cli/update/ | grep -c migrate_project_assets
0
```

**E-07-DISPO (AC-RC30-20, U-19: one disposition per rule-file hit; spec round for sync finding F4)**

Method. Each hit file's `.claude/skills/moai/` tokens were resolved against the template tree (E-07-INV). A literal path resolves when it exists; a glob resolves when at least one template file matches it; a brace list is checked per alternative. A file is `deployed` when every token resolves, `stale` when at least one token names a path the template does not ship, and `not-a-reference` when the hit names no path. The template tree is the deploy source: `internal/template/embed_manifest_gen.go:23-25` embeds `templates/...`, and `internal/template/catalog_tree_hash.go:17` reads "Deployed through //go:embed all:templates". The token column omits the `.claude/skills/moai/` prefix.

```text
$ grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules | wc -l
45
$ grep -rho "\.claude/skills/moai/[^ \`'\")]*" internal/template/templates/.claude/rules | wc -l
85
$ command ls internal/template/templates/.claude/agents/
moai
$ ls -d internal/template/templates/.claude/commands/harness
ls: internal/template/templates/.claude/commands/harness: No such file or directory   (exit 1)
```

Disposition per rule-file hit (paths relative to `internal/template/templates/.claude/rules/`):

| # | Rule file | Named target(s), prefix omitted | Disposition | REQ-RC30-025 correction |
|---|---|---|---|---|
| 01 | `moai/core/agent-common-protocol-reference.md` | `workflows/*.md`; `workflows/sync/quality-gates-quality.md` | deployed | — |
| 02 | `moai/core/askuser-protocol-reference.md` | `workflows/*.md` | deployed | — |
| 03 | `moai/core/moai-constitution-detail.md` | `workflows/*.md` | deployed | — |
| 04 | `moai/core/moai-mcp-tools-catalogue.md` | `workflows/run/external-delegation.md` (two hits) | deployed | — |
| 05 | `moai/core/security-decision-contract.md` | `workflows/sync/**` | deployed | — |
| 06 | `moai/design/constitution.md` | `**/design*.md` (matches `workflows/design.md`); `**/brand*.md` (matches no file) | stale | Drop `.claude/skills/moai/**/brand*.md` from the `paths:` scope. Keep `**/design*.md`, which names the template target `workflows/design.md`. |
| 07 | `moai/development/agent-authoring.md` | `workflows/run/external-delegation.md` | deployed | — |
| 08 | `moai/development/branch-origin-protocol.md` | `workflows/plan.md`; `workflows/plan/**/*.md` | deployed | — |
| 09 | `moai/development/coding-standards.md` | `workflows/` (directory) | deployed | — |
| 10 | `moai/development/karpathy-quickref.md` | `references/anti-patterns.md` | deployed | — |
| 11 | `moai/development/manager-develop-prompt-template.md` | `workflows/run.md` | deployed | — |
| 12 | `moai/development/skill-authoring.md` | `workflows/release-update.md`, `workflows/github.md`, `workflows/release.md` (lines 356–358, the retired-skill table) | stale | Remove the three rows from the shipped copy, or restate the retirement in prose with no `.claude/skills/moai/` or `.claude/agents/harness/` path token. Their current targets (`.claude/agents/harness/hns-*-specialist.md`, `.claude/commands/harness/*`) are absent from the template tree (command output above); no template replacement exists. |
| 13 | `moai/development/spec-frontmatter-schema.md` | `workflows/plan/spec-assembly.md` | deployed | — |
| 14 | `moai/NOTICE.md` | `references/anti-patterns.md` | deployed | — |
| 15 | `moai/workflow/archived-agent-rejection.md` | `workflows/run.md`, `workflows/plan.md`, `workflows/sync.md`; `workflows/{plan,run,sync}.md` | deployed | — |
| 16 | `moai/workflow/cadence-bridge.md` | `workflows/fix.md`, `workflows/gate.md`, `workflows/review.md`, `workflows/loop.md`, `workflows/sync/delivery.md` | deployed | — |
| 17 | `moai/workflow/context-clear-policy.md` | `workflows/run/**` | deployed | — |
| 18 | `moai/workflow/cross-session-messaging.md` | `workflows/gtd.md` | deployed | — |
| 19 | `moai/workflow/delivery-policy.md` | `workflows/**` | deployed | — |
| 20 | `moai/workflow/factory-dispatch-cards.md` | `workflows/gtd.md` | deployed | — |
| 21 | `moai/workflow/factory-dispatch-detail.md` | `workflows/gtd.md` | deployed | — |
| 22 | `moai/workflow/factory-dispatch-gates.md` | `workflows/gtd.md` | deployed | — |
| 23 | `moai/workflow/factory-dispatch-mechanics.md` | `workflows/gtd.md` | deployed | — |
| 24 | `moai/workflow/factory-dispatch.md` | `workflows/gtd.md` (two hits) | deployed | — |
| 25 | `moai/workflow/goal-directive-detail.md` | `workflows/goal.md`, `workflows/run.md` | deployed | — |
| 26 | `moai/workflow/goal-directive.md` | `workflows/goal.md` | deployed | — |
| 27 | `moai/workflow/graceful-exit-mutation-contract.md` | `workflows/sync/**` | deployed | — |
| 28 | `moai/workflow/language-routing-contract.md` | `workflows/sync/**` | deployed | — |
| 29 | `moai/workflow/mx-tag-protocol.md` | `references/mx-tag.md` (trailing period stripped) | deployed | — |
| 30 | `moai/workflow/native-invocation-model.md` | `SKILL.md` | deployed | — |
| 31 | `moai/workflow/orchestration-mode-selection.md` | `workflows/goal.md`, `workflows/run.md`, `workflows/plan.md` | deployed | — |
| 32 | `moai/workflow/phase-id-contract.md` | `workflows/run/**` | deployed | — |
| 33 | `moai/workflow/read-only-status-contract.md` | `workflows/sync/**` | deployed | — |
| 34 | `moai/workflow/resource-budget-contract.md` | `workflows/sync/**` | deployed | — |
| 35 | `moai/workflow/session-handoff-examples.md` | `workflows/goal.md` | deployed | — |
| 36 | `moai/workflow/skill-routing.md` | `SKILL.md` | deployed | — |
| 37 | `moai/workflow/snapshot-consumer-contract.md` | `workflows/sync/**` | deployed | — |
| 38 | `moai/workflow/spec-workflow.md` | `workflows/design.md`, `fix.md`, `mx.md`, `codemaps.md`, `clean.md`, `plan.md`, `run.md`, `sync.md`, `loop.md` | deployed | — |
| 39 | `moai/workflow/sync-backup-integrity.md` | `workflows/sync/**` | deployed | — |
| 40 | `moai/workflow/sync-coverage-scope-contract.md` | `workflows/sync/**` | deployed | — |
| 41 | `moai/workflow/tdd-result-contract.md` | `workflows/run/**` | deployed | — |
| 42 | `moai/workflow/team-capability-resolver.md` | `workflows/run/mode-orchestration.md` | deployed | — |
| 43 | `moai/workflow/trace-ledger-contract.md` | `workflows/sync*.md`, `workflows/sync/**`, `workflows/project*.md`, `workflows/project/**`, `workflows/plan*.md`, `workflows/plan/**`, `workflows/run*.md`, `workflows/run/**` | deployed | — |
| 44 | `moai/workflow/verification-batch-pattern.md` | `workflows/run.md`; `**/.claude/skills/moai/workflows/sync/**` | deployed | — |
| 45 | `moai/workflow/verification-plan-contract.md` | `workflows/run/**`, `workflows/sync/**` | deployed | — |

Count: 43 `deployed`, 2 `stale` (rows 06 and 12), 0 `not-a-reference`. Four token occurrences are stale (`**/brand*.md` once; the three retired paths once each), and 81 of 85 resolve. The stale rows carry their REQ-RC30-025 corrections in the last column; those are recorded here and are not yet applied to the template files (§4 item 19).

**Docs-line value (AC-RC30-20, U-20): `found-at-real-path`.** See E-08.

**E-07-INV (template target set: `internal/template/templates/.claude/skills/moai`)**

```text
$ find internal/template/templates/.claude/skills/moai -type f | sed 's#^internal/template/templates/##' | sort
.claude/skills/moai/references/anti-patterns.md
.claude/skills/moai/references/file-reading-optimization.md
.claude/skills/moai/references/mx-tag.md
.claude/skills/moai/references/reference.md
.claude/skills/moai/SKILL.md
.claude/skills/moai/workflows/clean.md
.claude/skills/moai/workflows/codemaps.md
.claude/skills/moai/workflows/design.md
.claude/skills/moai/workflows/e2e.md
.claude/skills/moai/workflows/factory.md
.claude/skills/moai/workflows/feedback.md
.claude/skills/moai/workflows/fix.md
.claude/skills/moai/workflows/gate.md
.claude/skills/moai/workflows/goal.md
.claude/skills/moai/workflows/gtd.md
.claude/skills/moai/workflows/harness-build-entry.md
.claude/skills/moai/workflows/harness-builder.md
.claude/skills/moai/workflows/harness.md
.claude/skills/moai/workflows/loop.md
.claude/skills/moai/workflows/moai.md
.claude/skills/moai/workflows/mx.md
.claude/skills/moai/workflows/plan.md
.claude/skills/moai/workflows/plan/.gitkeep
.claude/skills/moai/workflows/plan/clarity-interview.md
.claude/skills/moai/workflows/plan/context-discovery.md
.claude/skills/moai/workflows/plan/spec-assembly.md
.claude/skills/moai/workflows/project.md
.claude/skills/moai/workflows/project/.gitkeep
.claude/skills/moai/workflows/project/codebase-analysis.md
.claude/skills/moai/workflows/project/doc-generation.md
.claude/skills/moai/workflows/project/meta-harness.md
.claude/skills/moai/workflows/project/mode-detection.md
.claude/skills/moai/workflows/references/shadcn-component-catalog.md
.claude/skills/moai/workflows/review.md
.claude/skills/moai/workflows/run.md
.claude/skills/moai/workflows/run/.gitkeep
.claude/skills/moai/workflows/run/context-loading.md
.claude/skills/moai/workflows/run/external-delegation.md
.claude/skills/moai/workflows/run/mode-orchestration.md
.claude/skills/moai/workflows/run/phase-execution.md
.claude/skills/moai/workflows/run/task-decomposition.md
.claude/skills/moai/workflows/sync.md
.claude/skills/moai/workflows/sync/.gitkeep
.claude/skills/moai/workflows/sync/delivery.md
.claude/skills/moai/workflows/sync/doc-execution.md
.claude/skills/moai/workflows/sync/quality-gates-context.md
.claude/skills/moai/workflows/sync/quality-gates-quality.md
$ find internal/template/templates/.claude/skills/moai -type f | wc -l
47
```

**E-08 (AC-RC30-20, U-20: docs line)**

```text
$ grep -n "claude/agents/moai" docs-site/content/ko/cli-reference/update.md; echo "exit=$?"
exit=1
$ grep -n "moai/ " docs-site/content/ko/cli-reference/update.md
295:│   ├── moai/                # MoAI-ADK 에이전트 (업데이트 대상)
299:│   └── moai/                # MoAI-ADK 훅 스크립트 (업데이트 대상)
306:    └── moai/                # 규칙 파일 (moai 관리)
$ sed -n '286,298p' docs-site/content/ko/cli-reference/update.md
{{< /callout >}}

### moai 폴더 구조

MoAI-ADK는 다음 폴더에서만 파일을 관리합니다:

```
.claude/
├── agents/
│   ├── moai/                # MoAI-ADK 에이전트 (업데이트 대상)
│   └── harness/             # 사용자 하네스 에이전트 (업데이트 제외, 보존)
│
├── hooks/
```
```

Reading. The audit cites line 295, and that line is present: it names the `moai/` entry under `.claude/` › `agents/`. The literal pattern `claude/agents/moai` cannot match a tree diagram, so the literal grep reports no match even though the reference exists. The named path `.claude/agents/moai/` is a template target (12 files: `builder-harness.md`, `e2e-tester.md`, `manager-design.md`, `manager-develop.md`, `manager-docs.md`, `manager-git.md`, `manager-lead.md`, `manager-spec.md`, `manager-todo.md`, `plan-auditor.md`, `super-advisor.md`, `sync-auditor.md`). The docs line therefore resolves to a real path and needs no correction. Value recorded: `found-at-real-path`.

**E-18 / E-25 (AC-RC30-24, U-25: config-change audit log and the removal's attribution)**

Log and producer, in the card tree and the operator profile:

```text
$ ls -la .moai/logs/config-change-audit.log          (card tree)
ls: .moai/logs/config-change-audit.log: No such file or directory   (exit 1)
$ find /Users/goos/.moai/claude-profiles -name '*config-change*'
/Users/goos/.moai/claude-profiles/moai-adk/plugins/marketplaces/modu-ai-moai-adk..clone/.claude/hooks/moai/handle-config-change.sh
/Users/goos/.moai/claude-profiles/moai-adk/plugins/marketplaces/modu-ai-moai-adk..clone/internal/template/templates/.claude/hooks/moai/handle-config-change.sh.tmpl
$ grep -rl "config-change-audit" /Users/goos/.moai/claude-profiles | wc -l
117
$ grep -n "configChangeAuditRelPath" internal/hook/config_change.go    (card tree)
208:// configChangeAuditRelPath is the durable record of what the async reload
226:const configChangeAuditRelPath = ".moai/logs/config-change-audit.log"
267:	logPath := filepath.Join(projectDir, configChangeAuditRelPath)
$ ls -la /Users/goos/MoAI/mo.ai.kr/.moai/logs/config-change-audit.log
-rw-r--r--@ 1 goos  staff  261043 Oct 10 17:11 /Users/goos/MoAI/mo.ai.kr/.moai/logs/config-change-audit.log
```

Reading. No file named `config-change-audit.log` is in the operator profile. The 117 files that mention the name are transcripts, memory, lesson entries, and the plugin clone's copy of the producer. The producer is the ConfigChange hook: `handle-config-change.sh` execs `moai hook config-change`, and `internal/hook/config_change.go:226` names the log path. The log itself is in the mo.ai.kr project, which is outside this repository and outside the operator profile.

The log lines that bracket the removal (verbatim, `/Users/goos/MoAI/mo.ai.kr/.moai/logs/config-change-audit.log`):

```text
1231:[2026-10-09T02:26:43Z] session=89acb9f0-a359-421f-95c9-4daa9b054999 path=/Users/goos/MoAI/mo.ai.kr/.claude/skills/moai/SKILL.md source=skills result=rejected detail="invalid YAML: yaml: line 14: did not find expected whitespace or line break"
1241:[2026-10-09T02:26:46Z] session=65dafe61-913a-4217-a62c-bb91201efdcb path=/Users/goos/MoAI/mo.ai.kr/.claude/skills/moai/references/reference.md source=skills result=rejected detail="invalid YAML: yaml: line 32: mapping values are not allowed in this context"
1266:[2026-10-09T02:27:12Z] session=89acb9f0-a359-421f-95c9-4daa9b054999 path=/Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-lead.md source=skills result=rejected detail="read file: open /Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-lead.md: no such file or directory"
1267:[2026-10-09T02:27:15Z] session=65dafe61-913a-4217-a62c-bb91201efdcb path=/Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-docs.md source=skills result=rejected detail="read file: open /Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-docs.md: no such file or directory"
```

Reading. The skill files were still on disk at 02:26:43Z and 02:26:46Z (parse errors, so they existed). The agent files were absent at 02:27:12Z and 02:27:15Z. The audit's removal time, 2026-10-09 11:26:48 KST, is 02:26:48Z, which lies between those observations. This brackets the removal. It does not name the command that caused it.

Removal-command searches (all read-only; none attributes the removal):

```text
$ grep -rh --include='*.jsonl' '"timestamp":"2026-10-09T02:2' <operator-profile moai-adk/projects> | grep 'skills/moai'   → tool_use records whose `command` contains `skills/moai`: 0
$ grep -rh --include='*.jsonl' '"timestamp":"2026-10-09T02:2' <operator-profile mo.ai.kr/projects>                         → 10 tool_use records, none removes a path under .claude/skills/moai or .claude/agents/moai
$ grep -rh --include='*.jsonl' '"timestamp":"2026-10-09T' <operator-profile moai-adk/projects> | grep 'skills/moai'        → 51 tool_use records mention skills/moai; the one with a removal verb is a heredoc text note at 2026-10-09T17:37:03.819Z, not a removal
$ grep -rh --include='*.jsonl' '"timestamp":"2026-10-09T' <operator-profile> | grep -E 'moai (update|migrate|init)'     → 11 tool_use records on 2026-10-09; none falls in 02:20:00Z–02:35:00Z
$ ~/.zsh_history (2072 epoch-tagged entries, parsed)                                                                    → 0 entries in 2026-10-09T02:20:00Z–02:35:00Z
```

The one background task near the bracket is `b8g4eb8vq`. Its command, a `git merge --no-ff --no-commit WT-news-og-body-halt` in the mo.ai.kr session at 02:26:22Z, was moved to background at 02:26:23Z. Its output file (`/private/tmp/claude-501/-Users-goos-MoAI-mo-ai-kr/b4f5059b-9f5b-4d67-a0d3-35ad9970cbc5/tasks/b8g4eb8vq.output`, 542 bytes, mtime 11:26 local) holds a pre-commit hook failure message and `[exited with code 0]`, and it names no path. It cannot attribute the removal.

Verdict: the removal is bracketed and its producer is identified; the removing command is **not** attributed (§4 item 18). AC-RC30-24 stays FAIL until a command with observed output names the removal.

## 3. Baseline-attribution

- Each Evidence block names the command, the tree (`8fa0dc24b`, code paths equal main `2aab5f797`), and the output in this run. No figure is carried from the audit or from another tree.
- Audit figures (for example "423 files", "266 files", "369 preserved lines", "83 references", "17 commands", "12 gate.yaml comment lines", "29 of 54 `.agents/skills` dirs differ") are audit claims. They are cited only as audit claims, labelled as such in spec.md § C.
- The ranges used for overlap are file-level (`git diff --name-only main...<branch>`). Hunk-level ranges are measured from `git merge-base HEAD <absorbed-tip>` at M1 (plan § B.2), not from a literal SHA.
- The installed `moai` binary is not used. Measurements in later milestones use a build of this tree, invoked by path.
- Revision for the sync findings F4, F5, and F9 (card t1619, sync audit `ab0adb383`): E-07-DISPO, E-07-INV, and E-08 were run on card-tree HEAD `dc4d673b1` (branch `WT-moai-update-repair`), the same HEAD at the start and at the end of the measurements. E-18 reads the operator profile and the mo.ai.kr project, which are outside this tree; it is evidence about those locations, not a tree measurement.

## 4. Gaps

Not observed in this run. Each gap is listed with the next step.

1. **Init path (spec.md § D.3, C15).** `moai init` was not run from a build of this tree. Next step: `make build`, then `HOME=<scratch> <built-binary> init <scratch-project>`, then list the four roots and `~/.moai` under `<scratch>`.
2. **Operator home state (§ D.1).** `ls` of `~/.claude/skills`, `~/.claude/agents`, `~/.agents/skills`, `~/.codex/agents`, and `~/.moai` was not run. Next step: run each `ls` as a single command and record stdout.
3. **REQ-002 full clause.** Only spec.md:255 was read in part. Next step: read lines 255-260 in full, then judge whether an install into a profile directory is excluded.
4. **Audit-only code lines (C14).** `migrate_project_assets.go:154-168`, `:238-242`, `update_template_sync.go:247`, `:361`, `:619`, `:859`, `:1100-1114`, `:1143`, `:1157`, `:1183`, `install.go:612-624`, `backup.go:209`, `update_restore.go:43`, `update_reconcile.go:64-70`, and `user_asset_phase.go:105-125` are audit citations. They were not re-read. Next step: read each range in M0.
5. **Audit path error.** The audit cites `internal/cli/update/migrate_project_assets.go`. The function is in `internal/cli/migrate_project_assets.go` (E-MIGRATE-PATH). spec.md § C is corrected for U-02. Other rows cite the audit line and should be re-checked in M0.
6. **Migration call site.** The grep for `migrateProjectCommonAssets(` call sites failed because of a shell glob error (`--include=*.go` was not quoted). Next step: re-run with quoted globs.
7. **Make target and golden path.** `make tui-snapshot` was not confirmed in this run. Next step: `grep -n "tui-snapshot" Makefile`.
8. **Locale mechanism (U-30).** Not read. Next step: find how update output strings are localized, if at all.
9. **Queue card text.** t1527, t1594, and t1612 were not read. spec.md § F records their overlaps with the disposition undecided. Next step: read the three card texts before the overlap is decided.
10. **docs-site path (U-20).** Checked in this revision (E-08): the path exists, and the cited line 295 is present as a tree-diagram entry. The audit's claim is reproduced in tree form; the literal pattern is not. Closed by E-08.
11. **Runnable profile-visibility command (AC-RC30-01).** Not measured. Next step: card t1668 (operator-side profile session).
12. **Coverage baseline.** Not measured. The target is 85% or more per package. Next step: `go test -cover` on each touched package, in M7.
13. **Order deviation.** The SPEC-ID check ran after the write of spec.md, not before. The check now passes.
14. **Base ambiguity.** The worktree is cut from main, while `.claude/rules/local/gitflow-lane-protocol.md` § 1 names develop. The lane must confirm the base before absorption.
15. **Card t1616 status.** Landing waits for it. Its state was not checked.
16. **Absorption order and terminal state (leader intake, confirmed by the leader).** The absorption order is t1591, then t1578, then t1619. The absorption is done (merges `ff2bc876d` and `6f95917a1`; progress.md § G). Merge-ready is this card's terminal state until card t1616 lands. No merge into main is performed in this card.
17. **Integrity probe harness source (U-31).** The probe exists only on WT-update-migration-fixes (E-PROBE). Its codex-only RED cell (AC-RC30-26) can be captured only after M1 absorption. How the probe learns that a project is codex-only (the harness condition source) is not yet read in this tree, so the condition's exact source is a Gap.

18. **Removal attribution (U-25, AC-RC30-24).** Not attributed. E-18 identifies the log, its producer, and the bracket 02:26:46Z to 02:27:12Z. No command that removes `.claude/skills/moai/**` or the agent files is found in the operator profile's transcripts, the mo.ai.kr transcripts, or the shell history for the window. The audit counts 46 skill files and six agents; the log names two agent files, and no skill-file absence. Next step, from the mo.ai.kr operator terminal and not from this worktree: read that repository's reflog and terminal history for 2026-10-09 02:20–02:35Z UTC, and attribute only a command whose observed output shows the removal.
19. **Stale-row corrections are recorded, not applied.** E-07-DISPO records a REQ-RC30-025 correction for each of the two stale rows. The template files `internal/template/templates/.claude/rules/moai/design/constitution.md` and `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` are unchanged in this revision. Applying the corrections is a template edit outside this spec-artifact revision.
20. **Audit counts not reproduced.** The audit's "83 rule references and 17 commands" (plan D16) is not reproduced: the rules tree has 85 occurrences (E-07-DISPO). The commands tree is outside AC-RC30-20's grep, so the 17-command figure is not measured here.

## 5. Residual-risk

- **Premise risk (profile visibility).** The removal gate assumes that project-level `.claude/skills` is not affected by `CLAUDE_CONFIG_DIR`. That premise is measured only by AC-RC30-01. Until then it is a premise, not evidence.
- **Overlap risk.** Both branches change `internal/userassets/install.go`. A plain merge at M1 may conflict in ways the file list does not show.
- **Drift risk.** Branch tips move. Ranges stay correct only if they are re-measured after each absorption.
- **Audit-carry risk.** Most root causes are audit claims (§ 4 item 4). A cause that is wrong in this tree will still appear in the plan until M0 corrects it.
- **Budget risk.** Tier L allows 25 requirements and 25 criteria. SPEC A carries 17 requirements (REQ-RC30-003, -004 and -021 owned by card t1668; REQ-RC30-014, -016, -018, -019 and -020 owned by card t1669) and 22 criteria (AC-RC30-28 added in revision 0.1.4). Decision B1 is closed by the leader's split; no relaxation is recorded.
- **Decision risk.** Q2 is decided (operator decision d-20261010T060848Z-dc3b). Q1 is resolved from disk. U-30 waits for the locale-mechanism read. Q7 carries the applied default of 60 columns (board ruling d-20261010T101935Z-4bc4), which the operator has not confirmed. Q6 is partly measured.
- **Docs-line method (U-20).** AC-RC30-20's literal grep cannot match a tree-form reference. The recorded value (`found-at-real-path`) rests on reading line 295 (E-08); a reviewer should re-read that line rather than trust the grep count.
- **Removal recurrence (U-25).** While the removing command is unknown, nothing in this SPEC prevents the same removal in the mo.ai.kr project. The audit places the exposure in that project's tree.
- **Conflict surfaced by the spec.md line-206 correction.** REQ-UPM-030 (SPEC-UPDATE-MIGRATION-001) requires per-path listings in the update summary. REQ-RC30-015 (spec.md line 56) prints per-category counts in place of the file list when `--verbose` is not set. The default output cannot satisfy both. The REQ-RC30-015 decision is pending and is not made in this revision.
