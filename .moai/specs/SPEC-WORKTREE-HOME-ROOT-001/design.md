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
move:    classifier ── safety planner ── manifest ── git worktree move ── manifest ── card rebind ── journal ── move list
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

### D-6 / D-7 Migration trigger and who applies (DECIDED BY THE OPERATOR; ruling relayed by the leader 2026-10-19)

The lane's earlier judgment — `moai update` only previews, and the leader applies the live migration through the explicit verb — is reversed. **Decided:** `moai update` keeps applying the consolidation as it does today (`internal/cli/update.go:410`, `runUpdateWorktreeMigration(lockRoot, false, out)`); the call at `update.go:366` stays the preview that `moai update --dry-run` runs (the flag is defined at `update.go:85`). What changes is the destination — the home root instead of `<root>/.moai/worktrees` (`update_worktree_migration.go:88`) — and the selection — the classifier of REQ-WHR-013, so `agent-*` trees (dab5) and unclassified trees are no longer selected. The explicit verb stays for rehearsal, inspection and recovery: `moai worktree migrate` previews by default, `--apply` moves, `--verify` and `--rollback` remain.

**Why the non-destructive invariant still holds.** It is met by the devices on each move, not by withholding the move: (a) a move is `git worktree move` only — never a delete, never a bare filesystem move; (b) a tree with an active process or a live session is skipped (existing causes at `update_worktree_migration.go:118-139`); (c) a destination collision skips that tree, is reported and never overwrites; (d) the before/after manifest (HEAD, branch, unpushed commits, status hash, content digest) is compared for every move and a mismatch reverses it. Skipped trees stay usable at their old path, and the update's own retry (`update.go:407-409`) lets a later update take them. A failure in the consolidation is reported through the update ledger and does not fail the update (`update.go:410-414`).

**Who applies.** `moai update` applies the live migration; there is no separate applier and no gate in front of it. The leader's part is the scratch-repository rehearsal check (plan M6) and a confirmation of the move list after the first real run — an acceptance observation (acceptance.md O-1), not a gate.

**Rejected by the operator: a preview-only update with a leader-operated apply after landing.** Its two reasons — a partial migration on a loaded machine, and "one writer per tree" — are answered by the skip devices (a loaded tree is skipped, not moved) and by the retry on the next update. **Also rejected:** lane applies at landing.

### D-8 Config surface (Operator decision)

`MOAI_HOME` already relocates every `~/.moai` consumer (`paths.go:56-77`) and is the test seam. The git-strategy key `worktree_root` exists (`types.go:202`, `git-strategy.yaml:108` empty) but is documented in `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` as a dead, deprecated field with zero readers; reviving it could silently activate stale values on installations that set it. Recommended: leave it unread and add one new key, `worktree_project_key`. **Rejected:** reviving `worktree_root` as an absolute-root override (a second way to relocate the tree besides `MOAI_HOME`, with no validation story); an environment variable for the key (invisible in review, not reproducible).

### D-9 Integration worktree stays (Operator decision)

`.claude/worktrees/main` (the dedicated landing surface) is addressed by fixed path in `AGENTS.local.md` and the lane protocol, and `internal/mission/integration.go:41` checks the suffix `/.claude/worktrees/develop`. Relocating it is a separate decision with its own blast radius; out of scope.

### D-10 What does NOT change

- The `WT-` branch convention, the three traceability carriers, and the card-id-as-directory-leaf convention (the directory leaf remains the card id for card trees; the key sits one level above).
- The factory lease semantics: a card never moves to another tree (`card_worktree.go:42-44`) — the rebind is a new, narrow operation, not a relaxation of that rule.
- Disposal: `done`, `sweep`, `clean`, `remove` predicates and the hoist-before-dispose obligation.

### D-11 Move-list record (new with the operator's D-6/D-7 ruling)

