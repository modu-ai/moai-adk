# design.md — SPEC-WORKTREE-HOME-ROOT-001

> Architecture decisions with the alternatives rejected. Each decision that is the operator's to make is marked **Operator decision**, carries a recommended default and the reason, and is not applied silently. Facts used here are cited in research.md.

## §1. Shape of the change

One new idea (a location resolver plus a creation-class classifier), applied at three seams: where a tree is created (one chokepoint already exists), where a tree is looked up (thirteen known readers), and where a tree is moved (a new migration engine that reuses the existing safety planner). The data that must change with a move is small: a card record's `worktree_path`, the project worktree state entry, and nothing else that is authoritative.

```
create:  moai worktree new / factory next / session auto-entry / launcher -w <n> --branch
            └── worktree.WorktreeCreator (root.go:139) ── materializeSessionWorktree
                  └── destDir = resolver.PathFor(primaryRoot, name)      [was: <root>/.moai/worktrees/<name>]
lookup:  name or card id ── resolver.Candidates (home, legacy-moai, legacy-claude)  ◄── recorded worktree_path wins
classify: tree ── classifier (creation class, not path prefix) ── L1 / L2 / claude-native / integration / unclassified
move:    classifier ── safety planner ── manifest ── git worktree move ── manifest ── card rebind ── journal
```

## §2. Decisions

### D-1 Path form and key (DECIDED BY THE OPERATOR)

**Settled form:** `<MoaiHome>/worktrees/projects/<project-name>/<card>`, where the folder is the plain project name and carries a hash suffix only on a cross-repository collision. This is the operator's ruling relayed by the leader on 2026-10-19 (it replaces the provisional text of the first draft); it is a decision, not an open question.

**Reconciliation of the names in the record.** The card text (10-10) says `~/.moai/worktrees/<repo-id>`; the operator's wording on 10-19 (decision `cadf`) is `~/.moai/worktrees/projects/<project-name>/<name>`; `dab5` says `<project-key>`. The operator's wording governs. `<project-key>` in the SPEC is the project name, suffixed only on collision. `<repo-id>` is dropped as a term because the card text was never implemented anywhere (`git log --all -S'worktrees/projects'` matched nothing per the cadf record).

**Collision rule (exact).** When a folder under `<MoaiHome>/worktrees/projects/` has a name equal to the candidate project name compared **case-insensitively** and its owner marker names a different repository, the later-arriving repository gets `<project-name>-<first 8 hex of SHA-256 of the case-folded canonical primary root>`. The first repository keeps the plain name. The name test is case-insensitive on every volume (so a folder created on a case-sensitive volume cannot later collide silently on a case-insensitive one); the hash input is case-folded so two machines agree. If the suffixed folder is itself owned by a third repository, or a marker is unreadable, creation refuses naming the marker path.

**Stability.** Once a repository is assigned a key it never changes. The assignment is recorded in two places: a record in the repository's git common directory (it moves with the checkout and survives a rename of the clone's directory) and the owner marker `<key>/.moai-project.json` (schema version, owner identity, primary root last seen, creation time). Resolution reads the recorded key first and only derives a key when none is recorded, so re-running resolution or migration is idempotent and a later same-name arrival cannot displace the first. Residual risk: a byte-for-byte copy of a checkout duplicates the owner identity; this is recorded as a Gap, not solved here.

**Key derivation.** From the primary checkout root, never from the invoking worktree: the `git rev-parse --path-format=absolute --git-common-dir` parent (the derivation used at `session_worktree.go:218-222`, `done.go:186-205`). Basename, sanitized to `[A-Za-z0-9._-]`.

