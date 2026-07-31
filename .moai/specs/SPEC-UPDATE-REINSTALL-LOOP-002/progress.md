# SPEC-UPDATE-REINSTALL-LOOP-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

### Baseline

- **Code baseline** `d5336214e`; **worktree HEAD** `5468a4afc` (branch `plan/epic-update-config-audit`), a descendant of `d5336214e` that changes SPEC documents only.
- `git merge-base --is-ancestor d5336214e HEAD` → exit `0`. `git diff --name-only d5336214e...HEAD -- 'internal/cli/*.go' | wc -l` → `0`, confirming no Go source differs between the two.
- Every artifact that names a baseline names `5468a4afc`, citing `d5336214e` only as the code baseline it inherits.

### v0.1.0 (initial plan-phase authoring)

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- Every claimed `file:line` in the delegation brief was independently re-verified. All matched; see plan.md §C for the verification table.
- One additional defect was found during verification and folded into scope: clean-reinstall Step 4 never calls `backupDeprecatedPaths` (spec.md §A Defect 5, plan.md §B).
- The version matrix and the 9-entry preserve-root intersection were reproduced by an executed probe, not carried over from the brief.
- Defect 3 decision recorded and justified in plan.md §A D2 (narrow the override's consequence; do not defer).

### v0.2.0 (plan-audit revision — D1 through D5)

Plan-audit verdict on v0.1.0: **FAIL, 0.77** against the Tier M threshold 0.80; Must-Pass 7/7 passed; baseline integrity found sound (no unobserved-baseline defects). D1-D5 resolved below; D6 and D7-a deferred to the next audit.

- **D1** — `REQ-RIL2-003` narrowed to `v`/`V`-prefixed major-2 forms, so it no longer contradicts `NFR-RIL2-001`; the constraint takes no exception. `AC-RIL2-003` rebuilt on a **residue-free** fixture with `2.5.0`, `V2.5.0`, and `abc` in its input set. spec.md §A gained the residue-free widening table; §G risk row 2 rewritten.
- **D2** — plan.md §E M4's first option (relocate the dry-run early return) dropped as a confirmed defect and recorded as a rejected alternative with its evidence. `--dry-run` must never mutate; the fix hoists detection above the branch so the already-implemented non-mutating renderer at `update_clean_install.go:186-198` becomes reachable.
- **D3** — plan.md §E M3's `:384-390` corrected to `:406-412`, with the condition token given as the durable anchor.
- **D4** — `AC-RIL2-015` replaced: the tautological package-relative count became a fixed-literal assertion on `[dry-run] total:` and `use a worktree for isolation`, both read from the current sources.
- **D5** — acceptance.md §C.2 split into Batch A (compiles pre-fix, `--- FAIL` required) and Batch B (build failure is a valid falsification, with the undefined symbol named in stderr); `<pre-fix-commit>` bound to the run-phase base recorded in §E.2 below; new §C.3 gives `AC-RIL2-003` its own mutation-based falsification.
- **Coverage gaps closed** — `AC-RIL2-019` binds `REQ-RIL2-021`; `AC-RIL2-020` binds `NFR-RIL2-004`, diff-scoped after a package-wide draft was run and rejected (it fails on five pre-existing unformatted test files and one pre-existing camelCase filename).
- **Additional finding, not in the audit's D-list** — plan.md §A D1 step 4 read "unparseable → Signal 1 positive", ambiguous between "file unparseable" and "major digits unparseable". Under the second reading a residue-free `abc` moves `IsV2` false→true, an `NFR-RIL2-001` violation of the same class as D1. Step 4 and `REQ-RIL2-004` / `REQ-RIL2-006` now state the file/value distinction explicitly.

### Observed at revision time (2026-07-31, HEAD `5468a4afc`)

- `moai spec lint` → `0 error(s), 62 warning(s)`; `moai spec lint | grep -c 'SPEC-UPDATE-REINSTALL-LOOP-002'` → `0`. Zero findings for this SPEC; the 62 warnings all belong to other, grandfathered SPECs. This replaces the v0.1.0 line that asserted the output "recorded at plan-phase handoff" without recording it.
- `grep -n 'fpErr == nil && !fingerprint.IsV2 && isMoAIProject' internal/cli/update.go` → `406:` (D3 evidence).
- `grep -rn 'func TestUpdateDryRun' internal/cli/*_test.go` → no matches; `go test ./internal/cli/ -run 'TestUpdateDryRun' -v | grep -c -- '--- PASS'` → `0` (D4 evidence: the retired AC compared `0` against a tree-derived expectation of `0`).
- Residue-free classification probe (temporary, removed after use): `2.5.0` → `Signal 1=false, IsV2=false`; `V2.5.0` → `Signal 1=false, IsV2=false`; `abc` → `Signal 1=false, IsV2=false`; `v2.5.0` → `Signal 1=true, IsV2=true` (D1 evidence).
- Residue-carrying probe, run against **the v0.1.0 seven-row matrix** (not the revised nine-row table now in acceptance.md §B AC-RIL2-001): `IsV2=true` for all six non-v3-confirmed rows of that seven-row table, confirming the v0.1.0 matrix fixture cannot discriminate a Signal-1 change. Under the pre-change literal `"v3."` rule only `v3.0.1` was v3-confirmed there, so six of seven rows were non-v3-confirmed — **six is the accurate historical count for the v0.1.0 table and is deliberately retained.** The nine-row table has four such rows; that separate figure lives in acceptance.md §B.

### v0.3.0 (plan-audit iter3 revision — D12, D9, D15)

Plan-audit verdict on v0.2.0: **PASS, 0.88** against the Tier M threshold 0.80; Must-Pass 7/7. Three residual defects closed; no code changed (documentation-only revision).

- **D12 (major) — `AC-RIL2-014` command (a) was satisfiable by a comment alone.** The literal-match grep ran over the whole test file, and grep cannot distinguish code from prose, so a fixture whose `permissions.deny` array is empty and whose only occurrence of a `retiredV2DenyEntries` literal sat in a comment satisfied the check. This is the same defect class the sibling `SPEC-UPDATE-DATA-SURVIVAL-001` removed from its own AC-UDS-M5 command (b). Closed on two levels: (i) command (a) now strips leading-comment lines (`grep -v '^[[:space:]]*//' | grep -cE …`), verified to print `0`/exit 1 on a comment-only fixture and `1`/exit 0 on a genuinely-seeded one; (ii) the AC now **requires `TestUpdateDryRun_ZeroMutation` to assert at runtime** that the written `.claude/settings.json` parses and its `permissions.deny` array contains one of the twelve literals. The runtime assertion is the durable fix — the comment-stripping grep alone is still defeated by a *trailing* comment on a code line (observed: prints `1`, exit 0), and that residual gap is stated explicitly in the AC body. The §D Definition of Done clause was rewritten to reference all three checks rather than "both halves".
- **D9 (minor) — the two "six non-v3-confirmed rows" occurrences are NOT the same defect, and were handled differently.**
  - `acceptance.md:37` was genuinely wrong and is fixed. Its referent is the revised **nine-row** AC-RIL2-001 table, whose `V3VersionConfirmed = false` rows are `""`, `v2.5.0`, `2.5.0`, `V2.5.0` — **four**, not six. Four is the number that makes the sentence's own argument true, since exactly those four hold `IsV2 = true` on both sides of the change; the other four v3-confirmed rows (`3.0.1`, `V3.0.1`, `v4.0.0`, `4.0.0`) flip `true → false`.
  - `progress.md:37` (above, in the v0.2.0 observation block) was **correct as written and its number was deliberately NOT changed.** Its referent is explicitly the v0.1.0 **seven-row** matrix, where the pre-change literal `"v3."` rule left only `v3.0.1` v3-confirmed — so six of seven rows were non-v3-confirmed. **The iter3 audit classified this as a second occurrence of the stale figure; that classification is rejected.** Changing it to "four" would have corrupted an accurate historical observation by re-scoping it to a table it never described. The line was instead edited to name its referent (the v0.1.0 seven-row table) unambiguously, so no future reader or auditor mistakes it for the current nine-row table.
- **D15 (minor) — the sibling co-edit constraint is now recorded in this SPEC.** `SPEC-UPDATE-DOC-DRIFT-001/progress.md` §E.1 ("M1 versus E1") carried a constraint this SPEC did not: the two SPECs must not edit the `--dry-run` branch of `internal/cli/update.go` concurrently, and if the sibling lands first this SPEC's M4 becomes a no-op verification rather than an implementation. **Recorded in `plan.md` §E M4** (extending the existing consistency note) rather than here, because it changes *what M4 does* — degrading it from "implement the hoist" to "verify the hoist and record that no change was required" — and is therefore a plan constraint, not merely a sequencing note. The sibling's own claim that it settles `--dry-run` the same way was independently confirmed at `SPEC-UPDATE-DOC-DRIFT-001/spec.md:373` (option B selected) and `:376-378` (early return retained), and those line anchors are now cited in the plan.md note.

## §E.2 Run-phase Evidence

- `pre_fix_commit:` `3f0849239eedbaa3694a0b5b55821d1a618802c0` — captured with `git rev-parse HEAD` at run-phase entry on branch `feat/SPEC-UPDATE-REINSTALL-LOOP-002`, before M1's first implementation commit. This SHA binds acceptance.md §C.2 Batch A / Batch B and AC-RIL2-020.

### M1 — Version-signal normalization (complete)

Implementation commit: `542e9cdcf`. Touched `internal/cli/v2_detection.go`, new `internal/cli/v2_detection_matrix_test.go`, fixture migration in `internal/cli/deprecated_paths_collision_test.go`, plus the `draft → in-progress` frontmatter transition on `spec.md`.

#### RED evidence (pre-GREEN, captured before the implementation edit)

`go test ./internal/cli/ -run 'TestProbeVersionSignal_NormalizedMatrix|TestProbeVersionSignal_NoDestructiveWidening' -v` against the unmodified `probeVersionSignal`:

```
--- FAIL: TestProbeVersionSignal_NormalizedMatrix (0.02s)
    v2_detection_matrix_test.go:101: V3VersionConfirmed = false, want true (version "3.0.1", detail "")
    v2_detection_matrix_test.go:110: IsV2 = true, want false (version "3.0.1"; version=false agency=true deprecated=true v3=false)
    v2_detection_matrix_test.go:101: V3VersionConfirmed = false, want true (version "V3.0.1", detail "")
    v2_detection_matrix_test.go:101: V3VersionConfirmed = false, want true (version "v4.0.0", detail "")
    v2_detection_matrix_test.go:101: V3VersionConfirmed = false, want true (version "4.0.0", detail "")
    v2_detection_matrix_test.go:101: V3VersionConfirmed = false, want true (version "3.0.1-rc13", detail "")
    v2_detection_matrix_test.go:101: V3VersionConfirmed = false, want true (version "3.0.0+build.5", detail "")
--- PASS: TestProbeVersionSignal_NoDestructiveWidening (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.665s
```

Six of eleven matrix rows failed. `TestProbeVersionSignal_NoDestructiveWidening` reporting `--- PASS` here is the documented acceptance.md §C.2 Batch A exception, not a defect: against pre-fix sources the local reference implementation and the production rule are the same rule, so the implication holds trivially. Its falsification is §C.3 (recorded below).

#### AC matrix (M1 scope)

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-RIL2-001 | PASS | `go test ./internal/cli/ -run 'TestProbeVersionSignal_NormalizedMatrix' -v \| tail -20` | `--- PASS: TestProbeVersionSignal_NormalizedMatrix` + all 11 subtests `--- PASS` |
| AC-RIL2-002 | PASS | same, `\| grep -E 'rc13\|build'` | `--- PASS: .../3.0.1-rc13`, `--- PASS: .../3.0.0+build.5` |
| AC-RIL2-003 | PASS | `go test ./internal/cli/ -run 'TestProbeVersionSignal_NoDestructiveWidening' -v \| tail -5` | `--- PASS: TestProbeVersionSignal_NoDestructiveWidening` (12 inputs) |
| AC-RIL2-004 | PASS | `go test ./internal/cli/ -run 'TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop' -v \| tail -5` | `--- PASS: TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop (0.00s)`; `grep -n 'version:' internal/cli/deprecated_paths_collision_test.go` → `112: "moai:\n    version: \"1.9.0-legacy\"\n"` |

Fixture migration rationale (AC-RIL2-004): `1.9.0-legacy` normalizes to major `1` — neither `2` nor `>= 3` — so `probeVersionSignal` takes the REQ-RIL2-006 negative branch (Signal 1 false, `V3VersionConfirmed` false). Signal 3 therefore remains the sole driver of the test's assertions. The previous `3.0.0-rc13` normalizes to major `3` and would have been v3-confirmed, letting the test pass through the override while silently losing its isolation.

#### §C.2 Batch A replay (pre-fix sources)

Scratch worktree at `pre_fix_commit`, `v2_detection_matrix_test.go` copied in, production sources untouched:

```
--- FAIL: TestProbeVersionSignal_NormalizedMatrix (0.03s)
--- PASS: TestProbeVersionSignal_NoDestructiveWidening (0.02s)
```

No `build failed` line — both M1 tests compile against pre-fix sources, which is the Batch A requirement. Worktree disposed via `git worktree remove --force`.

#### §C.3 reference-rule mutation (falsification for AC-RIL2-003)

Both mutations applied to the new rule, run, and reverted:

1. Dropped the prefix requirement (`case major == 2 && prefix == 'v'` → `case major == 2`):

```
v2_detection_matrix_test.go:180: NFR-RIL2-001 violated for "2.5.0": reference rule classified IsV2=false, new rule classifies IsV2=true — the destructive path widened
v2_detection_matrix_test.go:180: NFR-RIL2-001 violated for "V2.5.0": reference rule classified IsV2=false, new rule classifies IsV2=true — the destructive path widened
--- FAIL: TestProbeVersionSignal_NoDestructiveWidening (0.02s)
```

2. Made a well-formed unrecognized string Signal-1 positive (`!parsed` branch returning `true`):

```
v2_detection_matrix_test.go:180: NFR-RIL2-001 violated for "abc": reference rule classified IsV2=false, new rule classifies IsV2=true — the destructive path widened
--- FAIL: TestProbeVersionSignal_NoDestructiveWidening (0.01s)
```

Both reverted; `grep -n 'MUTATION\|case major == 2' internal/cli/v2_detection.go` → single match `225: case major == 2 && prefix == 'v':`.

#### AC-RIL2-020 (D11 follow-up — re-run against a non-empty diff)

`B=3f0849239eedbaa3694a0b5b55821d1a618802c0`, run after commit `542e9cdcf`:

```
=== changed files ===
internal/cli/deprecated_paths_collision_test.go
internal/cli/v2_detection.go
internal/cli/v2_detection_matrix_test.go
=== (1) gofmt -l ===
(exit=0)                 # no output — all three files gofmt-clean
=== (2) non-snake_case count ===
0
=== (3) fmt.Errorf without %w ===
0
```

The AC is now meaningfully exercised: the diff carries 3 files where the plan-phase measurement had 0, so checks (1) and (2) operate on real input rather than an empty list.

#### Build, lint, coverage, suite

| Check | Observed |
|---|---|
| `go build ./...` | `BUILD_OK` (exit 0) |
| `GOOS=windows GOARCH=amd64 go build ./...` | `WINDOWS_BUILD_OK` (exit 0) |
| `go vet ./internal/cli/...` | `VET_OK` (exit 0) |
| `golangci-lint run --timeout=3m` | `0 issues.` — identical to the pre-flight baseline measured before any edit; zero NEW findings |
| `go test -cover ./internal/cli/` | `coverage: 75.7% of statements` post-change; the same command at `pre_fix_commit` in a clean scratch worktree also reported `75.7%` — unchanged at the reported precision |
| `go test ./...` | 0 packages reported `FAIL` |

Coverage note: `internal/cli` sits at 75.7%, below the 90% target for critical packages. This is the pre-existing package baseline (identical before and after M1), not a regression introduced here; raising it is outside M1's scope envelope.

#### Observation for manager-spec (documentation-only, no implementation impact)

REQ-RIL2-003 states the Signal-1-positive condition as major 2 "**and** the original string carried a leading `v` **or** `V`". The AC matrix (acceptance.md AC-RIL2-001) and the NFR-RIL2-001 residue-free widening table (AC-RIL2-003) both pin `V2.5.0` to Signal 1 **false**, and AC-RIL2-003 names "a `major == 2` rule that ignores the `v`/`V` prefix" as a widening. The uppercase form was also Signal-1 negative pre-change, since the literal test was the case-sensitive `strings.HasPrefix(v, "v2.")`.

The implementation follows the AC + NFR (lowercase `v` only), because NFR-RIL2-001 "admits no exception" and accepting `V` would flip a residue-free `V2.5.0` project from `IsV2=false` to `IsV2=true`. The `or V` clause in the REQ-RIL2-003 prose is the outlier and warrants a wording correction in a later scope-doc pass. No implementation change follows from it — the verification surface is unambiguous.

### M2 — PRESERVE-inventory exclusion + backup-before-delete (complete)

Both halves landed in ONE commit, per plan.md §B: removing the Step 6 resurrection without the Step 4 backup would convert a net-zero no-op into unbacked-up deletion of user-authored `.claude/commands/agency/*.md`.

#### RED evidence (captured before the exclusion and the backup were wired)

`go test ./internal/cli/ -run '<the four M2 tests>' -v` with `deprecatedInventoryCollisions` present as an unwired predicate and production behaviour unchanged:

```
    preserve_deprecated_exclusion_test.go:90: built PRESERVE inventory ∩ defs.DeprecatedPaths is non-empty: [.moai/project/brand/tokens.md .claude/commands/agency/agency.md .claude/commands/agency/brief.md .claude/commands/agency/build.md .claude/commands/agency/evolve.md .claude/commands/agency/learn.md .claude/commands/agency/profile.md .claude/commands/agency/resume.md .claude/commands/agency/review.md]
--- FAIL: TestDeprecatedPaths_NoPreserveInventoryCollision (0.01s)
--- PASS: TestPreserveInventory_GuardDetectsUnexcludedPath (0.00s)
    preserve_deprecated_exclusion_test.go:191: brief.md still present on disk after the clean-reinstall cycle (stat err=<nil>); Step 6 restored it from the PRESERVE backup
    preserve_deprecated_exclusion_test.go:206: post-run scanDeprecatedPaths still reports brief.md ([.claude/commands/agency/brief.md]); the deprecated-path v2 signal stays armed and the next `moai update` loops
--- FAIL: TestCleanReinstall_DeprecatedPathNotResurrected (0.02s)
    preserve_deprecated_exclusion_test.go:245: walk backup root ".../.moai/backup": no such file or directory — Step 4 wrote no deprecated-path backup at all
--- FAIL: TestCleanReinstall_BacksUpBeforeDeprecatedRemoval (0.03s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.093s
```

The RED run reproduces the exact 9-entry intersection, the Step 6 resurrection, and the absent Step 4 backup. `TestPreserveInventory_GuardDetectsUnexcludedPath` passes at RED by construction — it is the falsifier for the AC-RIL2-005 guard (it exercises the predicate against a deliberately poisoned inventory), not a behavioural assertion about production wiring.

#### AC matrix

| AC | Command | Actual output | Status |
|---|---|---|---|
| AC-RIL2-005 | `go test ./internal/cli/ -run 'TestDeprecatedPaths_NoPreserveInventoryCollision' -v` | `--- PASS: TestDeprecatedPaths_NoPreserveInventoryCollision (0.01s)` | PASS |
| AC-RIL2-006 | `go test ./internal/cli/ -run 'TestPreserveInventory_GuardDetectsUnexcludedPath' -v` | `--- PASS: TestPreserveInventory_GuardDetectsUnexcludedPath (0.00s)` | PASS |
| AC-RIL2-007 | `go test ./internal/cli/ -run 'TestCleanReinstall_DeprecatedPathNotResurrected' -v` | `--- PASS: TestCleanReinstall_DeprecatedPathNotResurrected (0.02s)` | PASS |
| AC-RIL2-008 | `go test ./internal/cli/ -run 'TestCleanReinstall_BacksUpBeforeDeprecatedRemoval' -v` + `grep -c 'backupDeprecatedPaths' internal/cli/update_clean_install.go` | `--- PASS: TestCleanReinstall_BacksUpBeforeDeprecatedRemoval (0.02s)`; grep count `5` (≥ 2 required; baseline was `1`, the header comment) | PASS |
| AC-RIL2-009 | `go test ./internal/cli/ -run 'TestCleanReinstall_RemovesOrphanedDBYaml' -v` | `--- PASS: TestCleanReinstall_RemovesOrphanedDBYaml (0.03s)` | PASS |

#### Quality gates

| Check | Observed |
|---|---|
| `go build ./...` | `BUILD_OK` (exit 0) |
| `GOOS=windows GOARCH=amd64 go build ./...` | `windows_exit=0` |
| `golangci-lint run --timeout=3m` | `0 issues.` — identical to the M2 pre-flight baseline; zero NEW findings |
| `go test -cover ./internal/cli/` | `coverage: 75.7% of statements` — unchanged from the M1 figure at the reported precision (measured WITH the new M2 test file present) |
| `go test ./...` | exit 0; zero packages reported `FAIL` |

#### Implementation notes

- The exclusion is applied AFTER both inventory sources are merged and deduped, so it covers the `collectUserOwnedFiles` scan as well as the `preserveInventoryRoots` walk.
- Nesting requires an explicit `/` boundary, so `.moai/project/brandX` is NOT excluded by the `.moai/project/brand` entry. A control assertion in `TestDeprecatedPaths_NoPreserveInventoryCollision` pins this, alongside `.moai/specs/…`, `.moai/project/product.md`, and `.claude/commands/mine.md` (over-exclusion guard).
- Comparison is on slash-normalized paths (`filepath.ToSlash`), so the predicate behaves identically on windows (NFR-RIL2-005).
- The Step 4 backup sentinel is **`DEPRECATED_BACKUP_FAILED`**. The error names the first unprotected path, the count of remaining paths, and wraps the underlying `backupDeprecatedPaths` error (which itself names the failing path). The backup is all-or-nothing, so a failure aborts before ANY path is removed.
- `backupDeprecatedPaths` copies file contents and errors on a directory, while `scanDeprecatedPaths` legitimately returns directory entries (`.agency`, `.moai/project/brand`). A new `expandDeprecatedBackupTargets` helper expands directory entries into their contained files before the backup call; files and symlinks pass through unchanged. Without it, every clean-reinstall on a tree carrying `.agency/` would abort at the new backup gate.
- The manifest manager was hoisted from Step 5 to Step 4 (the backup classifies files against the deployed hash). Side effect: a manifest load failure now aborts while the tree is still intact, before the first destructive step.
- REQ-RIL2-017 untouched: the removal count is still `len(deprecated) - len(remaining)` from the post-REMOVE re-scan, and `TestCleanReinstall_RemovesOrphanedDBYaml` confirms it.

#### `CleanupOldBackups` retention finding (verified, no fix — out of M2 scope)

The hazard does **not** apply to the backups M2 adds. Evidence:

- `backupDeprecatedPaths` writes to `<root>/.moai/backup/agency-<RFC3339-ish stamp>/` (`internal/cli/update_cleanup.go:208`, `BackupDirPrefix = "agency-"` at `:32`).
- `backup.CleanupOldBackups` scans `defs.BackupsDir` = `.moai-backups` (`internal/defs/dirs.go:12`) — a different directory tree — and deletes only entries whose name is exactly 15 characters matching `########_######` (`internal/cli/update/backup/backup.go:228-241`).
- `agency-2026-07-31T16-40-00Z` is 27 characters and does not match that pattern, and it does not live under `.moai-backups/`. It is therefore not a deletion candidate in this or any later run.

No other production writer or pruner of `.moai/backup/` exists (`grep -rn 'MoAIDir, "backup"' internal --include='*.go'` → one production hit, `update_cleanup.go:208`, plus three test-file hits). Whether `.moai/backup/` needs its own retention policy is `SPEC-UPDATE-DATA-SURVIVAL-001` scope.

#### Deferred items (recorded, NOT actioned in M2)

- **D16 — REQ-RIL2-003 `or V` wording.** `REQ-RIL2-003` reads "major exactly 2 **and** the original string carried a leading `v` **or `V`**", but `AC-RIL2-001`'s nine-row table and the `spec.md` §A residue-free table both pin `V2.5.0` to `Signal 1 = false`. Accepting `V` would flip a residue-free `V2.5.0` project from `IsV2=false` to `true` — an NFR-RIL2-001 widening. M1's implementation correctly used lowercase `v` only; the **requirement text** is what is wrong. The user has decided to batch this documentation correction after M4. No `spec.md` body edit was made.

### M3 — Residue cleanup for v3-confirmed projects (REQ-RIL2-018..023)

Base for this milestone: `e6fb18b6d` (M2). Files touched: `internal/cli/update.go` (the `fpErr == nil && !fingerprint.IsV2 && isMoAIProject` block), new `internal/cli/update_residue_cleanup.go`, new `internal/cli/update_residue_cleanup_test.go`.

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-RIL2-010 | PASS | `go test ./internal/cli/ -run 'TestResidueCleanup_NoDestructiveReinstall' -v` | `--- PASS: TestResidueCleanup_NoDestructiveReinstall (0.02s)` |
| AC-RIL2-011 | PASS | `go test ./internal/cli/ -run 'TestResidueCleanup_RemovesDeprecatedOnV3Project' -v` | `--- PASS: TestResidueCleanup_RemovesDeprecatedOnV3Project (0.07s)` + `--- PASS: .../v3_project_with_marker`, `--- PASS: .../no_project_marker` |
| AC-RIL2-012 | PASS | `go test ./internal/cli/ -run 'TestResidueCleanup_IdempotentAcrossTwoRuns' -v` | `--- PASS: TestResidueCleanup_IdempotentAcrossTwoRuns (0.04s)` |
| AC-RIL2-019 | PASS | `go test ./internal/cli/ -run 'TestResidueCleanup_PreservesAgencyMigrationPreStep' -v` | `--- PASS: TestResidueCleanup_PreservesAgencyMigrationPreStep (0.02s)` |

Implementation notes:

- The `:406` block is **extended**, not replaced. `runV3ResidueCleanup` retains the `.agency/` migration pre-step under its existing conditions (`.agency` present) and adds the deprecated-path sweep after it. The migration is not gated on the sweep's outcome and is not reordered past it (REQ-RIL2-021 / AC-RIL2-019).
- **The sweep operates on the PRE-migration residue snapshot.** The migration copies `.agency/` to `.agency.archived/` and leaves both in place; `.agency.archived` is itself a `defs.DeprecatedPaths` entry, so a sweep scanning AFTER the migration would delete the archive the migration had just written and fail AC-RIL2-019. Sweeping the pre-migration snapshot keeps the archive intact while still clearing the legacy `.agency/`. The snapshot is re-filtered for existence before backup, because backing up a path the migration consumed would abort the whole sweep.
- The project-marker gate (`isMoAIProject`) is checked INSIDE `runV3ResidueCleanup`, not only at the `:406` call site, so the gate holds for every caller (REQ-RIL2-023). Falsified: removing the gate makes `TestResidueCleanup_RemovesDeprecatedOnV3Project/no_project_marker` FAIL on all three assertions.
- `expandDeprecatedBackupTargets` (M2) is reused verbatim — `scanDeprecatedPaths` returns directory entries while `backupDeprecatedPaths` copies files and errors on a directory. The sentinel `DEPRECATED_BACKUP_FAILED` is reused for the abort-before-REMOVE path (REQ-RIL2-019).
- No PRESERVE snapshot, no tree wipe, no forced redeploy (REQ-RIL2-020). `runV3ResidueCleanup` takes no `template.Deployer` at all. `TestResidueCleanup_NoDestructiveReinstall` asserts this by tree-delta: every pre-existing file except the residue survives, no `.moai/backups/v2-to-v3-*` directory appears, and the only additions live under the residue path's own `.moai/backup/` directory.
- `Removed` is derived from the filesystem (post-`RemoveAll` `os.Lstat` absence check), never from the planned-list length — the REQ-RIL2-017 discipline carried forward.

Falsification (each new guard proven able to fail):

- Neutering the sweep (early return before backup+remove) → AC-RIL2-010/011/012/019 all FAIL with `residue ... not removed`, `BackupDir empty`, `run 1 removed 0 paths, want >= 1`.
- Removing the marker gate → `no_project_marker` FAILs: `Skipped = false`, `Removed = [.claude/agents/moai/planner.md]`, and the residue is gone from a non-MoAI directory.

Deferred (recorded, NOT actioned in M3):

- **D17 — `.agency.archived` converges in two runs, not one.** For a tree carrying `.agency/`: run 1 archives `.agency/` → `.agency.archived/` and sweeps `.agency/`; run 2 finds `.agency.archived` in its own pre-migration snapshot and removes it; run 3 removes zero. This is bounded and terminating (unlike the #1243 net-zero remove-then-restore loop), but it is a two-run convergence rather than the one-run convergence REQ-RIL2-022's general wording implies. `TestResidueCleanup_IdempotentAcrossTwoRuns` uses a residue fixture without `.agency/` (a plain deprecated file), where convergence is one run. Resolving the `.agency` case would require either permanently excluding `.agency.archived` from the sweep or accepting the archive's deletion on the second run — both are scope decisions, not implementation details, so neither was taken here.

### Epic run order (depends_on sequencing)

`SPEC-UPDATE-REINSTALL-LOOP-002` declares `related_specs`, not `depends_on`, so its own run-phase `depends_on` pre-flight is trivially satisfied. The sibling Epic SPECs are all `status: draft`, which is the expected state for an Epic whose members have not yet run — it is a sequencing fact, not a per-SPEC defect.

Intended order within the Epic, so that any dependency is satisfied by sequencing rather than by an `--ignore-deps` override:

1. `SPEC-UPDATE-REINSTALL-LOOP-002` (this SPEC) — the detection and clean-reinstall correctness base.
2. `SPEC-UPDATE-YAML-PRESERVE-001` — disjoint scope (YAML merge fidelity); may run in parallel.
3. Remaining Epic siblings, in the order the Epic entry point assigns.

Where a sibling later adds a `depends_on:` entry naming this SPEC, that entry is satisfied only when this SPEC reaches `status: completed` — the strict fulfilment definition. No sibling should be entered on `--ignore-deps` to work around draft status.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Decision: sub-agent** (Mode 5 — sequential `manager-develop`, one spawn per milestone,
`cycle_type = tdd`)

### Input parameters

| Parameter | Value |
|---|---|
| tier | M (frontmatter `tier: M`) |
| scope (file count) | 4 production files + 4 new `_test.go` files ≈ 8 |
| domain count | 1 (Go source under `internal/cli/`) |
| file language mix | 100% Go |
| concurrency benefit | LOW — M1 blocks M3; M2 supplies the backup wiring M3 reuses; M4 depends on M2+M3 |
| Agent Teams prereqs | n/a (Mode 3 RETIRED) |

### Mode evaluation

| Mode | Selected | Rationale |
|---|---|---|
| 1 `trivial` | no | Semantic change across 4 production files with a data-loss hazard (Defect 5) |
| 2 `background` | no | Write work, not read-only analysis |
| 3 `agent-team` | no | RETIRED tombstone — never selected |
| 4 `parallel` | no | Single domain (<3) and coding-heavy; Anthropic's coding-task parallelism caveat applies |
| 5 `sub-agent` | **yes** | Default fallback; sequential per-milestone delegation matches the M1→M2→M3→M4 dependency chain |
| 6 `workflow` | no | ~8 files (« ~30) and not a uniform mechanical transform |

### Justification

Coding-heavy work in one domain with a hard dependency chain — M1 widens the v3-confirmed
population that M3's residue path depends on, and M2 supplies the `backupDeprecatedPaths`
wiring M3 reuses. Anthropic's coding-task parallelism caveat ("most coding tasks involve
fewer truly parallelizable tasks than research") makes sequential `manager-develop` spawns
the correct default. M2 additionally must land its two halves in one commit (exclusion +
backup) because separating them opens an unbacked-up deletion path for user-authored
`.claude/commands/agency/*.md` — a constraint that a parallel fan-out could not honour.

`cycle_type = tdd` (quality.yaml `development_mode`), one `manager-develop` spawn per
milestone, Section A-E delegation template (REQUIRED at Tier M).

### Boundary case

Domain count 1 and file count ~8 are both below the Mode 4 auto-select thresholds
(≥3 domains / ≥10 files), so no tie-breaker was needed. Mode 6's `~30`-file soft boundary
is not approached.

### Gate record

- Plan Audit Gate: iter3 **PASS 0.88** (Tier M threshold 0.80), Must-Pass 7/7. Skip-eligibility
  was NOT taken — conditions 2 (score ≥ 0.90) and 3 (artifact-hash unchanged) both failed, so
  Phase 1 re-executed as iter3.
- Implementation Kickoff Approval: obtained (user selected iter3 re-audit, then approved
  proceeding to M1 after the D12/D9/D15 documentation revision landed).