**Problem.** Once `moai update` moves trees on its own, a lane that remembers a tree's old path (a dispatch pointer, a card note, a resumed session) needs a way to learn the new one. The factory card record's rebound `worktree_path` (REQ-WHR-011) is the primary pointer, updated in the journaled step order; it exists only for trees bound to a card. The move list is the fallback for the rest and the audit trail for all.

**Location.** `<primary>/.moai/state/worktree-migrate/move-list.jsonl`, beside the journals the engine keeps in `.moai/state/worktree-migrate/` (plan M6). The primary checkout's state area is the one place every tree of the repository can reach — a linked tree has its own `.moai/state/` (this tree's holds `.gitkeep` and runtime directories), which would hide the file — and it is machine-local (`.gitignore:401` ignores `.moai/state/`). A lane derives the primary root from `git rev-parse --path-format=absolute --git-common-dir`.

**Format.** JSON Lines, one object per line: `schema`, `ts`, `run`, `trigger` (`update` | `migrate` | `rollback`), `tree`, `card`, `old_path`, `new_path`, `result` (`pending` | `moved` | `skipped` | `failed` | `reversed` | `restored` | `refused`), `reason`. For a move `old_path` is the source and `new_path` the destination (a `pending` line, written just before the move, names the same two paths); for a skip, the source and the planned destination; for `reversed` and `restored`, the path the tree left and the path it arrived at. Example line:

```
{"schema":1,"ts":"2026-10-19T12:00:00Z","run":"20261019T120000Z","trigger":"update","tree":"t1821","card":"t1821","old_path":"/r/.moai/worktrees/t1821","new_path":"/h/.moai/worktrees/projects/r/t1821","result":"moved","reason":""}
```

**Idempotence and survival.** Append-only, one write per line, never rewritten or truncated; the writer runs under the migration lock (D-12) and starts each append on a fresh line (a newline first when the last byte is not one), so an unterminated tail left by a torn write becomes its own unparseable line instead of swallowing the next record. The dedupe key is the latest entry of the file that names either path of the candidate (`old_path` or `new_path` of any earlier entry): a candidate equal to it in `old_path`, `new_path`, `result` and `reason` is not appended, so an identical re-run adds nothing. The earlier key — the latest entry of the same `old_path` — dropped a re-applied move: after `moved` A→B and `restored` B→A, the latest entry of `old_path` A is still the old `moved` A→B, so a second A→B looked identical and was skipped, and the reader then answered A while the tree sat at B. Keyed on either path, the latest entry is the `restored` one and the re-applied move is appended; a tree moved from a path another tree once left differs in `new_path` and is appended. Rollback and automatic reversal append reverse entries, so a rollback does not erase history; a dry run writes nothing.

**Intent line.** Immediately before each `git worktree move` the engine appends a `pending` line naming the planned old and new paths. A kill between the move and the settling line leaves the move list with an unsettled `pending` line, the card record still on the old path and the old directory gone; the lookup answers from that line (below) instead of returning nothing until a later update converges.

**How a lane reads it (the lookup).** Match `old_path` with spelling-safe equality, and where the lane holds the tree name or a card id require the entry's `tree` or non-empty `card` to match too (two trees that used the same path at different times are then told apart). Take the latest `moved`, `restored`, `reversed` or `pending` entry for that path, follow its `new_path` through strictly later entries of those kinds, and skip every line that does not parse. If the entry reached last is an unsettled `pending`, answer with whichever of its two paths `git worktree list` registers (the tree is at the new path after a kill past the move, at the old path after a kill before it). Otherwise accept the final path only if `git worktree list` registers it. Worked cases: `moved` A→B then `restored` B→A then `moved` A→B answers B; after the first two it answers A; X moved A→B and a different tree Y later moved A→C answer B for X's card and C for Y's. The card record is read first.

**Rejected alternatives.** *Letting the lookup consult the journal instead of an intent line* (the journal is a per-run file keyed by timestamp that a lane would have to enumerate; the move list is the one file a lane is told about). *A symlink at the old path* (the D-4 cost list). *One JSON map rewritten per run* (a half-written file loses history, and a rollback would have to edit it). *A home-level file* (it would mix repositories; the project's own state area travels with the repository's tooling). *The card record alone* (trees without a card — the four non-`WT-` trees, `worker-*` trees — have none).

