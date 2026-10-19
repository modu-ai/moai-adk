# plan.md — SPEC-WORKTREE-HOME-ROOT-001

> Derived implementation plan. Ordered by decision-reversibility: the decisions most likely to change come first (path form and key, pre-run measurement gate and its failure branches, launcher and entry flow, stored-path rebind), mechanical rewrites last. No time estimates; priority and dependency ordering only. All code citations measured in worktree `.moai/worktrees/t1640` at HEAD `74ad197380`.

## §A. Context

- SPEC: `.moai/specs/SPEC-WORKTREE-HOME-ROOT-001/spec.md` (Tier L, 24 REQ / 24 AC). Card t1640, Class C, operator-approved, deferral lifted 10-19 (decision `cadf`).
- Binding decisions: `cadf` (layout, order, non-destructive), `dab5` (agent-* stay; no WorktreeCreate hook), `ef12` (non-binding defaults, mandatory pre-kickoff measurement).
- Order against sibling cards: t1791 (GH #1813) lands first, then this card absorbs its launcher behavior; t1879 and t1753 follow this card and rewrite the wider EnterWorktree guidance once against the new location.
- Methodology: `quality.yaml constitution.development_mode` governs; this plan is written for TDD with a characterization pass first for the readers that change behavior (M3).
- Class C plan audit by codex is required before the run phase; the lane runs it. Findings and verdict paths: `.moai/reports/t1640/`.

## §B. Pre-run measurement gate (blocks the run phase — REQ-WHR-021)

The plan makes no assumption about these results. The measurement owner is the lane in its own tree (or the leader), not this plan. Output goes to `.moai/reports/t1640/pre-run-measurements.md` with the command, the verbatim output, the tool versions (`claude --version` printed 2.1.303 in this planning session, `codex --version` printed codex-cli 0.163.0), and the tree SHA.

| Id | Measurement | Procedure (read-only apart from the entry itself) | PASS means | FAIL means and what changes |
|----|-------------|----------------------------------------------------|------------|------------------------------|
| PM-1a | `EnterWorktree` into a home-root tree from the primary checkout, default permission mode | Create a throwaway tree at the home root with the M2 materializer prototype (or `git worktree add` in a scratch repo — scratch repository only, never this one), then call `EnterWorktree(path)` | Enters without a prompt, or a prompt the operator accepts once | Prompt each time: lanes in bypass mode may behave differently (PM-1c); record and continue to PM-1c |
| PM-1b | Same, from inside another worktree session | Same, issued from a session already anchored in a different tree | Enters, or a refusal whose text is recorded verbatim | Refusal: in-session hopping to a home-root tree is unavailable; guidance (M7) says "exit to the primary, then enter" and the t1879 card owns the guard text. No design change to storage |
| PM-1c | Same, bypassPermissions mode (lane mode) | Same as 1a and 1b under bypass | Enters without prompt | Prompt or refusal in bypass: lanes cannot hop to a home-root tree in-session. **Design change:** lane card transitions use a terminal relaunch (`moai cc -w <abs>`), `factory-dispatch` guidance changes for lanes (t1879 scope), and D-5/REQ-WHR-006 is re-examined with the operator. Storage location is unchanged |
| PM-2 | Codex `/cd <abs>` into a home-root tree | In an idle Codex CLI session, `/cd` to the tree; record trust handling, AGENTS.md reload, `/pwd` | `/cd` succeeds or fails with a trust prompt the operator can satisfy once per tree | Refused as untrusted with no per-tree path: Codex entry guidance is `codex -C <abs>` only; `/cd` claim is dropped from the text |
| PM-3 | `claude --worktree <absolute path>` and a session started inside a home-root tree | Launch `claude` with an absolute path outside `.claude/worktrees` in a scratch repository; check hooks fire, statusline, trust, `.worktreeinclude` (expected not processed), project-memory key | Behaves as an existing tree | Backend rejects an absolute path to a new tree: REQ-WHR-006 narrows to "launcher creates the tree through the shared materializer, then enters it by absolute path" (this is what the plan already specifies), and the short-name-without-materialization path is removed; if even entry by absolute path degrades, escalate to the operator before M2 |
| PM-4 | `git worktree move` across filesystems | Scratch repository on a second volume if one exists; otherwise record "no second volume available" as a Gap | Move works or is refused with an error git reports | Cross-volume remains refused (REQ-WHR-016); no design change |
| PM-5 | Codex workspace-write behavior for a home-root tree | `codex exec` with `workspace-write` in a scratch tree at the home root, attempt a commit | Documented outcome | Same limitation as the 10-10 note: guidance says escalation or `--add-dir` is needed; no design change |
| PM-6 | Does a moved tree change the Claude project-memory key | Move a scratch tree, start `claude --continue` in the new path | Key unchanged or the loss is understood | Migration output carries an explicit residual-risk line; no design change |

A measurement that cannot be taken is recorded as a Gap in that file, never as a PASS (verification-claim-integrity §1.1).

## §C. Milestones (dependency order; priority labels instead of time)

Every milestone starts RED (a failing test or an observed-failing check) and ends GREEN with the verbatim output in the run-phase evidence. M0 is the only milestone that touches no production code.

### M0 — Baseline and measurements (Priority High; precedes everything)
- Record the RED-now cells for acceptance.md (commands and verbatim output at the run-phase start SHA), including the legacy-reader inventory re-count of research.md §R3.
- Execute §B measurements. The baseline artifact lands in its own commit before the first production change (verification-claim-integrity §2.3 ordering attribution).
- Exit: `.moai/reports/t1640/pre-run-measurements.md` exists; operator decisions D-1..D-9 are each confirmed or recorded as default-applied.

### M1 — Location resolver and config (Priority High; decisions D-1, D-8)
- New resolver in a small package (candidate: `internal/worktreeroot`, to be confirmed against the `internal/paths` import-light constraint noted at `internal/paths/paths.go:25-30`): `Root(repoRoot) (string, error)`, `PathFor(repoRoot, name)`, `Key(repoRoot)`, `Candidates(repoRoot, name)` (home first, then the two legacy roots), `Same(a, b)` spelling-safe equality (REQ-WHR-023). Calls `paths.MoaiHome()` only.
- Marker file read/write (`.moai-project.json`), collision selection and the corrupt-marker refusal (REQ-WHR-002/003).
- `git_strategy.worktree_project_key`: add the field beside `WorktreeBaseBranch` (`internal/config/types.go:176-186`), default `""`, loader helper modeled on `LoadWorktreeBaseBranch` (`internal/config/loader_worktree_base.go:30`), template mirror in `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`. The dead `worktree_root` key (`types.go:202`, `git-strategy.yaml:108`) is left untouched (D-8).
- Exit: resolver unit tests including the same-name-different-repo collision, corrupt marker, `MOAI_HOME` relocation, case-insensitive spelling equality.

### M2 — Creation paths (Priority High; decision D-5, PM-3)
- `materializeSessionWorktree` (`internal/cli/session_worktree.go:213-266`) computes `destDir` through the resolver instead of `filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` (line 223). All callers inherit: `worktree.WorktreeCreator` is wired once at `internal/cli/root.go:139`; the four callers of `factoryEnsureCardWorktree` (`factory_card.go:700-715`) and `moai worktree new` (`internal/cli/worktree/new.go:41-52`) need no per-caller change. The guard-and-notice behavior of lines 224-266 is preserved.
- `launcherWorktreeMaterializeReal` (`worktree_branch_flag.go:93-124`): `treePath` from the resolver; keep the idempotent re-entry and `hook.RegisterLauncherWorktree`.
- Launcher short-name path (REQ-WHR-006, D-5): `moai cc -w <name>` / `moai glm -w <name>` with no existing tree → shared materializer → absolute path handed to the backend; update `validateWorktreeL2Value` accepted-prefix list (`launcher.go:940-972`) to the home root plus the legacy roots. Condition: PM-3 and D-5 hold; otherwise record the exception and stop for operator input.
- Exit: no creation path names a legacy root (grep with count), `moai worktree new <name>` in a scratch `MOAI_HOME` creates at the home root, base-branch and `WT-` behaviors unchanged (existing tests stay green).

### M3 — Readers (Priority High; characterization first)
Characterization tests first, pinning today's behavior for legacy-rooted trees, then add home-root cases, then change the code. Each reader switches to the resolver with recorded `worktree_path` taking precedence:
1. `cardIDFromRegisteredWorktree` (`internal/hook/session_start_factory.go:231-259`) and `cardIDFromPath` (`internal/hook/session_start_record.go:226-241`): an ancestor-named-`worktrees` walk becomes marker-aware (`projects/<key>/<name>` resolves `<name>`; REQ-WHR-008).
2. `autoOwnerWorktrees` (`internal/cli/todo_auto.go:88-93`): candidates from the resolver plus the recorded path.
3. `DefaultWorktreeBases` (`internal/factory/status_read.go:63-69`) and the loop at line 117: scan the project's home subdirectory, not the bare `~/.moai/worktrees`.
4. Short-name resolvers: `resolveLaunchWorktreePath` (`launch_chain.go:60-72`), `resolveWorktreeLaunchTarget` (`factory_anchor_guard.go:98-106`), `seedAdmittedWorktreeHooksFor` (`codex_hooks_seed.go:61-75`), `ccWorktreeWriterPrecheck` (`session_worktree.go:1010-1033`), `resolveCodexWorktreeDir` (`codex_launcher.go:474-519`; ambiguity message extended to three candidates), `cleanupMoaiWorktrees` (`launcher.go:449-537`; it already scans per-subdirectory of `~/.moai/worktrees`, so the home-root layout adds one level).
5. Tier classifier: `isL1WorktreePath` (`internal/cli/worktree/done.go:230-247`) and `sweepTierOf` (`sweep.go:574-580`) classify by creation class: home-root moai-made trees are L1 (REQ-WHR-009). The disposal predicates are unchanged.
- Exit: every reader passes a home-root fixture and its legacy fixture; the M0 RED reproductions for misidentification (card id `projects`, owner-not-alive) turn GREEN.

### M4 — `moai worktree list` (Priority Medium)
- New `list` subcommand in `internal/cli/worktree/` registered in `root.go:42-54`; read-only; `--json`. Reuses the classifier of M5 (REQ-WHR-013) so list and migrate cannot disagree. Today `moai worktree list` prints the parent help text (observed in this session).
- Exit: fixture repository with one tree of each class lists each exactly once; the two spellings of one directory collapse (REQ-WHR-010/023).

### M5 — Classifier, stored-path rebind (Priority High; decisions D-2)
- Classifier (REQ-WHR-013) as a pure function over the `git worktree list --porcelain` records plus the integration-branch name; unclassified is a first-class result.
- Card record rebind (REQ-WHR-011): new `FactoryDB` method beside `RecordCardWorktree` (`internal/homestate/card_worktree.go:19-60`), version-checked like every record write, updating every run row carrying the old path (the cross-run reader is `PreviousCardWorktree`, line 73) and appending a `card.fields` event. `RecordCardWorktree`'s refusal at line 42-44 stays.
- Project worktree state re-key (`internal/hook/worktree_registry.go:29-65`, state file `.moai/state/worktrees.json`, REQ-WHR-012).
- Exit: classifier table test over the 6 measured shapes (home, legacy-moai WT-, legacy-moai non-WT such as `rc30-preview`, legacy-claude moai-made such as `t1497`, `agent-*`, integration `main`); rebind tests including the stale-old-path negative case.

### M6 — Migration engine (Priority High; invariant I-1; decisions D-6, D-7)
- `moai worktree migrate [--apply] [--verify] [--rollback <journal>] [--json]` in `internal/cli/worktree/`. It reuses the safety planning of `planLegacyWorktreeMigrationFor` (`internal/cli/update_worktree_migration.go:76-147`: prunable, locked, process cwd, registry readability, live anchor, destination exists) but takes its selection from the M5 classifier, not from the path prefix at lines 87, 105-108 (that prefix test is what currently selects every `.claude/worktrees/*` including `agent-*`).
- Per-tree manifest before and after (REQ-WHR-017), journal under `.moai/state/worktree-migrate/<timestamp>.json` (machine-local, gitignored), `git worktree move` as the only mutation (REQ-WHR-014), automatic reversal on mismatch or rebind failure, explicit `--rollback`.
- Tree-by-tree, continue past a skip or failure, final summary with moved / skipped / failed / reversed counts, exit non-zero when any reversal failed.
- `moai update` consolidation (REQ-WHR-019, D-6): the old function's selection switches to the classifier and its default becomes preview-and-print; its idempotence and primary-checkout-only behavior (lines 236-243) are preserved.
- Rehearsal in a scratch repository: a tree with committed-unpushed work, one with uncommitted and untracked and ignored files, one locked, one with an occupied cwd, plus an `agent-*` tree that must not move.
- Exit: dry-run makes zero filesystem writes (before/after tree hash), rehearsal proves the manifest equality, injected failure proves reversal, the old function no longer selects an `agent-*` tree (its current behavior: `grep -n 'agent-' internal/cli/update_worktree_migration.go` prints nothing).

### M7 — Retirement and guidance (Priority Medium; t1610 absorbed part)
- Retirement (REQ-WHR-020): remove the empty MoAI legacy directory by non-recursive removal only; no code path removes the Claude-owned directory.
- Guidance edits, Template-First (I-6), AGENTS.md net-neutral or smaller (I-5): AGENTS.md §0 table row `worktree-entry` and §3 text; `worktree-integration.md` Overview, Comparison Table, Glossary and Relationships; `worktree-integration-ops.md`; cc/glm `--help` text (`cc.go:41`, `glm.go:61`); `moai worktree` long help (`root.go:25-33`); `new.go:21` long text. The wording asserts only what the §B record shows; Codex `/cd` text comes from the memory reference and PM-2. The sweep for "run `moai cc -w` from inside a session" sentences carries a count-first grep (zero-sweep guard).
- The t1879 card rewrites the remainder after this card lands; M7 does not attempt a full docs rewrite.
- Exit: AGENTS.md byte count at or below 23,780; zero in-session `moai cc -w` instructions across the in-scope surfaces; template mirrors in sync (the repository's own mirror test).

### M8 — Verification and sync handoff (Priority Medium)
- Scoped test runs for the touched packages; the repository build step for embedded assets; cross-platform build check; golangci-lint on the touched packages (all per `AGENTS.md §4`: scoped verification, CI runs the full suite).
- Live apply is NOT done by the lane (D-7). The sync evidence includes the dry-run listing of the real repository (counts per class) and the exact `moai worktree migrate --apply` command for the leader.

## §D. File allow-list for the run phase (edits outside it are a blocker)

Go and tests (new files marked +): `internal/worktreeroot/*` (+, name to be confirmed), `internal/paths/paths.go` (only if a `WorktreesProjectsDir` accessor is preferred; stdlib-only constraint), `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/loader_worktree_project_key.go` (+), `internal/cli/session_worktree.go`, `internal/cli/worktree_branch_flag.go`, `internal/cli/launcher.go`, `internal/cli/launch_chain.go`, `internal/cli/factory_anchor_guard.go`, `internal/cli/codex_hooks_seed.go`, `internal/cli/codex_launcher.go`, `internal/cli/cc.go`, `internal/cli/glm.go`, `internal/cli/todo_auto.go`, `internal/cli/update_worktree_migration.go` and its tests, `internal/cli/update*.go` (caller of the consolidation, name to be confirmed), `internal/cli/worktree/root.go`, `internal/cli/worktree/new.go`, `internal/cli/worktree/done.go`, `internal/cli/worktree/sweep.go`, `internal/cli/worktree/list.go` (+), `internal/cli/worktree/migrate*.go` (+), `internal/cli/worktree/classify.go` (+), `internal/hook/session_start_factory.go`, `internal/hook/session_start_record.go`, `internal/hook/worktree_registry.go`, `internal/factory/status_read.go`, `internal/homestate/card_worktree.go` and its tests, matching `_test.go` files.

Templates and docs: `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`, the template mirrors of every rule/skill/agent file edited below, `.moai/config/sections/git-strategy.yaml`, `AGENTS.md`, `.claude/rules/moai/workflow/worktree-integration.md`, `.claude/rules/moai/workflow/worktree-integration-ops.md`, and the skills/agents files that M7's count-first grep names (list recorded in run-phase evidence, bounded to the 13 + 13 + 2 files research.md §R5 counted for `moai cc -w`).

Explicitly NOT touched: `.claude/worktrees/agent-*` (any), `~/.codex/worktrees`, the integration worktree `.claude/worktrees/main`, `internal/hook/worktree_create.go` and the unregistered WorktreeCreate wrapper (kept as dormant infrastructure), the dead `worktree_root` key, the disposal predicates of `done`/`sweep`/`clean`, `internal/mission/integration.go`, and any `.moai/specs/` directory other than this SPEC's own `progress.md` (run phase).

## §E. Technical approach notes

- **One materializer, one resolver.** The shared materializer is already the single creation chokepoint (`root.go:139`, `@MX:ANCHOR` at `factory_card.go:700-707` states four callers). The change is a destination swap there plus the readers; no new creation path is added.
- **Reader policy: recorded path first, resolver second, heuristics last.** The factory record already stores an absolute `worktree_path` (`internal/homestate/factory.go:66`); readers that re-derive a path from a name are the ones that break. The resolver exists for readers that have only a name (launchers) and for the transition.
- **Tier by class.** `isL1WorktreePath` answers "is this a session tree" from a path prefix today (`done.go:230-247`). Moving trees out of the prefix would silently turn session trees into L2 disposal candidates; the classifier makes the answer a property of how the tree was created. The sweep/done landing predicates are untouched.
- **Spelling-safe equality** is needed now: on this machine `/Users/goos/MoAI` and `/Users/goos/moai` are the same directory (same inode 105929094), and `git worktree list` reports 25 of the 192 MoAI-root trees under the lowercase spelling (research.md §R1).
- **No symlinks** at old paths (design.md D-4).
- **Windows**: path building uses `filepath` and the existing `paths.Home()` resolver; no Windows measurement was taken (OQ-4). Cross-platform build of the touched packages is a run-phase check (`GOOS=windows`).

## §F. Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| A reader that identifies trees by path shape is missed | Medium | M0 inventory re-count with a grep sweep for `".claude", "worktrees"` / `".moai", "worktrees"` / `Base(parent) == "worktrees"`; research.md §R3 lists the 12 known sites; AC-WHR-009 re-runs the sweep |
| Migration moves a tree a live session is using | Medium | Skip conditions REQ-WHR-016, D-7 (leader-operated), journal + reversal |
| Card record rebind leaves a stale path in an older run row | Medium | REQ-WHR-011(c), AC-WHR-013 negative case |
| `EnterWorktree`/`/cd` behave differently from the assumption | Medium | §B gate; branches stated per measurement |
| The lane's own tree and other lanes' trees move under their feet during rehearsal | Low | Rehearsal uses scratch repositories only |
| Guidance edits grow AGENTS.md past the Codex budget | Low | I-5 and AC-WHR-021 byte check |
| Collision key surprise (two repositories with one name) | Low | Marker guard, AC-WHR-003 |
| The old `moai update` consolidation keeps running with its old selection on older installed binaries | Low | The `agent-*` exclusion is also documented in the release note; binaries predating this change are the user's to update |

## §G. Anti-patterns to avoid

- Disposing or deleting a tree to "finish" the migration (REQ-WHR-014).
- Pinning every path with a symlink at the old location.
- Registering WorktreeCreate to get the short-name launcher case "for free" (dab5).
- Claiming a harness entry behavior that the §B record does not show.
- Reading `git worktree list` counts as a stable fact: the primary root has two spellings and the count changes every time a lane creates a tree; use a pinned SHA or a dated count when citing.

## §H. Cross-references

spec.md §E (decisions), §F (open measurements); design.md (alternatives); acceptance.md (criteria); research.md (evidence).

## Open clarification markers (must be settled before Implementation Kickoff)

- [NEEDS CLARIFICATION: D-1 path form — confirm `projects/<project-name>` with collision suffix versus always-suffixed key]
- [NEEDS CLARIFICATION: D-5 launcher short-name materialization — confirm after PM-3]
- [NEEDS CLARIFICATION: D-6/D-7 who applies the live migration, and whether `moai update` applies or previews]
- [NEEDS CLARIFICATION: OQ-1 release target label for this SPEC (`v3.2.0 target` is the nearest unreleased label found in the corpus)]
