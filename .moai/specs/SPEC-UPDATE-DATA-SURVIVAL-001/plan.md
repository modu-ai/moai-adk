# SPEC-UPDATE-DATA-SURVIVAL-001 — Implementation Plan

> Ordered by decision-reversibility. §B (the failure contract) and §C (the destructive-target
> registry) are the decisions most likely to change under review and are stated first. The
> mechanical work — guard rewrites, error-branch coverage — sits at the bottom.
>
> **Version header: intentionally none.** This plan carries no `Version:` line; `spec.md`'s
> frontmatter `version:` and its HISTORY table are the single version record for all four
> artifacts. (Closes D20's second half — the disposition is "intentionally none", not "missing".)

## §A Context

- **Repository**: `/Users/goos/MoAI/moai-adk-go`, worktree
  `.claude/worktrees/e2-data-survival`, branch `feat/SPEC-UPDATE-DATA-SURVIVAL-001`, HEAD
  `89b2e4772`, whose **code baseline is `8cc108ddb`** (verified an ancestor of HEAD; re-verified
  on HEAD `184a5bd222`). Only this SPEC's own artifacts differ from that baseline —
  `git diff --name-only origin/main HEAD | grep -c '\.go$'` returns `0` — so every code baseline
  recorded in `acceptance.md` reproduces at `8cc108ddb`.
  **`8cc108ddb` is no longer `origin/main`.** It was when recorded; `origin/main` has since
  advanced to `9ced435e9` (PR #1266), which entered this branch via merge `2255165f5`. The
  `= origin/main` parenthetical is therefore removed: `8cc108ddb` remains a valid **measurement
  anchor** (still an ancestor, so the figures still reproduce) but is NOT a comparison base. See
  `acceptance.md` §A.4 for the anchor-vs-base distinction and AC-UDS-019 for the merge-base-anchored
  template-neutrality check that replaced the fixed-SHA pin.
- **Re-baselined in v0.4.0.** v0.3.0 anchored to HEAD `a8b42e112` / code baseline `d5336214e` and
  claimed "no Go source differs". Both are wrong on this tree: `a8b42e112` is **not** an ancestor of
  HEAD, and `git diff --name-only d5336214e..HEAD | grep -v '^\.moai/' | wc -l` returns `19`. E1
  (`SPEC-UPDATE-REINSTALL-LOOP-002`) landed its run (`beeb0ebc2`, PR #1261) and sync (`8cc108ddb`,
  PR #1264) between the two rounds, adding a destructive call site and shifting coordinates across
  five files. Full detail and the re-measured figures: `acceptance.md` §A.4.
- **Tier**: M — 300-1000 LOC across 5-15 files. Justification in §H.
- **Depends on**: `SPEC-UPDATE-REINSTALL-LOOP-002` (E1). Only REQ-RIL2-015/016
  (backup-before-delete for `defs.DeprecatedPaths`) is inherited; every milestone below is
  otherwise independent of E1's merge and can proceed in parallel.
- **Development mode**: TDD (`quality.yaml` `constitution.development_mode`). Each milestone writes
  its guard first, observes it FAIL, then implements.

### A.1 PRESERVE list (do not modify)

- `internal/template/templates/**` — 16-language distribution surface.
- The `defs.DeprecatedPaths` slice body and `internal/defs/dirs_test.go` count assertions. Only the
  Category D **comment prose** changes (M2).
- `internal/cli/update/backup/merge.go` and the YAML merge semantics — owned by
  `SPEC-UPDATE-YAML-PRESERVE-001`.
- `probeVersionSignal`, `detectV2Fingerprint`, `preserveInventoryRoots`, and the `--dry-run` branch
  — owned by E1.
- The success-path outcome of an update (NFR-UDS-005): a run that succeeds today must produce the
  same tree afterwards.

## §B The F5 decision — report and assisted restore, not automatic rollback

**Decision: report + an explicitly-invoked restore command. No automatic rollback.**

### B.1 Rationale

1. **An automatic rollback runs the least-tested code in the worst state.** By the time rollback
   would fire, the tree is in an unknown partial state — some templates deployed, some paths
   removed, `.moai/config` possibly absent. A rollback is itself a multi-step destructive
   filesystem operation (delete the partial deploy, restore N backup trees, re-merge configs). A
   rollback that fails half-way leaves the tree strictly worse than the failure it was recovering
   from. This is the failure shape `runtime-recovery-doctrine.md` §3 invariant 3 names: the
   recovery action triggering the very class of error it exists to recover from. The doctrine's
   prescribed response is the last-resort escape (abort + preserve), not a retry loop.

2. **The backups already exist; only the pointer and the applicator are missing.** Once M3 lands,
   every destroyed file has a disk artifact. What a stranded user lacks is (a) which backup
   directory belongs to the failed run and (b) a command that applies it. Those are cheap, and
   both are outside the failure path.

3. **A user-invoked restore is inspectable and re-runnable.** The user can list the backup
   directory before applying it, apply it twice safely (REQ-UDS-023), and abort. An automatic
   rollback offers none of that and fires exactly when the operator is least able to supervise it.

4. **Cheapest-first (`runtime-recovery-doctrine.md` §2).** Reporting is rung-4 "abort + preserve":
   persist state, name it, stop. Automatic rollback is a heavier rung attempted before the cheaper
   one has been shown insufficient.

### B.2 What "report" concretely means

On any error after the first destructive step, on either path:

- Write a **recovery manifest** into the run-scoped backup directory recording: the failed step
  name, the error, the backup directory's absolute path, and the exact restore command.
- Print the manifest to the user's terminal, not only to a file.
- Return the original error wrapped, so exit status is unchanged.

### B.3 The lockout escape

The restore entry point must run on a tree whose `system.yaml` is gone — that absence is the damage
it repairs. So the project-marker gate — the `checkProjectMarker(cwd)` call at
`internal/cli/update.go:251`, whose implementation and error text M1 extracted into
`internal/cli/update_restore.go:21`/`:27` — is bypassed **for the restore
entry point only** (REQ-UDS-022/025). The ordinary update path keeps the gate unchanged, and an AC
asserts that a bare `moai update` on a marker-less tree still fails with the existing message.
(Coordinate note: this gate was recorded as `update.go:236` before M1; the M1 commit `4ddd35120`
both moved the call site and extracted the predicate into `update_restore.go`, so the old
single-coordinate citation no longer resolves to anything.)

Refusal criterion (REQ-UDS-024): the restore entry point accepts a directory only when it carries
the backup's own marker file. It never copies from an arbitrary directory.

### B.4 Rejected alternative

**Transactional rollback via a staged tree** — deploy into a shadow directory and swap on success.
Rejected: the swap is not atomic across the several roots the update touches (`.claude/`,
`.moai/`, repo-root files), cross-device renames are not guaranteed, and the change would rewrite
the entire deploy pipeline — far outside this SPEC's blast radius. Recorded so it is not silently
re-proposed.

## §C The destructive-target registry

The registry is the new shared artifact M2 introduces: a single in-repo table naming every
destructive filesystem operation in the update subsystem and its protection assignment. Below is
the enumeration derived from this tree; the implementation encodes it and guards it.

### C.0 Scan scope — what the guard enumerates, and why it is wider than §C.1's target list

REQ-UDS-007 requires the drift guard to enumerate destructive **call sites** by static source scan,
independently of the registry. That scan is mechanical, so its scope is fixed by grep semantics
rather than by judgement, and it finds materially more than the user-data targets §C.1 tabulates:

```
$ grep -rn 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | wc -l
      18
```

**18 call sites across 11 (file, function) pairs**, re-measured on HEAD `89b2e4772` and confirmed
unchanged on HEAD `2255165f5` (the M2 Step 0 re-scan — see the M1 coordinate-drift note below).
Two earlier
figures are superseded: a first draft implied 7 sites (omitting `update_archive.go`,
`update/backup/backup.go`, `update_cleanup.go`, `update_namespace_protect.go`), and v0.3.0 recorded
17 sites / 10 pairs against the retired `d5336214e` baseline, before E1 added
`internal/cli/update_residue_cleanup.go`. The registry must carry a row per scanned site, not per
user-data target, or the multiset comparison fails on first run.

| # | File | Function | Sites | Class |
|---|---|---|---|---|
| 1 | `update/deploy/deploy.go` | `CleanMoaiManagedPaths` | 3 (`:83`, `:105`, `:121`) | user data — §C.1 rows 1-10 |
| 2 | `update/deploy/deploy.go` | `MigrateLegacyMemoryDir` | 2 (`:169`, `:176`) | user data — §C.1 row 11 |
| 3 | `update_archive.go` | `archiveSkill` | 1 (`:101`) | user data — removes the archive destination before re-archiving |
| 4 | `update_archive.go` | `archiveLegacySkills` | 1 (`:322`) | user data — renames a skill dir into a backup location |
| 5 | `update/backup/backup.go` | `BackupMoaiConfig` | 3 (`:107`, `:135`, `:140`) | **exempt** — unwinds the backup dir this call just created, on its own error paths |
| 6 | `update/backup/backup.go` | `CleanupOldBackups` | 1 (`:259`) | **exempt** — retention pruning of moai-authored backup dirs |
| 7 | `update_clean_install.go` | `runCleanReinstall` | 1 (`:315`) | user data — §C.1 row 12 (E1-owned) |
| 8 | `update_cleanup.go` | `removeDeprecatedFile` | 1 (`:324`) | user data — deprecated-path removal |
| 9 | `update.go` | `ensureGlobalSettingsEnv` | 1 (`:853`) | user data, outside project root — §C.1 row 13 |
| 10 | `update_namespace_protect.go` | `backupUserOwnedNamespace` | 3 (`:225`, `:233`, `:243`) | **exempt** — defensive cleanup of the namespace backup dir this call created (`EC-UNP-007`) |
| 11 | `update_residue_cleanup.go` | `runV3ResidueCleanup` | 1 (`:135`) | user data — deprecated-path residue sweep; backup is E1's REQ-RIL2-019 (§C.1 row 12's cross-SPEC assignment) |

The three exempt rows (5, 6, 10) remove only directories the same run authored; no user data
predating the run is at risk. Each records that reason in the registry per REQ-UDS-007 rather than
carrying a protection-set assignment. Rows 3, 4, and 8 are destructive sites the original §C.1
enumeration missed; **row 11 did not exist when v0.3.0 was measured** — it arrived with E1's run PR
#1261. M2 assigns rows 3/4/8/11 protection or exemption during implementation, and §C.1 below is the
user-data target view that remains the design rationale.

**M2's first act is to re-run the scan above, not to trust this table.** The 17→18 drift between
v0.3.0 and v0.4.0 happened because a sibling Epic SPEC landed Go source between plan rounds; the
same can recur before run-phase entry. REQ-UDS-007's guard enumerates from source precisely so that
a stale table fails loudly rather than silently.

**M1 coordinate drift (self-inflicted, NOT a second external TOCTOU).** The M2 Step 0 re-scan was
executed on HEAD `2255165f5` and reported the divergence rather than silently reconciling it, per
the Step 0 obligation in §F. The result: **site total (18) and pair count (11) are unchanged, and
the (file, function) pair set is unchanged** — no site was added or removed. Only line coordinates
moved, in exactly two rows: row 7 `update_clean_install.go` `:307` → `:315`, and row 9 `update.go`
`:843` → `:853`. The cause is **this SPEC's own M1 commit `4ddd35120`** (recovery manifest +
restore entry point), which added code above both sites in `update_clean_install.go` and
`update.go`. It is *not* a foreign-SPEC TOCTOU like the v0.3.0→v0.4.0 event narrated above — a
reader must not mistake the two for a repeated external drift. The stale coordinates in this file
and in `acceptance.md` were corrected in place; the `§C.4` falsification requirement and every AC
verification command are untouched, because AC-UDS-005's guard keys on (file, enclosing function,
occurrence count) and never on line numbers.

