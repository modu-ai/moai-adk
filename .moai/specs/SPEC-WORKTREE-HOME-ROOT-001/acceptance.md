# acceptance.md — SPEC-WORKTREE-HOME-ROOT-001

> Verification layer. Every entry is a Given-When-Then criterion (`AC-WHR-NNN`), binary-testable. The GEARS obligation lives in spec.md §C; nothing here restates a requirement as GEARS. 24 criteria cover 24 requirements one-to-one (Tier L ceiling 25).
>
> **Two-cell discipline (verification-completeness §2).** Each criterion carries a RED-now cell (what is red on the pre-change tree and why) and a green path (which milestone flips it). Plan phase ran no Go commands, so a RED-now cell is marked **Observed (plan time)** only where a read-only command in this tree produced the quoted outcome, with the tree SHA `74ad197380`; every other cell is **To record at M0** — the run phase records command, verbatim stdout, exit code and tree SHA before the first production change, in a baseline commit that precedes the implementation commit (verification-claim-integrity §2.3).
>
> **Test layout.** Unless a criterion names another form, "test" means a Go test in the touched package using `t.TempDir()` repositories and a temporary `MOAI_HOME`; no test touches the real `~/.moai`, the real repository's worktrees, or global git config. Selectors are listed count-first (a `-list` count precedes each `-run`) so an absent test cannot pass as an empty sweep.

## §A. Coverage matrix

| REQ | AC | Kind | Severity |
|-----|----|------|----------|
| REQ-WHR-001 | AC-WHR-001 | Go test | MUST |
| REQ-WHR-002 | AC-WHR-002 | Go test | MUST |
| REQ-WHR-003 | AC-WHR-003 | Go test | MUST |
| REQ-WHR-004 | AC-WHR-004 | Go test | SHOULD |
| REQ-WHR-005 | AC-WHR-005 | Go test + grep | MUST |
| REQ-WHR-006 | AC-WHR-006 | Go test + measurement | MUST (conditional on PM-3, D-5) |
| REQ-WHR-007 | AC-WHR-007 | Go test + grep sweep | MUST |
| REQ-WHR-008 | AC-WHR-008 | Go test | MUST |
| REQ-WHR-009 | AC-WHR-009 | Go test | MUST |
| REQ-WHR-010 | AC-WHR-010 | Go test + CLI | MUST |
| REQ-WHR-011 | AC-WHR-011 | Go test | MUST |
| REQ-WHR-012 | AC-WHR-012 | Go test | MUST |
| REQ-WHR-013 | AC-WHR-013 | Go test | MUST |
| REQ-WHR-014 | AC-WHR-014 | Go test + static check | MUST |
| REQ-WHR-015 | AC-WHR-015 | Go test | MUST |
| REQ-WHR-016 | AC-WHR-016 | Go test | MUST |
| REQ-WHR-017 | AC-WHR-017 | Go test (rehearsal) | MUST |
| REQ-WHR-018 | AC-WHR-018 | Go test | MUST |
| REQ-WHR-019 | AC-WHR-019 | Go test | MUST |
| REQ-WHR-020 | AC-WHR-020 | Go test + grep | MUST |
| REQ-WHR-021 | AC-WHR-021 | Measurement record + grep + byte count | MUST (gate) |
| REQ-WHR-022 | AC-WHR-022 | Go test | MUST |
| REQ-WHR-023 | AC-WHR-023 | Go test | MUST |
| REQ-WHR-024 | AC-WHR-024 | Go test + grep | MUST |

Cross-cutting completion criteria stated on card t1640 map as: "new trees are never created outside the root" → AC-WHR-005, AC-WHR-006; "entry measured on both harnesses" → AC-WHR-021.

## §B. Acceptance criteria

