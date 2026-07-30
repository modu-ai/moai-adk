# SPEC-UPDATE-REINSTALL-LOOP-002 — Acceptance Criteria

Version: 0.2.0 · Status: draft

## §A Verification discipline

1. **No vacuous `-run`.** `go test -run <pattern>` exits 0 when the pattern matches zero tests. Every AC below that uses `-run` also asserts a literal `--- PASS: <exact test name>` line. An AC that would pass with the test deleted is a defect.
2. **Baselines are observed, not assumed.** Every "current baseline" line in §B was produced by running the stated command on 2026-07-31 against worktree HEAD `5468a4afc` (branch `plan/epic-update-config-audit`), whose code baseline is `d5336214e` — `5468a4afc` is a descendant that changes SPEC documents only, so no Go source differs between the two. Where a baseline names a commit, it names `5468a4afc`; `d5336214e` is cited only as the code baseline it inherits.
5. **A guard must be able to fail on the fixture it runs on.** A fixture where the asserted quantity is constant across the change is vacuous even when the assertion is well formed. This bites specifically on the version matrix: a project carrying `.agency/` yields `IsV2=true` for every non-v3-confirmed input under both the old and the new rule, so any monotonicity claim built on it holds trivially. Monotonicity is therefore asserted on a residue-free fixture (AC-RIL2-003).
3. **Falsification is required per new guard.** §C gives the runnable procedure that makes each new guard FAIL against unfixed code.
4. **`git stash` is prohibited.** The checkout is shared with concurrent sessions; `git stash` is repository-global and `git stash push` without `-u` refuses untracked files. Falsification uses a scratch `git worktree`.

## §B Acceptance criteria

### AC-RIL2-001 — Version matrix classifies as specified (REQ-RIL2-001..008)

```bash
go test ./internal/cli/ -run 'TestProbeVersionSignal_NormalizedMatrix' -v 2>&1 | tail -20
```

Expected: a `--- PASS: TestProbeVersionSignal_NormalizedMatrix` line, and each subtest present as `--- PASS: TestProbeVersionSignal_NormalizedMatrix/<case>`. Required rows and verdicts (nine rows; all three columns asserted per row):

| `moai.version` | `Signal 1` | `IsV2` | `V3VersionConfirmed` |
|---|---|---|---|
| `v3.0.1` | false | false | true |
| `3.0.1` | false | false | true |
| `V3.0.1` | false | false | true |
| `v4.0.0` | false | false | true |
| `4.0.0` | false | false | true |
| `""` | true | true | false |
| `v2.5.0` | true | true | false |
| `2.5.0` | false | true | false |
| `V2.5.0` | false | true | false |

Fixture: each row builds a `t.TempDir()` project carrying `.agency/` so Signals 2 and 3 are positive; only the version string varies. This is what makes the override observable.

The `Signal 1` column is mandatory. On this fixture `IsV2` is `true` for all six non-v3-confirmed rows both before and after the change, so an `IsV2`-only table cannot detect a Signal-1 regression. The `2.5.0` and `V2.5.0` rows exist to pin Signal 1 **false** for the bare form per REQ-RIL2-003 / REQ-RIL2-006 — an implementation that keys on major alone flips those two cells to `true` and fails here.

**Fails when:** any implementation that (a) keeps the literal `"v3."` prefix test — four rows' `V3VersionConfirmed` stay false; or (b) makes bare `2.x` Signal-1 positive — the `2.5.0` / `V2.5.0` Signal-1 cells flip.

**Current baseline (unfixed, observed at `5468a4afc`):** the test does not exist. A probe over the same fixture yields `IsV2=true, V3VersionConfirmed=false` for `3.0.1`, `V3.0.1`, `v4.0.0`, and `4.0.0` — four of the nine rows contradict the table — while `2.5.0` and `V2.5.0` already report `Signal 1=false`, matching the target and confirming those two rows encode "do not change this", not "fix this".

### AC-RIL2-002 — Prerelease and build metadata classify by major (REQ-RIL2-005)

```bash
go test ./internal/cli/ -run 'TestProbeVersionSignal_NormalizedMatrix' -v 2>&1 | grep -E 'rc13|build'
```

Expected: subtests covering `3.0.1-rc13` and `3.0.0+build.5`, both reported PASS, both classifying v3-confirmed.

### AC-RIL2-003 — Classification never widens the destructive path (NFR-RIL2-001)

```bash
go test ./internal/cli/ -run 'TestProbeVersionSignal_NoDestructiveWidening' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestProbeVersionSignal_NoDestructiveWidening`.