### D-12 Migration lock, reverse order and the state-file write (new in 0.4.0)

**One reverse order.** Forward: manifest, intent line, move, state re-key, card rebind, manifest comparison. Reverse (automatic reversal and explicit rollback alike): move back, state re-key back, card rebind back. The card rebind verifies at call time that its target is a registered worktree at the same HEAD and branch (REQ-WHR-011(b)); with the move reversed first the reverse target is the pre-move path, which is registered again by then. The original order (rebind first, move last) asked the rebind to accept a pre-move path while the tree still stood at the destination and was refused, so a failure after the forward rebind could not complete its own reversal. Between the reverse move and the reverse rebind the card record names a path that is no longer registered; the journal's pending status for the rebind step and the idempotent re-run close that window, and the three-store agreement is asserted after the whole sequence (AC-WHR-017a).

**One lock.** `migrate --apply`, `migrate --rollback` and the update-time apply take the same advisory lock file in `<primary>/.moai/state/worktree-migrate/` for the whole run. Without it two movers over different trees both read-modify-write `.moai/state/worktrees.json`, the card database and the move list; the registry's package mutex (`worktree_registry.go:22`) excludes goroutines of one process only, so one mover's re-key can be lost to the other's write. A mover that finds the lock held ends its run without moving, which the update treats as a reportable non-failure; the trees it did not reach are retried by the next update.

**Atomic state-file write.** The re-key writes a temp file in the same directory and renames it over the state file. Today the file is written by truncating `os.WriteFile` (`:122-143`) and `loadWorktreeEntries` returns an empty slice when the JSON does not parse (`:95`), so a crash mid-write followed by the next save erases every entry. The re-key therefore fails its step on a parse error and leaves the file as it is.

**Rejected:** a per-tree lock (two movers over different trees still race on the shared stores); a lock held only by the update (the explicit verb is the other mover).

## §3. Hook and guard impact (summary; evidence in research.md §R4)

| Surface | Impact | Handling |
|---------|--------|----------|
| Worktree-isolation guard text ("This session is isolated in the worktree <path> ...") | Emitted by the Claude Code runtime, compares against the session's own tree; not repository code (`SPEC-SESSION-ANCHOR-ATTR-001` spec records this). A home-root tree is its own absolute path; no change expected | PM-1b measures the in-session-hop refusal; t1879 owns text |
| `MCP project_root` validation (`mcp_worktree_root.go:60-93`) | Accepts a linked worktree by `git rev-parse` + primary identification, not by path prefix | No change; AC-WHR-007 includes a home-root `validateLinkedWorktreeRoot` case |
| BranchGuard | Keyed on primary checkout detection (`isPrimaryCheckout`, `branch_guard.go:1001`) | No change |
| `handleWorktreeMove` (EnterWorktree/ExitWorktree PostToolUse) | Stamps `MOAI_PROJECT_DIR` from the new cwd | No change; a home-root cwd passes through |
| Session registry path | Anchored to the primary checkout via git common dir (`registry.go:44-60`) | No change |
| `cleanupMoaiWorktrees` (every `moai cc` launch, removes `worker-*` trees by non-force `git worktree remove`) | Treats each direct child of `~/.moai/worktrees` as a base, so `projects` becomes a base. The match is `filepath.Rel(base, worktreePath)` rejecting only a `..` result (`launcher.go:505-518`), so a nested path `projects/<key>/worker-x` **is accepted** and removed (a clean tree with unpushed commits is removed, and its ignored reports go with it). Today `git worktree list` shows no `/worker-` tree, so there is no present exposure | Reader update plus a protection requirement (REQ-WHR-014, AC-WHR-014b): the cleanup removes a tree only when its branch is integrated and remotely landed by the predicate the disposal verbs use (a branch pushed but not integrated is kept), and never a tree the migration moved |
| Live-session containment (`internal/session/anchor.go`: `LiveAnchoredSessionsForProject` `:55`, `pathWithinTree` `:141`) | A cleaned-prefix test with a symlink-resolving fallback that does not fold case; on this machine `/Users/goos/MoAI` and `/Users/goos/moai` are one directory, so a session whose recorded cwd is a subdirectory spelled in the other case is not seen and its tree could be moved from under it | The test is made spelling-safe and subdirectory-aware (REQ-WHR-016/023, AC-WHR-016/019/023); `anchor.go` is on the run-phase allow-list |
| Statusline | `QuotaStateDirs` reads `<root>/.git/worktrees/*/gitdir` (`quota_dirs.go:30-70`), location-agnostic | No change |
| Chain ledger | Stores `worktree_path` as history | Untouched by migration; history is not rewritten |
| Factory landing directory (`factoryCardWorktreeDir` and its consumers `factoryRefuseForeignWorktree`, `factoryCarryPreviousWorktree`, `factory_card.go:499-540`) | Builds `<root>/.moai/worktrees/<card>`; after migration the REQ-SD-011 refusal and the t1521 carry-over would look at the legacy path and neither refuse nor reuse a home-root tree | Reader update in M3, REQ-WHR-007, AC-WHR-007 |
| Codex launcher | `resolveCodexWorktreeDir` resolves names against two in-repo roots; absolute paths go through `resolveWorktreeL2Path` | M3 adds the home-root candidate; PM-2 and PM-5 gate guidance |
| A lane holding only an old path after the routine update moved its tree | The old directory no longer exists | Card record rebind first (REQ-WHR-011), move list second (REQ-WHR-012, D-11) |

