# plan.md — SPEC-WORKTREE-HOME-ROOT-001

> Derived implementation plan. Ordered by decision-reversibility: the decisions most likely to change come first (path form and key, pre-kickoff measurement gate and its failure branches, launcher and entry flow, stored-path rebind), mechanical rewrites last. No time estimates; priority and dependency ordering only. Code citations measured in worktree `.moai/worktrees/t1640`; the correction round re-read them at HEAD `e5950b354f`.

## §A. Context

- SPEC: `.moai/specs/SPEC-WORKTREE-HOME-ROOT-001/spec.md` (Tier L, 25 REQ / 25 AC). Card t1640, Class C, operator-approved, deferral lifted 10-19 (decision `cadf`).
- Binding decisions: `cadf` (layout, order, non-destructive), `dab5` (agent-* stay; no WorktreeCreate hook), `ef12` (non-binding defaults, mandatory pre-kickoff EnterWorktree measurement), and the operator's relayed D-1 ruling (plain project name, case-insensitive collision suffix, stable key).
- Order against sibling cards: t1791 (GH #1813) lands first, then this card absorbs its launcher behavior; t1879 and t1753 follow this card and own the in-session wording (spec §G boundary).
- Methodology: `quality.yaml constitution.development_mode` governs; this plan is written for TDD with a characterization pass first for the readers that change behavior (M3).
- Class C plan audit by codex is required before the run phase; findings and verdict paths: `.moai/reports/t1640/` (iteration 1 returned FAIL, 0.62; this is the correction).

## §B. Pre-kickoff measurement gate (M-1; blocks the Kickoff — REQ-WHR-021)

**Placement.** M-1 is not a run-phase milestone. It runs after a plan-audit PASS and before the plan→run Kickoff, so a Gap or an unresolved consequence is found while the SPEC can still be amended. The Kickoff decision record cites the record file and its sha256.

**Executor.** The lane runs each measurement that a headless or scripted session can take; a measurement that needs an interactive session the lane cannot start is handed to the leader with the exact steps. A cell nobody took is a Gap and blocks (below).

**Where.** In a scratch git repository created under `<MoaiHome>` (never this repository, never a real card tree). The home-root tree in the scratch repository is made with `git worktree add`; AGENTS.md §3's reason for forbidding it (MoAI would not know the tree) does not apply to a throwaway repository with no card and no branch of value, and the record states this once per measurement. The M2 materializer is not used because it does not exist before the run phase.

**Record format.** `.moai/reports/t1640/pre-kickoff-measurements.md`, one line per cell `RESULT <id>: PASS|FAIL|GAP`, then the command, verbatim output, exit status, `claude --version` / `codex --version` and the tree SHA. Gate check, plain commands: `grep -c '^RESULT PM-\(1a\|1b\|1c\|1d\|2\|3\): \(PASS\|FAIL\)$'` must print 6 (count-first: an empty file prints 0 and does not pass), and `grep -c '^RESULT .*: GAP$'` over the gating ids must print 0.

**Classification rule.** Every observation maps to exactly one of PASS or FAIL per spec §F; a refusal is FAIL only (its text is recorded as evidence, never as a second PASS). The PASS/FAIL observables, the consequence of each FAIL, and the PM-3 branch selector are in spec §F and are not restated here.

**What a Gap does.** A Gap in PM-1a..PM-1d, PM-2 or PM-3 blocks the Kickoff. PM-4, PM-5 and PM-6 may be Gaps; each Gap is recorded with the consequence in spec §F.

**PM-3 selects D-5.** PASS → branch X (M2 implements materialize-then-absolute-path); FAIL → branch Y (M2 passes the short name through and records the exception). Both branches are specified in spec REQ-WHR-006, so no SPEC amendment follows the measurement.

## §C. Milestones (dependency order; priority labels instead of time)

Every run-phase milestone starts RED (a failing test or an observed-failing check) and ends GREEN with the verbatim output in the run-phase evidence.

### M-1 — Pre-kickoff measurements (Priority High; precedes the Kickoff, outside the run phase)
- Execute §B. Exit: the gate check of §B prints 6 and 0; each FAIL's consequence is carried into the Kickoff decision record.

### M0 — Baseline (Priority High; first run-phase milestone; touches no production code)
- Record the RED-now cells of acceptance.md (command, verbatim stdout, exit code, tree SHA) in a baseline commit that precedes the first production change (verification-claim-integrity §2.3 ordering attribution), including the legacy-reader inventory re-count of research.md §R3.
- Exit: baseline commit exists; operator-held decisions are recorded confirmed or default-applied (spec §E).

### M1 — Location resolver, key assignment and config (Priority High; decisions D-1, D-8)
- New resolver in a small package (candidate: `internal/worktreeroot`, confirmed against the `internal/paths` import-light constraint at `internal/paths/paths.go:25-30`): `Root`, `PathFor`, `Key`, `Candidates` (distinct trees across home root and both legacy roots, no root preference — REQ-WHR-022), `Same` spelling-safe equality (REQ-WHR-023). Calls `paths.MoaiHome()` only.
- Key assignment (REQ-WHR-002/003/004): first-assignment derivation from the primary root; assignment recorded in the repository-side record and the home owner marker; recorded key returned on every later resolution; case-insensitive same-name collision with a different owner yields the hash suffix for the later arrival only; refusal when the suffixed folder is owned by a third repository or a marker is unreadable; pin honored only at first assignment.
- `git_strategy.worktree_project_key`: field beside `WorktreeBaseBranch` (`internal/config/types.go:176-186`), default `""`, loader helper modeled on `LoadWorktreeBaseBranch` (`internal/config/loader_worktree_base.go:30`), template mirror in `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`. The dead `worktree_root` key is left untouched (D-8).
- Exit: resolver unit tests: plain name, collision with a differently cased name, first keeps plain name, stability across re-resolution, checkout move and clone rename, corrupt marker, `MOAI_HOME` relocation, spelling-safe equality.

### M2 — Creation paths (Priority High; decision D-5, PM-3)
- `materializeSessionWorktree` (`internal/cli/session_worktree.go:213-266`) computes `destDir` through the resolver instead of `filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` (line 223). All callers inherit: `worktree.WorktreeCreator` is wired once at `internal/cli/root.go:139`; the callers of `factoryEnsureCardWorktree` (`factory_card.go:700-715`) and `moai worktree new` (`internal/cli/worktree/new.go:41-52`) need no per-caller change in creation. The guard-and-notice behavior of lines 224-266 is preserved.
- `launcherWorktreeMaterializeReal` (`worktree_branch_flag.go:93-124`; the legacy destination is line 109): `treePath` from the resolver; keep the idempotent re-entry and `hook.RegisterLauncherWorktree`.
- Launcher short-name path, selected by PM-3: branch X materializes at the home root and passes the absolute path; branch Y passes the short name through unchanged. Update `validateWorktreeL2Value` accepted-prefix list (`launcher.go:940-972`) to the home root plus the legacy roots.
- Exit: no creation path names a legacy root except the branch-Y exception (grep with count), `moai worktree new <name>` in a scratch `MOAI_HOME` creates at the home root, base-branch and `WT-` behaviors unchanged.

### M3 — Readers and launcher cleanup (Priority High; characterization first)
Characterization tests first, pinning today's behavior for legacy-rooted trees, then add home-root cases, then change the code. Each reader switches to the resolver with the recorded `worktree_path` taking precedence:
1. `cardIDFromRegisteredWorktree` (`internal/hook/session_start_factory.go:231-259`) and `cardIDFromPath` (`internal/hook/session_start_record.go:226-241`): marker-aware walk (`projects/<key>/<name>` resolves `<name>`; REQ-WHR-008).
2. `autoOwnerWorktrees` (`internal/cli/todo_auto.go:88-93`): candidates from the resolver plus the recorded path.
3. `DefaultWorktreeBases` (`internal/factory/status_read.go:63-69`) and the loop at line 117: scan the project's home subdirectory, not the bare `~/.moai/worktrees`.
4. Factory landing directory (D4 of the audit): `factoryCardWorktreeDir` (`internal/cli/factory_card.go:499`, which returns `filepath.Join(root, sessionWorktreeSubdir, cardID)`) resolves through the resolver; its consumers `factoryRefuseForeignWorktree` (`:512`) and `factoryCarryPreviousWorktree` (`:528`) follow, reached from `factoryNextClaim` (call sites `:1152`, `:1162`) and `factoryNextValidate` (`:1450`, `:1455`). A replacement run must reuse the home-root tree of the same card (t1521 carry-over) and the foreign-tree refusal (REQ-SD-011) must fire before any lease edge.
5. Short-name resolvers: `resolveLaunchWorktreePath` (`launch_chain.go:60-72`), `resolveWorktreeLaunchTarget` (`factory_anchor_guard.go:98-106`), `seedAdmittedWorktreeHooksFor` (`codex_hooks_seed.go:61-75`), `ccWorktreeWriterPrecheck` (`session_worktree.go:1010-1033`), `resolveCodexWorktreeDir` (`codex_launcher.go:474-519`; duplicate-name rule extended to three roots, REQ-WHR-022).
6. `cleanupMoaiWorktrees` (`launcher.go:449-537`), corrected premise: it treats each direct child of `~/.moai/worktrees` as a base (so `projects` becomes a base) and matches a `worker-*` tree with `filepath.Rel(base, worktreePath)` rejecting only a `..` result (`:505-518`), which **accepts a nested path** such as `projects/<key>/worker-x`; it then runs non-force `git worktree remove` (`removeWorktree`, `:553-556`, called at `:518`) on every `moai cc` launch. A clean tree holding unpushed commits is removed by that command, and its ignored reports go with it. Required guard (REQ-WHR-014): the cleanup removes only a tree whose branch has no commits absent from every remote and which the migration did not move; otherwise it skips the tree and says why. Today `git worktree list | grep -c '/worker-'` prints 0 (no present exposure).
7. Tier classifier: `isL1WorktreePath` (`internal/cli/worktree/done.go:230-247`) and `sweepTierOf` (`sweep.go:574-580`) classify by creation class including witness (c) of REQ-WHR-013, so a non-`WT-` home-root tree is L1 (REQ-WHR-009). `isCardBranch` (`done.go:315`, `WT-` prefix only) is not the tier witness; the origin-landing check keeps its `WT-` scope. The disposal predicates are unchanged.
- Exit: every reader passes a home-root fixture and its legacy fixture; the M0 RED reproductions for misidentification (card id `projects`, owner-not-alive, factory landing directory) turn GREEN.

### M4 — `moai worktree list` (Priority Medium)
- New `list` subcommand in `internal/cli/worktree/` registered in `root.go:41-54`; read-only; `--json`. Reuses the classifier of M5 so list and migrate cannot disagree. Today `moai worktree list` prints the parent help text.
- Exit: fixture repository with one tree of each class lists each exactly once; the two spellings of one directory collapse (REQ-WHR-010/023).

### M5 — Classifier, stored-path rebind (Priority High; decision D-2)
- Classifier (REQ-WHR-013) as a pure function over the `git worktree list --porcelain` records, the integration-branch name and the owner marker; unclassified is a first-class result; witness (c) keeps a moved non-`WT-` tree (`rc30-preview`, `release-v314`, `t1515-absorb`, `t1868-publish`) moai-made.
- Card record rebind (REQ-WHR-011): new `FactoryDB` method beside `RecordCardWorktree` (`internal/homestate/card_worktree.go:19-60`), version-checked like every record write, updating every run row carrying the old path and appending a `card.fields` event. `RecordCardWorktree`'s refusal at lines 42-44 stays.
- Project worktree state re-key (`internal/hook/worktree_registry.go:29-65`; the file is keyed by `Path`): an operation that replaces an entry's path, and its inverse (REQ-WHR-012).
- Exit: classifier table test over the measured shapes (home, home non-`WT-` after a move, legacy-moai `WT-`, legacy-moai non-`WT-`, legacy-claude moai-made such as `t1497`, `agent-*`, integration `main`, unclassified); rebind tests including the stale-old-path negative case.

### M6 — Migration engine (Priority High; invariant I-1; decisions D-6, D-7)
- `moai worktree migrate [--apply] [--verify] [--rollback <journal>] [--json]` in `internal/cli/worktree/`. It reuses the safety planning of `planLegacyWorktreeMigrationFor` (`internal/cli/update_worktree_migration.go:76-147`) but takes its selection from the M5 classifier, not from the path prefix at lines 87, 105-108.
- Per-tree: manifest (including the content digest of REQ-WHR-017), journal under `.moai/state/worktree-migrate/<timestamp>.json` (machine-local, gitignored) written before the move with a done/pending status per step; step order manifest, `git worktree move`, state re-key, card rebind, manifest comparison; reverse order on any failure; idempotent `--rollback` and re-run that complete pending steps (REQ-WHR-017/018).
- Tree-by-tree, continue past a skip or failure, final summary with moved / skipped / failed / reversed counts, exit non-zero when any reversal failed or a step remains pending. An injectable device-id seam makes the cross-filesystem skip testable without a second volume.
- `moai update` consolidation (REQ-WHR-019): `runUpdateWorktreeMigration` (`update_worktree_migration.go:236`, called from `update.go:366` with `dryRun=true` and from `update.go:410` with `dryRun=false`) switches its selection to the classifier, names the home root as every move's destination (it currently targets `<root>/.moai/worktrees`, `:88`), and both call paths become preview-and-print. This is the recorded behavior change of D-6.
- Rehearsal on a scratch repository: a tree with committed-unpushed work, one with uncommitted/untracked/ignored files, one locked, one with an occupied cwd, a non-`WT-` tree, a `worker-*` tree, plus an `agent-*` tree that must not move; and one post-migration launcher cleanup pass.
- Exit: dry-run makes zero filesystem writes (before/after tree hash), rehearsal proves manifest equality, fault injection at each journaled step plus a kill proves the three stores agree afterwards, the old function no longer selects an `agent-*` tree.

### M7 — Retirement and location guidance (Priority Medium; t1610 absorbed part only)
- Retirement (REQ-WHR-020): remove the empty MoAI legacy directory by non-recursive removal only, placed in the migrate engine's package (`internal/cli/worktree/`), separate from the static no-remove scan scope of AC-WHR-014a; no code path removes the Claude-owned directory.
- Location guidance (REQ-WHR-025), Template-First (I-6), AGENTS.md net-neutral or smaller (I-5): the home-root location in AGENTS.md §3 (and the §0 `worktree-entry` row only where it states a location); `worktree-integration.md` Overview, Comparison Table, Glossary and Relationships; `worktree-integration-ops.md`; cc/glm `--help` location text (`cc.go:41`, `glm.go:61`); `moai worktree` long help (`root.go:25-33`); `new.go:21`. The wording asserts only what the §B record shows.
- **Boundary with card t1879 (sentence by sentence).** M7 writes: where trees live; the terminal commands that take an absolute path; the names of the in-session mechanisms; entry facts the record shows. M7 does not remove or rewrite existing sentences that tell a session to run `moai cc -w` from inside Claude Code, does not rewrite AGENTS.md §0/§3 in-session wording beyond the location text, does not touch lane in-session transition wording, and does not edit docs-site or README. Those belong to t1879, which runs after this card.
- Exit: AGENTS.md byte count at or below 23,780; the home-root location appears on each in-scope surface that names a worktree location; template mirrors in sync (the repository's own mirror test). No in-session-sentence count is claimed.

### M8 — Verification and sync handoff (Priority Medium)
- Scoped test runs for the touched packages; the repository build step for embedded assets; cross-platform build check; golangci-lint on the touched packages (all per `AGENTS.md §4`: scoped verification, CI runs the full suite).
- Live apply is NOT done by the lane (D-7). The sync evidence includes the dry-run listing of the real repository (counts per class, including the four non-`WT-` trees) and the exact `moai worktree migrate --apply` command for the leader.

## §D. File allow-list for the run phase (edits outside it are a blocker)

Go and tests (new files marked +): `internal/worktreeroot/*` (+, name to be confirmed), `internal/paths/paths.go` (only if a `WorktreesProjectsDir` accessor is preferred; stdlib-only constraint), `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/loader_worktree_project_key.go` (+), `internal/cli/session_worktree.go`, `internal/cli/worktree_branch_flag.go`, `internal/cli/launcher.go`, `internal/cli/launch_chain.go`, `internal/cli/factory_anchor_guard.go`, `internal/cli/codex_hooks_seed.go`, `internal/cli/codex_launcher.go`, `internal/cli/factory_card.go` and its tests, `internal/cli/cc.go`, `internal/cli/glm.go`, `internal/cli/todo_auto.go`, `internal/cli/update_worktree_migration.go` and its tests, `internal/cli/update.go` (the two call sites of the consolidation), `internal/cli/worktree/root.go`, `internal/cli/worktree/new.go`, `internal/cli/worktree/done.go`, `internal/cli/worktree/sweep.go`, `internal/cli/worktree/list.go` (+), `internal/cli/worktree/migrate*.go` (+), `internal/cli/worktree/classify.go` (+), `internal/cli/worktree/retire*.go` (+), `internal/hook/session_start_factory.go`, `internal/hook/session_start_record.go`, `internal/hook/worktree_registry.go`, `internal/factory/status_read.go`, `internal/homestate/card_worktree.go` and its tests, matching `_test.go` files.

Templates and docs: `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`, the template mirrors of every rule/skill/agent file edited below, `.moai/config/sections/git-strategy.yaml`, `AGENTS.md` (location text only), `.claude/rules/moai/workflow/worktree-integration.md`, `.claude/rules/moai/workflow/worktree-integration-ops.md`, and the skill/agent files that carry a worktree location statement (list recorded in run-phase evidence).

Explicitly NOT touched: `.claude/worktrees/agent-*` (any), `~/.codex/worktrees`, the integration worktree `.claude/worktrees/main`, `internal/hook/worktree_create.go` and the unregistered WorktreeCreate wrapper, the dead `worktree_root` key, the disposal predicates of `done`/`sweep`/`clean`, `internal/mission/integration.go`, docs-site and README files, lane in-session wording, and any `.moai/specs/` directory other than this SPEC's own `progress.md` (run phase).

## §E. Technical approach notes

- **One materializer, one resolver.** The shared materializer is already the single creation chokepoint (`root.go:139`). The change is a destination swap there plus the readers; no new creation path is added.
- **Reader policy: recorded path first, resolver second, heuristics last.** The factory record already stores an absolute `worktree_path` (`internal/homestate/factory.go:66`); readers that re-derive a path from a name are the ones that break.
- **Tier by class.** `isL1WorktreePath` answers "is this a session tree" from a path prefix today. Moving trees out of the prefix would turn session trees into L2 disposal candidates; the classifier makes the answer a property of how the tree was created, and witness (c) keeps that answer after the move for trees that do not carry a `WT-` branch.
- **Key stability.** The recorded assignment, not the current directory name, is the source of truth after the first resolution (spec I-8).
- **Spelling-safe equality** is needed now: on this machine `/Users/goos/MoAI` and `/Users/goos/moai` are the same directory (`stat -f '%i'` printed 105929094 for both), and `git worktree list` reports 25 of the 192 MoAI-root trees under the lowercase spelling (research.md §R1).
- **No symlinks** at old paths (design.md D-4).
- **Windows**: path building uses `filepath` and the existing `paths.Home()` resolver; no Windows measurement was taken (OQ-4). Cross-platform build of the touched packages is a run-phase check (`GOOS=windows`).

## §F. Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| A reader that identifies trees by path shape is missed | Medium | M0 inventory re-count with a grep sweep; research.md §R3 lists the 13 known sites; AC-WHR-007 re-runs the sweep |
| Migration moves a tree a live session is using | Medium | Skip conditions REQ-WHR-016, D-7 (leader-operated), journal + reversal |
| Launcher auto-cleanup removes a migrated unpushed `worker-*` tree | Medium after migration, none today | REQ-WHR-014 guard, AC-WHR-014b, post-migration launcher pass in the M6 rehearsal |
| Card record, state file and real location disagree after a failure | Medium | Journaled step order, reverse order, idempotent rollback, fault injection at each step (AC-WHR-017/018) |
| Card record rebind leaves a stale path in an older run row | Medium | REQ-WHR-011(c), AC-WHR-011 negative case |
| `EnterWorktree`/`/cd` behave differently from the assumption | Medium | §B gate; every FAIL has a stated consequence |
| A non-`WT-` tree becomes `done`-eligible after the move | Low after fix | Witness (c), AC-WHR-009/013 |
| Collision key surprise (two repositories with one name, differing case) | Low | Case-insensitive owner-marker guard, stability invariant, AC-WHR-002b/003a |
| Guidance edits grow AGENTS.md past the Codex budget | Low | I-5 and AC-WHR-025 byte check |
| The old `moai update` consolidation keeps running with its old selection on older installed binaries | Low | The behavior change is recorded in the release note; binaries predating this change are the user's to update |

## §G. Anti-patterns to avoid

- Disposing or deleting a tree to "finish" the migration (REQ-WHR-014).
- Pinning every path with a symlink at the old location.
- Registering WorktreeCreate to get the short-name launcher case "for free" (dab5).
- Claiming a harness entry behavior that the §B record does not show.
- Treating a Gap in a gating measurement as "to be recorded later".
- Re-deriving the project key from the current directory name after it has been assigned.
- Reading `git worktree list` counts as a stable fact: the primary root has two spellings and the count changes every time a lane creates a tree; use a pinned SHA or a dated count when citing.

## §H. Cross-references

spec.md §E (decisions), §F (measurements); design.md (alternatives); acceptance.md (criteria); research.md (evidence).

## §I. Settled clarifications (all earlier open markers resolved; none remains)

Each item is recorded in spec §E with its rationale and the rejected alternative.

- **D-1 path form** — DECIDED BY THE OPERATOR: `<MoaiHome>/worktrees/projects/<project-name>/<card>`, suffix only on a case-insensitive same-name folder owned by a different repository, key stable once assigned. Rejected alternative: always-suffixed key.
- **D-5 launcher short name** — CONDITIONAL on PM-3 (PASS → branch X materialize-then-absolute-path; FAIL → branch Y pass-through exception); PM-3's recorded result is the selector. Lane judgment under auto-semantics §6; the operator may override before kickoff.
- **D-6 / D-7** — DECIDED: dry-run is the default; the lane delivers the tool and a scratch-repository rehearsal; the leader applies it to the real repository; `moai update` previews and never applies, a recorded behavior change from `update.go:410`. Lane judgment under auto-semantics §6; the operator may override before kickoff.
- **OQ-1 release label** — DECIDED: `v3.2.0 target`. Lane judgment under auto-semantics §6; the operator may override before kickoff.
