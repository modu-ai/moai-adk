# SPEC-UPDATE-REINSTALL-LOOP-002 — Acceptance Criteria

Version: 0.1.0 · Status: draft

## §A Verification discipline

1. **No vacuous `-run`.** `go test -run <pattern>` exits 0 when the pattern matches zero tests. Every AC below that uses `-run` also asserts a literal `--- PASS: <exact test name>` line. An AC that would pass with the test deleted is a defect.
2. **Baselines are observed, not assumed.** Every "current baseline" line in §B was produced by running the stated command against the tree at `d5336214e` on 2026-07-31.
3. **Falsification is required per new guard.** §C gives the runnable procedure that makes each new guard FAIL against unfixed code.
4. **`git stash` is prohibited.** The checkout is shared with concurrent sessions; `git stash` is repository-global and `git stash push` without `-u` refuses untracked files. Falsification uses a scratch `git worktree`.

## §B Acceptance criteria

### AC-RIL2-001 — Version matrix classifies as specified (REQ-RIL2-001..008)

```bash
go test ./internal/cli/ -run 'TestProbeVersionSignal_NormalizedMatrix' -v 2>&1 | tail -20
```

Expected: a `--- PASS: TestProbeVersionSignal_NormalizedMatrix` line, and each subtest present as `--- PASS: TestProbeVersionSignal_NormalizedMatrix/<case>`. Required rows and verdicts:

| `moai.version` | `IsV2` | `V3VersionConfirmed` |
|---|---|---|
| `v3.0.1` | false | true |
| `3.0.1` | false | true |
| `V3.0.1` | false | true |
| `v4.0.0` | false | true |
| `4.0.0` | false | true |
| `""` | true | false |
| `v2.5.0` | true | false |

Fixture: each row builds a `t.TempDir()` project carrying `.agency/` so Signals 2 and 3 are positive; only the version string varies. This is what makes the override observable.

**Current baseline (unfixed, observed):** the test does not exist. The same fixture run through `detectV2Fingerprint` today yields `IsV2=true, V3Confirmed=false` for `3.0.1`, `V3.0.1`, `v4.0.0`, and `4.0.0` — four of the seven rows contradict the table.

### AC-RIL2-002 — Prerelease and build metadata classify by major (REQ-RIL2-005)

```bash
go test ./internal/cli/ -run 'TestProbeVersionSignal_NormalizedMatrix' -v 2>&1 | grep -E 'rc13|build'
```

Expected: subtests covering `3.0.1-rc13` and `3.0.0+build.5`, both reported PASS, both classifying v3-confirmed.

### AC-RIL2-003 — Classification never widens the destructive path (NFR-RIL2-001)

```bash
go test ./internal/cli/ -run 'TestProbeVersionSignal_NoDestructiveWidening' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestProbeVersionSignal_NoDestructiveWidening`. The test asserts that for every string in the matrix plus `1.9.0`, `abc`, and `3`, no input that yields `IsV2=false` under the pre-change literal-prefix rule yields `IsV2=true` under the new rule. The pre-change rule is reproduced inside the test as an explicit local reference implementation, so the assertion is a real comparison rather than a restatement.

### AC-RIL2-004 — Signal-3 isolation fixture survives the widened override (REQ-RIL2-009)

```bash
go test ./internal/cli/ -run 'TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop' -v 2>&1 | tail -5
grep -n 'version:' internal/cli/deprecated_paths_collision_test.go
```

Expected: `--- PASS: TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop`, and the fixture version string in `makeV3ProjectWithDesignDir` is one whose normalized major is neither 2 nor ≥3, so `V3VersionConfirmed` is false and the test still isolates Signal 3.

**Current baseline (observed):** `--- PASS: TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop (0.00s)`. The fixture uses `3.0.0-rc13`, which becomes v3-confirmed under D1 — the test would keep passing while silently losing its isolation. This AC exists precisely because the pass alone is not evidence.

### AC-RIL2-005 — Deprecated paths absent from the built PRESERVE inventory (REQ-RIL2-010, REQ-RIL2-012, REQ-RIL2-014)

```bash
go test ./internal/cli/ -run 'TestDeprecatedPaths_NoPreserveInventoryCollision' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestDeprecatedPaths_NoPreserveInventoryCollision`. The test seeds a `t.TempDir()` project containing all 9 intersecting paths, calls the production `buildPreserveInventory`, and asserts no returned entry equals or is nested under any `defs.DeprecatedPaths` entry.

The test must **not** assert emptiness of `defs.DeprecatedPaths` ∩ `preserveInventoryRoots`; that intersection is 9 by design.

**Current baseline (observed):** the test does not exist. A probe calling `buildPreserveInventory` on a tree seeded with `.claude/commands/agency/brief.md` returned an inventory containing that path — the assertion would fail today.

### AC-RIL2-006 — Guard detects a reintroduced un-excluded path (REQ-RIL2-013)

