# acceptance.md — SPEC-WORKTREE-HOME-ROOT-001

> Verification layer. Every entry is a Given-When-Then criterion (`AC-WHR-NNN`), binary-testable. The GEARS obligation lives in spec.md §C; nothing here restates a requirement as GEARS. 25 criteria cover 25 requirements one-to-one (Tier L ceiling 25); sub-labelled criteria (`a`/`b`) belong to their parent and count once.
>
> **Two-cell discipline (verification-completeness §2).** Each criterion carries a RED-now cell (what is red on the pre-change tree and why) and a green path (which milestone flips it). **Tree pin for every RED-now cell without its own pin: `e5950b354fe98b2c41ad8b6688da96f19e972d92`** (document-level pin; branch `WT-d2-moai-worktrees`). A cell is **Observed** where a plain read-only command in this tree produced the quoted outcome; the exit code of a no-match `grep` was not echoed by the harness (a no-match `grep` exits 1) and is re-recorded explicitly at M0. A cell is **To record at M0 (needs a Go test)** only where the red can be observed solely by running a new Go test that does not exist yet (plan time ran no Go command). 
>
> **Classification.** `RB` = release-blocking: the RED-now cell is Observed here, or becomes Observed at M0 before the first production change; if a cell cannot be re-executed on the current tree the criterion is demoted to `RG` (regression-guard) and is not recorded as a pass (verification-completeness §2.1). `RB*` = release-blocking once its M0 cell is recorded; until then unadopted.
>
> **Test layout.** Unless a criterion names another form, "test" means a Go test in the touched package using `t.TempDir()` repositories and a temporary `MOAI_HOME`; no test touches the real `~/.moai`, the real repository's worktrees, or global git config. Selectors are listed count-first (a `-list` count precedes each `-run`) so an absent test cannot pass as an empty sweep.

## §A. Coverage matrix

| REQ | AC | Kind | Severity | Class |
|-----|----|------|----------|-------|
| REQ-WHR-001 | AC-WHR-001 | Go test | MUST | RB |
| REQ-WHR-002 | AC-WHR-002a, AC-WHR-002b | Go test | MUST | RB* |
| REQ-WHR-003 | AC-WHR-003a, AC-WHR-003b | Go test | MUST | RB* |
| REQ-WHR-004 | AC-WHR-004 | Go test | SHOULD | RB |
| REQ-WHR-005 | AC-WHR-005 | Go test + grep | MUST | RB |
| REQ-WHR-006 | AC-WHR-006 | Go test + measurement | MUST (both branches) | RB |
| REQ-WHR-007 | AC-WHR-007 | Go test + grep sweep | MUST | RB |
| REQ-WHR-008 | AC-WHR-008 | Go test | MUST | RB* |
| REQ-WHR-009 | AC-WHR-009 | Go test | MUST | RB* |
| REQ-WHR-010 | AC-WHR-010 | Go test + CLI | MUST | RB |
| REQ-WHR-011 | AC-WHR-011 | Go test | MUST | RB* |
| REQ-WHR-012 | AC-WHR-012 | Go test | MUST | RB* |
| REQ-WHR-013 | AC-WHR-013 | Go test | MUST | RB |
| REQ-WHR-014 | AC-WHR-014a, AC-WHR-014b | Go test + static check | MUST | RB |
| REQ-WHR-015 | AC-WHR-015 | Go test | MUST | RB |
| REQ-WHR-016 | AC-WHR-016 | Go test | MUST | RB* |
| REQ-WHR-017 | AC-WHR-017 | Go test (rehearsal, fault injection) | MUST | RB* |
| REQ-WHR-018 | AC-WHR-018 | Go test | MUST | RB* |
| REQ-WHR-019 | AC-WHR-019 | Go test | MUST | RB |
| REQ-WHR-020 | AC-WHR-020 | Go test + grep | MUST | RB* |
| REQ-WHR-021 | AC-WHR-021 | Record check (plain greps) | MUST (gate) | RB |
| REQ-WHR-022 | AC-WHR-022 | Go test | MUST | RB* |
| REQ-WHR-023 | AC-WHR-023 | Go test | MUST | RB |
| REQ-WHR-024 | AC-WHR-024 | Go test + grep | MUST | RB |
| REQ-WHR-025 | AC-WHR-025 | grep + byte count | MUST | RB |