**Fixture — residue-free, and this is load-bearing.** Each case builds a `t.TempDir()` project containing **only** `.moai/config/sections/system.yaml`: no `.agency/`, no deprecated paths. Signals 2 and 3 are negative, so `IsV2` is a pure function of Signal 1 and the comparison can actually move. Running this assertion on the AC-RIL2-001 residue-carrying fixture would make it vacuous — `IsV2` is `true` there for every non-v3-confirmed input under both rules.

**Input set (all twelve strings required):**

`v3.0.1`, `3.0.1`, `V3.0.1`, `v4.0.0`, `4.0.0`, `""`, `v2.5.0`, **`2.5.0`**, **`V2.5.0`**, `1.9.0`, `abc`, `3`

`2.5.0` and `V2.5.0` are new in v0.2.0 — v0.1.0's input set had neither, which is why this AC could not fail. Together with `abc` they are the only three inputs capable of observing a widening:

| Input | `IsV2` today (residue-free, observed) | Widened by |
|---|---|---|
| `2.5.0` | false | a `major == 2` rule that ignores the `v`/`V` prefix |
| `V2.5.0` | false | the same |
| `abc` | false | reading REQ-RIL2-004's "unparseable" as "major digits unparseable" rather than "file unparseable" |

**Assertion.** The pre-change rule is reproduced inside the test as an explicit local reference implementation (a verbatim copy of the four-branch `switch` in `probeVersionSignal`), and for each input the test asserts: `reference_IsV2 == false` implies `new_IsV2 == false`. The comparison is against a real second implementation, not a restatement of the new rule's own output.

**Fails when:** the implementation makes bare `2.5.0` / `V2.5.0` Signal-1 positive, or makes a well-formed-but-unrecognized string (`abc`) Signal-1 positive. Either flips a residue-free `IsV2` from `false` to `true` and trips the implication. This is the AC's whole purpose — v0.1.0's input set could not fail on any implementation of REQ-RIL2-003.

