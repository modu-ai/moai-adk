# SPEC-UPDATE-DATA-SURVIVAL-001 — Acceptance Criteria

Version: 0.2.0 · Status: draft

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC. This binds every clause of an AC's expectation, not only
   its first: an AC whose headline assertion has no command behind it is defective even when its
   subordinate clauses do (the defect AC-UDS-014 carried at iteration 1).
2. **No vacuous `-run`.** `go test -run <pattern>` exits 0 when the pattern matches zero tests.
   Every AC below that uses `-run` therefore also asserts a literal `--- PASS: <exact test name>`
   line in the output. An AC that would still pass with its test deleted is a defect, and an AC
   whose only assertion is `exit 0` is rejected.
3. **A guard must be able to fail on the fixture it runs on.** A well-formed assertion over a
   quantity that is constant across the change is still vacuous. Two shapes of this bite here and
   are named so they are not re-introduced: (a) a **self-comparison** — a drift guard that both
   enumerates and validates from the same source yields `count == count` and can never fail
   (AC-UDS-005); (b) a **missing before-snapshot** — an invariance claim ("unchanged before and
   after") asserted from a single post-run observation, which passes whatever it observes
   (AC-UDS-013).
4. **Baselines are observed, not assumed.** Every "baseline" line in §B was produced by running the
   stated command on 2026-07-31 against worktree HEAD `a8b42e112` (branch
   `plan/epic-update-config-audit`), whose code baseline is `d5336214e` — every commit between the
   two changes SPEC documents only, so no Go source differs. Where a baseline names a commit, it
   names `a8b42e112`; `d5336214e` is cited only as the code baseline it inherits.
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
| REQ-UDS-006 | AC-UDS-005 | REQ-UDS-020 | AC-UDS-001 (no rollback asserted by absence of tree mutation) |
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

### M1 — Failure contract

#### AC-UDS-001 — a mid-run failure after the first destructive step writes and prints a recovery manifest (REQ-UDS-019, REQ-UDS-020)

```bash
go test -run 'TestUpdateFailure_WritesRecoveryManifest' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestUpdateFailure_WritesRecoveryManifest` line. The test injects a failing
step after `CleanMoaiManagedPaths`, then asserts on the fixture tree that a recovery manifest exists
inside the run-scoped backup directory naming the failed step and the restore command, and that the
same manifest text appears in the captured writer.

Baseline: the test does not exist; `go test -run 'TestUpdateFailure_WritesRecoveryManifest'
./internal/cli/` currently prints `ok … [no tests to run]` and exits 0 — which is exactly the
vacuity this AC's `--- PASS` requirement excludes.

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

Baseline: the gate exists at `internal/cli/update.go:236` and emits
`not a moai project: .moai/config/sections/system.yaml not found in the current directory`.

#### AC-UDS-004 — restore is idempotent and refuses an unrecognised directory (REQ-UDS-023, REQ-UDS-024)

```bash
go test -run 'TestRestore_IdempotentAndRefusesForeignDir' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestRestore_IdempotentAndRefusesForeignDir` line. The test applies the same
backup twice and asserts the resulting tree hashes are equal, then points the entry point at a
directory lacking the backup marker file and asserts a non-nil error and an unmodified tree.

Baseline: no restore entry point exists.

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

**Baseline, measured on this tree.** The scan finds **17 call sites across 10
(file, function) pairs** — materially more than the 7 an earlier draft of this AC assumed, because
the earlier figure counted only `deploy.go` plus two hand-picked sites and omitted the archive,
backup, cleanup, and namespace-protect files entirely. Command and verbatim output:

```
$ grep -rn 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | wc -l
      17
```

The pair count and the site total are independently re-derivable from the table below:

```
$ grep -c '^| [0-9]* | `internal' acceptance.md
10
$ grep '^| [0-9]* | `internal' acceptance.md | sed -E 's/.*\| ([0-9]+) \(.*/\1/' | awk '{s+=$1} END {print s}'
17
```

| # | File | Function | Sites |
|---|---|---|---|
| 1 | `internal/cli/update/deploy/deploy.go` | `CleanMoaiManagedPaths` | 3 (`:83`, `:105`, `:121`) |
| 2 | `internal/cli/update/deploy/deploy.go` | `MigrateLegacyMemoryDir` | 2 (`:169`, `:176`) |
| 3 | `internal/cli/update_archive.go` | `archiveSkill` | 1 (`:92`) |
| 4 | `internal/cli/update_archive.go` | `archiveLegacySkills` | 1 (`:304`) |
| 5 | `internal/cli/update/backup/backup.go` | `BackupMoaiConfig` | 3 (`:107`, `:135`, `:140`) |
| 6 | `internal/cli/update/backup/backup.go` | `CleanupOldBackups` | 1 (`:259`) |
| 7 | `internal/cli/update_clean_install.go` | `runCleanReinstall` | 1 (`:271`) |
| 8 | `internal/cli/update_cleanup.go` | `removeDeprecatedFile` | 1 (`:324`) |
| 9 | `internal/cli/update.go` | `ensureGlobalSettingsEnv` | 1 (`:766`) |
| 10 | `internal/cli/update_namespace_protect.go` | `backupUserOwnedNamespace` | 3 (`:225`, `:233`, `:243`) |

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

**REQ-UDS-002 run-scoped backup directories ARE in this pruning target set** when they are created
under the same backup root with a `YYYYMMDD_HHMMSS` name, because that is precisely the filter
`CleanupOldBackups` applies. M3 must therefore state whether its run-scoped directory adopts that
naming; if it does, a sufficiently old restore point can be pruned by a later run's rotation, and the
interaction between the retention window and the recovery contract is a follow-up review item.

Rows 1-4 and 7-9 are the genuine user-data destructive sites and carry protection assignments per
`plan.md` §C.

No registry exists today, so nothing fails when a new site is added unprotected.

#### AC-UDS-006 — `.moai/memory/` is backed up before the both-exist removal (REQ-UDS-008)

```bash
go test -run 'TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval' -count=1 -v ./internal/cli/update/deploy/
```

Expected: a `--- PASS: TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval` line. The fixture creates
both `.moai/memory/` (with a sentinel file) and `.moai/state/`, runs the migration, and asserts the
sentinel's bytes are present under the backup directory after `.moai/memory/` is gone.

Baseline: `internal/cli/update/deploy/deploy.go:176` calls `os.RemoveAll(legacyDir)` with no
preceding copy; `.moai/memory` appears in none of `preserveInventoryRoots`
(`update_preserve_inventory.go:66-70`), `BackupMoaiConfig` (`backup/backup.go:33-34`, `.moai/config`
only), or `userOwnedScanRoots` (`update_namespace_protect.go:39-43`).

#### AC-UDS-007 — the `.moai/db` group comment names its authorising SPEC, and Category D's accurate prose survives (REQ-UDS-009, NFR-UDS-007)

```bash
# (a) the brand+db group banner now names its authorising SPEC
sed -n '/brand + db directories/,/^\t{/p' internal/defs/dirs.go | grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001'
# (b) Category D's contrast clause is PRESERVED, not deleted
grep -c 'moai/project/db' internal/defs/dirs.go
# (c) the catalogue is otherwise untouched
go test -run 'TestDeprecatedPaths' -count=1 -v ./internal/defs/
```

Expected: (a) prints `≥1`; (b) prints **`2`** — unchanged; (c) a `--- PASS: TestDeprecatedPaths`
line with the entry-count assertions intact.

**Fails when:** the group banner is left bare (a → `0`), or the Category D contrast clause is
deleted (b → `0`), or the slice body changes (c fails on the count assertion).

Baseline, measured on this tree:

```
$ sed -n '/brand + db directories/,/^\t{/p' internal/defs/dirs.go | grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001'
0
$ grep -c 'moai/project/db' internal/defs/dirs.go
2
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
wc -l` → `0`. The only backup of that file is the `[]byte` at `update_template_sync.go:384-390`.

#### AC-UDS-009 — both paths use the same on-disk backup mechanism (REQ-UDS-005)

```bash
go test -run 'TestBackup_OnDiskCoverageParityAcrossPaths' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestBackup_OnDiskCoverageParityAcrossPaths` line, asserting that the set of
on-disk backed-up paths produced by the normal path equals the set produced by the clean-reinstall
path.

Baseline: the two paths build the in-memory slice independently — `update_template_sync.go:384-390`
versus `update_clean_install.go:312-317` — so parity is currently coincidental rather than asserted.

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

(a) is `0` because `ensureGlobalSettingsEnv` calls the plain `userHomeDir()` at `update.go:756`; the
injectable variable `userHomeDirFn` exists at `glm_tools.go:123` but is not used here. **This AC is
blocked on the seam**: with no injection point and `t.Setenv("HOME", …)` forbidden by NFR-UDS-002,
the guard has no way to redirect HOME, so command (a) is a precondition of (b) rather than a
decoration. `globalHooksDir` likewise appears only at `update.go:764-766`; no test observes it.
Falsification is AC-UDS-012.

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
> via `userHomeDir()` at `update.go:756`, so the indirection already exists." The coordinate is
> right and the conclusion is wrong: `userHomeDir` is a plain function (`homedir.go:14`), not a
> reassignable variable. The reassignable seam is `userHomeDirFn` (`glm_tools.go:123`), which this
> call site does not use. No indirection existed.

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
ok  	github.com/modu-ai/moai-adk/internal/cli	1.537s	coverage: 5.5% of statements in ./internal/cli/..., ./internal/cli/update/...
$ go tool cover -func=/tmp/uds-m5.out | grep -E 'CleanMoaiManagedPaths|MigrateLegacyMemoryDir|runCleanReinstall|BackupMoaiConfig'
.../update/backup/backup.go:27:      BackupMoaiConfig        0.0%
.../update/deploy/deploy.go:28:      CleanMoaiManagedPaths   0.0%
.../update_clean_install.go:137:     runCleanReinstall       0.0%
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
The `ok … 0.710s` duration and the aggregate `5.5%` figure vary per run and are not asserted; the
assertion is that at least one of the four named functions reports coverage above `0.0%`.

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

Baseline, measured on this tree:

```
github.com/modu-ai/moai-adk/internal/cli/update_preserve_inventory.go:330:	mergeBackPreserveInventory		64.3%
```

The uncovered blocks are the failure returns at `update_preserve_inventory.go:346`, `:350`, `:354`.

#### AC-UDS-017 — a partial restore names where it stopped (REQ-UDS-026)

```bash
go test -run 'TestMergeBackPreserveInventory_PartialRestore/reports_restored_count' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestMergeBackPreserveInventory_PartialRestore/reports_restored_count` line,
asserting the error text carries both the failing file's relative path and the count of files
already restored.

Baseline: the current errors at `:346`/`:350`/`:354` name the file but not the restored count, so
the boundary of a partial restore is unreported.

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
git diff --name-only d5336214e..HEAD -- internal/template/templates/ | wc -l
# (b) uncommitted modifications in the working tree
git status --porcelain internal/template/templates/ | wc -l
```

Expected: both print `0` — this SPEC modifies no template file, whether committed or not.

**Command (a) is the load-bearing one.** The earlier draft ran (b) alone. `git status --porcelain`
reports only the *working tree*: once run-phase modifies a template and commits it, the working
tree is clean again and (b) prints `0`, so an NFR-UDS-003 violation passes undetected. Diffing
against the code baseline catches the committed case, which is the case that actually ships.
`d5336214e` is the code baseline pinned in §A.4; every run-phase commit is a descendant of it.

**Fails when:** any file under `internal/template/templates/` is modified — (a) catches it once
committed, (b) catches it before.

Baseline, measured on this tree:

```
$ git diff --name-only d5336214e..HEAD -- internal/template/templates/ | wc -l
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