Cross-cutting completion criteria stated on card t1640 map as: "new trees are never created outside the root" → AC-WHR-005, AC-WHR-006; "entry measured on both harnesses" → AC-WHR-021.

## §B. Acceptance criteria

### AC-WHR-001 — Home-root path and `MOAI_HOME` relocation
**Given** a temporary `MOAI_HOME=/tmp/x/moai-home` and a temporary repository `proj` with no assignment. **When** the resolver is asked for the path of tree `t9`. **Then** it returns `/tmp/x/moai-home/worktrees/projects/proj/t9`, and a second call with `MOAI_HOME` unset resolves under `$HOME/.moai/worktrees/projects/proj/t9`. (The resolver's first call records the assignment of AC-WHR-002a; the path computation itself creates no tree.)
RED-now: Observed. Command `grep -n 'sessionWorktreeSubdir = ' internal/cli/session_worktree.go`; stdout `48:const sessionWorktreeSubdir = ".moai" + string(filepath.Separator) + "worktrees"`; the destination is built from it at line 223; no resolver exists, so the home path cannot be returned. Green path: M1.

### AC-WHR-002a — Key derivation and owner marker
**Given** a primary checkout at a path containing characters outside `[A-Za-z0-9._-]` (for example `my proj (v2)`), invoked from a linked worktree of that repository. **When** the resolver assigns the key. **Then** the key is the sanitized primary-checkout basename (`my-proj--v2-`), derived from the primary root and not from the invoking worktree, and an owner marker `.moai-project.json` is written at `<MoaiHome>/worktrees/projects/<key>/` holding the schema version, an owner identity, the primary root last seen and the creation time; the same assignment is recorded in the repository-side record.
RED-now: Observed. Command `grep -rn 'moai-project' internal`; stdout lists only unrelated `moai-project` skill names and an unrelated v2-detection comment (`internal/cli/update.go:479`, `internal/cli/v2_detection.go:315`, `internal/template/...`) — no marker code. Green path: M1.

### AC-WHR-002b — Stability: a key never changes once assigned
**Given** repository `/a/proj` with a recorded key `proj`. **When** (1) the resolver runs again with no change; (2) the checkout directory is moved to `/z/other-name` (the repository-side record moves with it) and resolved there; (3) a clone of a different repository named `proj` arrives later; (4) migration is run twice. **Then** every resolution of `/a/proj`'s repository returns `proj` (including from `/z/other-name`), the arrival in (3) receives the suffixed key and `proj` stays with the first repository, and the second migration run changes no assignment and moves nothing.
RED-now: To record at M0 (needs a Go test) — no assignment record exists (see AC-WHR-002a). Green path: M1, M6.

### AC-WHR-003a — Case-insensitive collision between two repositories
**Given** repository `/a/Proj` owning key `Proj` (owner marker present) and a later repository `/b/proj`. **When** the resolver assigns `/b/proj`. **Then** the folder-name test is case-insensitive (`Proj` equals `proj`), the later repository receives `proj-<8 hex of SHA-256 of the case-folded canonical root>`, `Proj` is untouched, no tree of `/b/proj` is placed under `Proj`, and on a case-sensitive fixture directory the same outcome holds (the name test is case-insensitive on every volume).
RED-now: Observed. Command `grep -rln 'worktrees/projects' internal`; stdout empty — no collision logic or home layout exists in code. Green path: M1.

### AC-WHR-003b — Third-owner and unreadable-marker refusal
**Given** the suffixed folder from AC-WHR-003a already owned by a third repository, and separately a corrupt or unparseable marker under the candidate key. **When** the resolver runs for the arriving repository. **Then** creation refuses with a message that contains the marker path, and no tree is created.
RED-now: To record at M0 (needs a Go test). Green path: M1.

### AC-WHR-004 — Operator key pin
**Given** `git_strategy.worktree_project_key: "fixedkey"` and a repository with no recorded assignment. **When** the resolver assigns the key. **Then** the key is `fixedkey`; with a differing pin set later on an assigned repository the recorded key is kept and a mismatch notice is printed; a pin naming a folder owned by another repository makes creation refuse (no suffix); a value containing a separator or `..` is rejected; with the key empty the result equals AC-WHR-002a's derivation.
RED-now: Observed. Command `grep -rn 'worktree_project_key' internal .moai/config`; stdout empty (key absent). Green path: M1.

### AC-WHR-005 — No creation path names a legacy root
**Given** the run-phase tree after M2. **When** `moai worktree new <n>`, the factory card materializer, the session-worktree auto-entry and the launcher `-w <n> --branch <b>` path each create a tree in a temporary `MOAI_HOME` fixture. **Then** each tree is registered under `<MOAI_HOME>/worktrees/projects/<key>/<n>`, has the `WT-` branch convention, the configured base branch (existing `SPEC-WORKTREE-BASEREF-001` tests stay green), the global `safe.directory` registration (via the existing seam), and a seeded `.codex/hooks.json`; and a static scan finds zero creation call sites constructing `".claude","worktrees"` or `".moai","worktrees"` destinations in production code, except the REQ-WHR-006 branch-Y pass-through (count-first: the scan's expected-site list is printed before its result).
RED-now: Observed. Command `grep -n 'filepath.Join(projectRoot, ".claude", "worktrees", name)' internal/cli/worktree_branch_flag.go`; stdout `109:	treePath := filepath.Join(projectRoot, ".claude", "worktrees", name)`; plus AC-WHR-001's `session_worktree.go:48`. Green path: M2.

### AC-WHR-006 — Launcher short name: both PM-3 branches
**Given** the pre-kickoff record's PM-3 result and a stub backend that records its argv. **When** `moai cc -w newname` (and `moai glm -w newname`) runs with no tree of that name in any root. **Then** — **branch X (PM-3 PASS):** the shared materializer ran, the backend received the absolute home-root path, and no directory `.claude/worktrees/newname` exists afterwards; **branch Y (PM-3 FAIL):** the backend received the short name `newname` unchanged and the launcher created nothing; **both branches:** with a tree already present at the home root or a legacy root the launcher enters that tree and creates nothing. The test selects its branch from the recorded PM-3 line, and fails when that line is absent or GAP.
RED-now: Observed. Command `grep -n 'claude then resolves the name' internal/cli/launcher.go`; stdout `913://     claude then resolves the name against .claude/worktrees/<name>. This` — today the short name is passed through for Claude Code to create under `.claude/worktrees` (branch X would flip this; branch Y keeps it and adds the existing-tree resolution). Green path: M2.

### AC-WHR-007 — Readers resolve through the resolver; sweep finds no stray derivation
**Given** a fixture repository with one tree at the home root (card `t7`), one at `.moai/worktrees/t7b`, one at `.claude/worktrees/t7c`, and recorded `worktree_path` for `t7`. **When** each reader named in REQ-WHR-007 resolves the tree: `cardIDFromRegisteredWorktree`, `cardIDFromPath`, `autoOwnerWorktrees`/`ownerAlive`, `DefaultWorktreeBases`+`ReadCardStatus`, `factoryCardWorktreeDir`, `resolveLaunchWorktreePath`, `resolveWorktreeLaunchTarget`, `seedAdmittedWorktreeHooksFor`, `ccWorktreeWriterPrecheck`, `resolveCodexWorktreeDir`, `cleanupMoaiWorktrees`; **and** a replacement run of card `t7` (its predecessor recorded the home-root tree) and a foreign directory at the card's landing path. **Then** each reader finds the tree at its actual location, the recorded path wins over a derived candidate, the replacement run reuses the home-root tree (previous-worktree carry-over), the foreign-tree refusal fires before any lease edge and leaves the card row unchanged, and a grep sweep over production Go for the literal pairs `".claude", "worktrees"`, `".moai", "worktrees"` and `Base(parent) == "worktrees"` returns only the sites on a reviewed allow-list (resolver internals, legacy-candidate list), its count printed first.
RED-now: Observed. Command `grep -n 'func factoryCardWorktreeDir' internal/cli/factory_card.go`; stdout `499:func factoryCardWorktreeDir(root, cardID string) string {` and the next lines return `filepath.Join(root, sessionWorktreeSubdir, cardID)` (a home-root tree is invisible to the refusal and the carry-over); the other readers: To record at M0 (needs a Go test); inventory in research.md §R3. Green path: M3.

### AC-WHR-008 — No misidentification at the new depth
**Given** a session whose cwd is `<MOAI_HOME>/worktrees/projects/proj/t7/internal/cli`. **When** SessionStart derives the card id (both derivations) and `todo --auto` judges the owner of card `t7` whose recorded tree is that path with a live registry entry. **Then** the card id is `t7` (not `projects`, not `proj`), and the owner is judged alive.
RED-now: To record at M0 (needs a Go test) — expected red because `cardIDFromPath` returns the first ancestor whose parent is named `worktrees` (`session_start_record.go:226-241`), which for the new layout is the directory `projects`, and `autoOwnerWorktrees` lists only the two in-repo candidates (`todo_auto.go:88-93`). Green path: M3.

### AC-WHR-009 — Tier by creation class, including a non-`WT-` tree
**Given** a home-root moai-made tree with a landed branch and a clean state, a home-root moai-made tree whose branch is not `WT-` (a migrated `release-v314`-shaped fixture with branch `release/v3.1.4`), and an L2 registry-style tree. **When** `moai worktree done <tree>` runs on each home-root tree with and without `--force` in a fresh process, and `moai worktree sweep --json` runs. **Then** `done` refuses both home-root trees both times with the L1 refusal text, `sweep` reports their tier as `L1` and applies only its existing landing predicates, and the legacy-rooted fixtures behave exactly as before (existing `done_l1_tier_guard_test.go` stays green).
RED-now: To record at M0 (needs a Go test) — expected red because `isL1WorktreePath` answers by prefix (`done.go:230-247`) and a home-root tree is outside both prefixes; `sweepTierOf` (`sweep.go:574-580`) delegates to it. Green path: M3.

### AC-WHR-010 — `moai worktree list`
**Given** a fixture repository with one tree of each class (primary, home, legacy-moai WT-, legacy-moai non-WT, legacy-claude moai-made, `agent-*`, integration, Codex-managed path) and one tree registered under two spellings of the same directory. **When** `moai worktree list --json` and `moai worktree list` run. **Then** exit code is 0, every tree appears exactly once (the two spellings collapse to one row), each row has name, path, root kind, class, branch, short HEAD, dirty flag, unpushed-commit count and anchored flag, and the run mutates nothing (tree hash before equals after).
RED-now: Observed. Command `grep -n 'AddCommand' internal/cli/worktree/root.go`; stdout `41:	WorktreeCmd.AddCommand(` — registration block at lines 41-54 lists no `list` verb; at the audited SHA `moai worktree list` printed the parent help text (plan-audit iteration 1 reproduction). Green path: M4.

### AC-WHR-011 — Card record rebind
**Given** a factory run with card `t7` recording worktree path `P_old`, a replacement run row for `t7` also carrying `P_old`, and a moved tree at `P_new`. **When** the rebind operation runs with `(card=t7, old=P_old, new=P_new)`. **Then** every run row of `t7` carrying `P_old` now carries `P_new`, one `card.fields` event records both paths, `PreviousCardWorktree("t7", otherRun)` returns `P_new`, a rebind whose `old` does not equal the current value is refused, a rebind to an unregistered path or a different HEAD/branch is refused, and `RecordCardWorktree` still refuses to move a card onto a different tree (existing behavior test unchanged).
RED-now: To record at M0 (needs a Go test) — no rebind operation exists; `RecordCardWorktree` refuses any path change (`card_worktree.go:42-44`). Green path: M5.

### AC-WHR-012 — Other path-bearing state
**Given** `.moai/state/worktrees.json` holding an entry for `P_old`, a session registry entry whose cwd is `P_old` with a dead pid, and a chain-ledger node naming `P_old`. **When** the migration moves the tree. **Then** the state-file entry is keyed to `P_new`, the session registry and ledger entries are unmodified byte for byte, and the journal lists each store as touched or deliberately untouched, with a done status per step.
RED-now: To record at M0 (needs a Go test); the state file is keyed by `Path` (`internal/hook/worktree_registry.go:29-65`) and nothing re-keys it. Green path: M5, M6.

### AC-WHR-013 — Classifier, including non-`WT-` trees after the move
**Given** the measured shapes: home tree; legacy-moai `WT-` tree; legacy-moai tree with a non-`WT-` branch (`rc30-preview`, `release-v314` branch `release/v3.1.4`, `t1515-absorb`, `t1868-publish` are such trees in this repository); legacy-claude moai-made tree (`t1497`); `agent-<hex>` tree; integration worktree holding the integration branch; a tree whose branch is not `WT-` and whose directory is not under `.moai/worktrees` (`audit-gate-async-worktree-doc`); **and a home-root fixture whose branch is not `WT-`, whose owner marker names this repository (a `release-v314` tree after migration), and a home-root fixture whose owner marker names a different repository**. **When** the classifier runs in a fresh process. **Then** the results are, respectively: moai-made, moai-made, moai-made (directory witness), moai-made (branch witness), claude-native, integration, unclassified, moai-made (marker witness), unclassified (foreign marker); unclassified is listed and counted and never moved; and `moai worktree list`, `moai worktree migrate`, the tier decision of AC-WHR-009 and the `moai update` consolidation return the same classification for the same input (one shared function).
RED-now: Observed. Command `grep -n 'agent-' internal/cli/update_worktree_migration.go`; stdout empty — the existing consolidation selects by path prefix only (`update_worktree_migration.go:87,105-108`) and therefore includes `agent-*` trees; and `isCardBranch` (`done.go:315`) is `WT-`-prefix only. Green path: M5, M6.

### AC-WHR-014a — Non-destructive migration (static and behavioral)
**Given** the migration engine source (the files of `internal/cli/worktree/migrate*.go`; the retirement file is outside this scan and removes only an empty directory non-recursively). **When** a static check scans it for the forbidden operations (`worktree remove`, `worktree prune`, `branch -D`, `reset`, `clean`, `checkout`, `stash`, `os.Remove`, `os.RemoveAll`, `rm`), and a behavioral test runs `--apply` against the rehearsal repository. **Then** the static check finds none (its swept set non-empty — the file list is printed and counted), the only git subcommand the engine runs that mutates a tree is `worktree move`, and after the rehearsal every branch tip that existed before still exists with the same commit id and `git worktree list` registers the same number of trees as before.
RED-now: To record at M0 (needs a Go test; the engine is absent: `grep -rln 'worktree migrate' internal/cli` printed nothing). The static check must also be shown red on a deliberately planted `git worktree remove` line before it is trusted (verification-completeness §1.1). Green path: M6.

### AC-WHR-014b — Launcher auto-cleanup never removes an unlanded or migrated tree
**Given** a registered tree `worker-review` under the home root (a migrated tree) with one commit present on no remote and a clean working state, a second `worker-old` tree whose branch is fully on a remote and which the migration did not move, and a `worker-*` tree under a legacy root. **When** `cleanupMoaiWorktrees` runs (the launch-time cleanup of `moai cc`) after a migration. **Then** `worker-review` is not removed and is reported as skipped with the reason, and `git branch --list` still shows its branch; the legacy-root `worker-*` tree with no unlanded commits keeps today's removal behavior; the unlanded-commit guard also protects a `worker-*` tree under a legacy root.
RED-now: Observed. Command `grep -n 'removeWorktree(projectRoot, worktreePath)' internal/cli/launcher.go`; stdout `518:			if err := removeWorktree(projectRoot, worktreePath); err == nil {` — the removal is reached for any `worker-*` tree whose `filepath.Rel` to a base does not start with `..` (lines 505-518), with no unlanded-commit check; the nested-path acceptance is the audit's iteration-1 finding D5. Green path: M3, M6 (post-migration rehearsal pass).

### AC-WHR-015 — Dry-run is read-only; apply is idempotent; verify exists
**Given** the rehearsal repository (committed-unpushed tree, dirty tree with untracked and ignored files, locked tree, occupied-cwd tree, `agent-*` tree). **When** `moai worktree migrate` runs with no flag. **Then** exit code is 0, the listed plan shows each tree's class, source, destination and skip reason, no path under the repository or `MOAI_HOME` is created, modified or removed (before/after content hash equal), and `--apply` is required to move anything; **and** after `--apply`, `--verify` reports each recorded move as `intact`, and a second `--apply` moves nothing and changes no assignment.
RED-now: To record at M0 (needs a Go test; no verb: `grep -rln 'worktree migrate' internal/cli` printed nothing). Green path: M6.

### AC-WHR-016 — Skip conditions
**Given** the rehearsal fixtures for each skip cause: locked, prunable, process cwd inside the tree, live anchored session, source missing, destination exists, destination on another filesystem (exercised through an injectable device-id seam, so no second volume is needed), process-cwd inspection unavailable. **When** `--apply` runs. **Then** every such tree stays at its old path, remains a registered worktree with unchanged HEAD and working state, and the output names the cause for each; when process cwd inspection is unavailable no tree moves at all.
RED-now: To record at M0 (needs a Go test). Green path: M6.

### AC-WHR-017 — Manifest with content digest, ordered steps, automatic reversal at every step
**Given** a rehearsal tree on branch `WT-unpushed` with two commits present on no remote, a modified tracked file, an untracked file, an ignored file, a populated `.moai/reports`, a card record and a `.moai/state/worktrees.json` entry. **When** `--apply` moves it. **Then** the post-move manifest (HEAD, branch, unpushed commit list, status hash, ignored-file count, reports file count, content digest) equals the pre-move manifest field for field; `git -C <new> log --not --remotes` lists the same two commit ids; the contents of the modified, untracked, ignored and reports files are byte-identical; a content mutant (one byte changed in an ignored report file, no other field changed) is detected by the digest and the move is reversed; **and** with a fault injected separately after the move, after the state re-key and after the card rebind, and with the process killed between steps, the completed steps are reversed (or the re-run converges), after which the registered worktree path, the `.moai/state/worktrees.json` entry and the card record all name the same real location; a forced reverse-step failure leaves that step pending in the journal, makes the exit code non-zero and names the tree.
RED-now: To record at M0 (needs a Go test). Green path: M6. This is the operator's "nothing lost" proof for the unpushed-branch invariant, within the hashed scope stated in the manifest (I-2).

### AC-WHR-018 — Explicit rollback and convergence
**Given** a journal from an applied migration of three trees, one of whose source paths has since been re-occupied, and a fourth journal entry whose reverse card rebind is recorded pending after a reverse move. **When** `moai worktree migrate --rollback <journal>` runs, then runs again. **Then** the two free trees return to their sources with matching manifests, restored state-file entries and rebound card records; the re-occupied one is refused with the reason and the exit code is non-zero; the pending rebind is completed by the run and the second run changes nothing (idempotent); afterwards the registered path, state-file entry and card record agree for every restored tree.
RED-now: To record at M0 (needs a Go test). Green path: M6.

### AC-WHR-019 — `moai update` consolidation: classifier, destination, preview only
**Given** a primary checkout with legacy-rooted moai-made trees (including a non-`WT-` one) and `agent-*` trees. **When** the update consolidation runs through both call paths (the dry-run path and the non-dry-run path of `moai update`). **Then** no `agent-*` tree appears in its plan, every planned destination is under `<MoaiHome>/worktrees/projects/<key>/` and none under `<primary>/.moai/worktrees`, the output prints the dry-run summary plus the exact `moai worktree migrate --apply` command, and **nothing is moved on either path** (before/after `git worktree list` byte-identical). Its idempotence and primary-checkout-only behavior (a linked-worktree update leaves siblings alone) are unchanged.
RED-now: Observed. Command `grep -n 'runUpdateWorktreeMigration' internal/cli/update.go`; stdout includes `366:...runUpdateWorktreeMigration(lockRoot, true, out)` and `410:...runUpdateWorktreeMigration(lockRoot, false, out)` — the second call applies today; and `grep -n 'agent-' internal/cli/update_worktree_migration.go` printed nothing (agent trees are selected). This is the recorded behavior change of decision D-6. Green path: M6.

### AC-WHR-020 — Retirement
**Given** a fixture in which the last tree under the MoAI legacy root has been migrated and the directory is empty, then a second fixture where the directory still contains an untracked file. **When** the retirement step runs. **Then** the empty directory is removed with a non-recursive removal and the non-empty one is left in place with a notice; `.claude/worktrees` and every claude-native tree in it are untouched in both fixtures; and a repository-wide scan of production creation code finds no legacy-root destination other than the REQ-WHR-006 branch-Y pass-through (this scan shares AC-WHR-005's mechanism).
RED-now: To record at M0 (needs a Go test; no retirement code exists). Green path: M7.

### AC-WHR-021 — Pre-kickoff measurement gate cannot pass on Gaps
**Given** three fixture records and the plain gate check of plan §B (`grep -c '^RESULT PM-\(1a\|1b\|1c\|1d\|2\|3\): \(PASS\|FAIL\)$'` and `grep -c '^RESULT PM-\(1a\|1b\|1c\|1d\|2\|3\): GAP$'`): (i) a record in which all six gating lines are `GAP`; (ii) a record with five PASS/FAIL lines and one `GAP` (for example PM-1c); (iii) a record with six PASS/FAIL lines, each carrying command, verbatim output, exit status, tool versions and tree SHA, in which each FAIL line states the consequence of spec §F. **When** the gate check is applied. **Then** (i) and (ii) fail the gate (the PASS/FAIL count is not 6 and the GAP count is not 0), (iii) satisfies it, an absent file fails it, a refusal in PM-1b is classified on exactly one line (FAIL), and a Kickoff decision record cites the record path and its sha256. The measurement itself is taken before the Kickoff and outside the run-phase milestones.
RED-now: Observed. Command `ls .moai/reports/t1640/pre-kickoff-measurements.md`; stdout `ls: .moai/reports/t1640/pre-kickoff-measurements.md: No such file or directory`; exit code 1 — the record is absent, so the gate (which treats absence as unsatisfied) is red now. Green path: M-1.

### AC-WHR-022 — Transition resolver and ambiguity across all three roots
**Given** a name `t7` present both at the home root and at `.moai/worktrees/t7` (two different trees), a name `t6` present at both `.moai/worktrees/t6` and `.claude/worktrees/t6` (the GH #1813 shape settled by card t1791), a name `t5` present at the home root under two spellings of one directory, and a name `t8` at only one root. **When** a launcher short-name lookup and `resolveCodexWorktreeDir` run for each. **Then** `t8` and `t5` resolve to their single tree, `t7` and `t6` are reported ambiguous with a message instructing an absolute path and naming both candidates, no root is preferred, and neither lookup guesses.
RED-now: To record at M0 (needs a Go test) — today the Codex resolver reports ambiguity for its two in-repo roots only (`internal/cli/codex_launcher.go:474-519`) and does not know the home root. Green path: M3.

### AC-WHR-023 — Spelling-safe equality
**Given** a case-insensitive volume where `/x/MoAI` and `/x/moai` are one directory (the machine under test reports one inode for `/Users/goos/MoAI` and `/Users/goos/moai`), and a tree registered under the lowercase spelling. **When** equality and containment tests run (`isL1WorktreePath`, resolver `Same`, the classifier, the migration skip checks, `moai worktree list`). **Then** the two spellings are equal and the tree is reported once; and on a case-sensitive fixture directory two directories differing only by case remain distinct. The project-folder name test of AC-WHR-003a remains case-insensitive on that case-sensitive fixture (name-level rule).
RED-now: Observed. Command `stat -f '%i' /Users/goos/MoAI /Users/goos/moai`; stdout `105929094` and `105929094` (one directory, two spellings); and 25 of the 192 MoAI-root trees are registered under the lowercase spelling in `git worktree list` (plan-audit iteration 1 measurement, date-sensitive). Green path: M1 (helper), M3-M6 (use).

### AC-WHR-024 — Claude-owned and Codex-owned state untouched; no WorktreeCreate hook
**Given** the rehearsal repository containing `agent-*` trees under `.claude/worktrees`, a tree under a fake `~/.codex/worktrees`, and the integration worktree. **When** `list`, `migrate --apply`, `--rollback`, retirement and `update` all run. **Then** none of those trees is moved, renamed, pruned or removed, their `git worktree list` records are byte-identical before and after, and `grep -n 'WorktreeCreate' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` finds no registered hook entry.
RED-now: Observed (second clause). Command `grep -n 'WorktreeCreate' .claude/settings.json`; stdout empty — no registration today, so the clause is a regression guard (`RG`) for that part; the first clause fails today because the old consolidation selects `agent-*` (AC-WHR-013, same command). Green path: M6.

### AC-WHR-025 — Location guidance only; budget; no t1879 criterion
**Given** the run phase after M7. **When** a count-first grep over the in-scope surfaces (AGENTS.md, the two worktree rule files, the skills and agents that carry a worktree location statement, their template mirrors, the cc/glm/worktree help strings) is run for the home-root location, and `wc -c AGENTS.md` is run. **Then** (a) the home-root location appears in each in-scope surface that names a worktree location, (b) every entry-behavior statement added by this card is traceable to a `RESULT` line of the measurement record, (c) `wc -c AGENTS.md` prints at most 23780, (d) the template-mirror consistency test passes, and (e) this criterion neither requires nor checks the count of sentences that instruct a session to run `moai cc -w` from inside Claude Code (that count belongs to card t1879).
RED-now: Observed. Command `wc -c AGENTS.md`; stdout `23780 AGENTS.md` (budget baseline 23780; the home-root location is absent from the guidance surfaces today, which the M0 count-first grep records). Green path: M7.

## §C. Edge cases (each is covered by an AC above or recorded as a Gap)

- Repository with zero legacy trees: `migrate` prints "0 moai-made trees" and exits 0; `update` prints nothing new.
- Tree whose directory name contains a space or non-ASCII: the sanitizer touches only the project key; tree names are validated as today (`validateNewWorktreeName`, `new.go:56-63`).
- `MOAI_HOME` relative or unset: `paths.MoaiHome()` already disregards a relative value and falls back to `$HOME/.moai` (`paths.go:67-76`).
- A tree being created while a migration runs: creation and migration serialize on the existing factory step lock for card trees (`factory.AcquireFactoryStepLock`, `factory_card.go:762`); non-card creation during `--apply` is recorded as a Gap, not assumed safe.
- Interrupted `--apply` (process killed between steps): the journal entry is written before the move with a status per step, so `--verify` classifies the tree and `--rollback` or a re-run converges (AC-WHR-017/018).
- Windows paths: unmeasured (OQ-4).

## §D. Quality gates and Definition of Done

- TRUST 5: scoped tests green for every touched package with coverage at or above the repository threshold for the new packages; lint clean on touched packages; cross-platform build of touched packages; no `AskUserQuestion` in subagent-domain code; conventional commits carrying `card t1640`.
- Every `RB` criterion has its RED-now cell recorded at M0 and its green output recorded at the milestone named, with command, verbatim output, exit code and SHA in `progress.md §E.2` (written by the run-phase specialist, not by this plan); every `RB*` criterion gains its RED cell at M0 before the first production change or is demoted to `RG`.
- The pre-kickoff measurement record exists, its gate check passes (six gating lines, no Gap), and every FAIL in it has been resolved by the stated consequence.
- The operator-held decisions are recorded as confirmed or default-applied.
- Sync phase evidence includes the real repository's dry-run listing (counts per class) and the leader-operated apply command; no live tree was moved by the lane.
- Plan audit by codex PASS recorded at `.moai/reports/t1640/` before the run phase (Class C requirement).