**Current baseline (unfixed, observed at `5468a4afc`):** the test does not exist. A probe over the residue-free fixture reports `IsV2=false` for `3.0.1`, `V3.0.1`, `v4.0.0`, `4.0.0`, `2.5.0`, `V2.5.0`, `1.9.0`, `abc`, and `3`, and `IsV2=true` for `""` and `v2.5.0`. The post-change rule must reproduce every `false` in that list.

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
go test ./internal/cli/ -run 'TestUpdateDryRun_PreservesExistingOutput' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestUpdateDryRun_PreservesExistingOutput`.

The test captures the `--dry-run` output and asserts **both** of the following literal substrings are present:

| Source | Required literal substring |
|---|---|
| `dryRunArchiveLegacySkills` (`internal/cli/update_archive.go:351`) | `[dry-run] total:` |
| `emitWorktreeAdvisory` (`internal/cli/worktree_advisory.go:41-45`, default `auto_create: false` wording) | `use a worktree for isolation` |

Both literals were read from the current sources at `5468a4afc`. The archive line's format string is `[dry-run] total: %d skills archived, 0 user customizations modified`; asserting the stable `[dry-run] total:` prefix avoids coupling the AC to the skill count, which varies with the fixture. That line is emitted through `tui.Pill(...)`, so the implementer must confirm at test-authoring time that the label survives rendering into the captured writer verbatim — if the pill decorates *within* the label, the assertion narrows to the longest contiguous literal that does survive, and this AC is updated to name it. The advisory has two phrasings selected by `workflow.worktree.auto_create`; the fixture must leave that key at its default so the `use a worktree for isolation` branch is the one taken. Should the implementation need to change either literal, the AC is updated in the same commit — the point is that the AC names a fixed string, not a tree-derived quantity.

**Fails when:** the M4 reordering displaces `emitWorktreeAdvisory` or `dryRunArchiveLegacySkills` from the `--dry-run` path — the corresponding substring disappears from the captured output. This is REQ-RIL2-027's actual subject.

**Why this replaces the v0.1.0 form.** The prior command was `go test -run 'TestUpdateDryRun' -v | grep -c -- '--- PASS'` with expected value "a count matching the number of `TestUpdateDryRun*` tests in the package". That expectation is defined by the tree it is measured against, so it is satisfied by every possible value — and it was observed to pass at `5468a4afc` with a count of `0`, because no `TestUpdateDryRun*` test exists (`grep -rn 'func TestUpdateDryRun' internal/cli/*_test.go` returns nothing). It also never referenced the archive summary or the advisory, so REQ-RIL2-027 was effectively unverified.

**Current baseline (unfixed, observed at `5468a4afc`):** `grep -rn 'func TestUpdateDryRun_PreservesExistingOutput' internal/cli/` returns nothing — the test does not exist, so the AC fails today.

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

> Note on (b): the working-tree check is environment-sensitive — on a shared checkout an unrelated concurrent edit also produces a non-zero count. Treat a non-zero result as a prompt to inspect the listed paths, and read (b) as satisfied when no listed path was created by the test run.

### AC-RIL2-019 — Agency-migration pre-step still fires alongside residue cleanup (REQ-RIL2-021)

```bash
go test ./internal/cli/ -run 'TestResidueCleanup_PreservesAgencyMigrationPreStep' -v 2>&1 | tail -5
```

Expected: `--- PASS: TestResidueCleanup_PreservesAgencyMigrationPreStep`. The test builds a v3-confirmed project (`3.0.1`) that carries **both** an `.agency/` directory and a deprecated path, runs the update path, and asserts:

- the `.agency/` migration pre-step fired — its archive destination exists on disk;
- the deprecated path was removed by the residue sweep;
- both happened in the same run.

**Why this AC exists.** REQ-RIL2-021 had zero AC coverage in v0.1.0 (`grep -c 'REQ-RIL2-021' acceptance.md` → `0`). The requirement is a *preservation* constraint on M3, and M3's plan.md instruction is to **extend** the `fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd)` block at `update.go:406-412` — the block that currently invokes `runAgencyMigrationAdapter` at `:408`. An implementation that replaces the block's body rather than extending it silently drops the migration, and nothing else in the suite would notice.

**Fails when:** M3 replaces rather than extends the `:406-412` block, or gates the migration behind the residue-cleanup outcome, or reorders so the residue sweep short-circuits before `runAgencyMigrationAdapter` is reached.

**Current baseline (unfixed, observed at `5468a4afc`):** the test does not exist. The migration pre-step itself does fire today — `runAgencyMigrationAdapter` is called at `internal/cli/update.go:408` inside the `:406-412` block — so the AC's second assertion (deprecated path removed) is the half that fails before M3 lands.

### AC-RIL2-020 — Go conventions on the files this SPEC touches (NFR-RIL2-004)

The check is scoped to **the files this change set touches**, resolved from the diff rather than from a hardcoded list, so it is not polluted by pre-existing violations elsewhere in the package:

```bash
B=<pre-fix-commit>   # SHA recorded in progress.md §E.2
git diff --name-only "$B"...HEAD -- 'internal/cli/*.go' > /tmp/ril2-changed.txt

xargs -r gofmt -l < /tmp/ril2-changed.txt                                  # (1)
grep -vcE '/[a-z0-9]+(_[a-z0-9]+)*\.go$' /tmp/ril2-changed.txt             # (2)
git diff --unified=0 "$B"...HEAD -- internal/cli/ \
  | grep -E '^\+.*fmt\.Errorf\(' | grep -vc '%w'                           # (3)
```

Expected: (1) prints **nothing** — every file this SPEC adds or edits is gofmt-clean; (2) is `0` — every such filename is `snake_case.go`; (3) is `0` — no `fmt.Errorf` call added by this change set omits `%w`. English comments and godoc are reviewed at PR time; that clause of NFR-RIL2-004 is not mechanically asserted here.

**Why this AC exists.** NFR-RIL2-004 had zero AC coverage in v0.1.0 (`grep -c 'NFR-RIL2-004' acceptance.md` → `0`). AC-RIL2-016's `go vet` does **not** subsume it: `go vet` checks neither formatting nor filename casing, and its analyzers do not require `%w` in `fmt.Errorf`. Absorbing NFR-004 into AC-016 would have been coverage in name only.

**Why diff-scoped rather than package-wide.** A package-wide form was drafted and rejected after being run: `gofmt -l internal/cli/*_test.go` names five pre-existing unformatted files (`glm_env_parity_test.go`, `launcher_worktree_l2_test.go`, `tuxiu_characterization_test.go`, `update_characterization_test.go`, `update_hygiene_characterization_test.go`), and `ls internal/cli/*.go | grep -cE '<snake_case>'` returns `312` against `313` files because `profile_setup_acceptEdits_test.go` predates the convention. A package-wide AC would fail on the baseline for reasons this SPEC neither caused nor is permitted to fix under its scope exclusions.

**Fails when:** a file this SPEC touches is left unformatted (1 names it), a new file is added with a non-`snake_case` name (2 rises above `0`), or a new error path wraps with `%v` or string concatenation instead of `%w` (3 rises above `0`).

**Current baseline (observed at `5468a4afc`, with `B=d5336214e`):** `git diff --name-only d5336214e...HEAD -- 'internal/cli/*.go' | wc -l` → `0` (this branch changes no Go source), so all three quantities are trivially empty/`0`. The AC is therefore **not** meaningfully exercised until run-phase produces a non-empty diff — which is the point: it measures the change set, and there is no change set yet. Its can-fail property was verified against the package-wide variant above, which does produce non-zero counts on real inputs.

## §C Falsification procedure

Each new guard must be shown to FAIL against unfixed code. Two mechanisms are used.

### C.1 Poisoned-input tests (self-contained, permanent)

AC-RIL2-006 is itself the falsification for AC-RIL2-005: it feeds the guard's comparison a deliberately non-excluded inventory and asserts detection. This mechanism lives in the suite permanently, requires no git manipulation, and follows the precedent of `TestDeprecatedPaths_CollisionGuardDetectsReinsertion` (`internal/cli/deprecated_paths_collision_test.go:68`).

### C.2 Scratch-worktree replay (for behaviour tests)

**Binding `<pre-fix-commit>`.** It is the run-phase base: the commit that is `HEAD` immediately before M1's first implementation commit. Capture it at run-phase entry with `git rev-parse HEAD` and record the SHA in `progress.md` §E.2; every replay below uses that recorded SHA. It is not a free placeholder — a replay against an unrecorded commit proves nothing about this change set.

**Why this runs as two batches.** A Go test binary is linked per package. If any copied `_test.go` references a production symbol that does not exist in the pre-fix tree, the whole package fails to compile and `go test` emits `FAIL <pkg> [build failed]` and **no per-test result lines at all** — the `--- FAIL` lines of the other, perfectly good tests in the same batch are swallowed with it. A single all-five batch therefore cannot produce the output the v0.1.0 procedure predicted. The batches are split so that compile-clean tests are never masked by a compile-broken sibling.

This is not a hypothesis. It was observed on this tree during plan-audit: one probe file that mis-assumed the element type of `defs.DeprecatedPaths` took the entire package down —

```
$ go test ./internal/cli/ -run 'TestZZAuditProbe2' -v
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

— with zero `--- FAIL` / `--- PASS` lines emitted.

#### Setup (both batches)

```bash
git worktree add /tmp/ril2-falsify <pre-fix-commit>   # SHA recorded in progress.md §E.2
```

#### Batch A — M1 + M2 (must compile pre-fix; `--- FAIL` is required)

These tests reference only symbols that already exist at the pre-fix commit — `probeVersionSignal`, `detectV2Fingerprint`, `buildPreserveInventory`, `runCleanReinstall`, `scanDeprecatedPaths`. They compile, run, and fail on behaviour.

```bash
cp internal/cli/v2_detection_matrix_test.go            /tmp/ril2-falsify/internal/cli/
cp internal/cli/preserve_deprecated_exclusion_test.go  /tmp/ril2-falsify/internal/cli/

go -C /tmp/ril2-falsify test ./internal/cli/ \
  -run 'TestProbeVersionSignal_NormalizedMatrix|TestProbeVersionSignal_NoDestructiveWidening|TestCleanReinstall_DeprecatedPathNotResurrected|TestCleanReinstall_BacksUpBeforeDeprecatedRemoval' \
  -v 2>&1 | grep -E '^--- (FAIL|PASS)|build failed'
```

Required outcome: a `--- FAIL` line for `TestProbeVersionSignal_NormalizedMatrix`, `TestCleanReinstall_DeprecatedPathNotResurrected`, and `TestCleanReinstall_BacksUpBeforeDeprecatedRemoval`.

`TestProbeVersionSignal_NoDestructiveWidening` (AC-RIL2-003) is the exception in this batch and **must report `--- PASS`** against pre-fix sources. It asserts an implication about the *reference* rule versus the *new* rule; against the pre-fix tree both sides are the same rule, so the implication holds trivially. Its falsification is C.3 below, not this replay. A `--- FAIL` here would mean the reference implementation copied into the test does not match the pre-fix `switch` and must be corrected.

A `build failed` line in Batch A is itself a defect: it means a Batch-A test referenced a not-yet-existing symbol and belongs in Batch B.

#### Batch B — M3 + M4 (build failure is the expected falsification)

M3 introduces the residue-cleanup path and M4 changes the dry-run ordering. Two outcomes are admissible, and the run must record which one occurred:

- **B-i — the test references a new production symbol.** The package does not compile. Required evidence: `go -C /tmp/ril2-falsify build ./internal/cli/` exits non-zero **and** its stderr names the undefined symbol the new test depends on. This is a valid falsification: the guard cannot even link against unfixed code, which is a stronger discrimination than a behavioural failure.
- **B-ii — the test drives only pre-existing entry points** (for example the cobra command via `runUpdate`). The package compiles and each named test reports `--- FAIL`.

```bash
cp internal/cli/residue_cleanup_test.go      /tmp/ril2-falsify/internal/cli/
cp internal/cli/update_dryrun_reach_test.go  /tmp/ril2-falsify/internal/cli/

# B-i evidence (record exit code and stderr verbatim):
go -C /tmp/ril2-falsify build ./internal/cli/ 2>&1; echo "build_exit=$?"

# B-ii evidence (only meaningful when the build above exits 0):
go -C /tmp/ril2-falsify test ./internal/cli/ \
  -run 'TestResidueCleanup_IdempotentAcrossTwoRuns|TestUpdateDryRun_EmitsCleanReinstallPlan|TestUpdateDryRun_PreservesExistingOutput' \
  -v 2>&1 | grep -E '^--- (FAIL|PASS)|build failed'
```

**Inadmissible in either batch:** a `--- PASS` for any named test other than `TestProbeVersionSignal_NoDestructiveWidening`. A pass against pre-fix sources means the test does not discriminate the fix and must be rewritten. Equally inadmissible is a bare `[build failed]` reported as success without the accompanying `go build` stderr naming the undefined symbol — "it did not compile" is only evidence when the reason is shown.

#### Dispose

```bash
git worktree remove /tmp/ril2-falsify
```

Notes on why this form:

- `git stash` is prohibited (§A.4). `git stash push` without `-u` refuses untracked files — the new test files are untracked at this point — and with `-u` it is repository-global and can absorb a concurrent session's work.
- `go -C <dir>` is used rather than `cd`, because a `cd` inside a compound Bash command changes the working directory only for that invocation and silently reads the wrong tree when the pattern is copied.
- Copying only `_test.go` files keeps the production sources at their pre-fix state, which is what makes the FAIL meaningful.

### C.3 Reference-rule mutation (falsification for AC-RIL2-003)

AC-RIL2-003 cannot be falsified by the C.2 replay — see the Batch A note. Its falsification is a bounded, reverted mutation of the *new* rule:

```bash
# 1. In probeVersionSignal, temporarily admit the bare form:
#    make a major-2 string Signal-1 positive WITHOUT requiring the v/V prefix.
# 2. Run:
go test ./internal/cli/ -run 'TestProbeVersionSignal_NoDestructiveWidening' -v 2>&1 | tail -5
# 3. Revert the mutation.
```

Required outcome at step 2: `--- FAIL: TestProbeVersionSignal_NoDestructiveWidening`, with the failure message naming `2.5.0` or `V2.5.0` as an input that moved from `IsV2=false` to `IsV2=true`. A `--- PASS` means the fixture still carries residue (so `IsV2` cannot move) or the input set lost the bare-form rows — both of which reproduce the exact defect this AC was rewritten to close.

Repeat with a second mutation that makes a well-formed unrecognized string Signal-1 positive; required outcome is `--- FAIL` naming `abc`.

## §D Definition of Done

- All AC-RIL2-001 … AC-RIL2-020 pass with their stated observable output.
- §C.2 Batch A produces `--- FAIL` for its three behaviour tests (and `--- PASS` for `TestProbeVersionSignal_NoDestructiveWidening`, per the note there) against pre-fix sources.
- §C.2 Batch B records which outcome occurred — B-i (build failure with the undefined symbol named in stderr) or B-ii (`--- FAIL` per named test) — with the evidence quoted verbatim.
- §C.3 produces `--- FAIL: TestProbeVersionSignal_NoDestructiveWidening` under both named mutations, each reverted afterwards.
- The `<pre-fix-commit>` SHA is recorded in `progress.md` §E.2 and is the SHA used by every §C replay and by AC-RIL2-020.
- `moai spec lint` reports no findings for this SPEC. **Observed at plan-phase (2026-07-31, HEAD `5468a4afc`):** `0 error(s), 62 warning(s)` catalog-wide, with zero findings naming `SPEC-UPDATE-REINSTALL-LOOP-002` (`moai spec lint | grep -c 'SPEC-UPDATE-REINSTALL-LOOP-002'` → `0`). The 62 warnings all belong to other, grandfathered SPECs.
- No file under `internal/template/templates/**` modified.
- `--dry-run` performs no filesystem mutation on any path (AC-RIL2-014), and the M4 implementation did not relocate the dry-run early return past `update.go:306-326` (plan.md §E M4, rejected option).
