# SPEC-UPDATE-DATA-SURVIVAL-001 — Acceptance Criteria

Version: 0.5.1 · Status: in-progress

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC. This binds every clause of an AC's expectation, not only
   its first: an AC whose headline assertion has no command behind it is defective even when its
   subordinate clauses do (the defect AC-UDS-014 carried at iteration 1).
2. **No vacuous `-run`.** `go test -run <pattern>` exits 0 when the pattern matches zero tests.
   Every AC below that uses `-run` therefore also asserts a literal `--- PASS: <exact test name>`
   line in the output. An AC that would still pass with its test deleted is a defect, and an AC
   whose only assertion is `exit 0` is rejected.
3. **A guard must be able to fail on the fixture it runs on.** The test is not "does this quantity
   change?" but "**could any change move it?**" A constant-valued assertion is vacuous only when no
   change to the implementation could move it. Two shapes of genuine vacuity bite here and are
   named so they are not re-introduced: (a) a **self-comparison** — a drift guard that both
   enumerates and validates from the same source yields `count == count` and can never fail
   (AC-UDS-005); (b) a **missing before-snapshot** — an invariance claim ("unchanged before and
   after") asserted from a single post-run observation, which passes whatever it observes
   (AC-UDS-013).

   **Preservation-guard carve-out.** A constant assertion that *would* move in the direction an
   unwanted change pushes it is a **preservation guard**, not a vacuous one — it is the mechanism
   by which "this must not change" becomes observable. Four ACs below are preservation guards and
   are classified as such rather than as violations of this clause:

   | AC | Constant | What moves it (so it can fail) |
   |---|---|---|
   | AC-UDS-007 (b) | `2` → `2` | Deleting the Category D `.moai/project/db/` contrast clause (`dirs.go:349`) drives it to `1`; deleting the whole Category D banner (`dirs.go:344-351`) drives it to `0` |
   | AC-UDS-010 | exit 0 → exit 0 | Any success-path regression (NFR-UDS-005) drives it non-zero |
   | AC-UDS-018 | exit 0 → exit 0 | A path-handling change that breaks the windows cross-build drives it non-zero |
   | AC-UDS-019 | `0` → `0` | Committing or staging any `internal/template/templates/**` edit drives it ≥ 1 |

   **Relation to §A.2.** §A.2 rejects an AC whose only assertion is `exit 0`; that rejection is
   scoped to `go test -run <pattern>`, where exit 0 is also the outcome when the pattern matches
   **zero** tests — the exit code there carries no information about whether anything ran.
   AC-UDS-010 (whole-package `go test`, no `-run`) and AC-UDS-018 (`go build`) have no such
   zero-match mode: their exit 0 means the suite or the build actually completed. The two clauses
   therefore do not conflict.
4. **Baselines are observed, not assumed, and are re-baselined when the tree moves.** Every
   "baseline" line in §B was re-measured on 2026-08-01 against the **code baseline `8cc108ddb`**
   (= `origin/main`, verified an ancestor of this branch's HEAD), on branch
   `feat/SPEC-UPDATE-DATA-SURVIVAL-001`. Where a baseline names a commit, it names `8cc108ddb`.

   **On the HEAD anchor.** The measurements were recorded by artifact commit `89b2e4772`; every
   commit from `89b2e4772` to the current HEAD changes SPEC documents only
   (`git diff --name-only 89b2e4772 HEAD` returns no `.go` path), so the code baseline these
   figures describe is unchanged by them. A reader comparing `git rev-parse HEAD` against a
   recorded SHA will see a mismatch — that is the self-referential-commit hazard (a commit cannot
   name its own SHA), not a measurement error. The load-bearing anchor is `8cc108ddb`.

   **Why the previous anchor was retired.** v0.3.0 anchored these baselines to worktree HEAD
   `a8b42e112` / code baseline `d5336214e`, and asserted "every commit between the two changes SPEC
   documents only, so no Go source differs." That premise is **measurably false on this tree**, and
   the anchor is unreachable from it:

   ```
   $ git merge-base --is-ancestor a8b42e112 HEAD && echo "ANCESTOR: yes" || echo "ANCESTOR: NO"
   ANCESTOR: NO
   $ git merge-base --is-ancestor 8cc108ddb HEAD && echo "ANCESTOR: yes" || echo "ANCESTOR: NO"
   ANCESTOR: yes
   $ git diff --name-only d5336214e..HEAD | grep -v '^\.moai/' | wc -l
         19
   ```

   The 19 differing files are E1's (`SPEC-UPDATE-REINSTALL-LOOP-002`) landing — run PR #1261
   (`beeb0ebc2`) and sync PR #1264 (`8cc108ddb`) — which added `internal/cli/update_residue_cleanup.go`
   (a new destructive site, §C.0 row 11) and shifted line coordinates across `update.go`,
   `update_preserve_inventory.go`, `update_archive.go`, `update_clean_install.go`, and
   `update_cleanup.go`. Every coordinate and count in this file was re-measured against
   `89b2e4772`; nothing is carried over from the retired anchor.

   **Time-of-check-to-time-of-use.** These baselines describe HEAD `89b2e4772`. If another Epic SPEC
   lands Go changes before run-phase entry, the same drift recurs — so M2's first act is to re-run
   the §C.0 source scan rather than trust the recorded table (see AC-UDS-005).
5. **Every AC cites the REQ it verifies.** An AC with no REQ citation is untraceable, and a REQ with
   no citing AC is uncovered. §B.0 carries the coverage map.
6. **Every new guard needs a falsification** proving it FAILS against unfixed code. §C gives the
   runnable procedure.
7. **`git stash` is prohibited.** The checkout is shared with concurrent sessions: `git stash push`
   refuses untracked files without `-u`, and with `-u` it is repository-global and can swallow a
   parallel session's work. Falsification uses a scratch `git worktree` driven by `go -C`, or
   `go test -overlay`.
8. **Crash windows are reached deterministically, never by racing a crash** (NFR-UDS-001).

## §B Acceptance criteria

### §B.0 REQ ↔ AC coverage map

Every REQ has at least one citing AC; every AC names the REQ it verifies.

| REQ | Verified by | REQ | Verified by |
|---|---|---|---|
| REQ-UDS-001 | AC-UDS-008 | REQ-UDS-015 | AC-UDS-014, AC-UDS-015 |
| REQ-UDS-002 | AC-UDS-008 | REQ-UDS-016 | AC-UDS-014 |
| REQ-UDS-003 | **AC-UDS-020** | REQ-UDS-017 | AC-UDS-015 |
| REQ-UDS-004 | AC-UDS-010 | REQ-UDS-018 | AC-UDS-014 |
| REQ-UDS-005 | AC-UDS-009 | REQ-UDS-019 | AC-UDS-001 |
| REQ-UDS-006 | AC-UDS-005 | REQ-UDS-020 | AC-UDS-001 (clauses 3+4 — a non-empty planted set is destroyed, and stays absent on return) |
| REQ-UDS-007 | AC-UDS-005, §C.4 | REQ-UDS-021 | AC-UDS-002 |
| REQ-UDS-008 | AC-UDS-006 | REQ-UDS-022 | AC-UDS-002 |
| REQ-UDS-009 | AC-UDS-007 | REQ-UDS-023 | AC-UDS-004 |
| REQ-UDS-010 | AC-UDS-005 (row 12 cross-SPEC assignment) | REQ-UDS-024 | AC-UDS-004 |
| REQ-UDS-011 | AC-UDS-011, AC-UDS-012 | REQ-UDS-025 | AC-UDS-003 |
| REQ-UDS-012 | AC-UDS-013 | REQ-UDS-026 | AC-UDS-017 |
| REQ-UDS-013 | AC-UDS-011, AC-UDS-013 | REQ-UDS-027 | AC-UDS-016 |
| REQ-UDS-014 | AC-UDS-012 | REQ-UDS-028 | AC-UDS-016 |

NFR coverage: NFR-UDS-001 → AC-UDS-008 (separate-call reach into the crash window);
NFR-UDS-002 → AC-UDS-013; NFR-UDS-003 → AC-UDS-019; NFR-UDS-005 → AC-UDS-010;
NFR-UDS-006 → AC-UDS-018; NFR-UDS-007 → AC-UDS-007.
NFR-UDS-004 (Go conventions — `snake_case.go` filenames, `fmt.Errorf("…: %w", err)` wrapping) is
deliberately **not** mapped to an AC in this SPEC: it is enforced by the project toolchain
(`gofmt`, `go vet`, `golangci-lint`) on every milestone, not by a test this SPEC authors. The map
above is complete under that exemption.

### M1 — Failure contract

#### AC-UDS-001 — a mid-run failure after the first destructive step writes and prints a recovery manifest (REQ-UDS-019, REQ-UDS-020)

```bash
# (a) the guard itself
go test -run 'TestUpdateFailure_WritesRecoveryManifest' -count=1 -v ./internal/cli/
# (b) the fixture's planted set is a NON-EMPTY literal declared in the test source
sed -n '/^var plantedMoaiManagedPaths = \[\]string{/,/^}/p' \
  internal/cli/update_recovery_manifest_test.go | grep -cE '^\s+"'
```

Expected: (a) a `--- PASS: TestUpdateFailure_WritesRecoveryManifest` line; (b) prints **`≥ 1`**.
The test plants each path in `plantedMoaiManagedPaths` into a `t.TempDir()` fixture, injects a
failing step after `CleanMoaiManagedPaths`, then asserts **four** things:

1. a recovery manifest exists inside the run-scoped backup directory naming the failed step and the
   restore command;
2. the same manifest text appears in the captured writer;
3. **every path in `plantedMoaiManagedPaths` existed on disk immediately before
   `CleanMoaiManagedPaths` ran, and was gone immediately after it returned** — the destruction
   actually happened, so the removed set is non-empty by construction;
4. **every path in `plantedMoaiManagedPaths` is still absent when the outer call returns** — the
   update did not silently restore them (REQ-UDS-020, no automatic rollback).

**Clauses 3+4 are REQ-UDS-020's only mechanical coverage, and clause 3 is what makes clause 4
non-vacuous.** An earlier draft mapped REQ-UDS-020 to this AC while the AC body asserted nothing
about rollback (covered on paper only); v0.3.0 added the absence assertion but left its quantifier
unbounded, so on a fixture containing no moai-managed paths "all of ∅ are still absent" held
trivially and **adding an automatic rollback could not move it** — vacuous by §A.3's own criterion,
which is the exact condition the original defect was raised to fix.

**Why the non-emptiness pin is sourced from the fixture, not from the call.** Asserting "≥ 1 path
was removed" from a count the test computes by observing `CleanMoaiManagedPaths`' own output would
compare the function against itself — §A.3 shape (a), a self-comparison that can never fail. The
count therefore comes from **known-planted paths**: a literal `[]string` in the test source, whose
non-emptiness command (b) verifies by reading that source, and whose actual planting clause 3
verifies against the filesystem.

**Falsification (why the criterion is unsatisfiable on an empty removed-set).** Three independent
directions move it:

| Mutation | Which clause fails | Why |
|---|---|---|
| Empty `plantedMoaiManagedPaths` | (b) → `0` | the AC requires `≥ 1`; the empty-fixture escape is closed at the source level |
| Fixture declares paths but does not create them | clause 3 pre-assertion | a planted path that never existed cannot be observed present before the call |
| An automatic rollback is added | clause 4 | the recorded paths reappear on return |

A rollback added to a tree whose fixture is empty no longer passes silently: (b) fails first.

Baseline: the test does not exist; `go test -run 'TestUpdateFailure_WritesRecoveryManifest'
./internal/cli/` currently prints `ok … [no tests to run]` and exits 0 — which is exactly the
vacuity this AC's `--- PASS` requirement excludes. Command (b) is likewise unmeasurable until M1
creates `internal/cli/update_recovery_manifest_test.go`; that is expected for a new-guard AC and is
not a baseline gap (same precedent as AC-UDS-013's `t.Setenv` grep).

#### AC-UDS-002 — the restore entry point runs on a tree whose project marker was destroyed (REQ-UDS-021, REQ-UDS-022)

```bash
go test -run 'TestRestore_ProceedsWithoutProjectMarker' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestRestore_ProceedsWithoutProjectMarker` line. The fixture deletes
`.moai/config/sections/system.yaml`, invokes the restore entry point against a backup directory,
and asserts the file is restored — proving the marker-gate lockout is escaped.

Baseline: `go run ./cmd/moai update --help 2>&1 | grep -ci restore` → `0`. No restore surface
exists on the update command today.

#### AC-UDS-003 — the ordinary update path still refuses a marker-less tree (REQ-UDS-025)

```bash
go test -run 'TestUpdate_RejectsTreeWithoutProjectMarker' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestUpdate_RejectsTreeWithoutProjectMarker` line, asserting the returned
error contains `not a moai project`. This pins REQ-UDS-025: the bypass is scoped to restore alone.

Baseline, re-measured on HEAD `2255165f5`: the gate is the `checkProjectMarker(cwd)` call at
`internal/cli/update.go:251`; the predicate and its message live in
`internal/cli/update_restore.go:21`/`:27` and emit
`not a moai project: %s not found in the current directory` (`%s` =
`.moai/config/sections/system.yaml`). The pre-M1 citation was `update.go:236`; M1 commit
`4ddd35120` both moved the call site and extracted the predicate into `update_restore.go`, so that
coordinate no longer resolves. This AC's verification command is unchanged — it asserts on the
returned error text, not on a line number.

#### AC-UDS-004 — restore is idempotent and refuses an unrecognised directory (REQ-UDS-023, REQ-UDS-024)

```bash
go test -run 'TestRestore_IdempotentAndRefusesForeignDir' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestRestore_IdempotentAndRefusesForeignDir` line. The test applies the same
backup twice and asserts the resulting tree hashes are equal, then points the entry point at a
directory lacking the backup marker file and asserts a non-nil error and an unmodified tree.

Baseline: no restore entry point with **marker-gate bypass + idempotency + foreign-directory
refusal** (REQ-UDS-022/023/024) exists, and none is reachable from the CLI. A restore surface does
exist and must not be duplicated: `backup.RestoreMoaiConfig` (`backup/restore.go:40`) restores from
a backup directory but is **mid-run only** — its two live callers are `update_clean_install.go:437`
and `update_template_sync.go:406` (re-measured on HEAD `2255165f5`; recorded pre-M1 as `:429` and
`:397`, both shifted by M1 commit `4ddd35120`) — and `moai migrate restore-skill` is the only user-invocable
restore. **M1 must state explicitly whether the new entry point extends `RestoreMoaiConfig` or is
separate**, per the two-owners hazard in `plan.md` §H.

### M2 — Destructive-target registry

#### AC-UDS-005 — every destructive site in the update subsystem is registered (REQ-UDS-006, REQ-UDS-007, REQ-UDS-010)

```bash
go test -run 'TestDestructiveTargetRegistry_CoversAllSites' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestDestructiveTargetRegistry_CoversAllSites` line. The guard **enumerates
the sites itself** by statically scanning `internal/cli/update/**` and `internal/cli/update*.go`
(excluding `_test.go`) for `os.RemoveAll(` and `os.Rename(` call sites, keyed on
(file, enclosing function, occurrence count), then asserts that scanned multiset equals the
registry's recorded multiset, and that every registry row carries either a protection-set
assignment or a recorded exemption reason.

**The enumeration source must not be the registry.** Per REQ-UDS-007, a guard that derives both
sides of the comparison from the registry yields `count == count` and passes forever; the source
scan is what makes this AC able to fail. §C.4 is its runnable falsification.

**Baseline, re-measured on HEAD `89b2e4772` (code baseline `8cc108ddb`), and re-measured again on
HEAD `2255165f5` at M2 Step 0.** The scan finds
**18 call sites across 11 (file, function) pairs** — both figures identical across the two
measurements, and the (file, function) pair set identical too. The M2 Step 0 re-scan moved **only
line coordinates**, in rows 7 and 9 (see the row-7/row-9 note below the table). Command and
verbatim output:

```
$ grep -rn 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | wc -l
      18
```

**The pair count and the site total are re-derived from the Go source, not from this document.**
v0.3.0 published two "independent re-derivation" commands that grepped `acceptance.md` itself
(`grep -c '^| [0-9]* | \`internal' acceptance.md`), so they returned `10` / `17` forever regardless
of the tree — §A.3 shape (a), a self-comparison, applied to the document's own consistency check.
That is why a 17→18 drift produced no signal. Both are replaced with source-tree scans:

```
$ grep -rn 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | cut -d: -f1 | sort | uniq -c
   2 internal/cli/update_archive.go
   1 internal/cli/update_clean_install.go
   1 internal/cli/update_cleanup.go
   3 internal/cli/update_namespace_protect.go
   1 internal/cli/update_residue_cleanup.go
   1 internal/cli/update.go
   4 internal/cli/update/backup/backup.go
   5 internal/cli/update/deploy/deploy.go

$ grep -rl 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | while read -r f; do
      awk -v F="$f" '/^func /{fn=$2; sub(/\(.*/,"",fn)} /os\.RemoveAll\(|os\.Rename\(/{print F" "fn}' "$f"
    done | sort -u | wc -l
      11
```

The table below MUST equal those two figures (18 sites, 11 pairs). A tree that drifts moves the
commands' output and the equality fails — which is the signal the replaced commands could never
produce.

| # | File | Function | Sites |
|---|---|---|---|
| 1 | `internal/cli/update/deploy/deploy.go` | `CleanMoaiManagedPaths` | 3 (`:83`, `:105`, `:121`) |
| 2 | `internal/cli/update/deploy/deploy.go` | `MigrateLegacyMemoryDir` | 2 (`:169`, `:176`) |
| 3 | `internal/cli/update_archive.go` | `archiveSkill` | 1 (`:101`) |
| 4 | `internal/cli/update_archive.go` | `archiveLegacySkills` | 1 (`:322`) |
| 5 | `internal/cli/update/backup/backup.go` | `BackupMoaiConfig` | 3 (`:107`, `:135`, `:140`) |
| 6 | `internal/cli/update/backup/backup.go` | `CleanupOldBackups` | 1 (`:259`) |
| 7 | `internal/cli/update_clean_install.go` | `runCleanReinstall` | 1 (`:315`) |
| 8 | `internal/cli/update_cleanup.go` | `removeDeprecatedFile` | 1 (`:324`) |
| 9 | `internal/cli/update.go` | `ensureGlobalSettingsEnv` | 1 (`:853`) |
| 10 | `internal/cli/update_namespace_protect.go` | `backupUserOwnedNamespace` | 3 (`:225`, `:233`, `:243`) |
| 11 | `internal/cli/update_residue_cleanup.go` | `runV3ResidueCleanup` | 1 (`:135`) |

**Row 11 is new in v0.4.0.** `internal/cli/update_residue_cleanup.go` was added by E1's run PR #1261
(`beeb0ebc2`, `git log --oneline --diff-filter=A -- internal/cli/update_residue_cleanup.go`) after
v0.3.0's registry was measured, so it was absent from the 17-site table. Rows 3, 4, 7, and 9 also
carry re-measured coordinates (`:92`→`:101`, `:304`→`:322`, `:271`→`:307`, `:766`→`:843`); rows 1,
2, 5, 6, 8, 10 were re-measured and are unchanged.

**Rows 7 and 9 moved a second time at M2 Step 0 — a DIFFERENT, self-inflicted event.** The M2
Step 0 re-scan on HEAD `2255165f5` shifted row 7 `:307`→`:315` and row 9 `:843`→`:853`. This is
**not** a second external TOCTOU: the cause is this SPEC's own **M1 commit `4ddd35120`** (recovery
manifest + restore entry point), which added code above both sites. The two drift events are
therefore distinguishable and must stay so — the v0.3.0→v0.4.0 shifts recorded in the paragraph
above were caused by **E1** (`SPEC-UPDATE-REINSTALL-LOOP-002`, a sibling SPEC landing between plan
rounds), while these two are caused by **M1 of this SPEC**. The full drift chain per row is
therefore row 7 `:271`→`:307`→`:315` and row 9 `:766`→`:843`→`:853`. Site total (18), pair count
(11), and the (file, function) pair set were all unchanged by the M1 event — only coordinates
moved, which is why this AC's guard (keyed on file + enclosing function + occurrence count, never
on line numbers) required no change and why its verification command is untouched.

Rows 5, 6, and 10 carry an **exemption reason** rather than a protection-set assignment, but for two
materially different reasons — an earlier form of this paragraph collapsed them into one ("directories
the same run created") and was false for row 6:

- **Rows 5 and 10 — same-call rewind.** `BackupMoaiConfig` (`backup.go:107/135/140`) unwinds the
  `backupDir` this very call created; `backupUserOwnedNamespace`
  (`update_namespace_protect.go:225/233/243`) is defensive cleanup of the staging directory this
  very call created on a failed namespace backup. Neither can touch data that predates the run.
- **Row 6 — retention pruning of moai-authored backup directories.** `CleanupOldBackups`
  (`backup.go:259`) deletes `backups[:len(backups)-keepCount]` — the **oldest excess backups of
  previous runs**, not anything this run created. This matches `plan.md` §C.0 row 6
  ("retention pruning of moai-authored backup dirs"); the two artifacts now agree. Its exemption
  rests on a different ground: every directory it deletes was authored by `moai` itself under the
  backup root and matched the `YYYYMMDD_HHMMSS` name filter (`backup.go:228-241`, `len(name) == 15`,
  8 digits + `_` + 6 digits), so it destroys moai-owned restore points under a declared retention
  policy — never user-authored data. It is exempt from the *user-data* protection set, NOT harmless
  to the recovery contract.

**Row 11 classification — user data, NOT exempt.** `runV3ResidueCleanup`
(`update_residue_cleanup.go:65`) removes `sweep` — the existence-refiltered subset of
`scanDeprecatedPaths`' return (`update_residue_cleanup.go:95-100`), i.e. `defs.DeprecatedPaths`
entries **that predate the run and still exist on disk** — user-tree residue, not directories this
call authored. It therefore fails the same-call-rewind test that exempts rows 5/6/10. It is not
unprotected either: the function backs up before deleting (`:114-131`, `backupDeprecatedPaths`,
aborting on failure with the `DEPRECATED_BACKUP_FAILED` sentinel at `:116`/`:125`), which is E1's
REQ-RIL2-019/015 contract. Its registry assignment is therefore the **cross-SPEC** one — the same
assignment row 12 of `plan.md` §C.1 records for `update_clean_install.go:315` — and this SPEC does
not re-specify it (§C Scope Exclusions, REQ-UDS-010).

**REQ-UDS-002 run-scoped backup directories ARE in this pruning target set** when they are created
under the same backup root with a `YYYYMMDD_HHMMSS` name, because that is precisely the filter
`CleanupOldBackups` applies. M3 must therefore state whether its run-scoped directory adopts that
naming; if it does, a sufficiently old restore point can be pruned by a later run's rotation, and the
interaction between the retention window and the recovery contract is a follow-up review item.

Rows 1-4, 7-9, and 11 are the genuine user-data destructive sites and carry protection assignments
per `plan.md` §C.

No registry exists today, so nothing fails when a new site is added unprotected.

#### AC-UDS-006 — `.moai/memory/` is backed up before the both-exist removal (REQ-UDS-008)

```bash
go test -run 'TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval' -count=1 -v ./internal/cli/update/deploy/
```

Expected: a `--- PASS: TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval` line. The fixture creates
both `.moai/memory/` (with a sentinel file) and `.moai/state/`, runs the migration, and asserts the
sentinel's bytes are present under the backup directory after `.moai/memory/` is gone.

Baseline, re-measured on HEAD `89b2e4772`: `internal/cli/update/deploy/deploy.go:176` calls
`os.RemoveAll(legacyDir)` with no preceding copy; `.moai/memory` appears in none of
`preserveInventoryRoots` (`update_preserve_inventory.go:68-72`), `BackupMoaiConfig`
(`backup/backup.go:28`, `.moai/config` only), or `userOwnedScanRoots`
(`update_namespace_protect.go:39-43`).

#### AC-UDS-007 — the `.moai/db` group comment names its authorising SPEC, and Category D's accurate prose survives (REQ-UDS-009, NFR-UDS-007)

```bash
# (a) the brand+db group banner now names its authorising SPEC.
#     Range: the group heading line through the group's first registered path.
#     Both anchors are content, not indentation or line numbers — see the range note below.
sed -n '/brand + db directories/,/Path: *"\.moai\/project\/brand"/p' internal/defs/dirs.go \
  | grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001'
# (b) Category D's contrast clause is PRESERVED, not deleted
grep -c 'moai/project/db' internal/defs/dirs.go
# (c) the catalogue is otherwise untouched
go test -run 'TestDeprecatedPaths' -count=1 -v ./internal/defs/
```

Expected: (a) prints `≥1`; (b) prints **`2`** — unchanged; (c) a `--- PASS: TestDeprecatedPaths`
line with the entry-count assertions intact.

**Fails when:** the group banner is left bare (a → `0`), or the Category D contrast clause is
deleted (b → `1`), or the whole Category D banner is deleted (b → `0`), or the slice body changes
(c fails on the count assertion).

**What (b) actually guards — the whole Category D banner, not the contrast clause alone.** `grep -c
'moai/project/db'` matches **two** lines: `dirs.go:349` (the contrast clause REQ-UDS-009 names) and
`dirs.go:346` (an unrelated descriptive mention of the removed `.moai/project/db/` scaffold in the
same banner). v0.3.0 stated the falsifier as `2 → 0`; that arithmetic is wrong. Measured:

```
$ grep -n 'moai/project/db' internal/defs/dirs.go
346:	// .moai/project/db/ scaffold) was fully removed. db.yaml is no longer
349:	// .moai/project/db/ docs (a preserve root, manual deletion per CHANGELOG),
$ sed '349d' internal/defs/dirs.go | grep -c 'moai/project/db'
1
$ sed '344,351d' internal/defs/dirs.go | grep -c 'moai/project/db'
0
```

The guard still functions — the AC expects exactly `2`, so `1` fails it — but the pattern binds
`:346` as well, which REQ-UDS-009 does not separately protect. That wider binding is **accepted
deliberately** rather than narrowed: `:346` is accurate prose in the same banner, deleting it is
also an unwanted change, and a narrower anchor would couple the AC to one line's exact wording. (b)
is therefore documented as guarding the Category D banner's `.moai/project/db/` prose as a whole.

Baseline, measured on this tree:

```
$ sed -n '/brand + db directories/,/Path: *"\.moai\/project\/brand"/p' internal/defs/dirs.go | grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001'
0
$ grep -c 'moai/project/db' internal/defs/dirs.go
2
```

**Range note — why this window, and why it is not widened backward.** An earlier draft ended the
range at `/^\t{/`, which captures exactly two lines (the heading and the opening brace) and is
coupled to both the tab indentation and the assumption that the SPEC-ID sits on or below the
heading. Two changes fix that. First, the end anchor is now the group's first registered path
(`.moai/project/brand`), which is content and survives a reindent or an added blank line. Second,
REQ-UDS-009 pins the SPEC-ID's **placement** to the region this window covers — on the
`// brand + db directories` line or between it and the group's first entry — so a conforming
implementation cannot land outside the window.

The window is deliberately **not** extended backward. `dirs.go:302` already contains an unrelated
`SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` mention (the note recording that a *different* entry was
reversed). Measured: a ten-line lookback from the heading captures that line, which would drive
(a)'s baseline from `0` to `1` and make the AC pass before REQ-UDS-009 is implemented at all — a
false green. The placement requirement is what makes the forward-only window sufficient.

Falsification of (a), measured on this tree: inserting a conforming SPEC-ID line immediately below
the heading flips (a) from `0` to `1`, and the range still does not reach the `dirs.go:302` note:

```
$ sed -e 's|// brand + db directories|// brand + db directories\n\t\t// Deprecated by SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001; not under any preserve root.|' internal/defs/dirs.go \
    | sed -n '/brand + db directories/,/Path: *"\.moai\/project\/brand"/p' | grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001'
1
$ sed -n '/brand + db directories/,/Path: *"\.moai\/project\/brand"/p' internal/defs/dirs.go | grep -c 'This reverses'
0
```

(a) is `0` because the group carries only the bare comment `// brand + db directories`
(`dirs.go:305`) — that is the gap REQ-UDS-009 closes. (b) is `2` and **must stay `2`**: an earlier
draft of this AC expected `0`, on the retracted premise that the Category D comment misnamed its
path. `sed -n '305,358p' internal/defs/dirs.go` refutes that premise — `.moai/db` (`:313`) belongs
to the brand+db group, Category D's sole entry is `.moai/config/sections/db.yaml` (`:354`), and the
comment names it exactly. The `.moai/project/db/` mentions are a deliberate contrast clause stating
what is *not* removed. The old expectation would have deleted accurate documentation; this AC now
guards it instead. See §A Defect 2's retracted finding in `spec.md`.

### M3 — On-disk backup before destruction

#### AC-UDS-008 — the three in-memory-only files exist on disk before the first destructive step (REQ-UDS-001, REQ-UDS-002, NFR-UDS-001)

```bash
go test -run 'TestBackup_OnDiskBeforeFirstDestructiveStep' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestBackup_OnDiskBeforeFirstDestructiveStep` line. The test invokes the
backup step and the destructive step as **separate calls** and asserts, between them, that
`.claude/settings.json`, `.moai/status_line.sh`, and `.gitignore` each exist under the run-scoped
backup directory with bytes identical to the originals. This is the deterministic reach into the
crash window required by NFR-UDS-001 — no crash is raced.

Baseline: `grep -rn 'settings.json' internal/cli/update/backup/ --include='*.go' | grep -v _test |
wc -l` → `0`. The only backup of that file is the `[]byte` slice `mergeableBackups`
(declared `update_template_sync.go:300`, populated at `:397` — re-measured on HEAD `2255165f5`;
recorded pre-M1 as `:294`/`:388`).

#### AC-UDS-009 — both paths use the same on-disk backup mechanism (REQ-UDS-005)

```bash
go test -run 'TestBackup_OnDiskCoverageParityAcrossPaths' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestBackup_OnDiskCoverageParityAcrossPaths` line, asserting that the set of
on-disk backed-up paths produced by the normal path equals the set produced by the clean-reinstall
path.

Baseline, re-measured on HEAD `2255165f5`: the two paths build the in-memory slice independently —
`update_template_sync.go:300`/`:397` versus `update_clean_install.go:356`/`:359` — so parity is
currently coincidental rather than asserted. (Recorded pre-M1 as `:294`/`:388` and `:348`/`:351`;
all four shifted by M1 commit `4ddd35120`.)

#### AC-UDS-010 — the success path is unchanged (REQ-UDS-004, NFR-UDS-005)

```bash
go test -count=1 ./internal/cli/... ./internal/cli/update/...
```

Expected: exit 0 with no `--- FAIL` line. The pre-existing update tests are the regression surface
for NFR-UDS-005.

Baseline: the same command exits 0 on this tree.

#### AC-UDS-020 — a failed backup write aborts before the first destructive step (REQ-UDS-003)

```bash
go test -run 'TestBackup_AbortsBeforeDestructionOnWriteFailure' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestBackup_AbortsBeforeDestructionOnWriteFailure` line. The test injects a
failing on-disk backup writer for one of the three files, runs the update, and asserts **three
things**:

1. the returned error contains the grep-able sentinel `backup-write-failed:` followed by the
   relative path of the file that could not be backed up;
2. `CleanMoaiManagedPaths` was **not** invoked — observed via a call-recording spy passed to the
   step, not inferred from the tree;
3. the fixture tree is byte-identical to its pre-run state, so nothing was destroyed.

The failure is injected, never raced (§A.8). Clause 2 is the load-bearing one: clause 3 alone would
also hold if the destructive step ran and happened to remove nothing, so the spy is what makes the
"aborts *before*" claim observable rather than inferred.

**Fails when:** the abort is implemented as "log and continue" (clause 2 flips), the sentinel is
absent or omits the filename (clause 1), or the abort fires after partial destruction (clause 3).

Baseline: no such abort exists and no test names it —

```
$ grep -rn 'backup-write-failed' internal/cli/ --include='*.go' | wc -l
0
$ go test -run 'TestBackup_AbortsBeforeDestructionOnWriteFailure' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	0.539s [no tests to run]   # exit 0
```

— the second command is precisely the vacuity §A.2 excludes, which is why the `--- PASS` line is
required above. REQ-UDS-003 had **no** citing AC before this revision: it is the only failure-path
requirement in M3's safety contract, and leaving it unverified would ship the SPEC's central
promise (nothing is destroyed before it exists on disk) untested on the one path where it matters.

### M4 — HOME deletion-radius pinning

#### AC-UDS-011 — widening the HOME removal target fails the suite (REQ-UDS-011, REQ-UDS-013)

```bash
# (a) the seam substitution REQ-UDS-013 requires — ensureGlobalSettingsEnv must call the variable
sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
# (b) the guard itself
go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/
```

Expected: (a) prints `1`; (b) a `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius` line.
The test overrides `userHomeDirFn` to a `t.TempDir()`, creates both `<tmp>/.claude/hooks/moai/` and
a sibling `<tmp>/.claude/hooks/user-hook.sh`, runs the function, and asserts the sibling survives
while the `moai` subdirectory is gone.

**Fails when:** the seam substitution is skipped (a → `0`, and (b) cannot be written at all — see
the baseline), or the removal radius is widened (b fails; §C.1 demonstrates it).

Baseline, measured on this tree:

```
$ sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
0
$ grep -rn 'globalHooksDir' internal/cli/ --include='*_test.go' | wc -l
       0
```

(a) is `0` because `ensureGlobalSettingsEnv` (`update.go:842`) calls the plain `userHomeDir()` at
`update.go:843`; the injectable variable `userHomeDirFn` exists at `glm_tools.go:123` but is not
used here. **This AC is blocked on the seam**: with no injection point and `t.Setenv("HOME", …)`
forbidden by NFR-UDS-002, the guard has no way to redirect HOME, so command (a) is a precondition of
(b) rather than a decoration. `globalHooksDir` likewise appears only at `update.go:851-853`; no test
observes it. Falsification is AC-UDS-012.

(Coordinates re-measured on HEAD `2255165f5`; recorded pre-M1 as `:832`/`:833`/`:841-843`, all
shifted by +10 by M1 commit `4ddd35120`. Note that `update.go:843` is the `userHomeDir()` call —
it is **not** the `os.RemoveAll` site AC-UDS-005 row 9 registers, which moved to `update.go:853`.
The pre-M1 tables happened to record the `os.RemoveAll` site at `:843`, so the two values can be
confused when comparing old and new records.)

#### AC-UDS-012 — the radius guard fails against a widened radius (REQ-UDS-011, REQ-UDS-014)

Runnable falsification via `go test -overlay` — see §C.1. Expected: a `--- FAIL:
TestEnsureGlobalSettingsEnv_HooksRemovalRadius` line when the overlay widens the target to
`~/.claude/hooks`.

Baseline: the audit lens ran precisely this mutation against the current suite and observed
`ok github.com/modu-ai/moai-adk/internal/cli 18.305s` — every test passed with the widened radius.

#### AC-UDS-013 — the HOME test does not touch the operator's real home (REQ-UDS-012, REQ-UDS-013, NFR-UDS-002)

```bash
# before-snapshot — MUST run first
find ~/.claude/hooks -mindepth 1 2>/dev/null | sort > /tmp/uds-home-before.txt

go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/

# after-snapshot + comparison
find ~/.claude/hooks -mindepth 1 2>/dev/null | sort > /tmp/uds-home-after.txt
diff /tmp/uds-home-before.txt /tmp/uds-home-after.txt; echo "diff-exit=$?"

# the mechanism that makes the isolation real, not incidental
grep -c 't.Setenv("HOME"' internal/cli/update_home_radius_test.go
sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
```

Expected: a `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius` line; `diff` produces **no
output** and `diff-exit=0`; the `t.Setenv("HOME"` grep prints `0`; the `userHomeDirFn()` grep
prints `1`.

**Fails when:** the test writes into or deletes from the operator's real `~/.claude/hooks` (the
diff becomes non-empty), or it achieves isolation by mutating the process environment instead of
the seam (the `t.Setenv` grep flips to `≥1`, which NFR-UDS-002 forbids because a process-wide HOME
mutation is visible to every parallel test in the package).

**Why the before-snapshot is mandatory (§A.3b).** An earlier draft of this AC ran a single
post-test `ls ~/.claude/hooks | head` and expected the listing to be "unchanged before and after".
With nothing captured beforehand there was nothing to compare against, so the AC passed on whatever
it printed — including a listing the test had just corrupted. Two snapshots plus `diff` is what
makes the invariance claim observable.

**Independence from AC-UDS-011.** That earlier draft also ran the *same* test as AC-UDS-011 with no
additional assertion, so it verified nothing AC-UDS-011 did not already cover. The four
expectations above are all absent from AC-UDS-011: the filesystem diff, the `t.Setenv` prohibition,
and the mechanism grep are this AC's own.

Baseline, measured on this tree:

```
$ sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
0
```

The seam is absent — see AC-UDS-011's baseline. The test file
`internal/cli/update_home_radius_test.go` does not yet exist, so its `t.Setenv` grep is
unmeasurable until M4 creates it; that is expected for a new-guard AC and is not a baseline gap.

> **Retracted baseline claim.** The earlier draft asserted "`ensureGlobalSettingsEnv` resolves HOME
> via `userHomeDir()`, so the indirection already exists." The conclusion is wrong: `userHomeDir` is
> a plain function (`homedir.go:14`), not a reassignable variable. The reassignable seam is
> `userHomeDirFn` (`glm_tools.go:123`), which this call site does not use. No indirection existed.
> (The coordinate cited there, `update.go:756`, was also correct only against the retired
> `d5336214e` baseline; the call was at `update.go:833` on HEAD `89b2e4772`, and is at
> `update.go:843` on HEAD `2255165f5` after M1 commit `4ddd35120`.)

### M5 — Non-vacuous user-area safety guard

#### AC-UDS-014 — the replacement guard drives real production code (REQ-UDS-015, REQ-UDS-016, REQ-UDS-018)

```bash
# (a) the fake's DEFINITION is gone, not merely renamed
grep -cE '^func simulateMoaiUpdate|^func simulate[A-Za-z]*Update' internal/cli/update_safety_test.go
# (b) a real production entry point is EXECUTED — the load-bearing assertion
go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -covermode=set \
  -coverpkg=./internal/cli/...,./internal/cli/update/... \
  -coverprofile=/tmp/uds-m5.out ./internal/cli/
go tool cover -func=/tmp/uds-m5.out \
  | grep -E 'CleanMoaiManagedPaths|MigrateLegacyMemoryDir|runCleanReinstall|BackupMoaiConfig' \
  | grep -vE '[[:space:]]0\.0%$' > /tmp/uds-covered.txt
test -s /tmp/uds-covered.txt || { echo "VACUOUS: guard executed no registry production function"; exit 1; }
cat /tmp/uds-covered.txt
# (c) the guard passes
go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/
```

Expected: (a) prints `0`; (b) the `test -s` guard passes silently and `cat` prints **at least one**
line naming a registry production function with coverage **strictly above `0.0%`**; (c) a
`--- PASS: TestMoaiUpdate_PreservesUserArea` line.

**Command (b) is the whole point of this AC, and it must be execution evidence — not a grep.** Two
successive drafts got this wrong in different ways. The first stated "while the test body invokes a
real update entry point" as prose with no command behind it, which left the AC satisfiable by
renaming the fake: `simulateMoaiUpdate` → `simulateUpdate` drives (a) to `0`, the untouched
tautological test still emits `--- PASS`, and nothing checks that production code runs. The second
promoted a production-symbol **text grep** from baseline to expected postcondition — which is
satisfiable by a single comment, because grep cannot distinguish code from prose. Observed:

```
$ printf '// drives deploy.CleanMoaiManagedPaths\n' | grep -cE 'runUpdate|RunUpdate|CleanMoaiManagedPaths|BackupMoaiConfig|runCleanReinstall|buildPreserveInventory|deploy\.|backup\.'
1
```

A replacement guard that deletes the fake and leaves behind one comment naming the production
function would therefore have passed (a), (b), and (c) while executing zero lines of production
code — the exact defect §A Defect 4 exists to remove, relocated from prose into a grep. Coverage is
the only form of (b) that cannot be satisfied by text: `> 0.0%` on a named function is evidence the
test actually entered it.

**`-coverpkg` is required, not optional.** `CleanMoaiManagedPaths` and `BackupMoaiConfig` live in
`./internal/cli/update/deploy/` and `./internal/cli/update/backup/`, which are **not** the package
under test. Without `-coverpkg` they are absent from the profile entirely and the assertion silently
narrows to the one same-package function. Observed on this tree:

```
$ go test -run 'TestMoaiUpdate_PreservesUserArea' -covermode=set -coverprofile=/tmp/nopkg.out ./internal/cli/
$ go tool cover -func=/tmp/nopkg.out | grep -E 'CleanMoaiManagedPaths|runCleanReinstall|BackupMoaiConfig'
.../internal/cli/update_clean_install.go:137:  runCleanReinstall   0.0%        # only 1 of 3 visible
```

(a)'s `^func` anchor is retained: it closes the rename escape by asserting the fake's *definition*
is deleted rather than the identifier merely absent — the defect this AC removes is the fake itself
(REQ-UDS-018), not its name.

**Fails when:** the fake is renamed rather than deleted (a), the replacement still executes no
production code — including the case where it only *mentions* one in a comment (b) — or the guard
does not pass (c).

Baseline, measured on this tree:

```
$ grep -cE '^func simulateMoaiUpdate|^func simulate[A-Za-z]*Update' internal/cli/update_safety_test.go
1
$ grep -c 'simulateMoaiUpdate' internal/cli/update_safety_test.go        # unanchored, for context
4
$ go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -covermode=set \
    -coverpkg=./internal/cli/...,./internal/cli/update/... \
    -coverprofile=/tmp/uds-m5.out ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1.038s	coverage: 5.4% of statements in ./internal/cli/..., ./internal/cli/update/...
$ go tool cover -func=/tmp/uds-m5.out | grep -E 'CleanMoaiManagedPaths|MigrateLegacyMemoryDir|runCleanReinstall|BackupMoaiConfig'
github.com/modu-ai/moai-adk/internal/cli/update/backup/backup.go:27:		BackupMoaiConfig			0.0%
github.com/modu-ai/moai-adk/internal/cli/update/deploy/deploy.go:28:		CleanMoaiManagedPaths			0.0%
github.com/modu-ai/moai-adk/internal/cli/update/deploy/deploy.go:143:		MigrateLegacyMemoryDir			0.0%
github.com/modu-ai/moai-adk/internal/cli/update_clean_install.go:137:		runCleanReinstall			0.0%
$ go tool cover -func=/tmp/uds-m5.out | grep -E 'CleanMoaiManagedPaths|MigrateLegacyMemoryDir|runCleanReinstall|BackupMoaiConfig' | grep -vE '[[:space:]]0\.0%$' > /tmp/uds-covered.txt
$ test -s /tmp/uds-covered.txt || echo "VACUOUS: guard executed no registry production function"
VACUOUS: guard executed no registry production function
$ go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/
=== RUN   TestMoaiUpdate_PreservesUserArea
--- PASS: TestMoaiUpdate_PreservesUserArea (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.710s
```

Note the two greps differ: (a)'s `^func` anchor counts the **definition** (`1`), while the
unanchored form counts every mention (`4` — one definition plus three call sites). (a)'s
expectation of `0` therefore means the function is deleted, which the unanchored form could also
reach by renaming.

(b)'s all-`0.0%` profile against the final `PASS` is precisely the defect: the test passes today
without executing a single line of production code, and the coverage profile says so mechanically.
The `ok … 0.710s` duration and the aggregate `5.4%` figure vary per run and are not asserted; the
assertion is that at least one of the four named functions reports coverage above `0.0%`.

**All four named symbols appear in the profile — the transcript is complete.** v0.3.0 recorded only
three lines, silently dropping `MigrateLegacyMemoryDir 0.0%`. Since §A.4 binds this SPEC to verbatim
observed baselines, that lossy transcription is corrected above: the four-line block is the verbatim
output of the command on HEAD `89b2e4772`. The verdict is unchanged (all four are `0.0%`), but the
four-line form is also the evidence that `-coverpkg` genuinely reaches `internal/cli/update/deploy`
and `internal/cli/update/backup` — packages that are **not** the package under test.

The (b) command discriminates in both directions — verified by a positive control on a test that
does execute one of these functions:

```
$ go test -run 'TestCleanMoaiManagedPaths' -count=1 -covermode=set \
    -coverpkg=./internal/cli/...,./internal/cli/update/... \
    -coverprofile=/tmp/pos.out ./internal/cli/update/deploy/
$ go tool cover -func=/tmp/pos.out | grep CleanMoaiManagedPaths | grep -vE '[[:space:]]0\.0%$'
.../update/deploy/deploy.go:28:      CleanMoaiManagedPaths   94.4%
```

Zero surviving lines on the current tautological guard; one surviving line at 94.4% on a guard that
genuinely enters the function. The `test -s` gate therefore fails today and passes after M5 only if
production code actually runs.

#### AC-UDS-015 — the replacement guard fails when production code touches a user-owned path (REQ-UDS-015, REQ-UDS-017)

Runnable falsification via `go test -overlay` — see §C.2. Expected: a `--- FAIL:
TestMoaiUpdate_PreservesUserArea` line when the overlay makes the production entry point write into
`.claude/agents/harness/`.

Baseline: the same mutation against the current test passes, because `simulateMoaiUpdate` never
calls production code.

### M6 — Partial-restore reporting and coverage

#### AC-UDS-016 — the three failure branches are covered and distinguishable (REQ-UDS-027, REQ-UDS-028)

```bash
go test -run 'TestMergeBackPreserveInventory_PartialRestore' -count=1 -v ./internal/cli/
go test -covermode=set -coverprofile=/tmp/uds-cov.out ./internal/cli/ && \
  go tool cover -func=/tmp/uds-cov.out | grep mergeBackPreserveInventory
```

Expected: `--- PASS: TestMergeBackPreserveInventory_PartialRestore` subtests for the stat,
`MkdirAll`, and `copyFile` branches, each asserting a distinct error substring; and a coverage line
for `mergeBackPreserveInventory` strictly greater than the recorded baseline.

Baseline, re-measured on HEAD `89b2e4772`:

```
$ go test -covermode=set -coverprofile=/tmp/uds-cov.out ./internal/cli/ && \
    go tool cover -func=/tmp/uds-cov.out | grep mergeBackPreserveInventory
github.com/modu-ai/moai-adk/internal/cli/update_preserve_inventory.go:400:	mergeBackPreserveInventory		64.3%
```

The uncovered blocks are the failure returns at `update_preserve_inventory.go:416` (stat), `:420`
(`MkdirAll`), and `:424` (`copyFile`). The coverage figure (`64.3%`) is unchanged from v0.3.0, but
the function's coordinate moved `:330` → `:400` and its three failure returns moved
`:346`/`:350`/`:354` → `:416`/`:420`/`:424` when E1 landed — this AC's grep is symbol-anchored
(`grep mergeBackPreserveInventory`), so it survived the drift, but the prose coordinates did not and
are corrected here.

#### AC-UDS-017 — a partial restore names where it stopped (REQ-UDS-026)

```bash
go test -run 'TestMergeBackPreserveInventory_PartialRestore/reports_restored_count' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestMergeBackPreserveInventory_PartialRestore/reports_restored_count` line,
asserting the error text carries both the failing file's relative path and the count of files
already restored.

Baseline: the current errors at `update_preserve_inventory.go:416`/`:420`/`:424` name the file but
not the restored count, so the boundary of a partial restore is unreported.

### Cross-cutting

#### AC-UDS-018 — cross-platform build (NFR-UDS-006)

```bash
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

Expected: both exit 0.

Baseline: both exit 0 on this tree.

#### AC-UDS-019 — template neutrality (NFR-UDS-003)

```bash
# (a) committed modifications since the code baseline
git diff --name-only 8cc108ddb..HEAD -- internal/template/templates/ | wc -l
# (b) uncommitted modifications in the working tree
git status --porcelain internal/template/templates/ | wc -l
```

Expected: both print `0` — this SPEC modifies no template file, whether committed or not.

**Command (a) is the load-bearing one.** The earlier draft ran (b) alone. `git status --porcelain`
reports only the *working tree*: once run-phase modifies a template and commits it, the working
tree is clean again and (b) prints `0`, so an NFR-UDS-003 violation passes undetected. Diffing
against the code baseline catches the committed case, which is the case that actually ships.
`8cc108ddb` is the code baseline pinned in §A.4 and is verified an ancestor of HEAD; every
run-phase commit is a descendant of it. (v0.3.0 pinned `d5336214e`, which is a descendant-of-nothing
here — see §A.4's re-baseline note.)

**Fails when:** any file under `internal/template/templates/` is modified — (a) catches it once
committed, (b) catches it before.

Baseline, re-measured on HEAD `89b2e4772`:

```
$ git diff --name-only 8cc108ddb..HEAD -- internal/template/templates/ | wc -l
       0
$ git status --porcelain internal/template/templates/ | wc -l
       0
```

## §C Falsification procedure

Two mechanisms. Neither uses `git stash` (§A.5).

### C.1 Mutation via `go test -overlay` (deletion radius, backup coverage)

`go test -overlay` substitutes a file's content for the duration of one test run without touching
the working tree — the mechanism that originally proved Defect 3 and Defect 4. It is the preferred
falsification here because it needs no branch, no worktree, and no cleanup.

```bash
# 1. Copy the production file and apply the widening mutation to the copy.
mkdir -p /tmp/uds-falsify
cp internal/cli/update.go /tmp/uds-falsify/update_widened.go
#    edit /tmp/uds-falsify/update_widened.go so globalHooksDir drops the "moai" segment:
#      filepath.Join(homeDir, defs.ClaudeDir, "hooks")

# 2. Declare the overlay.
cat > /tmp/uds-falsify/overlay.json <<'JSON'
{"Replace": {"internal/cli/update.go": "/tmp/uds-falsify/update_widened.go"}}
JSON

# 3. Run the new guard against the mutated source.
go test -overlay=/tmp/uds-falsify/overlay.json \
  -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/

# Expected: --- FAIL: TestEnsureGlobalSettingsEnv_HooksRemovalRadius
# Observed against the CURRENT suite with the same mutation: every test passed.

# 4. Clean up.
rm -rf /tmp/uds-falsify
```

The overlay paths in `Replace` are relative to the module root; the replacement paths are absolute.

### C.2 Mutation for the user-area guard

Identical shape, mutating the production entry point the replacement guard drives so that it writes
into `.claude/agents/harness/`, then:

```bash
go test -overlay=/tmp/uds-falsify/overlay.json \
  -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/
# Expected: --- FAIL: TestMoaiUpdate_PreservesUserArea
```

### C.3 Scratch-worktree replay (behaviour tests)

For AC-UDS-001, AC-UDS-002, AC-UDS-004, AC-UDS-006, AC-UDS-008, and AC-UDS-017, the falsification is
to run the new test files against the pre-fix source in a disposable worktree:

```bash
# 1. Scratch worktree at the commit BEFORE the fix.
git worktree add /tmp/uds-prefix <pre-fix-commit>

# 2. Copy ONLY the new test files across.
cp internal/cli/update_recovery_manifest_test.go   /tmp/uds-prefix/internal/cli/
cp internal/cli/update_restore_test.go             /tmp/uds-prefix/internal/cli/
cp internal/cli/update_ondisk_backup_test.go       /tmp/uds-prefix/internal/cli/
cp internal/cli/update_preserve_partial_test.go    /tmp/uds-prefix/internal/cli/

# 3. Drive with go -C, never cd.
go -C /tmp/uds-prefix test ./internal/cli/ \
  -run 'TestUpdateFailure_WritesRecoveryManifest|TestRestore_|TestBackup_OnDisk|TestMergeBackPreserveInventory_PartialRestore' \
  -count=1 -v

# Expected: at least one --- FAIL line per new test. A run with zero --- FAIL lines means the
# tests do not actually depend on the fix and must be strengthened.

# 4. Dispose.
git worktree remove /tmp/uds-prefix
```

Notes:

- `go -C <dir>` is used rather than `cd`: a `cd` inside a compound Bash command changes the working
  directory only for that invocation and silently reads the wrong tree when the pattern is copied.
- The worktree lives under `/tmp`, so it never collides with a concurrent session's checkout.
- Only new test files are copied — never production files — so the replay isolates the behaviour
  change from incidental drift.

### C.4 Drift-guard falsification (AC-UDS-005)

The registry guard's whole value rests on its enumeration being independent of the registry
(REQ-UDS-007). That independence is only demonstrable by adding a destructive site the registry
does not know about and observing the guard FAIL. Without this procedure a `count == count`
self-comparison would look identical to a working guard.

```bash
# 1. Copy a scanned production file and inject one unregistered destructive site.
mkdir -p /tmp/uds-falsify-c4
cp internal/cli/update_cleanup.go /tmp/uds-falsify-c4/update_cleanup_extra.go
#    append inside an existing function body, or add a new one:
#      func udsFalsifyProbe(p string) error { return os.RemoveAll(p) }

# 2. Declare the overlay (Replace keys are module-root-relative; values absolute).
cat > /tmp/uds-falsify-c4/overlay.json <<'JSON'
{"Replace": {"internal/cli/update_cleanup.go": "/tmp/uds-falsify-c4/update_cleanup_extra.go"}}
JSON

# 3. Run the drift guard against the mutated source.
go test -overlay=/tmp/uds-falsify-c4/overlay.json \
  -run 'TestDestructiveTargetRegistry_CoversAllSites' -count=1 -v ./internal/cli/

# Expected: --- FAIL: TestDestructiveTargetRegistry_CoversAllSites
#           reporting a scanned site (update_cleanup.go / udsFalsifyProbe) with no registry row.
# A PASS here means the guard enumerates from the registry rather than from source, and the
# guard must be rewritten before AC-UDS-005 may be marked PASS.

# 4. Clean up.
rm -rf /tmp/uds-falsify-c4
```

The directory `/tmp/uds-falsify-c4` is distinct from §C.1's `/tmp/uds-falsify` so the two
procedures can be run in either order without one removing the other's overlay.

> **Caveat — scan mechanism.** If the guard scans the on-disk source tree directly (via
> `go/parser` over the repository path) rather than the build's overlaid inputs, `go test -overlay`
> will not be visible to it and step 3 will PASS for the wrong reason. In that case the guard is
> still falsifiable, but the mutation must be applied to a scratch `git worktree` and driven with
> `go -C` per §C.3 instead. Whichever mechanism the implementation uses, an observed `--- FAIL`
> against an injected unregistered site is required before AC-UDS-005 is marked PASS.

## §D Definition of Done

- Every AC in §B has been executed and its verbatim output cited in the milestone's §E matrix.
- Every new guard has an executed falsification from §C showing the required `--- FAIL` line —
  including §C.4 for the drift guard, whose independence is not otherwise observable.
- The §B.0 coverage map holds: no REQ lacks a citing AC, and no AC lacks a REQ citation.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test ./internal/cli/... ./internal/cli/update/... ./internal/defs/` exits 0 with no `--- FAIL`.
- `golangci-lint run --timeout=2m` introduces no new finding relative to the recorded baseline.
- `git status --porcelain internal/template/templates/` is empty.
- The destructive-target registry (§C of `plan.md`) is encoded in code and its drift guard passes.