### C.1 Destructive targets (user-data view)

| # | Target | Site | Protection today | Action |
|---|---|---|---|---|
| 1 | `.claude/settings.json` | `deploy.go:40-42` → `:105` | **in-memory only** (`update_template_sync.go:300`/`:397`) | M3 adds disk backup |
| 2 | `.moai/status_line.sh` | same mergeable set | **in-memory only** | M3 adds disk backup |
| 3 | `.gitignore` | `update_template_sync.go:298`/`:388` | **in-memory only** | M3 adds disk backup |
| 4 | `.claude/commands/moai` | `deploy.go:44-46` → `:105` | template-managed, regenerated | exempt (recorded) |
| 5 | `.claude/agents/moai` | `deploy.go:48-50` → `:105` | template-managed, regenerated | exempt (recorded) |
| 6 | `.claude/skills/moai*` (glob) | `deploy.go:52-55` → `:83` | template-managed | exempt, **with a noted hazard** — see §C.1a |
| 7 | `.claude/rules/moai` | `deploy.go:57-59` → `:105` | template-managed | exempt (recorded) |
| 8 | `.claude/output-styles/moai` | `deploy.go:61-63` → `:105` | template-managed | exempt (recorded) |
| 9 | `.claude/hooks/moai` | `deploy.go:65-67` → `:105` | template-managed | exempt (recorded) |
| 10 | `.moai/config` (wholesale) | `deploy.go:121` | `BackupMoaiConfig` (`backup.go:27`) | covered |
| 11 | `.moai/memory/` | `deploy.go:176` | **none** | M2 adds backup (REQ-UDS-008) |
| 12 | `defs.DeprecatedPaths` incl. `.moai/db` | `update_clean_install.go:315` **and** `update_residue_cleanup.go:135` | E1 REQ-RIL2-015/019 (both sites back up before deleting) | inherited from E1; registry records the cross-SPEC assignment on both sites |
| 13 | `~/.claude/hooks/moai` | `update.go:853` | **none**, outside project | M4 pins the radius; backup out of scope (outside project root) |

