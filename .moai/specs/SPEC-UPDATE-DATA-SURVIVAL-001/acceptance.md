# SPEC-UPDATE-DATA-SURVIVAL-001 — Acceptance Criteria

## §A Discipline

1. **Every AC states a command and its expected observable output.** A criterion phrased as a
   property with no command is not an AC.
2. **`go test -run <pattern>` exits 0 on zero matches.** Every `-run` AC therefore also requires a
   verbatim `--- PASS: <exact test name>` line in the output. A `-run` AC whose only assertion is
   `exit 0` is vacuous and is rejected.
3. **Baselines were recorded from this tree while authoring** (HEAD `d5336214e`). Each AC below
   carries its observed pre-change baseline so a reviewer can tell a real change from a no-op.
4. **Every new guard needs a falsification** proving it FAILS against unfixed code. §C gives the
   runnable procedure.
5. **`git stash` is prohibited.** The checkout is shared with concurrent sessions: `git stash push`
   refuses untracked files without `-u`, and with `-u` it is repository-global and can swallow a
   parallel session's work. Falsification uses a scratch `git worktree` driven by `go -C`, or
   `go test -overlay`.
6. **Crash windows are reached deterministically, never by racing a crash** (NFR-UDS-001).

## §B Acceptance criteria

### M1 — Failure contract

#### AC-UDS-001 — a mid-run failure after the first destructive step writes and prints a recovery manifest

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

#### AC-UDS-002 — the restore entry point runs on a tree whose project marker was destroyed

```bash
go test -run 'TestRestore_ProceedsWithoutProjectMarker' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestRestore_ProceedsWithoutProjectMarker` line. The fixture deletes
`.moai/config/sections/system.yaml`, invokes the restore entry point against a backup directory,
and asserts the file is restored — proving the marker-gate lockout is escaped.

Baseline: `go run ./cmd/moai update --help 2>&1 | grep -ci restore` → `0`. No restore surface
exists on the update command today.

#### AC-UDS-003 — the ordinary update path still refuses a marker-less tree

```bash
go test -run 'TestUpdate_RejectsTreeWithoutProjectMarker' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestUpdate_RejectsTreeWithoutProjectMarker` line, asserting the returned
error contains `not a moai project`. This pins REQ-UDS-025: the bypass is scoped to restore alone.

Baseline: the gate exists at `internal/cli/update.go:236` and emits
`not a moai project: .moai/config/sections/system.yaml not found in the current directory`.

#### AC-UDS-004 — restore is idempotent and refuses an unrecognised directory

```bash
go test -run 'TestRestore_IdempotentAndRefusesForeignDir' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestRestore_IdempotentAndRefusesForeignDir` line. The test applies the same
backup twice and asserts the resulting tree hashes are equal, then points the entry point at a
directory lacking the backup marker file and asserts a non-nil error and an unmodified tree.

Baseline: no restore entry point exists.

### M2 — Destructive-target registry

#### AC-UDS-005 — every destructive site in the update subsystem is registered

```bash
go test -run 'TestDestructiveTargetRegistry_CoversAllSites' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestDestructiveTargetRegistry_CoversAllSites` line. The guard asserts that
each registry row carries either a protection-set assignment or a recorded exemption, and that the
registry's row count equals the number of destructive sites the guard enumerates.

Baseline: `grep -c 'RemoveAll\|os.Rename' internal/cli/update/deploy/deploy.go` → `5`
(`deploy.go:83`, `:105`, `:121`, `:169`, `:176`). Plus `update_clean_install.go:271` and
`update.go:766`. No registry exists; nothing fails when a new site is added unprotected.

#### AC-UDS-006 — `.moai/memory/` is backed up before the both-exist removal

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

#### AC-UDS-007 — the Category D comment names the registered path

```bash
grep -c 'moai/project/db' internal/defs/dirs.go
go test -run 'TestDeprecatedPaths' -count=1 ./internal/defs/
```

Expected: the grep prints `0`, and the `defs` suite exits 0 with the entry-count assertions intact.

Baseline: the grep currently prints `2` — the Category D comment (`dirs.go:344-351`) describes
`.moai/project/db/` while the registered directory entry is `.moai/db` (`dirs.go:313`).

### M3 — On-disk backup before destruction

#### AC-UDS-008 — the three in-memory-only files exist on disk before the first destructive step

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

#### AC-UDS-009 — both paths use the same on-disk backup mechanism

```bash
go test -run 'TestBackup_OnDiskCoverageParityAcrossPaths' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestBackup_OnDiskCoverageParityAcrossPaths` line, asserting that the set of
on-disk backed-up paths produced by the normal path equals the set produced by the clean-reinstall
path.

Baseline: the two paths build the in-memory slice independently — `update_template_sync.go:384-390`
versus `update_clean_install.go:312-317` — so parity is currently coincidental rather than asserted.