### AC-WHR-001 — Home-root path and `MOAI_HOME` relocation
**Given** a temporary `MOAI_HOME=/tmp/x/moai-home` and a temporary repository `proj` with no marker. **When** the resolver is asked for the path of tree `t9`. **Then** it returns `/tmp/x/moai-home/worktrees/projects/proj/t9`, creates nothing on disk, and a second call with `MOAI_HOME` unset resolves under `$HOME/.moai/worktrees/projects/proj/t9`.
RED-now: To record at M0 — no resolver exists (the destination is `filepath.Join(projectRoot, sessionWorktreeSubdir, branch)` at `internal/cli/session_worktree.go:223`, constant at line 48; Observed (plan time): `grep -n 'sessionWorktreeSubdir = ' internal/cli/session_worktree.go` printed line 48). Green path: M1.

### AC-WHR-002 — Key derivation and marker
**Given** a primary checkout at a path containing characters outside `[A-Za-z0-9._-]` (for example `my proj (v2)`), invoked from a linked worktree of that repository. **When** the resolver derives the key. **Then** the key is the sanitized primary-checkout basename (`my-proj--v2-`), derived from the primary root and not from the invoking worktree, and a marker `.moai-project.json` is written at `<MoaiHome>/worktrees/projects/<key>/` holding the schema version, primary root and creation time.
RED-now: To record at M0 (no marker code). Green path: M1.

