# SPEC-UPDATE-REPAIR-RC30-001 — Research and verification appendix

Five sections, per `verification-claim-integrity.md` § 3: Claim, Evidence, Baseline-attribution, Gaps, Residual-risk. Tree for every measurement in this appendix: `2aab5f797b75983e132af451da68f69e3426557b`, unless a row names another ref.

## 1. Claim

| # | Claim | Status |
|---|-------|--------|
| C1 | The measured tree is main tip `2aab5f797b75983e132af451da68f69e3426557b`. | Observed |
| C2 | The installed binary (`4f8aba061`) descends from the tree, so no claim here relies on it. | Observed |
| C3 | `WT-user-asset-bundle` (t1591) changes 75 files and `WT-update-migration-fixes` (t1578) changes 13 files, relative to main. | Observed |
| C4 | Branch tips at measurement: t1591 `cae7e28fe`, t1578 `317ed31b6`. | Observed |
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

**E-BRANCH (C3, C4)**
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
$ git rev-parse --short WT-user-asset-bundle        -> 01b519c21
$ git rev-parse --short WT-update-migration-fixes   -> 40ab940f4
$ git rev-parse --short main                        -> 2aab5f797
$ git diff --stat main...01b519c21 | tail -1        -> 78 files changed, 8653 insertions(+), 157 deletions(-)
$ git diff --stat main...40ab940f4 | tail -1        -> 13 files changed, 2068 insertions(+), 7 deletions(-)
$ git diff --name-only main...01b519c21 -- internal/cli/update_tux.go internal/cli/update_archive.go internal/cli/update/report/outcome.go internal/cli/update_restore.go internal/cli/bundle.go internal/cli/update_template_sync.go internal/config/manager.go internal/config/types.go
(empty: no IN-NOW file is changed at 01b519c21)
$ git diff --name-only main...40ab940f4 -- (the same eight IN-NOW paths)
(empty: no IN-NOW file is changed at 40ab940f4)
$ git diff --name-only main...01b519c21 -- internal/cli/migrate_project_assets.go internal/cli/user_asset_phase.go internal/userassets/install.go internal/userassets/paths.go internal/cli/update.go internal/template/harness_fs.go
internal/cli/migrate_project_assets.go
internal/cli/user_asset_phase.go
internal/template/harness_fs.go
internal/userassets/install.go
internal/userassets/paths.go
$ git diff --name-only main...40ab940f4 -- (the same six paths)
internal/cli/update.go
internal/userassets/install.go
```
Re-derivation at M1 is required after absorption (spec.md § G.2). The audit's tip `2022e4633` for t1591 is superseded by `01b519c21`.

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

**E-R5 (item-register count, run after the revision)**
```text
$ grep -Ec "^\| U-[0-9]{2} \|" .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md
31
$ grep -E "^\| U-[0-9]{2} \|" .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md | grep -Eo "IN-NOW|BLOCKED-ON|GAP-INVESTIGATE|DECISION-PENDING" | wc -l
31
$ grep -E "^\| U-[0-9]{2} \|" .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md | grep -Evc "IN-NOW|BLOCKED-ON|GAP-INVESTIGATE|DECISION-PENDING"
0
(counts by disposition cell: IN-NOW 13, BLOCKED-ON 10, GAP-INVESTIGATE 7, DECISION-PENDING 1)
```
A first count returned 32 tokens. The extra token was in the Source cell of U-14 ("Lane brief IN-NOW example"). That repeat was removed, and the count above is the one after the removal.

**E-MIGRATE-PATH (C7)**
```text
$ grep -rln "func migrateProjectCommonAssets" internal/cli
internal/cli/migrate_project_assets.go
$ ls internal/cli/update/ | grep -c migrate_project_assets
0
```

## 3. Baseline-attribution

- Each Evidence block names the command, the tree (`2aab5f797`), and the output in this run. No figure is carried from the audit or from another tree.
- Audit figures (for example "423 files", "266 files", "369 preserved lines", "83 references", "17 commands", "12 gate.yaml comment lines", "29 of 54 `.agents/skills` dirs differ") are audit claims. They are cited only as audit claims, labelled as such in spec.md § C.
- The ranges used for overlap are file-level (`git diff --name-only main...<branch>`). Hunk-level ranges are measured from `git merge-base HEAD <absorbed-tip>` at M1 (plan § B.2), not from a literal SHA.
- The installed `moai` binary is not used. Measurements in later milestones use a build of this tree, invoked by path.

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
10. **docs-site path (U-20).** `docs-site/content/ko/cli-reference/update.md` was not checked. Next step: `ls` of the path.
11. **Runnable profile-visibility command (AC-RC30-01).** Not measured. Next step: M0.
12. **Coverage baseline.** Not measured. The target is 85% or more per package. Next step: `go test -cover` on each touched package, in M7.
13. **Order deviation.** The SPEC-ID check ran after the write of spec.md, not before. The check now passes.
14. **Base ambiguity.** The worktree is cut from main, while `.claude/rules/local/gitflow-lane-protocol.md` § 1 names develop. The lane must confirm the base before absorption.
15. **Card t1616 status.** Landing waits for it. Its state was not checked.
16. **Absorption order and terminal state (leader intake, confirmed by the leader).** The absorption order is t1591, then t1578, then t1619. Merge-ready is this card's terminal state until card t1616 lands. No merge is performed in this card.
17. **Integrity probe harness source (U-31).** The probe exists only on WT-update-migration-fixes (E-PROBE). Its codex-only RED cell (AC-RC30-26) can be captured only after M1 absorption. How the probe learns that a project is codex-only (the harness condition source) is not yet read in this tree, so the condition's exact source is a Gap.

## 5. Residual-risk

- **Premise risk (profile visibility).** The removal gate assumes that project-level `.claude/skills` is not affected by `CLAUDE_CONFIG_DIR`. That premise is measured only by AC-RC30-01. Until then it is a premise, not evidence.
- **Overlap risk.** Both branches change `internal/userassets/install.go`. A plain merge at M1 may conflict in ways the file list does not show.
- **Drift risk.** Branch tips move. Ranges stay correct only if they are re-measured after each absorption.
- **Audit-carry risk.** Most root causes are audit claims (§ 4 item 4). A cause that is wrong in this tree will still appear in the plan until M0 corrects it.
- **Budget risk.** Tier L allows 25 requirements (25 used, REQ-RC30-001 to REQ-RC30-025) and 25 criteria (26 used, AC-RC30-01 to AC-RC30-26; OPEN LEADER DECISION B1). The leader decides between a split and a merge; no relaxation is recorded.
- **Decision risk.** Q2 is pending (operator decision). Q1 is resolved from disk. U-30 waits for the locale-mechanism read, and U-01 is DECISION-PENDING (Q2). Q7 is pending with no applied default.
