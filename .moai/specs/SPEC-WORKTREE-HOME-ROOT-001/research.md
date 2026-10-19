# research.md — SPEC-WORKTREE-HOME-ROOT-001

> Evidence record. Everything here was observed in this planning session (and re-read at HEAD `e5950b354f` in the correction round) in worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1640`, branch `WT-d2-moai-worktrees`, HEAD `74ad197380`, with read-only commands and file reads; no Go command was run. Items not observed are listed in §R8. Counts that move (worktree totals) are dated 2026-10-19 and re-measured at run-phase M0.

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
| G-11 | `moai update` caller name/location for the consolidation (`runUpdateWorktreeMigration` is the function; its caller was not traced) | plan §D marks `internal/cli/update*.go` "to be confirmed" |
| G-12 | Whether any running Claude Code lane's project-memory key survives a tree move | PM-6 |

## §R9. Residual-risk summary

- R-1 A reader not found by §R3 exists. Mitigation: standing sweep, recorded-path precedence, dual-read.
- R-2 The apply step meets a live tree (skip) so the migration is partial on a busy machine. Mitigation: leader-operated, repeatable, per-tree.
- R-3 `EnterWorktree`/`/cd` differ from the plan's assumption. Mitigation: PM gate and stated branches.
- R-4 The classifier leaves `audit-gate-async-worktree-doc` unclassified (one known tree). Operator may reclassify by hand.
- R-5 The lowercase-spelling alias could also affect the recorded `worktree_path` of cards created through the lowercase spelling; the rebind compares via spelling-safe equality (REQ-WHR-023).