### AC-WHR-003 — Collision between two repositories with one name
**Given** repositories `/a/proj` and `/b/proj`, and a marker under key `proj` recording `/a/proj`. **When** the resolver runs for `/b/proj`. **Then** it selects `proj-<8 hex of SHA-256 of the case-folded canonical root>`, writes a marker for `/b/proj` under that key, never places a tree under `proj`, returns the same key on a second call, and a corrupt or unparseable marker under the candidate key makes creation refuse with a message containing the marker path.
RED-now: To record at M0 (no collision logic; today both repositories would share nothing because the layout does not exist — the hazard is introduced by the new layout, so the criterion's RED is "no guard exists"). Green path: M1.

### AC-WHR-004 — Operator key pin
**Given** `git_strategy.worktree_project_key: "fixedkey"` in the repository's `git-strategy.yaml` and the repository later renamed. **When** the resolver runs before and after the rename. **Then** both runs return the key `fixedkey` (after the marker guard), and with the key empty the result equals AC-WHR-002's derivation. A value containing a separator or `..` is rejected.
RED-now: Observed (plan time): `grep -rn 'worktree_project_key' internal .moai/config` printed no line (key absent). Green path: M1.

### AC-WHR-005 — No creation path names a legacy root
**Given** the run-phase tree after M2. **When** `moai worktree new <n>`, the factory card materializer, the session-worktree auto-entry and the launcher `-w <n> --branch <b>` path each create a tree in a temporary `MOAI_HOME` fixture. **Then** each tree is registered under `<MOAI_HOME>/worktrees/projects/<key>/<n>`, has the `WT-` branch convention, the configured base branch (existing `SPEC-WORKTREE-BASEREF-001` tests stay green), the global `safe.directory` registration (via the existing seam), and a seeded `.codex/hooks.json`; and a static scan finds zero creation call sites constructing `".claude","worktrees"` or `".moai","worktrees"` destinations in production code (count-first: the scan's expected-site list is printed before its result).
RED-now: Observed (plan time): `sessionWorktreeSubdir` is `.moai/worktrees` (`session_worktree.go:48`) and `launcherWorktreeMaterializeReal` builds `filepath.Join(projectRoot, ".claude", "worktrees", name)` (`worktree_branch_flag.go:109`). Green path: M2.

### AC-WHR-006 — Launcher short name creates at the home root (conditional)
**Given** pre-run measurement PM-3 passed and decision D-5 is "materialize". **When** `moai cc -w newname` (and `moai glm -w newname`) runs with no tree of that name in any root, using a stub backend that records its argv. **Then** the shared materializer ran, the backend received the absolute home-root path, no directory `.claude/worktrees/newname` exists afterwards, and with a tree already present at a legacy root the launcher enters that tree and creates nothing. **If PM-3 failed or D-5 is "leave native"**, this criterion is replaced by a recorded exception and an operator-approved SPEC amendment before M2 proceeds; it is not silently skipped.
RED-now: To record at M0 (today the short name is passed through for Claude Code to create under `.claude/worktrees`: `launch_chain.go:56-71`, `launcher.go:911-914`). Green path: M2.

### AC-WHR-007 — Readers resolve through the resolver; sweep finds no stray derivation
**Given** a fixture repository with one tree at the home root (card `t7`), one at `.moai/worktrees/t7b`, one at `.claude/worktrees/t7c`, and recorded `worktree_path` for `t7`. **When** each reader named in REQ-WHR-007 resolves the tree: `cardIDFromRegisteredWorktree`, `cardIDFromPath`, `autoOwnerWorktrees`/`ownerAlive`, `DefaultWorktreeBases`+`ReadCardStatus`, `resolveLaunchWorktreePath`, `resolveWorktreeLaunchTarget`, `seedAdmittedWorktreeHooksFor`, `ccWorktreeWriterPrecheck`, `resolveCodexWorktreeDir`, `cleanupMoaiWorktrees`. **Then** each finds the tree at its actual location, the recorded path wins over a derived candidate, and a grep sweep over production Go for the literal pairs `".claude", "worktrees"`, `".moai", "worktrees"` and `Base(parent) == "worktrees"` returns only the sites on a reviewed allow-list (resolver internals, legacy-candidate list), its count printed first.
RED-now: To record at M0; the 12-site inventory is in research.md §R3. Green path: M3.

### AC-WHR-008 — No misidentification at the new depth
**Given** a session whose cwd is `<MOAI_HOME>/worktrees/projects/proj/t7/internal/cli`. **When** SessionStart derives the card id (both derivations) and `todo --auto` judges the owner of card `t7` whose recorded tree is that path with a live registry entry. **Then** the card id is `t7` (not `projects`, not `proj`), and the owner is judged alive.
RED-now: To record at M0 — expected red because `cardIDFromPath` returns the first ancestor whose parent is named `worktrees` (`session_start_record.go:226-241`), which for the new layout is the directory `projects`, and `autoOwnerWorktrees` lists only the two in-repo candidates (`todo_auto.go:88-93`). Green path: M3.

### AC-WHR-009 — Tier by creation class
**Given** a home-root moai-made tree with a landed branch and a clean state, and an L2 registry-style tree. **When** `moai worktree done <home-root tree>` runs with and without `--force`, and `moai worktree sweep --json` runs. **Then** `done` refuses the home-root tree both times with the L1 refusal text, `sweep` reports its tier as `L1` and applies only its existing landing predicates, and the legacy-rooted fixtures behave exactly as before (existing `done_l1_tier_guard_test.go` stays green).
RED-now: To record at M0 — expected red because `isL1WorktreePath` answers by prefix (`done.go:230-247`) and a home-root tree is outside both prefixes. Green path: M3.

### AC-WHR-010 — `moai worktree list`
**Given** a fixture repository with one tree of each class (primary, home, legacy-moai WT-, legacy-moai non-WT, legacy-claude moai-made, `agent-*`, integration, Codex-managed path) and one tree registered under two spellings of the same directory. **When** `moai worktree list --json` and `moai worktree list` run. **Then** exit code is 0, every tree appears exactly once (the two spellings collapse to one row), each row has name, path, root kind, class, branch, short HEAD, dirty flag, unpushed-commit count and anchored flag, and the run mutates nothing (tree hash before equals after).
RED-now: Observed (plan time): `moai worktree list` printed the parent command's help text (first line "Manage Git worktrees for parallel SPEC development: new, sync, remove, clean, sweep, recover, done, hoist, plus the guard verbs...") — there is no `list` verb registered in `internal/cli/worktree/root.go:42-54`. Green path: M4.

### AC-WHR-011 — Card record rebind
**Given** a factory run with card `t7` recording worktree path `P_old`, a replacement run row for `t7` also carrying `P_old`, and a moved tree at `P_new`. **When** the rebind operation runs with `(card=t7, old=P_old, new=P_new)`. **Then** every run row of `t7` carrying `P_old` now carries `P_new`, one `card.fields` event records both paths, `PreviousCardWorktree("t7", otherRun)` returns `P_new`, a rebind whose `old` does not equal the current value is refused, a rebind to an unregistered path or a different HEAD/branch is refused, and `RecordCardWorktree` still refuses to move a card onto a different tree (existing behavior test unchanged).
RED-now: To record at M0 — no rebind operation exists; `RecordCardWorktree` refuses any path change (`card_worktree.go:42-44`). Green path: M5.

### AC-WHR-012 — Other path-bearing state
**Given** `.moai/state/worktrees.json` holding an entry for `P_old`, a session registry entry whose cwd is `P_old` with a dead pid, and a chain-ledger node naming `P_old`. **When** the migration moves the tree. **Then** the state-file entry is keyed to `P_new`, the session registry and ledger entries are unmodified byte for byte, and the journal lists each store as touched or deliberately untouched.
RED-now: To record at M0. Green path: M5, M6.

### AC-WHR-013 — Classifier
**Given** the six measured shapes: home tree, legacy-moai `WT-` tree, legacy-moai tree with a non-`WT-` branch (`rc30-preview`, `release-v314`, `t1515-absorb`, `t1868-publish` are such trees in this repository), legacy-claude moai-made tree (`t1497`), `agent-<hex>` tree, integration worktree holding the integration branch, plus a tree whose branch is not `WT-` and whose directory is not under `.moai/worktrees` (`audit-gate-async-worktree-doc`). **When** the classifier runs. **Then** the results are, respectively: moai-made, moai-made, moai-made (directory under the MoAI legacy root), moai-made (branch `WT-`), claude-native, integration, unclassified; unclassified is listed and counted and never moved; and `moai worktree list`, `moai worktree migrate` and the `moai update` consolidation return the same classification for the same input (one shared function).
RED-now: Observed (plan time): `grep -n 'agent-' internal/cli/update_worktree_migration.go` printed nothing — the existing consolidation selects by path prefix only (`update_worktree_migration.go:87,105-108`) and therefore includes `agent-*` trees. Green path: M5, M6.

### AC-WHR-014 — Non-destructive invariant (static and behavioral)
**Given** the migration engine source. **When** a static check scans it for the forbidden operations (`worktree remove`, `worktree prune`, `branch -D`, `reset`, `clean`, `checkout`, `stash`, `os.Remove`, `os.RemoveAll`, `rm`), and a behavioral test runs `--apply` against the rehearsal repository. **Then** the static check finds none (its swept set non-empty — the file list is printed and counted), the only git subcommand the engine runs that mutates a tree is `worktree move`, and after the rehearsal every branch tip that existed before still exists with the same commit id and `git worktree list` registers the same number of trees as before.
RED-now: To record at M0 (engine absent; the static check must also be shown red on a deliberately planted `git worktree remove` line before it is trusted — verification-completeness §1.1). Green path: M6.

### AC-WHR-015 — Dry-run is read-only; apply and verify exist
**Given** the rehearsal repository (committed-unpushed tree, dirty tree with untracked and ignored files, locked tree, occupied-cwd tree, `agent-*` tree). **When** `moai worktree migrate` runs with no flag. **Then** exit code is 0, the listed plan shows each tree's class, source, destination and skip reason, no path under the repository or `MOAI_HOME` is created, modified or removed (before/after content hash equal), and `--apply` is required to move anything; **and** after `--apply`, `--verify` reports each recorded move as `intact`.
RED-now: To record at M0 (no verb). Green path: M6.

### AC-WHR-016 — Skip conditions
**Given** the rehearsal fixtures for each skip cause: locked, prunable, process cwd inside the tree, live anchored session, source missing, destination exists, destination on another filesystem (where a second volume exists; otherwise recorded as a Gap), process-cwd inspection unavailable. **When** `--apply` runs. **Then** every such tree stays at its old path, remains a registered worktree with unchanged HEAD and working state, and the output names the cause for each; when process cwd inspection is unavailable no tree moves at all.
RED-now: To record at M0. Green path: M6.

### AC-WHR-017 — Manifest equality and automatic reversal; unpushed branch proof
**Given** a rehearsal tree on branch `WT-unpushed` with two commits present on no remote, a modified tracked file, an untracked file, an ignored file, and a populated `.moai/reports`. **When** `--apply` moves it. **Then** the post-move manifest (HEAD, branch, unpushed commit list, status hash, ignored-file count, reports file count) equals the pre-move manifest field for field; `git -C <new> log --not --remotes` lists the same two commit ids as before; the file contents of the modified, untracked and ignored files are byte-identical; **and** with a fault injected after the move (a failing card rebind, then separately a tampered manifest) the tree is reversed to its source, the manifest at the source again matches, and the command reports the tree as not migrated. A forced reversal failure makes the exit code non-zero and names the tree left at the destination.
RED-now: To record at M0. Green path: M6. This is the operator's "nothing lost" proof for the unpushed-branch invariant.

### AC-WHR-018 — Explicit rollback
**Given** a journal from an applied migration of three trees, one of whose source paths has since been re-occupied. **When** `moai worktree migrate --rollback <journal>` runs. **Then** the two free trees return to their sources with matching manifests and rebound card records, the re-occupied one is refused with the reason, and the exit code is non-zero because one tree was refused.
RED-now: To record at M0. Green path: M6.

### AC-WHR-019 — `moai update` consolidation
**Given** a primary checkout with legacy-rooted moai-made trees and `agent-*` trees. **When** the update consolidation runs (the existing entry point `runUpdateWorktreeMigration`). **Then** no `agent-*` tree appears in its plan or is moved, under the default decision D-6 nothing is moved and the output prints the dry-run summary plus the exact `moai worktree migrate --apply` command, and under the alternative decision (apply) only classifier-selected trees move. Its idempotence and primary-checkout-only behavior (a linked-worktree update leaves siblings alone) are unchanged.
RED-now: To record at M0 — expected red: today the consolidation moves every `.claude/worktrees/*` tree including `agent-*` (selection at `update_worktree_migration.go:87,105-108`; the function documentation at lines 233-235 states the primary-only rule). Green path: M6.

### AC-WHR-020 — Retirement
**Given** a fixture in which the last tree under the MoAI legacy root has been migrated and the directory is empty, then a second fixture where the directory still contains an untracked file. **When** the retirement step runs. **Then** the empty directory is removed with a non-recursive removal and the non-empty one is left in place with a notice; `.claude/worktrees` and every claude-native tree in it are untouched in both fixtures; and a repository-wide scan of production creation code finds no legacy-root destination (this scan shares AC-WHR-005's mechanism).
RED-now: To record at M0. Green path: M7.

### AC-WHR-021 — Pre-run measurement record, entry on both harnesses, guidance
**Given** the run phase has not started. **When** the gate is checked. **Then** `.moai/reports/t1640/pre-run-measurements.md` exists and records, for PM-1a, PM-1b, PM-1c, PM-2 (and PM-3..PM-6, each PASS, FAIL with its stated consequence, or Gap) the command, verbatim output, exit status, tool versions and tree SHA; **and after M7**, (a) a count-first grep over the in-scope guidance surfaces (AGENTS.md, the two worktree rule files, the skills and agents named in plan §D, their template mirrors, the cc/glm/worktree help strings) returns zero sentences instructing a session to run `moai cc -w` from inside Claude Code, (b) every entry-behavior statement in the guidance is traceable to a line of the measurement record, (c) the home-root location appears in each surface that names a worktree location, (d) `wc -c AGENTS.md` is at most 23780, and (e) the template-mirror consistency test passes.
RED-now: Observed (plan time): `wc -c AGENTS.md` printed 23780; `.moai/reports/t1640/` did not exist in this tree before the SPEC (the directory is created by the run phase). Green path: M0 (gate), M7 (guidance).

### AC-WHR-022 — Transition resolver and ambiguity
**Given** a name `t7` present both at the home root and at `.moai/worktrees/t7` (two different trees), and a name `t8` at only one legacy root. **When** a launcher short-name lookup and `resolveCodexWorktreeDir` run for each. **Then** `t8` resolves to its single location, `t7` is reported ambiguous with a message instructing an absolute path and naming both candidates, and neither lookup guesses.
RED-now: To record at M0 (today the Codex resolver reports ambiguity for two roots, `codex_launcher.go:491-496`; the three-root case is new). Green path: M3.

### AC-WHR-023 — Spelling-safe equality
**Given** a case-insensitive volume where `/x/MoAI` and `/x/moai` are one directory (the machine under test reports one inode for `/Users/goos/MoAI` and `/Users/goos/moai`: Observed (plan time) `stat -f '%i'` printed 105929094 for both), and a tree registered under the lowercase spelling. **When** equality and containment tests run (`isL1WorktreePath`, resolver `Same`, the classifier, the migration skip checks, `moai worktree list`). **Then** the two spellings are equal and the tree is reported once; and on a case-sensitive fixture directory two directories differing only by case remain distinct.
RED-now: Observed (plan time): `git worktree list` printed 25 of the 192 MoAI-root trees under `/Users/goos/moai/moai-adk-go/.moai/worktrees` while the other 167 print under `/Users/goos/MoAI/...`. Green path: M1 (helper), M3-M6 (use).

### AC-WHR-024 — Claude-owned and Codex-owned state untouched; no WorktreeCreate hook
**Given** the rehearsal repository containing `agent-*` trees under `.claude/worktrees`, a tree under a fake `~/.codex/worktrees`, and the integration worktree. **When** `list`, `migrate --apply`, `--rollback`, retirement and `update` all run. **Then** none of those trees is moved, renamed, pruned or removed, their `git worktree list` records are byte-identical before and after, and `grep -rn 'WorktreeCreate' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` finds no registered hook entry.
RED-now: Observed (plan time) for the second clause: the WorktreeCreate handler exists and is documented as not registered (`.claude/rules/moai/workflow/worktree-integration-ops.md`, "WorktreeCreate and WorktreeRemove Hooks (Not Registered by Default)"); the first clause fails today because the old consolidation selects `agent-*` (AC-WHR-013). Green path: M6.

## §C. Edge cases (each is covered by an AC above or recorded as a Gap)

- Repository with zero legacy trees: `migrate` prints "0 moai-made trees" and exits 0; `update` prints nothing new.
- Tree whose directory name contains a space or non-ASCII: the sanitizer touches only the project key; tree names are validated as today (`validateNewWorktreeName`, `new.go:56-63`).
- `MOAI_HOME` relative or unset: `paths.MoaiHome()` already disregards a relative value and falls back to `$HOME/.moai` (`paths.go:67-76`).
- A tree being created while a migration runs: creation and migration serialize on the existing factory step lock for card trees (`factory.AcquireFactoryStepLock`, `factory_card.go:762`); non-card creation during `--apply` is recorded as a Gap, not assumed safe.
- Interrupted `--apply` (process killed between `git worktree move` and verification): the journal entry is written before the move, so `--verify` classifies the tree as `drifted` or `intact` and `--rollback` can restore it.
- Windows paths: unmeasured (OQ-4).

## §D. Quality gates and Definition of Done

- TRUST 5: scoped tests green for every touched package with coverage at or above the repository threshold for the new packages; lint clean on touched packages; cross-platform build of touched packages; no `AskUserQuestion` in subagent-domain code; conventional commits carrying `card t1640`.
- Every release-blocking criterion above has its RED-now cell recorded at M0 and its green output recorded at the milestone named, with command, verbatim output, exit code and SHA in `progress.md §E.2` (written by the run-phase specialist, not by this plan).
- The operator decisions D-1..D-9 are each recorded as confirmed or default-applied.
- The pre-run measurement record exists and every FAIL in it has been resolved by the stated branch.
- Sync phase evidence includes the real repository's dry-run listing (counts per class) and the leader-operated apply command; no live tree was moved by the lane.
- Plan audit by codex PASS recorded at `.moai/reports/t1640/` before the run phase (Class C requirement).