```bash
go test ./internal/cli/ -run 'TestPreserveInventory_GuardDetectsUnexcludedPath' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestPreserveInventory_GuardDetectsUnexcludedPath`. Modelled on the existing `TestDeprecatedPaths_CollisionGuardDetectsReinsertion`: the test feeds the guard's comparison an inventory that deliberately retains a deprecated path and asserts the violation is reported. This proves AC-RIL2-005 is a real check.

### AC-RIL2-007 — No resurrection across a clean-reinstall cycle (REQ-RIL2-011)

```bash
go test ./internal/cli/ -run 'TestCleanReinstall_DeprecatedPathNotResurrected' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestCleanReinstall_DeprecatedPathNotResurrected`. The test seeds `.claude/commands/agency/brief.md` plus `.agency/` (to force the clean path), runs `runCleanReinstall` with the stub deployer, then asserts:

- `brief.md` is absent from disk after the cycle;
- `brief.md` appears in `result.RemovedPaths`;
- a post-run `scanDeprecatedPaths` does not report `brief.md`.

The third assertion is the loop-termination proof and is the one the current code fails.

**Current baseline (observed):** on a synthetic tree, `scanDeprecatedPaths(before)` returned `[.claude/commands/agency/brief.md .agency]` and the path was present in the PRESERVE inventory, so Step 6 restores it and the post-run scan still reports it.

### AC-RIL2-008 — Deprecated paths are backed up before deletion (REQ-RIL2-015, REQ-RIL2-016)

```bash
go test ./internal/cli/ -run 'TestCleanReinstall_BacksUpBeforeDeprecatedRemoval' -v 2>&1 | tail -5
grep -c 'backupDeprecatedPaths' internal/cli/update_clean_install.go
```

Expected: `--- PASS: TestCleanReinstall_BacksUpBeforeDeprecatedRemoval`, and the grep count is ≥ 2 (the file header reference plus at least one call site). The test asserts a backup copy of `brief.md` exists on disk after the cycle even though the live path is gone.

**Current baseline (observed):** `grep -c backupDeprecatedPaths` over `update_clean_install.go` lines 250-276 returns `0` — Step 4 calls `os.RemoveAll` with no backup. The whole-file count today is 1, and that single occurrence is the header comment at line 12, not a call.

### AC-RIL2-009 — Removal count derived from filesystem diff (REQ-RIL2-017)

```bash
go test ./internal/cli/ -run 'TestCleanReinstall_RemovesOrphanedDBYaml' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestCleanReinstall_RemovesOrphanedDBYaml` — the existing test continues to pass, confirming the diff-based count survives the M2 change.

**Current baseline:** this test exists at `internal/cli/deprecated_paths_collision_test.go:200` and is expected to pass before and after.

### AC-RIL2-010 — v3-confirmed project stays out of the destructive path (REQ-RIL2-018, REQ-RIL2-020)

```bash
go test ./internal/cli/ -run 'TestResidueCleanup_NoDestructiveReinstall' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestResidueCleanup_NoDestructiveReinstall`. The test builds a v3-confirmed project (`3.0.1`) with residue and asserts the stub deployer's force-deploy was **not** invoked and no `.moai/backups/v2-to-v3-*` directory was created, while the residue is nonetheless gone.

### AC-RIL2-011 — Residue is cleaned on a v3-confirmed project (REQ-RIL2-019, REQ-RIL2-023)

```bash
go test ./internal/cli/ -run 'TestResidueCleanup_RemovesDeprecatedOnV3Project' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestResidueCleanup_RemovesDeprecatedOnV3Project`. Asserts the deprecated path is removed, a backup copy exists, and that the same run against a directory lacking `.moai/config/sections/system.yaml` performs no removal (the project-marker gate).

### AC-RIL2-012 — Second consecutive run removes zero (REQ-RIL2-022)

```bash
go test ./internal/cli/ -run 'TestResidueCleanup_IdempotentAcrossTwoRuns' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestResidueCleanup_IdempotentAcrossTwoRuns`. The test runs the residue path twice against one tree and asserts run 1 removes ≥ 1 and run 2 removes exactly 0. This is the direct anti-loop assertion for issue #1243.

### AC-RIL2-013 — `--dry-run` reaches the clean-reinstall plan (REQ-RIL2-024, REQ-RIL2-025)

```bash
go test ./internal/cli/ -run 'TestUpdateDryRun_EmitsCleanReinstallPlan' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestUpdateDryRun_EmitsCleanReinstallPlan`. The captured output contains the literal substrings `DRY-RUN` and `Would remove` with a non-zero count.

**Current baseline (observed):** `internal/cli/update.go:294-304` returns from the dry-run branch before the v2-detection block at `:328`, so `runCleanReinstall`'s dry-run branch at `update_clean_install.go:186-198` is unreachable from the CLI and neither substring can be produced.

### AC-RIL2-014 — `--dry-run` mutates nothing (REQ-RIL2-026)