Rows 1-3, 11, 12, 13 are the complete set of destructive targets in **no** protection set. Of those,
row 12 is E1's; rows 1-3 are M3; row 11 is M2; row 13 is M4 (radius pinning rather than backup,
because a HOME-scoped backup would itself write outside the project).

### C.1a Noted hazard, not a deliverable

Row 6's glob `.claude/skills/moai*` also matches a user-authored `moai-my-notes` skill — precisely
the ambiguous case `collectUserOwnedFilesConservative`
(`update_namespace_protect.go:167`) was introduced to back up. The conservative namespace backup
therefore already covers it, so this SPEC records the interaction in the registry and does **not**
narrow the glob. Narrowing it is a behaviour change belonging to a namespace SPEC.

### C.2 Measured baseline for M6

`mergeBackPreserveInventory` statement coverage was re-measured on HEAD `89b2e4772` with
`go test -covermode=set -coverprofile=... ./internal/cli/` followed by `go tool cover -func`,
yielding `64.3%` — unchanged from v0.3.0. The figure is recorded verbatim in `acceptance.md`
**AC-UDS-016** as that AC's baseline; the three uncovered blocks are the failure returns at
`update_preserve_inventory.go:416` (stat), `:420` (`MkdirAll`), and `:424` (`copyFile`). The
function itself moved `:330` → `:400` when E1 landed; the coverage grep is symbol-anchored
(`grep mergeBackPreserveInventory`) and survived, but the prose coordinates did not and are
corrected here. (Two stale cross-references corrected earlier and retained: the milestone is M6, not
M7 — §F defines M1-M6 only — and the coverage baseline lives in AC-UDS-016, not AC-UDS-013, which
covers HOME-test isolation.)