**Alternatives rejected**
- *Always-suffixed key* (`moai-adk-go-1f3a9c2b`): collision-proof and needs no marker for ownership, but the paths stop matching the operator's wording and are unreadable in `ps`, statusline and error messages. Rejected as an alternative; not an open question.
- *Hash-only directory name*: same readability loss, no gain.
- *Key from `git remote get-url origin`*: absent for local-only repositories and changes with remote renames.
- *Key from the repository root path with separators replaced* (Claude's project-directory encoding): long and machine-specific.
- *Nesting by owner/name*: no owner concept exists locally.
- *Recomputing the key on every call from the current directory name*: a renamed or moved checkout would silently start a second key; replaced by the stability invariant.

**Operator pin** (`git_strategy.worktree_project_key`) names the first-assigned key only; it never re-keys an assigned repository and, if its folder is owned by another repository, refuses rather than suffixes. Rename and move survival is provided by the stability invariant, so the pin is a naming convenience (OQ-2 answered: kept as first-assignment pin).

### D-2 What is "moai-made" (Operator decision, from `dab5`)

Rule: not primary, not integration, not `agent-<hex>`, not Codex-managed, and a creation-class witness: (a) directory under `.moai/worktrees`, (b) branch begins `WT-`, or (c) directory under `<MoaiHome>/worktrees/projects/<key>/` with an owner marker that names this repository. Witness (c) is what keeps the four non-`WT-` trees moai-made after the move: without it they would fall out of (a) and become `unclassified` at the home root, and the tier decision would let `moai worktree done` accept them (its origin-landing check is guarded by `isCardBranch`, `done.go:315`, `WT-` only). Measured shapes at HEAD `74ad197380`:

| Shape | Count | Class |
|-------|-------|-------|
| `.moai/worktrees/*` with `[WT-…]` branch | 188 | moai-made |
| `.moai/worktrees/*` without `WT-` (`rc30-preview`, `release-v314`, `t1515-absorb`, `t1868-publish`) | 4 | moai-made by directory |
| `.claude/worktrees/*` named `agent-<hex>` | 31 | claude-native — untouched |
| `.claude/worktrees/*` other, with `WT-` branch (`t1241`, `t1356`, `t1452`, `t1459`, `t1497`, `t1502`, `t1507`, `t1508`, `t1510`, `t1511`, `t1513`, `t1868-publish-cc`, `agents-md-primary`) | 13 | moai-made by branch |
| `.claude/worktrees/audit-gate-async-worktree-doc` (`worktree-` branch) | 1 | unclassified — listed, never moved |
| `.claude/worktrees/main` (branch `main`) | 1 | integration — untouched |
| `~/.codex/worktrees/*` | 12 | codex-managed — untouched |
| primary checkout | 1 | primary |

(188 + 4 + 31 + 13 + 1 + 1 + 12 + 1 = 251, matching the `git worktree list` line count.) The two classifier outliers matter: `audit-gate-async-worktree-doc` is probably a Claude-created `EnterWorktree` tree (its branch is the `worktree-<name>` form that `worktree-integration.md` says `EnterWorktree` auto-names) and would be silently moved by any path-prefix rule; failing closed (unclassified) is the safe answer, and the operator can classify it by renaming its branch or moving it by hand.

**Alternatives rejected:** *move everything under `.claude/worktrees` except `agent-*`* (would move the integration worktree `main`, which docs and `internal/mission/integration.go:41` address by fixed path, and the unclassified tree); *classify by branch only* (misses the four non-`WT-` trees in `.moai/worktrees`); *classify by who created the tree via the state file* (the file is incomplete: only launcher and hook paths write to it, `worktree_registry.go:29-65`).

### D-3 Tier is a property of the creation class (REQ-WHR-009)

`isL1WorktreePath` (`done.go:230-247`) and `sweepTierOf` (`sweep.go:574-580`) decide L1 versus L2 by "is the path under `<main>/.claude/worktrees` or `<main>/.moai/worktrees`". The L2 class (SPEC-scoped persistent tree under `~/.moai/worktrees/<project>/<SPEC>`) is described in `worktree-integration.md` but, measured, the home directory holds no L2 tree today (seven near-empty directories). If the home-root tree were classified by path alone it would become an L2 disposal-eligible tree: `done` would accept it. The classifier decouples tier from location, and its witness (c) keeps a moved non-`WT-` tree L1. **Rejected:** keeping the prefix test and adding the home root to the L1 list (works for now, but hard-codes a third prefix into the safety guard; the guard exists precisely to refuse by class, and `--force`-immunity is its invariant, `done.go:227-229` annotation); making `done` refuse everything under `~/.moai/worktrees` (breaks any real L2 tree a user created with the older layout).

### D-4 Transition mechanism: dual-read, no symlinks (Operator decision)

**Chosen:** the resolver returns candidates (home, legacy-moai, legacy-claude); recorded paths win; the transition ends when a later release removes the legacy candidates, not by date.

**Alternative rejected — symlink at the old path to the new location.** Costs: (1) `git worktree move` already updates git's own admin files; a symlink at the old path does not help git but does make two spellings of one tree, which the classifier, `sweep`, `done` and the session anchor logic would have to unify; (2) the L1 guard resolves symlinks (`canonicalTierPath`), so a symlinked legacy path would classify as the home root and defeat the legacy-path fixtures; (3) `launcher.go:923-948` rejects out-of-prefix paths after symlink resolution, so behavior would depend on which spelling the user typed; (4) directory symlinks need privilege on Windows; (5) a symlink inside `<root>/.moai/worktrees` re-enters the repository tree, which is what the unification removes; (6) `moai update`'s existing migration explicitly refuses a symlinked destination parent (`update_worktree_migration.go:89-95`), so the codebase already treats this shape as a hazard. Benefit would be zero-change for readers not yet migrated; the dual-read resolver gives that benefit explicitly for the 13 known readers, and the grep-sweep acceptance (AC-WHR-007) guards against a thirteenth.

**Alternative rejected — leave the old trees where they are forever (only new trees at the home root).** Satisfies "no new tree outside the root" but contradicts the card's "retire the old directories for MoAI-made trees". Kept available: the migration is explicit and per-tree, so an operator who never applies it still has a working system.

### D-5 Launcher short name for a new tree (Operator decision; depends on PM-3)

Claude Code creates `.claude/worktrees/<name>` itself when `claude --worktree <name>` receives a bare name; the only mechanisms to change that location are the WorktreeCreate hook (rejected by `dab5`) or not passing a bare name. `moai cc -w <name>` is MoAI's launcher, so it can materialize the tree itself and pass an absolute path, which `launcher.go:911-919` already documents Claude Code accepts for an existing tree. Decided as a conditional (spec §E D-5): if PM-3 records PASS, materialize at the home root and pass the absolute path (branch X); if PM-3 records FAIL, pass the short name through and record the single documented exception (branch Y). PM-3 is the selector; a Gap blocks the Kickoff.

**Alternatives rejected:** *WorktreeCreate hook* (replaces native creation for every isolation spawn; empty stdout or non-zero exit aborts every isolation spawn; `.worktreeinclude` and `baseRef` stop being processed; the shipped wrapper is unregistered by design — worktree-integration-ops.md "Not Registered by Default"; changelog snapshot lines 7215, 5540). *Leave the short-name path native* (the completion criterion "new trees never appear outside the root" would carry a documented exception; acceptable only if the operator chooses it or PM-3 fails). *Intercept `EnterWorktree` creation* (it is a Claude-side tool call the harness cannot redirect).

**Effect on `.worktreeinclude`:** the file is processed by Claude for natively created trees; a MoAI-materialized tree does not get it. This is already true of `moai worktree new` trees and is not a regression; the shared materializer applies git identity, safe.directory and Codex hook seeding instead (`session_worktree.go:251-264`). Whether a `.worktreeinclude` equivalent should be added to the materializer is not decided here (not in scope).

### D-6 / D-7 Migration trigger and live apply (DECIDED, lane judgment under auto-semantics §6; the operator may override before kickoff)

Two separate questions. **Trigger:** the card (and `ef12`) read as "the `moai update` consolidation changes destination". Decided: `moai update` **previews** — it prints the plan (destination is the home root), the exact `moai worktree migrate --apply` command, and moves nothing; the explicit verb moves. **This is a behavior change.** Today the non-dry-run path of `moai update` applies the consolidation (`internal/cli/update.go:410`, `runUpdateWorktreeMigration(lockRoot, false, out)`; the dry-run path at `update.go:366` previews). Reason: the non-destructive invariant — an automatic update must not move trees that hold the only copy of unpushed work, and the migration now has a non-git side effect (card rebind, state re-key) at a scale of ~200 trees on machines with live lanes. Rejected: `update` applies (the ef12 literal reading; a one-line change in the caller). **Who applies to the real repository:** the leader, once, after landing, with lanes quiescent; the lane delivers the tool, the scratch-repository rehearsal and the real-repository dry-run listing only. Rationale: the skip rules would skip any tree with a live anchor so a loaded machine yields a partial migration, and "one writer per tree" is violated by a lane reaching into another lane's tree. Rejected: lane applies at landing.

### D-8 Config surface (Operator decision)

`MOAI_HOME` already relocates every `~/.moai` consumer (`paths.go:56-77`) and is the test seam. The git-strategy key `worktree_root` exists (`types.go:202`, `git-strategy.yaml:108` empty) but is documented in `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` as a dead, deprecated field with zero readers; reviving it could silently activate stale values on installations that set it. Recommended: leave it unread and add one new key, `worktree_project_key`. **Rejected:** reviving `worktree_root` as an absolute-root override (a second way to relocate the tree besides `MOAI_HOME`, with no validation story); an environment variable for the key (invisible in review, not reproducible).

### D-9 Integration worktree stays (Operator decision)

`.claude/worktrees/main` (the dedicated landing surface) is addressed by fixed path in `AGENTS.local.md` and the lane protocol, and `internal/mission/integration.go:41` checks the suffix `/.claude/worktrees/develop`. Relocating it is a separate decision with its own blast radius; out of scope.

### D-10 What does NOT change

- The `WT-` branch convention, the three traceability carriers, and the card-id-as-directory-leaf convention (the directory leaf remains the card id for card trees; the key sits one level above).
- The factory lease semantics: a card never moves to another tree (`card_worktree.go:42-44`) — the rebind is a new, narrow operation, not a relaxation of that rule.
- Disposal: `done`, `sweep`, `clean`, `remove` predicates and the hoist-before-dispose obligation.

## §3. Hook and guard impact (summary; evidence in research.md §R4)

| Surface | Impact | Handling |
|---------|--------|----------|
| Worktree-isolation guard text ("This session is isolated in the worktree <path> ...") | Emitted by the Claude Code runtime, compares against the session's own tree; not repository code (`SPEC-SESSION-ANCHOR-ATTR-001` spec records this). A home-root tree is its own absolute path; no change expected | PM-1b measures the in-session-hop refusal; t1879 owns text |
| `MCP project_root` validation (`mcp_worktree_root.go:60-93`) | Accepts a linked worktree by `git rev-parse` + primary identification, not by path prefix | No change; AC-WHR-007 includes a home-root `validateLinkedWorktreeRoot` case |
| BranchGuard | Keyed on primary checkout detection (`isPrimaryCheckout`, `branch_guard.go:1001`) | No change |
| `handleWorktreeMove` (EnterWorktree/ExitWorktree PostToolUse) | Stamps `MOAI_PROJECT_DIR` from the new cwd | No change; a home-root cwd passes through |
| Session registry path | Anchored to the primary checkout via git common dir (`registry.go:44-60`) | No change |
| `cleanupMoaiWorktrees` (every `moai cc` launch, removes `worker-*` trees by non-force `git worktree remove`) | Treats each direct child of `~/.moai/worktrees` as a base, so `projects` becomes a base. The match is `filepath.Rel(base, worktreePath)` rejecting only a `..` result (`launcher.go:505-518`), so a nested path `projects/<key>/worker-x` **is accepted** and removed (a clean tree with unpushed commits is removed, and its ignored reports go with it). Today `git worktree list` shows no `/worker-` tree, so there is no present exposure | Reader update plus a protection requirement (REQ-WHR-014, AC-WHR-014b): the cleanup never removes a registered tree with commits on no remote and never a tree the migration moved |
| Statusline | `QuotaStateDirs` reads `<root>/.git/worktrees/*/gitdir` (`quota_dirs.go:30-70`), location-agnostic | No change |
| Chain ledger | Stores `worktree_path` as history | Untouched by migration; history is not rewritten |
| Factory landing directory (`factoryCardWorktreeDir` and its consumers `factoryRefuseForeignWorktree`, `factoryCarryPreviousWorktree`, `factory_card.go:499-540`) | Builds `<root>/.moai/worktrees/<card>`; after migration the REQ-SD-011 refusal and the t1521 carry-over would look at the legacy path and neither refuse nor reuse a home-root tree | Reader update in M3, REQ-WHR-007, AC-WHR-007 |
| Codex launcher | `resolveCodexWorktreeDir` resolves names against two in-repo roots; absolute paths go through `resolveWorktreeL2Path` | M3 adds the home-root candidate; PM-2 and PM-5 gate guidance |

## §4. Failure modes designed for

1. **Half-done move.** Journal written before the move with a done/pending status per step; step order manifest, move, state re-key (`.moai/state/worktrees.json`), card rebind, manifest comparison; reversal runs in the opposite order; a failed reverse step stays pending and `--rollback` / a re-run is idempotent and converges, so the registered path, the state-file entry and the card record always agree with the real location; `--verify` exists. The manifest carries a content digest (tracked diff, untracked files, `.moai/reports`, ignored files within a recorded bound) so a content change is visible; what it does not hash is stated in its output (I-2).
2. **Two repositories, one name (any case).** Case-insensitive owner-marker guard, deterministic suffix for the later arrival only, stable key, corrupt-marker refusal.
3. **Reader missed.** Recorded path precedence limits the blast radius; the grep sweep (AC-WHR-007) is a standing check, not a one-time audit.
4. **Entry behavior differs from assumption.** PM gate before run; every guidance sentence traced to a line of the record.
5. **Case-insensitive volumes.** Spelling-safe equality is a requirement, not a nicety: 25 of 192 trees are registered under the lowercase spelling today.

## §5. Rejected scope expansions (for the record)

- Registering WorktreeCreate/WorktreeRemove (dab5).
- Moving `agent-*` trees or cleaning them (dab5; OQ-3 keeps the question visible through `list`).
- A new SQLite table or registry for trees: the factory record plus `git worktree list` already carry everything needed; a parallel registry would be a second source of truth that disagrees under rollback.
- Deleting trees as part of "retirement" (REQ-WHR-014).
- Rewriting the wider EnterWorktree guidance (t1879).