#### AC-UDS-010 — the success path is unchanged

```bash
go test -count=1 ./internal/cli/... ./internal/cli/update/...
```

Expected: exit 0 with no `--- FAIL` line. The pre-existing update tests are the regression surface
for NFR-UDS-005.

Baseline: the same command exits 0 on this tree.

### M4 — HOME deletion-radius pinning

#### AC-UDS-011 — widening the HOME removal target fails the suite

```bash
go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius` line. The test redirects the
home lookup to a `t.TempDir()`, creates both `~/.claude/hooks/moai/` and a sibling
`~/.claude/hooks/user-hook.sh`, runs the function, and asserts the sibling survives while the
`moai` subdirectory is gone.

Baseline: `grep -rn 'globalHooksDir' internal/cli/ --include='*_test.go' | wc -l` → `0`. The symbol
appears only at `internal/cli/update.go:764-766`; no test observes it. Falsification is
AC-UDS-012.

#### AC-UDS-012 — the radius guard fails against a widened radius

Runnable falsification via `go test -overlay` — see §C.1. Expected: a `--- FAIL:
TestEnsureGlobalSettingsEnv_HooksRemovalRadius` line when the overlay widens the target to
`~/.claude/hooks`.

Baseline: the audit lens ran precisely this mutation against the current suite and observed
`ok github.com/modu-ai/moai-adk/internal/cli 18.305s` — every test passed with the widened radius.

#### AC-UDS-013 — the HOME test does not touch the operator's real home

```bash
go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 ./internal/cli/
ls ~/.claude/hooks 2>/dev/null | head
```

Expected: the test exits 0 and the operator's `~/.claude/hooks` listing is unchanged before and
after. The test redirects through the existing `userHomeDir` indirection
(`internal/cli/update.go:756`) rather than `t.Setenv("HOME", …)`, per `CLAUDE.local.md` §13.

Baseline: `ensureGlobalSettingsEnv` resolves HOME via `userHomeDir()` at `update.go:756`, so the
indirection already exists.

### M5 — Non-vacuous user-area safety guard

#### AC-UDS-014 — the replacement guard drives real production code

```bash
grep -c 'simulateMoaiUpdate' internal/cli/update_safety_test.go
go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/
```

Expected: the grep prints `0`, and the test output carries a `--- PASS:
TestMoaiUpdate_PreservesUserArea` line while the test body invokes a real update entry point.

Baseline: the grep currently prints `4`; a grep of that file for
`runUpdate|RunUpdate|CleanMoaiManagedPaths|BackupMoaiConfig|runCleanReinstall|buildPreserveInventory|deploy\.|backup\.`
returns **no matches**; and the test passes today —

```
=== RUN   TestMoaiUpdate_PreservesUserArea
--- PASS: TestMoaiUpdate_PreservesUserArea (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.542s
```

— which is precisely the problem: it passes without executing any production code.

#### AC-UDS-015 — the replacement guard fails when production code touches a user-owned path

Runnable falsification via `go test -overlay` — see §C.2. Expected: a `--- FAIL:
TestMoaiUpdate_PreservesUserArea` line when the overlay makes the production entry point write into
`.claude/agents/harness/`.

Baseline: the same mutation against the current test passes, because `simulateMoaiUpdate` never
calls production code.

### M6 — Partial-restore reporting and coverage

#### AC-UDS-016 — the three failure branches are covered and distinguishable

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

#### AC-UDS-017 — a partial restore names where it stopped

```bash
go test -run 'TestMergeBackPreserveInventory_PartialRestore/reports_restored_count' -count=1 -v ./internal/cli/
```

Expected: a `--- PASS: TestMergeBackPreserveInventory_PartialRestore/reports_restored_count` line,
asserting the error text carries both the failing file's relative path and the count of files
already restored.

Baseline: the current errors at `:346`/`:350`/`:354` name the file but not the restored count, so
the boundary of a partial restore is unreported.

### Cross-cutting

#### AC-UDS-018 — cross-platform build

```bash
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

Expected: both exit 0.

Baseline: both exit 0 on this tree.

#### AC-UDS-019 — template neutrality

```bash
git status --porcelain internal/template/templates/ | wc -l
```

Expected: `0` — this SPEC modifies no template file (NFR-UDS-003).

Baseline: `0`.

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

## §D Definition of Done

- Every AC in §B has been executed and its verbatim output cited in the milestone's §E matrix.
- Every new guard has an executed falsification from §C showing the required `--- FAIL` line.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test ./internal/cli/... ./internal/cli/update/... ./internal/defs/` exits 0 with no `--- FAIL`.
- `golangci-lint run --timeout=2m` introduces no new finding relative to the recorded baseline.
- `git status --porcelain internal/template/templates/` is empty.
- The destructive-target registry (§C of `plan.md`) is encoded in code and its drift guard passes.