## §4. Failure modes designed for

1. **Half-done move.** Journal written before the move with a done/pending status per step; step order manifest, intent line, move, state re-key (`.moai/state/worktrees.json`), card rebind, manifest comparison; reversal runs the move first, then the re-key, then the rebind (D-12); a failed reverse step stays pending and `--rollback` / a re-run is idempotent and converges, so the registered path, the state-file entry and the card record always agree with the real location; `--verify` exists. The manifest carries a content digest (tracked diff, untracked files, `.moai/reports`, ignored files within a recorded bound) so a content change is visible; what it does not hash is stated in its output (I-2).
2. **Two repositories, one name (any case).** Case-insensitive owner-marker guard, deterministic suffix for the later arrival only, stable key, corrupt-marker refusal.
3. **Reader missed.** Recorded path precedence limits the blast radius; the grep sweep (AC-WHR-007) is a standing check, not a one-time audit.
4. **Entry behavior differs from assumption.** PM gate before run; every guidance sentence traced to a line of the record.
5. **Case-insensitive volumes.** Spelling-safe equality is a requirement, not a nicety: 25 of 192 trees are registered under the lowercase spelling today.
6. **A lane does not know its tree moved.** The routine update skips a tree with an active process or a live session, so a running lane's own tree stays where it is; a parked lane finds the new path through the rebound card record or the move list; a rolled-back move appends a reverse entry, so the chain resolves back to the original path.
7. **Two movers at once.** One migration lock for all three mutating entry points (D-12); the second mover moves nothing.
8. **Kill between the move and the settling line.** The intent line (D-11) lets the lookup answer from the registered one of its two paths; a re-run converges the stores from the journal.
9. **A torn move-list tail.** The next append starts on a fresh line and the reader skips every unparseable line.

## §5. Rejected scope expansions (for the record)

- Registering WorktreeCreate/WorktreeRemove (dab5).
- Moving `agent-*` trees or cleaning them (dab5; OQ-3 keeps the question visible through `list`).
- A new SQLite table or registry for trees: the factory record plus `git worktree list` already carry everything needed; a parallel registry would be a second source of truth that disagrees under rollback.
- Deleting trees as part of "retirement" (REQ-WHR-014).
- Rewriting the wider EnterWorktree guidance (t1879).