## §D Constraints

- `t.TempDir()` only; no test writes to the operator's real `~/.claude`.
- HOME redirection goes through the injectable seam `userHomeDirFn` (`glm_tools.go:123`), not
  `t.Setenv("HOME", …)` — per `CLAUDE.local.md` §13, a process-wide HOME mutation pollutes parallel
  tests. **M4 must first route `ensureGlobalSettingsEnv` through that seam** (REQ-UDS-013): it
  currently calls the plain function `userHomeDir()` at `update.go:843`, which is not reassignable,
  so no redirection point exists today. An earlier draft of this constraint asserted the
  indirection already existed; it does not — see `spec.md` §A Defect 3.
  **Do not conflate `:843` with §C.0 row 9.** `update.go:843` is now the `userHomeDir()` call; the
  `os.RemoveAll` site that §C.0 row 9 registers moved to `update.go:853`. The two were distinct all
  along, but the pre-M1 tables recorded the `os.RemoveAll` site at `:843` and the `userHomeDir()`
  call at `:833`, so a reader comparing the old and new values could mistake them for the same
  line. Both moved by +10 in M1 commit `4ddd35120`.
- No `internal/template/templates/**` edit.
- Falsification uses a scratch `git worktree` regenerated from a pre-fix ref and driven with
  `go -C`, or `go test -overlay`. **`git stash` is prohibited** — it refuses untracked files without
  `-u`, and with `-u` it is repository-global and can swallow a concurrent session's work.