```bash
go test ./internal/cli/ -run 'TestUpdateDryRun_ZeroMutation' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestUpdateDryRun_ZeroMutation`. The test hashes every file under the `t.TempDir()` project before and after the dry-run invocation and asserts the two maps are equal, including that no `.moai/backups/` directory was created.

### AC-RIL2-015 — Existing dry-run output preserved (REQ-RIL2-027)

```bash
go test ./internal/cli/ -run 'TestUpdateDryRun' -v 2>&1 | grep -c -- '--- PASS'
```

Expected: a count matching the number of `TestUpdateDryRun*` tests in the package, with no `--- FAIL` line. The legacy-skill archive summary and the worktree advisory continue to be emitted.

### AC-RIL2-016 — Build, suite, and cross-platform

```bash
go build ./... && echo "BUILD_OK"
GOOS=windows GOARCH=amd64 go build ./... && echo "WINDOWS_BUILD_OK"
go test ./internal/cli/... 2>&1 | tail -5
go vet ./internal/cli/... && echo "VET_OK"
```

Expected: `BUILD_OK`, `WINDOWS_BUILD_OK`, `VET_OK`, and an `ok github.com/modu-ai/moai-adk/internal/cli` line with no `FAIL`.

### AC-RIL2-017 — No template mutation (NFR-RIL2-003)

```bash
git diff --name-only origin/main...HEAD -- internal/template/templates/ | wc -l
```

Expected: `0`.

### AC-RIL2-018 — Test isolation (NFR-RIL2-002)

```bash
grep -rn 'os.MkdirTemp\|ioutil.TempDir' internal/cli/*_test.go | grep -v 't.TempDir' | wc -l
git status --porcelain | grep -v '^?? .moai/specs/' | wc -l
```

Expected: `0` for the first (all temp dirs go through `t.TempDir()`), and the second run immediately after a full `go test ./internal/cli/...` shows no test-created files in the project root.

## §C Falsification procedure

Each new guard must be shown to FAIL against unfixed code. Two mechanisms are used.

### C.1 Poisoned-input tests (self-contained, permanent)

AC-RIL2-006 is itself the falsification for AC-RIL2-005: it feeds the guard's comparison a deliberately non-excluded inventory and asserts detection. This mechanism lives in the suite permanently, requires no git manipulation, and follows the precedent of `TestDeprecatedPaths_CollisionGuardDetectsReinsertion` (`internal/cli/deprecated_paths_collision_test.go:68`).

### C.2 Scratch-worktree replay (for behaviour tests)

For AC-RIL2-001, AC-RIL2-007, AC-RIL2-008, AC-RIL2-012, and AC-RIL2-013, the falsification is to run the new test file against the pre-fix source:

```bash
# 1. Create a scratch worktree at the commit BEFORE the fix.
git worktree add /tmp/ril2-falsify <pre-fix-commit>

# 2. Copy ONLY the new test files in (production sources stay pre-fix).
cp internal/cli/v2_detection_matrix_test.go            /tmp/ril2-falsify/internal/cli/
cp internal/cli/preserve_deprecated_exclusion_test.go  /tmp/ril2-falsify/internal/cli/
cp internal/cli/residue_cleanup_test.go                /tmp/ril2-falsify/internal/cli/
cp internal/cli/update_dryrun_reach_test.go            /tmp/ril2-falsify/internal/cli/

# 3. Observe FAIL.
go -C /tmp/ril2-falsify test ./internal/cli/ \
  -run 'TestProbeVersionSignal_NormalizedMatrix|TestCleanReinstall_DeprecatedPathNotResurrected|TestCleanReinstall_BacksUpBeforeDeprecatedRemoval|TestResidueCleanup_IdempotentAcrossTwoRuns|TestUpdateDryRun_EmitsCleanReinstallPlan' \
  -v 2>&1 | grep -E '^--- (FAIL|PASS)'

# 4. Dispose.
git worktree remove /tmp/ril2-falsify
```

Expected at step 3: a `--- FAIL` line for **every** one of the five named tests. A `--- PASS` for any of them means that test does not actually discriminate the fix and must be rewritten.

Notes on why this form:

- `git stash` is prohibited (§A.4). `git stash push` without `-u` refuses untracked files — the new test files are untracked at this point — and with `-u` it is repository-global and can absorb a concurrent session's work.
- `go -C <dir>` is used rather than `cd`, because a `cd` inside a compound Bash command changes the working directory only for that invocation and silently reads the wrong tree when the pattern is copied.
- Copying only `_test.go` files keeps the production sources at their pre-fix state, which is what makes the FAIL meaningful.

## §D Definition of Done

- All AC-RIL2-001 … AC-RIL2-018 pass with their stated observable output.
- §C.2 produces `--- FAIL` for all five named tests against pre-fix sources.
- `moai spec lint` reports no findings for this SPEC.
- No file under `internal/template/templates/**` modified.
