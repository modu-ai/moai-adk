# research.md — SPEC-WORKTREE-HOME-ROOT-001

> Evidence record. Everything here was observed in this planning session (re-read at HEAD `e5950b354f` in the first correction round at `40c2e127bf` in the operator-decision correction, §R10, at `bf78f5541a` in the plan-audit iteration 2 correction, §R11, at `9495566efa` in the plan-audit iteration 3 correction, §R12, and at `090b7d51f9` in the plan-audit iteration 4 correction, §R13 — §R13, §R12 and §R11 are placed above §R10 as the newest blocks) in worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640`, branch `WT-d2-moai-worktrees`, HEAD `74ad197380`, with read-only commands and file reads; no Go command was run. Items not observed are listed in §R8. Counts that move (worktree totals) are dated 2026-10-19 and re-measured at run-phase M0.

## §R1. Where trees live now (measured)

`git worktree list` at HEAD `74ad197380`, 251 lines, 0 prunable, 0 locked:

| Root | Count | Notes |
|------|-------|-------|
| `<primary>/.moai/worktrees` | 192 | 167 listed under `/Users/goos/MoAI/…`, 25 under `/Users/goos/moai/…` (lowercase). `stat -f '%i'` on `/Users/goos/moai` and `/Users/goos/MoAI` both printed 105929094 — one directory, two spellings (case-insensitive volume). 188 have a `WT-…` branch; 4 do not (`rc30-preview`, `release-v314`, `t1515-absorb`, `t1868-publish`). Includes this tree (`t1640`). |
| `<primary>/.claude/worktrees` | 46 | 31 named `agent-<hex>`; 13 moai-made with `WT-` branches (`agents-md-primary`, `t1241`, `t1356`, `t1452`, `t1459`, `t1497`, `t1502`, `t1507`, `t1508`, `t1510`, `t1511`, `t1513`, `t1868-publish-cc`); `audit-gate-async-worktree-doc` (branch `worktree-audit-gate-async-worktree-doc`); `main` (integration, branch `main`) |
| `~/.codex/worktrees` | 12 | Codex-app-managed (`pr-ci-codemaps-fix` … `pr1787-tests-fix`) |
| primary checkout | 1 | `/Users/goos/MoAI/moai-adk-go`, detached HEAD `478285ab4c` |
| `~/.moai/worktrees` (home) | 0 registered | `ls ~/.moai/worktrees` shows seven directories (`claude.mo.ai.kr`, `mink`, `mo-ai-kr`, `mo.ai.kr`, `MoAI-ADK`, `moai-adk-go`, `moai-studio`), all empty or near-empty (`mink/SPEC-MINK-BRAND-RENAME-001`, `MoAI-ADK/fix` and `.moai-worktree-registry.json` of 2 bytes) |

Operator figures in decision `cadf` (191 and 46) differ from this session's 192 by one tree: this card's own worktree, created after that measurement. Sum 192 + 46 + 12 + 1 = 251 ✓.

Same filesystem device: `stat -f '%d' . ~/.moai/worktrees /Users/goos/MoAI/moai-adk-go/.git` printed 16777230 for all three, so `git worktree move` is a rename on this machine; cross-volume behavior is unmeasured (PM-4).

## §R2. Creation paths (code)

- `internal/cli/session_worktree.go:48` `const sessionWorktreeSubdir = ".moai" + sep + "worktrees"`; `:223` `destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` inside `materializeSessionWorktree` (`:213-266`), which runs `git worktree add -b <branch> <dest> [<base>]` (`:276-299`), then `init.defaultBranch`, `applyWorktreeGitConfig` (global `safe.directory`, identity), and `sessionWorktreeSeedCodexHooks` (`:236-264`).
- `internal/cli/root.go:139` `worktree.WorktreeCreator = materializeSessionWorktree` — the single wiring for `moai worktree new` (`internal/cli/worktree/new.go:41-52`, long text names `.moai/worktrees/<name>` at `:21`) and for `factoryEnsureCardWorktree` (`internal/cli/factory_card.go:700-770`; its `@MX:ANCHOR` states four callers: `factory_card`, `mcp_factory_card`, `factory_lane_relaunch`, `codex_launcher`). `factoryCreateAndRenameWorktree` (`:761-792`) serializes on `factory.AcquireFactoryStepLock` and renames the branch to `WT-<slug>`.
- `internal/cli/worktree_branch_flag.go:109` `treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)` — launcher `-w <name> --branch <existing>` path (`:93-124`), registers via `hook.RegisterLauncherWorktree` (`internal/hook/worktree_registry.go:65`).
- Bare `moai cc -w <name>`: `internal/cli/launcher.go:911-914` documents that a short name is rewritten to `--worktree <name>` and "claude then resolves the name against .claude/worktrees/<name>"; `cc.go:41` and `glm.go:61` help text say the same.
- `internal/hook/worktree_create.go:40-41` `agentWorktreeParentDir = filepath.Join(".claude", "worktrees")` — the shipped WorktreeCreate handler; unregistered by design (`.claude/rules/moai/workflow/worktree-integration-ops.md`, section "WorktreeCreate and WorktreeRemove Hooks (Not Registered by Default)", lines 320-334 in this tree).
- Configured base branch: `session_worktree.go:236-239` via `config.LoadWorktreeBaseBranch` (`internal/config/loader_worktree_base.go:30`), key `git_strategy.worktree_base_branch` (`types.go:176-186`).
- Dead config key: `internal/config/types.go:202` `WorktreeRoot string yaml:"worktree_root"`, `.moai/config/sections/git-strategy.yaml:108` `worktree_root: ""`, default `internal/config/defaults.go:1392`; `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001/spec.md:50,134` records it as a dead field with zero readers. A grep for `WorktreeRoot|worktree_root` over non-test Go found only the type, default and these SPEC/doc mentions.
- Home root accessor already exists: `internal/paths/paths.go:88` `WorktreesDir()` returns `<MoaiHome>/worktrees`; `MoaiHome()` honors an absolute `MOAI_HOME` (`:67-77`). The package is stdlib-only by design (`:25-30`).

## §R3. Readers that identify a tree by path shape (the migration hazard)

| # | Site | What it assumes | Effect once a tree is at `<home>/worktrees/projects/<key>/<name>` |
|---|------|-----------------|---------------------------------------------------------------------|
| 1 | `internal/hook/session_start_record.go:226-241` `cardIDFromPath` | card = first ancestor whose parent is named `worktrees` | resolves `projects` as the card id |
| 2 | `internal/hook/session_start_factory.go:231-259` `cardIDFromRegisteredWorktree` | `worktrees` whose parent is `.moai` or `.claude` | resolves no card (new root is neither) — falls through to "no card tree" |
| 3 | `internal/cli/todo_auto.go:88-93` `autoOwnerWorktrees` | owner trees = `<root>/.claude/worktrees/<card>` and `<root>/.moai/worktrees/<card>` | a live owner at the home root is judged not alive; takeover risk |
| 4 | `internal/factory/status_read.go:63-69` `DefaultWorktreeBases`, loop `:117-143` | bases `<primary>/.claude/worktrees` and `~/.moai/worktrees`; tree = `<base>/<specID>` | `<home>/worktrees/<specID>` does not exist; the project/key level is missed |
| 5 | `internal/cli/launch_chain.go:60-72`, `internal/cli/factory_anchor_guard.go:98-106`, `internal/cli/codex_hooks_seed.go:61-75`, `internal/cli/session_worktree.go:1010-1033` (`ccWorktreeWriterPrecheck`) | short name → `<root>/.claude/worktrees/<name>` | short name never resolves to a home-root tree |
| 6 | `internal/cli/codex_launcher.go:474-519` `resolveCodexWorktreeDir` | two in-repo candidates `.moai/worktrees` and `.claude/worktrees`; ambiguity message for both | `moai codex -w <name>` cannot find a home-root tree by name |
| 7 | `internal/cli/launcher.go:940-972` `validateWorktreeL2Value` | accepted absolute prefixes: `paths.WorktreesDir()` plus two in-repo roots | home-root path is under `WorktreesDir()` and is accepted — but the error text names the old prefixes |
| 8 | `internal/cli/launcher.go:449-537` `cleanupMoaiWorktrees` | each direct child of `~/.moai/worktrees` is a base; removes `worker-*` trees by non-force `git worktree remove` (`removeWorktree`, `:553-556`, called `:518`) on every `moai cc` launch; the match is `filepath.Rel(base, worktreePath)` rejecting only `..` (`:505-518`) | `projects` becomes a base, and a nested path `projects/<key>/worker-x` **is accepted** (corrects the first draft, which said it would not be found); a migrated `worker-*` tree with unpushed commits and a clean state would be removed on the next launch. Today `git worktree list` shows 0 `/worker-` trees (measured at `e5950b354f`), so no present exposure |
| 9 | `internal/cli/worktree/done.go:230-247` `isL1WorktreePath`; `internal/cli/worktree/sweep.go:574-580` `sweepTierOf` | L1 = under `<main>/.claude/worktrees` or `<main>/.moai/worktrees`, else L2 | a moved session tree is classified L2 (disposal-eligible class); `done` would accept it |
| 10 | `internal/cli/update_worktree_migration.go:87,105-108` | selects every registered tree under `<root>/.claude/worktrees`; `grep -n 'agent-' internal/cli/update_worktree_migration.go` printed nothing, i.e. no agent exclusion | conflicts with dab5 (moves `agent-*` trees); destination is `<root>/.moai/worktrees` (`:88`) |
| 11 | `internal/homestate/card_worktree.go:19-60` `RecordCardWorktree` (store `internal/homestate/factory.go:66` `worktree_path TEXT`) | refuses any change of a recorded path (`:42-44`) | a moved tree leaves the card record naming the old path; readers `verifyPushed(ctx, cur.WorktreePath, …)` etc. (`card_transition.go:614,708,716,726,751`, `card_pr_states.go:68,74,78,82`, `card_evidence_readers.go:188,194,203,222`, `card_audit_kickoff.go:24,31`) operate on the old path |
| 12 | `internal/homestate/card_worktree.go:73-92` `PreviousCardWorktree` | newest binding from another run | stale path carried into a replacement run unless every run row is rebound |
| 13 | `internal/cli/factory_card.go:499` `factoryCardWorktreeDir`; consumers `factoryRefuseForeignWorktree` (`:512`), `factoryCarryPreviousWorktree` (`:528`); callers `factoryNextClaim` (`:1152`, `:1162`) and `factoryNextValidate` (`:1450`, `:1455`) | landing directory = `<root>/.moai/worktrees/<card>` (via `sessionWorktreeSubdir`) | a home-root tree of the card is neither refused as foreign nor carried into a replacement run (t1521); found by plan-audit iteration 1 (D4) |

Not affected (checked): `internal/hook/post_tool_worktree.go:24-60` (`handleWorktreeMove` re-stamps from the new cwd), `internal/session/registry.go:44-60` (registry anchored to the primary via git common dir), `internal/statusline/quota_dirs.go:30-70` (reads `<common>/worktrees/*/gitdir`), `internal/cli/mcp_worktree_root.go:60-93` (accepts a linked worktree by git facts, not prefix), `internal/hook/branch_guard.go:1001` (`isPrimaryCheckout` by git-dir vs common-dir).

Other files naming the legacy paths (comments, help text): `internal/cli/hook.go:788` (a comment already describing the L2 layout), `internal/cli/spawn.go:16,198`, `internal/cli/doctor_served_model.go:41`, `internal/harness/rosterguard/check.go:346`, `internal/hook/quality/astgrep_gate.go:148,221` (scan exclusions), `internal/mx/scan_ignore.go:18`, `internal/mission/integration.go:41`.

## §R4. Guards and entry behavior (what is known, what is not)

- The worktree-isolation guard ("This session is isolated in the worktree …") belongs to the Claude Code runtime; the repository only classifies its refusal text. `internal/hook/post_tool_failure.go:40-48` states it verbatim per `SPEC-SESSION-ANCHOR-ATTR-001/research.md:24`; the classifier anchor is `isolated in the worktree`. During this planning session the guard refused five of my commands (a `sed` with a shell-variable operand, a `sed` whose program came from a loop variable, an `awk` program, and two compound commands that also contained `git`), each with the message "…cannot be shown not to be git / too complex to verify"; the same reads re-issued as plain separate commands passed. This corroborates `AGENTS.md §3` ("a refusal is the guard reading the command") and bears on acceptance-command authoring: criteria use plain single commands.
- CC changelog snapshot `.moai/research/cc-changelog-snapshot-2.1.300.md`: line 4231 "`EnterWorktree` now asks for confirmation before entering a git worktree outside the project's `.claude/worktrees/` directory"; 4974 "`EnterWorktree` can now switch between Claude-managed worktrees mid-session"; 5283 `worktree.bgIsolation: "none"`; 5540 `worktree.baseRef` (`fresh` | `head`); 6786 `worktree.sparsePaths`; 7215 "Added `WorktreeCreate` and `WorktreeRemove` hook events". Reading: no location setting exists; hooks replace native creation.
- `claude --version` printed `2.1.303 (Claude Code)` in this session (the ef12 note says 2.1.302); `codex --version` printed `codex-cli 0.163.0` (the memory reference was measured on 0.161.0).
- `EnterWorktree` behavior for a home-root tree: **not measured** (PM-1). t1611 body states: "from the primary checkout, an outside-path EnterWorktree succeeded without an approval prompt but was refused inside a worktree session; the official doc says approval is required".
- Codex facts (memory `reference_codex_in_session_worktree`, 8 days old, measured 2026-10-10 on 0.161.0, claims not re-verified here): `/cd <abs>` since 0.149 changes cwd and write roots, reloads AGENTS.md/`.codex`/hooks/permissions, refuses running work, background terminals, remote sessions, untrusted destinations and different permission profiles; managed worktrees pool at `$CODEX_HOME/worktrees` with detached HEAD from the remote default branch; writes outside the project need `--add-dir` or `writable_roots`; `git commit` in `workspace-write` is refused at `.git/index.lock` in any layout; `spawn_agent` takes no cwd. Used for PM-2 and PM-5 design only; no guidance sentence relies on it without a measurement.

## §R5. Guidance surfaces naming the old layout (scope of the t1610 part)

Counted with `grep -rl` over this tree (files, not occurrences; a file may be a template mirror of another):

| Surface | Files containing `moai cc -w` | Files containing `.claude/worktrees` or `.moai/worktrees` |
|---------|------------------------------|------------------------------------------------------------|
| `.claude/rules` | 13 | 7 |
| `.claude/skills` | 13 | 7 |
| `.claude/agents` | 2 | 1 |
| `.claude/output-styles` | 0 | not counted |
| `internal/template/templates` | 32 | 18 |
| `internal/cli` (Go sources, comments and help) | 24 | see §R3 |
| `docs-site/content` (ko/ja/zh/en) | 46 | not counted |
| repository root | `AGENTS.md`, `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `CHANGELOG.md` | — |

`AGENTS.md` is 23,780 bytes (`wc -c`), against the Codex budget 24,576 quoted in card t1610's body; headroom 796 bytes. Lines naming the topic: `AGENTS.md:38` (capability table `worktree-entry`), `:131-136` (§3). `.claude/rules/moai/workflow/worktree-integration.md` lines 13-15 (Overview), 41-42 (Glossary), 45, 51-57 carry the L1/L2 definitions and the `WT-` rules. `AGENTS.local.md` has 4 lines naming `worktrees`. The docs-site and README surfaces (46 + 5 files) are not in the allow-list of plan.md §D; see OQ-5.

## §R6. Safety machinery that already exists and is reused

- `planLegacyWorktreeMigrationFor` (`update_worktree_migration.go:76-147`) already refuses: prunable, locked, current process inside the tree, any active process cwd inside the tree (`activeProcessCWDs`, `processUsesWorktree` `:149-156`), unreadable session registry (`:64-72`), live anchored session (`session.LiveAnchoredSessionsForProject`), missing source, destination exists, symlinked destination parent (`:89-95`).
- `moveLegacyWorktree` (`:161-231`) already re-plans immediately before the move, runs `git worktree move`, re-lists, compares HEAD/branch/registered path and `git status --porcelain=v1 --untracked-files=all` before and after, and reverses with `git worktree move` back on mismatch. The new engine extends the comparison (unpushed commit list, ignored-file count, reports count) and adds journal, `--verify`, `--rollback`, card rebind.
- `hoistWorktreeReports` (`internal/cli/worktree/hoist.go:50-…`) copies `<tree>/.moai/reports` into the primary's `.moai/reports/worktrees/<tree-name>/` — evidence preservation before any disposal; migration does not dispose, so it does not call it, but the manifest counts the reports files so a move cannot silently lose them.
- Landing predicates `LandedByPatchID`, `LandedByCommitPatchIDs` (`internal/cli/worktree/landing_predicate.go:178,257`) and the session-exit guards (`session_worktree.go:651-733`) are untouched.
- `gitMainRootFromTargetFunc` (`done.go:186-205`) derives the main root from a target path via `git rev-parse --path-format=absolute --git-common-dir`, falling back to the first porcelain stanza; the same derivation is the model for the project key.

## §R7. Decision-board input read for this SPEC

`moai decision read --scope card:t1640` returned 135 records; the three that bind this card were read in full from the persisted output: `d-20261018T221252Z-cadf` (operator, AskUserQuestion), `d-20261018T221922Z-dab5` (operator, chat 10-19), `d-20261018T221839Z-ef12` (leader research note, "non-binding default, operator may override"). Their text is the source for the REQ/decision mapping in spec.md. Card bodies read with `moai todo show`: t1640 (picked), t1611, t1610, t1501 (all `dropped`, merged into t1640 on 2026-10-10). The leader candidate "core.hooksPath=/dev/null in the shared .git/config" is already in effect: `git config --show-origin --get-all core.hooksPath` printed `file:/Users/goos/MoAI/moai-adk-go/.git/config  /dev/null` from this tree. `.moai design §8.5` (named by the card): `ls -d .moai/design` printed "No such file or directory"; `grep -rln "8\.5" .moai/docs` matched two files, `harness-delivery-strategy.md` (line 226 "### 8.5 Tier system 비변경 (REQ-CGA-006)") and `git-workflow-doctrine.md` (line 33 and 163, "§18.5 Hotfix Workflow"), neither about worktree location; `grep -rl "worktrees/projects\|moai-32-master" .moai/docs .moai/project` matched nothing. The document could not be located in this tree (Gap G-3). `.moai/reports/moai-32-master-upgrade-plan-20261010.html`: `ls .moai/reports | grep -ci "master-upgrade"` printed 0 over a 483-entry listing, so §2 stage 2 of the plan could not be read (Gap G-4; `.moai/reports` is gitignored and machine-local, so the file may exist in another checkout). The card text itself carries the scope sentence used here.

## §R8. Not measured (Gaps) — each maps to an item in spec.md §F or a decision

| Id | Not measured | Where handled |
|----|--------------|---------------|
| G-1 | `EnterWorktree` into a home-root tree (primary / in-worktree; default / bypass) | PM-1a..PM-1d (pre-kickoff gate) |
| G-2 | Codex `/cd` into a home-root tree on 0.163.0 | PM-2 |
| G-3 | `.moai design §8.5` — document not found in this tree | OQ list; the plan does not rely on it |
| G-4 | Master upgrade plan HTML §2 stage 2 — file not present in this tree | the card text is the scope source |
| G-5 | `claude --worktree <absolute outside path>` creation semantics, hooks, trust, project-memory key | PM-3, PM-6 |
| G-6 | Cross-volume `git worktree move` | PM-4 |
| G-7 | Codex sandbox commit behavior in the new layout | PM-5 |
| G-8 | Windows behavior of the resolver, marker and move | OQ-4 |
| G-9 | Whether each of the 238 in-repo trees' branches exists on a remote (the "only copy" claim is the operator's invariant, not re-counted here) | M0 dry-run listing records `unpushed commit count` per tree; the plan relies on the invariant, not on a count |
| G-10 | Whether `todo --auto`, lease renewal or other lane machinery reads a tree path from a source not found by the sweep in §R3 | standing grep sweep AC-WHR-007; Residual-risk R-1 |
| G-11 | Callers of the consolidation — traced in the 0.3.0 correction: `runUpdateWorktreeMigration` is called at `update.go:366` (`dryRun=true`, inside the `--dry-run` branch) and `update.go:410` (`dryRun=false`, after the dry-run return and under the update lock) | closed; see §R10 |
| G-12 | Whether any running Claude Code lane's project-memory key survives a tree move | PM-6 |
| G-13 | That git deletes a worktree's administrative directory, and the instance token file in it, when it removes or prunes the tree, and that it reuses the admin id for the next tree of that name (the premise of REQ-WHR-012 / design D-11 identity); not run — a scratch repository sits outside this tree, and AGENTS.md §3 records that the guard refuses git against any path outside it (not attempted here) | AC-WHR-012b step (9d) asserts both at M1/M6; if git kept the directory, the token would match a recreated tree and the run phase returns a blocker |
| G-14 | Cross-process behavior of `internal/lockfile` on Windows (a process-local mutex there) and the bounded wait under a stalled holder | OQ-4; AC-WHR-017b step (7) on the platforms that provide the lock |

## §R9. Residual-risk summary

- R-1 A reader not found by §R3 exists. Mitigation: standing sweep, recorded-path precedence, dual-read.
- R-2 The routine update meets live trees and skips them, so the migration is partial on a busy machine. Mitigation: a skipped tree stays usable at its old path, a later update retries it (`update.go:407-409`), and the move list records each skip with its reason.
- R-3 `EnterWorktree`/`/cd` differ from the plan's assumption. Mitigation: PM gate and stated branches.
- R-4 The classifier leaves `audit-gate-async-worktree-doc` unclassified (one known tree). Operator may reclassify by hand.
- R-5 The lowercase-spelling alias could also affect the recorded `worktree_path` of cards created through the lowercase spelling; the rebind compares via spelling-safe equality (REQ-WHR-023).
- R-6 A lane holds a pre-move path after the routine update moved a parked tree. Mitigation: card record rebind first, move list second (REQ-WHR-011/012); the move-list file is machine-local (gitignored), so a reader on another machine has no copy of it; that case is not designed for here.

## §R15. Measurements for the plan-audit iteration 6 correction (version 0.8.0; tree `6f63a1494b11469cb8109caf0917e87b5a5ceb2f`)

All by read-only commands, no Go command. Judging build for the one CLI cell: `moai version` printed `moai_cp/20261007_195927-1888-g7705b365e2`; `git merge-base --is-ancestor 7705b365e2 HEAD` exited 0 (no output), so that build is a strict ancestor of the tree and its CLI result describes older code; every other cell is a grep, `ls` or `wc`, independent of that binary. As in §R13, the harness prints no status line for a `grep` that exits 1, so those codes are derived from the stdout. The first run of the four `--exclude` grep cells with the pattern unquoted failed on a zsh glob (`no matches found`), not on the worktree guard; they were re-run with the pattern quoted, and the table shows those runs. `git diff --stat 49951ae514 6f63a1494b` listed the five SPEC artifacts only (103 insertions, 46 deletions), so the document pin of acceptance.md moved from `49951ae514` to this tree with only SPEC files differing.

| AC | Command | Observed stdout | Exit |
|----|---------|-----------------|------|
| 001 | `grep -rln --exclude='*_test.go' 'worktrees/projects' internal` | (empty) | 1 |
| 002a | `grep -rln --exclude='*_test.go' 'moai-project.json' internal` | (empty) | 1 |
| 004 | `grep -rn --exclude='*_test.go' 'worktree_project_key' internal .moai/config` | (empty) | 1 |
| 005a | `grep -n 'destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)' internal/cli/session_worktree.go` | `223:	destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` | 0 |
| 005b | `grep -n 'filepath.Join(projectRoot, ".claude", "worktrees", name)' internal/cli/worktree_branch_flag.go` | `109:	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)` | 0 |
| 010 | `moai worktree list --json` | `ERROR  Unknown flag: --json.  Try --help for usage.` | 1 |
| 012b | `grep -rln --exclude='*_test.go' "move-list\|movelist\|move_list" internal` | (empty) | 1 |
| 013 | `ls internal/cli/worktree/classify.go` | `ls: internal/cli/worktree/classify.go: No such file or directory` | 1 |
| 019 | `grep -c 'Join(root, sessionWorktreeSubdir)' internal/cli/update_worktree_migration.go` | `1` | 0 |
| 021 | `ls .moai/reports/t1640/pre-kickoff-measurements.md` | `ls: .moai/reports/t1640/pre-kickoff-measurements.md: No such file or directory` | 1 |
| 025 | `grep -c 'worktrees/projects' AGENTS.md` | `0` | 1 |
| 025 baseline | `wc -c AGENTS.md` | `23780 AGENTS.md` | 0 |
| 025 template | `grep -c 'worktrees/projects' internal/template/templates/AGENTS.md.tmpl` | `0` | 1 |
| 024 (RG) | `grep -n 'WorktreeCreate' .claude/settings.json` | (empty) | 1 |

Every cell reproduces the §R14 value, and the plan-audit iteration 6 report's MP-8 table (same commands, same tree).

Kickoff precondition (D2): `internal/homestate/card_audit_kickoff.go:31-38` was read: the factory Kickoff edge calls `auditReadyRecorded(<spec dir>/progress.md)`, which cuts the file at `## §E.1`, reads that section up to the next `\n## ` and accepts only a bare `audit_ready: true` line (a look-alike key is named in the refusal), so the edge refuses without the file. `ls .moai/specs/SPEC-WORKTREE-HOME-ROOT-001` printed `acceptance.md design.md plan.md research.md spec.md`: `progress.md` is absent at this tree. The schema section map (`spec-frontmatter-schema.md`) gives `§E.1` to the plan owner and `§E.2` to the run-phase specialist, which fixes who writes what: the plan owner emits the file at the plan-audit PASS, M0 appends `§E.2` (plan §B, §C). The earlier sentence that M0 creates the file is withdrawn (edited in place in §R14).

Gate check, iteration 6 shape (D4). The first character of COMMAND, OUTPUT and TOOLS is `[^[:space:];]` in place of `[^ ]`: it rejects a tab and a lone `;`, and unlike `[[:alnum:]]` it keeps a pointer such as `@PM-1c-output` (fixture g03, PM-1c) valid. Two fixtures added next to g01 to g14 in `…/scratchpad/gate5/`: `g15-tab.md` (fixture (iii) with the command, output and tool-version fields of `PM-1b` each a single tab; `grep -c -E` of a tab character printed `1`, and a `[[:space:]]+` form of the `PM-1b` line printed `1`, so the tab bytes are in the file) and `g16-semicolon.md` (the same fields of `PM-1c` each a lone `;`). Red observed on a known failing input: the iteration 5 form (first character `[^ ]`) printed `1` for the `PM-1b` complete-record count on g15 and `1` for the `PM-1c` count on g16, so it passes both. The new form, run as separate plain commands over all sixteen fixtures (one command per id), printed these complete-record counts (fixture: 1a 1b 1c 1d 2 3):

| Fixture | 1a | 1b | 1c | 1d | 2 | 3 |
|---------|----|----|----|----|---|---|
| g01 all GAP | 0 | 0 | 0 | 0 | 0 | 0 |
| g02 one GAP (1c) | 1 | 1 | 0 | 1 | 1 | 1 |
| g03 valid | 1 | 1 | 1 | 1 | 1 | 1 |
| g04 six identical 1a | 6 | 0 | 0 | 0 | 0 | 0 |
| g05 2 twice, no 3 | 1 | 1 | 1 | 1 | 2 | 0 |
| g06 1a PASS and FAIL, no 1b | 2 | 0 | 1 | 1 | 1 | 1 |
| g07 valid + PM-4 GAP | 1 | 1 | 1 | 1 | 1 | 1 |
| g08 valid + 1a ERROR | 1 | 1 | 1 | 1 | 1 | 1 |
| g09 1c ERROR only | 1 | 1 | 0 | 1 | 1 | 1 |
| g10 six bare lines | 0 | 0 | 0 | 0 | 0 | 0 |
| g11 1b command blank | 1 | 0 | 1 | 1 | 1 | 1 |
| g12 1c other SHA | 1 | 1 | 0 | 1 | 1 | 1 |
| g13 2 output blank | 1 | 1 | 1 | 1 | 0 | 1 |
| g14 3 exit absent | 1 | 1 | 1 | 1 | 1 | 0 |
| g15 1b tabs | 1 | 0 | 1 | 1 | 1 | 1 |
| g16 1c lone `;` | 1 | 1 | 0 | 1 | 1 | 1 |

Reading: g15 and g16 each fail exactly the id they damage (`0`), the fourteen earlier fixtures give the values of §R14 (g03 and g07 pass with every count `1`; g08's second line shows only in the all-lines count of `PM-1a`, `2`), and the tab and the lone `;` are no longer passable. The all-lines counts (unchanged form) were re-run on g15 and g16 for `PM-1b` and `PM-1c` only (`1` each); they were not re-run on the other fourteen fixtures, whose files and counts are unchanged since §R14 (Gap). Residual: the shape proves that each of the five fields is present and does not start with whitespace or `;`, not that it is true; a value such as `-` passes, and the Kickoff reader still reads the record (spec.md §F, plan §B). The residual of §R14 about a lone `;` no longer holds.

By-hand ordering check, re-done for version 0.8.0 (the table of §R13 below was edited in place for M1). Additions and the milestone each is built in, with the milestone whose tests first name it: `MovedBy` decided from the tree's latest location line, with the `pending` and `refused` cases (built M1; named by the M1 fixtures, which are hand-written move-list lines and trees, and by the M3 cleanup-guard tests for `worker-pending`, `worker-refused`, `worker-premove`; the M6 writer that emits those lines is built later and no earlier test names it); the no-token lookup clause (reader built M1; named by M1 fixtures); `progress.md` with `§E.1`, emitted by the plan owner before the Kickoff, and `§E.2` appended from M0 (a documentation file, no symbol); the `spec.md` `draft -> in-progress` flip in the first production-file commit of M1 (a frontmatter edit, no symbol; M0 and M1's RED commit leave `spec.md` untouched); the regenerated `internal/template/catalog.yaml` hashes in M7 (generated, consumed by the repository's mirror test at M7's exit); the gate regex (M-1, before any milestone). No test names a symbol built in a later milestone, M0 still commits no Go file and touches `progress.md` only, and every milestone still ends GREEN on its own. Result: no ordering conflict found.

## §R14. Measurements for the plan-audit iteration 5 correction (version 0.7.0; tree `49951ae51488544f109be7805ea804c4319c6c79`)

All by read-only commands, no Go command. Judging build for the one CLI cell: `moai version` printed `moai_cp/20261007_195927-1888-g7705b365e2`; `git merge-base --is-ancestor 7705b365e2 HEAD` exited 0 (no output), so that build is a strict ancestor of the tree and its CLI result describes older code; every other cell is a grep, `ls` or `wc`, independent of that binary. As in §R13, the harness prints no status line for a `grep` that exits 1, so those codes are derived from the stdout; the `ls` and `moai` cells showed `Exit code 1` themselves.

| AC | Command | Observed stdout | Exit |
|----|---------|-----------------|------|
| 001 | `grep -rln --exclude='*_test.go' 'worktrees/projects' internal` | (empty) | 1 |
| 002a | `grep -rln --exclude='*_test.go' 'moai-project.json' internal` | (empty) | 1 |
| 004 | `grep -rn --exclude='*_test.go' 'worktree_project_key' internal .moai/config` | (empty) | 1 |
| 005a | `grep -n 'destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)' internal/cli/session_worktree.go` | `223:	destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` | 0 |
| 005b | `grep -n 'filepath.Join(projectRoot, ".claude", "worktrees", name)' internal/cli/worktree_branch_flag.go` | `109:	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)` | 0 |
| 010 | `moai worktree list --json` | `ERROR  Unknown flag: --json.  Try --help for usage.` | 1 |
| 012b | `grep -rln --exclude='*_test.go' "move-list\|movelist\|move_list" internal` | (empty) | 1 |
| 013 | `ls internal/cli/worktree/classify.go` | `ls: internal/cli/worktree/classify.go: No such file or directory` | 1 |
| 019 | `grep -c 'Join(root, sessionWorktreeSubdir)' internal/cli/update_worktree_migration.go` | `1` | 0 |
| 021 | `ls .moai/reports/t1640/pre-kickoff-measurements.md` | `ls: .moai/reports/t1640/pre-kickoff-measurements.md: No such file or directory` | 1 |
| 025 | `grep -c 'worktrees/projects' AGENTS.md` | `0` | 1 |
| 025 baseline | `wc -c AGENTS.md` | `23780 AGENTS.md` | 0 |
| 025 template | `grep -c 'worktrees/projects' internal/template/templates/AGENTS.md.tmpl` | `0` | 1 |
| 024 (RG) | `grep -n 'WorktreeCreate' .claude/settings.json` | (empty) | 1 |

Every cell reproduces the §R13 value (the plan-audit iteration 5 report re-executed the same cells on this tree). Document pin of acceptance.md moved from `090b7d51f9` to this tree; only SPEC files differ between the two (it moved on to `6f63a1494b` in the iteration 6 round, §R15).

Baseline carrier (D9): `git check-ignore -v .moai/reports/t1640/pre-kickoff-measurements.md` printed `.gitignore:235:.moai/reports/*	.moai/reports/t1640/pre-kickoff-measurements.md`, so card reports are not tracked; `git check-ignore -v .moai/specs/SPEC-WORKTREE-HOME-ROOT-001/progress.md` printed nothing and exited 1, so `progress.md` is tracked-eligible. The file itself does not exist in the SPEC directory at this tree; the plan owner emits it with `§E.1` at the plan-audit PASS, before the Kickoff, and M0 appends the `§E.2` lines to it (this round first wrote "M0 creates it"; the factory Kickoff edge reads `§E.1`, so that was withdrawn in §R15; plan §B, §C).

Gate check, iteration 5 shape (D4). Fixtures `g01` to `g14` in the session scratch area `…/scratchpad/gate5/`, one per fixture of AC-WHR-021 ((i) to (xiv)); `<audited>` is the sha written in them, `49951ae51488544f109be7805ea804c4319c6c79`. Check forms: all-lines `grep -c '^RESULT PM-1a:' <f>`; complete record `grep -c -E '^RESULT PM-1a: (PASS|FAIL) ;; COMMAND: [^ ].* ;; OUTPUT: [^ ].* ;; EXIT: [0-9]+ ;; TOOLS: [^ ].* ;; TREE: <audited>$' <f>`. On `g03` (fixture (iii)) all twelve counts were run as separate commands and each printed `1`. On the other fixtures the per-id counts were read from an id listing (`grep -o -E '^RESULT PM-[0-9a-d]+:'` for all lines; the same `grep -o -E` with the complete-record shape for complete records) plus the targeted counts named; the twelve separate counts were not run on each of them (Gap).

| Fixture | Observed |
|---------|----------|
| g01 (i) all GAP | all-lines listing: the six ids once each; complete-record listing: empty |
| g02 (ii) five complete + `PM-1c` GAP | complete-record listing: `PM-1a`, `PM-1b`, `PM-1d`, `PM-2`, `PM-3`; `PM-1c` absent |
| g03 (iii) valid | all twelve counts `1` |
| g04 (iv) six identical `PM-1a` | all-lines listing: `PM-1a` six times; complete-record count `PM-1a` `6`, `PM-1b` `0` |
| g05 (v) `PM-2` twice, no `PM-3` | all-lines `PM-2` `2`; complete-record `PM-3` `0` |
| g06 (vi) `PM-1a` PASS and FAIL, `PM-1b` absent | all-lines `PM-1b` `0`; complete-record `PM-1a` `2` |
| g07 (vii) valid + `PM-4` GAP | all-lines listing: seven ids including `PM-4`; complete-record listing: the six gating ids once each |
| g08 (viii) valid + second `PM-1a: ERROR` line | all-lines `PM-1a` `2`; complete-record `PM-1a` `1` |
| g09 (ix) `PM-1c` replaced by `RESULT PM-1c: ERROR` | all-lines `PM-1c` `1`; complete-record `PM-1c` `0` |
| g10 (x) six bare `PASS` lines | complete-record listing: empty; all-lines `PM-1a` `1`; the iteration 5 form `grep -c '^RESULT PM-1a: \(PASS\|FAIL\)$'` `1`; the old single count `grep -c '^RESULT PM-\(1a\|1b\|1c\|1d\|2\|3\): \(PASS\|FAIL\)$'` `6` |
| g11 (xi) `PM-1b` command blank | all-lines `PM-1b` `1`; complete-record `PM-1b` `0`; the other five ids listed once |
| g12 (xii) `PM-1c` tree = `090b7d51f945f7ce78c753f39312d68f93829e55` | all-lines `PM-1c` `1`; complete-record `PM-1c` `0`; the other five ids listed once |
| g13 (xiii) `PM-2` output blank | complete-record `PM-2` `0`; the other five ids listed once |
| g14 (xiv) `PM-3` exit field absent | all-lines `PM-3` `1`; complete-record `PM-3` `0` |

Reading: the bare-line record (g10) passes every check of iteration 5 (the audit's D4) and fails the complete-record count for all six ids; each of the four single-field mutants fails exactly the id it damages. Residual of that round: the shape proved the five fields present and non-blank, not true; its first-character class `[^ ]` passed a tab and a lone `;`, which §R15 closes with `[^[:space:];]`.

Package-name search (iteration 5 M1): pattern `sys[c]all` (the bracketed letter keeps this sentence from matching itself), `grep -c 'sys[c]all' <file>` over spec.md, plan.md, acceptance.md, design.md and research.md after the edits: `0` for each of the five, and the same search with the letter unbracketed (the form the plan-audit report uses) printed `0` for each (no section of the SPEC names the package, so no section needs a `//go:build` split; the §R13 sentence that printed `1` for research.md was this file's own report of the search).

By-hand ordering check, re-done for version 0.7.0 (the table of §R13 below was edited in place for M1, M5 and M6). Additions and the milestone each is built in, with the milestone whose tests first name it: requester-identity parameter and expected-identity rule of `MovedTo` (built M1; named by M1 fixtures, M3 readers, M6 end-to-end); `MovedBy` keyed on the tree's own `ident` (built M1; named by M1 fixtures and the M3 cleanup guard); journal `ident` and the identity check before every reverse move (built M6 from `TreeIdent` of M1; named by the M6 RED commit only); lapse error and `stateLockLapseBudget` (built M5; named by M5 tests for the error and constant, by M6 for the run ending); the complete-record gate check (M-1, before any milestone); the `progress.md §E.2` carrier (M0 onward, a documentation file, no symbol). No test names a symbol built in a later milestone: every addition is consumed by the same or a later milestone than the one that builds it, so M0 still commits no Go file and every milestone still ends GREEN on its own. The RED commit of each milestone holds test files and `progress.md` lines only. Result: no ordering conflict found.

## §R13. Measurements for the plan-audit iteration 4 correction (version 0.6.0; tree `090b7d51f945f7ce78c753f39312d68f93829e55`)

All by read-only commands, no Go command. The harness prints no status line for a command that exits 1 when it also prints nothing or a bare count (`grep`); those codes below are derived from the stdout, as in §R12, while the `ls` and `moai` cells showed `Exit code 1` themselves. Judging build for the one CLI cell: `moai version` printed `moai_cp/20261007_195927-1888-g7705b365e2`; `git merge-base --is-ancestor 7705b365e2 HEAD` exited 0 (no output), so the build is a strict ancestor of the tree and its CLI result describes older code; every other cell is a grep, `ls` or `wc`, independent of that binary.

| AC | Command | Observed stdout | Exit |
|----|---------|-----------------|------|
| 001 | `grep -rln --exclude='*_test.go' 'worktrees/projects' internal` | (empty) | 1 |
| 002a | `grep -rln --exclude='*_test.go' 'moai-project.json' internal` | (empty) | 1 |
| 004 | `grep -rn --exclude='*_test.go' 'worktree_project_key' internal .moai/config` | (empty) | 1 |
| 005b | `grep -n 'filepath.Join(projectRoot, ".claude", "worktrees", name)' internal/cli/worktree_branch_flag.go` | `109:	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)` | 0 |
| 005a (new cell) | `grep -n 'destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)' internal/cli/session_worktree.go` | `223:	destDir := filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` | 0 |
| 010 | `moai worktree list --json` | `ERROR  Unknown flag: --json.  Try --help for usage.` | 1 |
| 012b | `grep -rln --exclude='*_test.go' "move-list\|movelist\|move_list" internal` | (empty) | 1 |
| 013 | `ls internal/cli/worktree/classify.go` | `ls: internal/cli/worktree/classify.go: No such file or directory` | 1 |
| 019 | `grep -c 'Join(root, sessionWorktreeSubdir)' internal/cli/update_worktree_migration.go` | `1` | 0 |
| 021 | `ls .moai/reports/t1640/pre-kickoff-measurements.md` | `ls: .moai/reports/t1640/pre-kickoff-measurements.md: No such file or directory` | 1 |
| 025 | `grep -c 'worktrees/projects' AGENTS.md` | `0` | 1 |
| 025 baseline | `wc -c AGENTS.md` | `23780 AGENTS.md` | 0 |
| 024 (RG) | `grep -n 'WorktreeCreate' .claude/settings.json` | (empty) | 1 |
| 019 context | `grep -n 'runUpdateWorktreeMigration' internal/cli/update.go` | `366:		if migrationErr := runUpdateWorktreeMigration(lockRoot, true, out); migrationErr != nil {` and `410:	if migrationErr := runUpdateWorktreeMigration(lockRoot, false, out); migrationErr != nil {` | 0 |
| 019 | `grep -n 'sessionWorktreeSubdir' internal/cli/update_worktree_migration.go` | `88:	newRoot := filepath.Join(root, sessionWorktreeSubdir)` | 0 |
| 025 template | `grep -c 'worktrees/projects' internal/template/templates/AGENTS.md.tmpl` | `0` | 1 |

Every cell reproduces the previous round's value. The AC-WHR-005 cell split in 0.6.0: the launcher cell (005b) is the earlier cell unchanged; the 005a cell is new, with the same expected-green direction (exit 1, empty stdout once the destination comes from the resolver).

Lock primitive (D11): `ls internal/lockfile` printed `lockfile_contract_test.go lockfile_unix.go lockfile_unix_test.go lockfile_windows.go`. Read: `lockfile_unix.go` `Lock(f)` is a blocking exclusive `flock` (`LOCK_EX`, no non-blocking flag) and `Unlock(f)` releases it; `lockfile_windows.go` `Lock(f)` takes a process-local mutex keyed by file name and states "Multi-process locking is not supported", `Unlock` releases it. Neither file offers a try-acquire or a timeout, so the try-and-end migration lock and the 5-second bounded state-file wait need one added non-blocking `TryLock` per platform file (design D-12); `Lock`/`Unlock` stay as they are.

Launcher path (D2), `internal/cli/worktree_branch_flag.go`: `:23-24` states the branch must exist and the flag "never" creates one; `:109` builds `treePath` under `.claude/worktrees`; `:116` is `gitcore.NewWorktreeManager(projectRoot).Add(treePath, branch)`; `:122` is `hook.RegisterLauncherWorktree(...)`. The materializer's `WT-` rename, configured base and hook seeding are in `session_worktree.go:213-266`, not here.

A search of the five artifacts for the OS-call package name found no use of it in the SPEC body; the exact counts are in §R14 (the sentence that first reported this one counted itself, plan-audit iteration 5 M1).

### By-hand ordering check (D1; milestones in plan order, all five artifacts re-read; table and per-AC list kept current through version 0.8.0, re-done in §R14 and §R15)

Rule under test: no milestone's tests name a symbol first built in a later milestone; `go build` of every package, tests excluded, stays green at every commit; every milestone ends GREEN. "RED commit" = the test-only commit that opens the milestone.

| Milestone | Builds (non-test) | Tests in its RED commit name | Cells recorded in the RED commit | Packages that must pass at exit | Later symbol named? |
|-----------|-------------------|------------------------------|----------------------------------|----------------------------------|---------------------|
| M-1 | no Go; the measurement record | — | AC-021 (record) | — | no |
| M0 | no Go file; evidence only | — (no test) | observed cells: AC-001, 002a, 004, 005a, 005b, 010, 012b, 013, 019 (first), 025 | every package as before (nothing changed) | no |
| M1 | `internal/worktreeroot`, `classify.go`, key assignment and marker, `git_strategy.worktree_project_key`, `movelist*.go` reader (`TreeIdent`, token writer, `MovedTo` taking the requester's identity, `MovedBy` keyed on the tree's own `ident` and its latest location line: `moved`, `pending` or `refused` at the path, released by `restored` or `reversed`) | M1 symbols; git, config, session registry (existing) | AC-002b (1-3), 003a, 003b, 023 (`Same`) | `internal/worktreeroot`, `internal/config`, `internal/cli/worktree` | no; move-list fixtures are hand-written lines, not the M6 writer |
| M2 | destination swap in `session_worktree.go`, `worktree_branch_flag.go`, `launcher.go` | resolver (M1); existing materializer and `RegisterLauncherWorktree` | AC-006 | `internal/cli` | no |
| M3 | reader changes (`internal/hook`, `internal/cli`, `internal/factory`, `done.go`, `sweep.go`), cleanup guard | resolver, classifier, `MovedBy` (M1); existing readers | AC-007, 008, 009, 014b, 022 | `internal/hook`, `internal/cli`, `internal/factory`, `internal/cli/worktree` | no |
| M4 | `list.go`, registration in `root.go` | classifier, `Same` (M1) | (AC-010 Go clauses; its observed cell is M0's) | `internal/cli/worktree` | no |
| M5 | card rebind (`internal/homestate`), state-file lock, atomic write, re-key, lapse error and `stateLockLapseBudget` (`internal/hook`), `TryLock` (`internal/lockfile`) | M5 symbols only | AC-011, 012a (re-key), 017b (state-file clauses) | `internal/homestate`, `internal/hook`, `internal/lockfile` | no |
| M6 | `migrate*.go` (journal `ident`, identity check before every reverse move, run ending on the first lapse), move-list writer, `update_worktree_migration.go`, `internal/session/anchor.go` | M1 (classifier, `TreeIdent`, token writer), M5 (rebind, re-key, `TryLock`, lapse error, budget constant) and M6 symbols | AC-002b (4), 012a (journal), 014a, 015, 016, 017a (identity clause), 017b (migration-lock, reversal and budget halves), 018 (identity entries), 019 (second), 023 (anchor) | `internal/cli/worktree`, `internal/cli`, `internal/session` | no |
| M7 | `retire*.go`, guidance text | M6 engine package (earlier); no new symbol after it | AC-020 | `internal/cli/worktree`, template mirror test | no |
| M8 | nothing | — | — | scoped runs, `GOOS=windows` build, lint | — |

Per-AC check, RED point then Green path (every RED point is at or before the first milestone on the Green path line, and no AC names a symbol of a milestone after its RED point): 001 M0/M1; 002a M0/M1; 002b M1+M6/M1,M6; 003a, 003b M1/M1; 004 M0/M1; 005a, 005b M0/M2; 006 M2/M2; 007, 008 M3/M3; 009 M3/M3; 010 M0/M4; 011 M5/M5; 012a M5+M6/M5,M6; 012b M0 (+ reader clauses in M1's tests, writer clauses in M6's)/M1,M6; 013 M0/M1,M3,M4,M6; 014a M6/M6; 014b M3/M1,M3,M6; 015, 016, 017a M6/M6; 017b M5+M6/M5,M6; 018 M6/M6; 019 M0+M6/M6; 020 M7/M7; 021 M-1/M-1; 022 M3/M3; 023 M1+M6/M1,M3-M6; 024 (RG)/M6; 025 M0/M7. One finding of the check, fixed in this round: AC-WHR-017b clauses (6) and (7) mix a file-unchanged half (M5) and a tree-reversal half (M6), so those cells split by half; AC-WHR-012a's journal clause is M6's. Cross-milestone consumers (the M1 classifier in M3, M4 and M6; `TreeIdent` and `MovedBy` in M3 and M6, `TreeIdent` also in M6's journal and identity check; the M5 rebind, re-key, lapse error and budget constant in M6 only; the resolver in M2 and M3) all consume a symbol from an earlier milestone.

## §R12. Measurements for the plan-audit iteration 3 correction (version 0.5.0; tree `9495566efa15331f10135469d86313e8b1c1733c`)

All by read-only commands in the session scratch area or the tree, no Go command. The harness prints no status line for a grep that exits 1 (it did print `Exit code 1` for `ls internal/cli/worktree/classify.go`), so a grep cell's code below is derived from its stdout.

- Grep cells with the test-file exclusion (D2), each printed nothing (stdout empty, stderr empty): `grep -rln --exclude='*_test.go' 'worktrees/projects' internal`; `grep -rln --exclude='*_test.go' 'moai-project.json' internal`; `grep -rn --exclude='*_test.go' 'worktree_project_key' internal .moai/config`; `grep -rln --exclude='*_test.go' "move-list\|movelist\|move_list" internal`. The same patterns without the exclusion printed nothing in the iteration 3 audit (exit 1), and the exclusion only narrows the file set, so these are red for the stated reasons, and a test file cannot turn them green.
- AC-WHR-019 cell (D2): `grep -n 'sessionWorktreeSubdir' internal/cli/update_worktree_migration.go` printed `88:	newRoot := filepath.Join(root, sessionWorktreeSubdir)`; `grep -c 'sessionWorktreeSubdir' internal/cli/update_worktree_migration.go` printed `1`; `grep -c 'Join(root, sessionWorktreeSubdir)' internal/cli/update_worktree_migration.go` printed `1` (exit 0). The old cell `grep -n 'newRoot := '` stays red after a correct edit that keeps the variable name; the new cell reads green (`0`, exit 1) only when the destination expression is gone.
- Unpushed-commit forms (D3): `git log --not --remotes --format=%h -n 3` printed nothing; `git rev-list --count --not --remotes` printed `0`; `git rev-list --count --branches --not --remotes` printed `899`; `git log --branches --not --remotes --format=%h -n 3` printed `a4c26c1e0d`, `9a3c01d6b4`, `ea31b214fb`. The no-positive-revision form is empty where the positive form lists 899 commits, which is the vacuity the iteration 3 audit reported. In this tree `git log HEAD --not --remotes --format=%h -n 3` and `git rev-list --count HEAD --not --remotes` now print nothing and `0`: `git branch -r --contains 9495566efa` printed `origin/WT-d2-moai-worktrees`, i.e. the branch was pushed after the audit, so the audit's three-SHA listing for `HEAD` no longer reproduces here; the form is correct and the AC fixture keeps two unpushed commits on purpose.
- Gate-check mutant (D9), fixtures `gate-valid.md` (six valid lines plus a non-gating `PM-4: GAP`), `gate-error-extra.md` (valid six plus a second line `RESULT PM-1a: ERROR`) and `gate-error-only.md` (`PM-1c` replaced by the single line `RESULT PM-1c: ERROR`) in the session scratch area. On `gate-error-extra.md`: the iteration 2 `PM-1a` PASS-or-FAIL count printed `1` and the gating GAP count printed `0` (the record passes that check); the new all-lines count `grep -c '^RESULT PM-1a:'` printed `2` and for `PM-1b` `1`. On `gate-error-only.md`: `grep -c '^RESULT PM-1c:'` printed `1` and `grep -c '^RESULT PM-1c: \(PASS\|FAIL\)$'` printed `0`. On `gate-valid.md`: all twelve counts (all-lines and PASS-or-FAIL for each of `PM-1a`, `PM-1b`, `PM-1c`, `PM-1d`, `PM-2`, `PM-3`) printed `1`.
- Not run: the state-file lost-update, the killed lock holder and the reuse-of-the-old-path cases (D1, D5, D6) need Go fixtures; they rest on code reading of `internal/hook/worktree_registry.go` (§R11) and the audit's Codex reproductions, and their cells are `RB*` recorded at M0.

## §R11. Measurements for the plan-audit iteration 2 correction (version 0.4.0; tree `bf78f5541af77860230364a8f3db655f1d1b88fb`)

All by read-only commands, no Go command. Installed `moai` build for the one CLI cell: `moai_cp/20261007_195927-1888-g7705b365e2` (`moai version`), an ancestor of the tree HEAD; every other cell is a grep or `ls`, independent of that binary.

- State-file write (D4): `grep -n 'worktreeMu\|func saveWorktreeEntries\|func loadWorktreeEntries\|os.WriteFile' internal/hook/worktree_registry.go` printed `22:var worktreeMu sync.Mutex`, the `Lock` calls at `:30` and `:72`, `95:func loadWorktreeEntries`, `122:func saveWorktreeEntries`, `137:	if err := os.WriteFile(path, data, 0o644); err != nil {` — a package mutex (one process) and a truncating write. The plan-audit report read `loadWorktreeEntries` returning an empty slice on a parse error at `:109-115`.
- Live-session containment (D5): `grep -n 'func pathWithinTree\|func LiveAnchoredSessionsForProject' internal/session/anchor.go` printed `55:func LiveAnchoredSessionsForProject` and `141:func pathWithinTree`.
- Template source of the guidance (D10): `grep -n -i 'worktrees' internal/template/templates/AGENTS.md.tmpl` printed lines 130 (`## 3. Worktrees`), 133 and 153; `grep -c 'worktrees/projects' internal/template/templates/AGENTS.md.tmpl` printed `0` (exit 1), and the same count on `AGENTS.md` printed `0` (exit 1); `wc -c AGENTS.md` printed `23780`.
- Gate-check mutants (D9), fixtures in the session scratch area, six lines or fewer each. The old single-count check `grep -c '^RESULT PM-\(1a\|1b\|1c\|1d\|2\|3\): \(PASS\|FAIL\)$'` printed 6 on all three of: six identical `PM-1a: PASS` lines; five distinct ids plus `PM-2` twice and no `PM-3`; a `PM-1a` PASS/FAIL pair with `PM-1b` absent — each a Gap that passes. The per-id counts printed: duplicate fixture `PM-1a` 6 and `PM-1b` 0; missing fixture `PM-3` 0 (`PM-2` 2); conflict fixture `PM-1a` 2 and `PM-1b` 0; a valid six-id fixture (plus a non-gating `PM-4: GAP`) printed 1 for each of the six ids and 0 for the gating GAP count.
- RED-now cells re-observed with exit codes are in acceptance.md; the `moai worktree list --json` run printed `Unknown flag: --json.` with exit 1 on the installed build.

## §R10. Measurements for the operator-decision correction (version 0.3.0; tree `40c2e127bf`)

- `grep -n "runUpdateWorktreeMigration" internal/cli/update.go` printed line 366 (`..., true, out)`) and line 410 (`..., false, out)`). Lines 360-366 sit in the `if getBoolFlag(cmd, "dry-run")` branch; lines 398-414 follow its return. `grep -n "dry-run" internal/cli/update.go` printed the flag definition at line 85 (`Bool("dry-run", false, "Show planned archive and install operations without modifying the filesystem")`).
- Failure handling of the apply path (read at `update.go:405-414`): `updateLedger.referencef("worktree migration failed: %v", migrationErr)` — reported, not returned. The comment at 407-409 states the call also runs when template sync is version-matched, "so a later update can retry trees that were locked or active on an earlier pass".
- Existing skip causes (read at `update_worktree_migration.go:118-139`): prunable, locked, current process inside the tree, an active process inside, unreadable session registry, an active session anchored in the tree, source missing or not a directory, destination already exists, destination cannot be inspected. The destination root is `newRoot := filepath.Join(root, sessionWorktreeSubdir)` at line 88 (`grep -n "newRoot := "` printed exactly that line); the selection is the `claudeNativeWorktreeSubdir` prefix at lines 87 and 105-108.
- State area: `ls .moai/state` in this tree printed `.gitkeep audit-multi audit-receipts context-usage github landed` (each tree has its own `.moai/state/`); `git check-ignore -v .moai/state/worktree-migrate/move-list.jsonl` printed `.gitignore:401:.moai/state/	.moai/state/worktree-migrate/move-list.jsonl` — machine-local.
- `grep -rln "move-list\|movelist\|move_list" internal` printed nothing (exit status not echoed; a no-match `grep` exits 1) — no move-list code exists today.