- Conventional Commits; `fmt.Errorf("…: %w", err)`; English comments and godoc.

## §E Self-Verification

Every milestone reports the Section E matrix from
`.claude/rules/moai/development/manager-develop-prompt-template.md`: the AC PASS/FAIL table with
verbatim command output, cross-platform build result, coverage, lint delta, and the RED failing
output captured before GREEN.

## §F Milestones

### M1 — Failure contract: recovery manifest + restore entry point

Implements §B. Highest reversibility — the shape of the manifest and the restore command's surface
are the decisions most likely to be revised under review, so they land first and are reviewed
before anything depends on them.

- Recovery-manifest writer invoked from both paths' error returns after the first destructive step.
- Restore entry point accepting a backup directory; marker-file validation (REQ-UDS-024);
  idempotent application (REQ-UDS-023).
- Project-marker gate bypass scoped to the restore entry point (REQ-UDS-022/025).
- **Fixture obligation (REQ-UDS-020's non-vacuity precondition).** The recovery-manifest test's
  fixture MUST plant at least one moai-managed path that `CleanMoaiManagedPaths` will actually
  remove, declared as a **literal `[]string` named `plantedMoaiManagedPaths` in the test source**.
  The test asserts each planted path exists before the destructive call and is gone after it, then
  asserts each is still gone when the outer call returns.

  This is a **plan obligation, not a test-author choice**. REQ-UDS-020 ("no automatic rollback") is
  a negative requirement whose only mechanical coverage is that final absence assertion — and on an
  empty fixture "all of ∅ are still absent" holds trivially, so adding the very rollback the
  requirement forbids could not move it. The requirement would revert to covered-on-paper, which is
  the defect the clause was added to close.

  The non-emptiness count MUST come from the planted literal, **not** from observing
  `CleanMoaiManagedPaths`' own return value or output. A count the test derives from the function it
  is checking compares the function against itself — `acceptance.md` §A.3 shape (a), a
  self-comparison that can never fail.
- Covers REQ-UDS-019 through REQ-UDS-025.

### M2 — Destructive-target registry + `.moai/memory/` backup + comment reconciliation

Introduces the registry of §C.0 as code (**11 rows / 18 sites** as measured on HEAD `89b2e4772` —
not the 7 an earlier draft assumed, and not the 10/17 v0.3.0 recorded before E1 landed), its drift
guard, the missing `.moai/memory/` backup, and the `dirs.go` brand+db group comment fix.

- **Step 0: re-run the §C.0 source scan before encoding anything.** The table is a plan-time
  snapshot and has already drifted once (17→18) between audit rounds. Encode what the scan returns
  at implementation time, and if it differs from §C.0, report the divergence rather than silently
  reconciling.
- The guard enumerates sites by static source scan, independently of the registry (REQ-UDS-007).
  Its independence is proven by the §C.4 falsification in `acceptance.md`, which injects an
  unregistered destructive site and requires an observed `--- FAIL`. A guard that passes that
  procedure is enumerating from the registry and must be rewritten.
- The `dirs.go` change is confined to the `// brand + db directories` group comment. Category D's
  comment is **accurate** and is not touched — AC-UDS-007 guards its `.moai/project/db/` contrast
  clause against deletion. See `spec.md` §A Defect 2's retracted finding.
- Covers REQ-UDS-006 through REQ-UDS-010.

### M3 — On-disk backup for the three in-memory-only files

Writes `.claude/settings.json`, `.moai/status_line.sh`, and `.gitignore` to the run-scoped backup
directory before the first destructive step, on both paths, retaining the in-memory merge-back.

- Includes the **failure path**: a backup write that fails aborts before the first destructive step
  and surfaces the `backup-write-failed:` sentinel naming the file (REQ-UDS-003, AC-UDS-020). This
  path had no AC before the iteration-1 revision; it is the one branch where the SPEC's central
  promise is actually load-bearing, so the abort must be verified by a call-recording spy proving
  `CleanMoaiManagedPaths` was never invoked — not inferred from an unchanged tree.
- Covers REQ-UDS-001 through REQ-UDS-005.

### M4 — HOME deletion-radius pinning

Routes `ensureGlobalSettingsEnv`'s HOME lookup through the injectable `userHomeDirFn` seam, extracts
the removal target to a named symbol, adds the radius guard, and demonstrates its failure against a
widened radius via `go test -overlay`.

- **The seam substitution is a precondition, not a nicety.** `update.go:843` calls the plain
  `userHomeDir()`; with no injection point and `t.Setenv("HOME", …)` forbidden (NFR-UDS-002), the
  radius guard cannot be written at all. Do this first, then the guard.
- Covers REQ-UDS-011 through REQ-UDS-014.

### M5 — Non-vacuous user-area safety guard

Replaces `TestMoaiUpdate_PreservesUserArea` with a guard driving a real production entry point,
deletes `simulateMoaiUpdate`, and demonstrates the replacement failing under a mutation that
touches `.claude/agents/harness/`.

- Covers REQ-UDS-015 through REQ-UDS-018.

### M6 — `mergeBackPreserveInventory` partial-restore reporting and coverage

Adds the restored-count and stopping-file detail to the three failure returns and covers each
branch deterministically.

- Covers REQ-UDS-026 through REQ-UDS-028.

## §G Anti-Patterns

- Re-specifying deprecated-path backup. It is E1's REQ-RIL2-015; duplicating it creates two owners
  for one behaviour.
- Testing the crash window by racing a real crash. Reach it deterministically (NFR-UDS-001).
- Using `git stash` for falsification. Prohibited by §D.
- Widening the marker-gate bypass beyond the restore entry point.
- Asserting the new guards pass without first observing them fail against unfixed code.
- Narrowing the `moai*` glob as a side effect of M2. It is a recorded hazard, not a deliverable.
- Changing the `defs.DeprecatedPaths` entry count while fixing its comment.
- Building the drift guard so it enumerates from the registry. That yields `count == count`, passes
  forever, and reproduces in M2 exactly the vacuous-guard defect §A Defect 3/4 exist to remove.
- Deleting the Category D `.moai/project/db/` contrast prose. It is accurate; the retracted
  iteration-1 finding that called it a mismatch is refuted in `spec.md` §A Defect 2.
- Satisfying AC-UDS-014 by renaming `simulateMoaiUpdate` instead of deleting it and driving real
  production code.

## §H Tier Justification

Tier **M**. Six milestones across roughly ten files: `deploy.go`, `update_template_sync.go`,
`update_clean_install.go`, `update_preserve_inventory.go`, `update.go`, `dirs.go`, plus four to five
new `_test.go` files and one new registry source file. No file exceeds a few hundred lines of
change; the estimate sits inside the 300-1000 LOC / 5-15 file band. It is not Tier S: it spans two
execution paths, adds a new user-facing entry point, and carries a design decision (§B). It is not
Tier L: the design decision is a single binary choice with a recorded rejected alternative, not a
design space warranting `design.md`, and the codebase analysis that would populate `research.md` is
already inlined in §C as a verified enumeration.

## §I Cross-References

- `.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md` — the inherited deprecated-path backup
  requirement and the scratch-worktree falsification precedent (`acceptance.md` §C.2).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §2, §3 — the cheapest-first ladder and
  the invariant behind §B.1.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` §E — the self-verification
  matrix each milestone reports.
- `CLAUDE.local.md` §13 (HOME in tests), §24 (user-owned namespace contract), §25 (template
  neutrality).
