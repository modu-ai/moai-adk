# SPEC-UPDATE-DATA-SURVIVAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]
code_baseline: 8cc108ddb          # = origin/main; verified ancestor of HEAD
worktree_head_at_revision: 89b2e4772
worktree: .claude/worktrees/e2-data-survival
branch: feat/SPEC-UPDATE-DATA-SURVIVAL-001
superseded_baselines:              # retired in the v0.4.0 re-baseline (D17)
  code_baseline: d5336214e         # 19 non-.moai files differ from HEAD
  worktree_head: a8b42e112         # NOT an ancestor of HEAD
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.71
    threshold: 0.80
    dimensions: {clarity: 0.75, completeness: 0.90, testability: 0.60, traceability: 0.65}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7]
    deferred: [D8, D9, D10]
    report: .moai/reports/plan-audit/SPEC-UPDATE-DATA-SURVIVAL-001.md
  iteration_2:
    verdict: PASS
    score: 0.84            # harmonic 0.837 / arithmetic 0.840
    threshold: 0.80
    dimensions: {clarity: 0.80, completeness: 0.90, testability: 0.78, traceability: 0.88}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7]
    opened: [D11, D12, D13, D14, D15, D16]
    deferred: [D8, D9, D10]
    report: .moai/reports/plan-audit/SPEC-UPDATE-DATA-SURVIVAL-001-epic-update-config-iter2.md
  iteration_2_delta:
    version: 0.3.0
    applied: [D13, D14, D15]
    already_present: [D11, D12, D16]
    resolved_by_fact: [D10]
    still_deferred: [D8, D9]
  iteration_3:
    verdict: FAIL
    score: 0.758           # harmonic 0.758 / arithmetic 0.760
    threshold: 0.80
    dimensions: {clarity: 0.75, completeness: 0.75, testability: 0.72, traceability: 0.82}
    scope: delta
    blockers: [D17, D18, D19]
    advisories: [D20, D21, D22]
    resolved_this_round: [D11, D12, D13, D14, D15, D16]   # D13 partially → reopened as D19
    still_deferred: [D8]
    resolved_by_fact: [D9, D10]                            # D9 escalated into D17; D10 by E1 completion
    stop_signal: true      # 0.84 → 0.758 regression; see re_baseline below
    report: .moai/reports/plan-audit/SPEC-UPDATE-DATA-SURVIVAL-001-e2-delta-iter3.md
  re_baseline:
    version: 0.4.0
    ceiling_override: user-approved            # 3-iteration ceiling exceeded by explicit user approval
    trigger: external-tree-movement            # NOT a revision failure
    cause: >
      E1 (SPEC-UPDATE-REINSTALL-LOOP-002) landed run PR #1261 (beeb0ebc2) and sync PR #1264
      (8cc108ddb) between iteration 2 and iteration 3. The plan-phase artifacts described a tree
      that no longer existed: a new destructive call site appeared and coordinates shifted across
      five files. The auditor's own diagnosis and the orchestrator's independent confirmation agree
      that the iteration-3 regression is external, and that the one iteration-2 departure (D15's
      rejection of the suggested sed-range remedy) was the correct call.
    applied: [D17, D18, D19, D20, D21, D22]
    still_deferred: [D8]
    registry_delta: {sites: 17 -> 18, pairs: 10 -> 11, added_row: update_residue_cleanup.go:135}
    go_source_changed: false                   # git diff --name-only origin/main HEAD | grep -c '\.go$' -> 0
  iteration_4:
    verdict: PASS
    score: 0.892           # harmonic 0.89173 / arithmetic 0.8925
    threshold: 0.80
    dimensions: {clarity: 0.90, completeness: 0.92, testability: 0.85, traceability: 0.90}
    must_pass: 7/7
    scope: full            # NOT a delta — iteration-2 scores deliberately not carried forward,
                           # because iteration 3 diagnosed its own regression as external drift,
                           # so a carried score is not attributable to a measured baseline
    blockers: []           # all three iteration-3 blockers cleared and independently re-verified
    resolved_this_round: [D17, D18, D19, D20, D21, D22]
    opened: [D23, D24, D25, D26]               # all minor; none blocks run-phase entry
    still_deferred: [D8]                       # M5 must resolve before AC-UDS-015 can be PASS
    gap_recovery:
      iter3_gap_1: closed                      # D1-D7 re-measured against the moved tree; none regressed
      iter3_gap_7: closed                      # every registry row's enclosing function re-derived (awk scan)
      iter3_gap_8: closed                      # full §B baseline re-run: 0 drift / 19 of 19 reproduce
    falsification_performed:
      registry_re_derivation: >
        The two replacement Go-source scan commands were falsified in a scratch copy by injecting a
        19th destructive site: both moved (18 -> 19, 11 -> 12), while the retired self-referential
        form stayed pinned at the table's own row count. Reachability proven, not assumed.
      ac_uds_001_fixture_pin: >
        Three independent mutations move the assertion; the count is sourced from the literal
        plantedMoaiManagedPaths fixture, never from CleanMoaiManagedPaths' own output.
    report: .moai/reports/plan-audit/SPEC-UPDATE-DATA-SURVIVAL-001-e2-iter4.md
  iteration_4_delta:
    version: 0.5.0
    applied: [D23, D24, D25, D26]
    also_applied: informational-row-11-sweep   # runV3ResidueCleanup removes `sweep`, the
                                               # existence-refiltered subset, not the raw
                                               # scanDeprecatedPaths return
    re_audit_required: false                   # documentation-precision only; delta-scope contract
    still_deferred: [D8]
```

### Iteration-2 delta round — divergence from the audit report

Three of the six iteration-2 findings were **already applied** in the artifacts when this delta
round opened. The audit report was written against worktree `epic-update-config` HEAD `6bd568a41`;
the artifacts that landed on `main` in PR #1258 (commit `9a6b6c854`, the sole commit touching this
SPEC directory) already carry those fixes. Re-measured on this tree rather than carried over:

| Finding | Report claim | Measured here | Disposition |
|---|---|---|---|
| D11 registry-pair-count | `11` in three places | `grep -rn '11 rows\|11 pairs\|11 (file'` → no match; both tables carry exactly `10` rows summing to `17` sites | already correct — no edit |
| D12 exemption-premise-false | rows 5/6/10 collapsed into "the same run created" | `acceptance.md` already splits row 6 as retention pruning and carries the run-scoped pruning-scope line | already correct — no edit |
| D16 grep-satisfiable-by-comment | AC-UDS-014 (b) is a text grep | (b) is already coverage-based (`-coverpkg`, `go tool cover -func`, `test -s` gate, positive control at 94.4%) | already correct — no edit |
| D10 depends_on target draft | E1 was `draft` | `git show origin/main:.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md` → `status: completed` | resolved by fact |

D13, D14, and D15 were open and are applied in v0.3.0. D15's fix deliberately departs from the
report's suggested remedy: a backward-widened window would capture the unrelated
`SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` mention at `dirs.go:302` and turn AC-UDS-007 (a)'s `0`
baseline into a false `1`. The applied fix pins the SPEC-ID's placement in REQ-UDS-009 and
content-anchors the window forward instead. See `acceptance.md` AC-UDS-007 § Range note.

### Iteration-3 re-baseline round — what moved and why

Iteration 3 scored **0.758** (clarity 0.75 / completeness 0.75 / testability 0.72 / traceability
0.82), down from iteration 2's 0.84 — a STOP-signal regression under the Retry Loop Contract. The
regression's cause is **external**: the code baseline moved under the plan.

| Measurement | v0.3.0 recorded | Re-measured on HEAD `89b2e4772` |
|---|---|---|
| worktree head anchor | `a8b42e112` | **not an ancestor of HEAD** (`git merge-base --is-ancestor` → non-zero) |
| code baseline anchor | `d5336214e` | superseded by `8cc108ddb` (verified ancestor) |
| "no Go source differs" premise | asserted | **false** — `git diff --name-only d5336214e..HEAD \| grep -v '^\.moai/' \| wc -l` → `19` |
| destructive call sites | 17 | **18** |
| (file, function) pairs | 10 | **11** |
| new site | — | `update_residue_cleanup.go:135` (`runV3ResidueCleanup`), added by `beeb0ebc2` |
| `ensureGlobalSettingsEnv` | `update.go:755` | `update.go:832` |
| `globalHooksDir` removal | `update.go:764-766` | `update.go:841-843` |
| `mergeBackPreserveInventory` | `:330` (returns `:346`/`:350`/`:354`) | `:400` (returns `:416`/`:420`/`:424`) |
| `preserveInventoryRoots` | `:66-70` | `:68-72` |
| AC-UDS-014 coverage transcript | 3 lines | **4 lines** (`MigrateLegacyMemoryDir` was dropped) |
| AC-UDS-007 (b) falsifier | `2 → 0` | `2 → 1` (contrast clause alone); `2 → 0` only on whole-banner deletion |

**Why the drift went undetected.** The two commands v0.3.0 published beneath the registry as
"independently re-derivable" grepped `acceptance.md` itself, so they returned `10` / `17` forever
regardless of the tree — §A.3 shape (a), a self-comparison, applied to the document's own
consistency check. Both are replaced in v0.4.0 with Go-source scans whose output moves when the
tree moves.

**D19 is not a regression either — it is D13's fix landing one clause short.** v0.3.0 added the
"destroyed paths are still absent on return" assertion, but left its quantifier unbounded: on an
empty removed-set "all of ∅ are absent" holds trivially, and adding the automatic rollback
REQ-UDS-020 forbids could not move it. v0.4.0 pins the removed set non-empty from a literal
fixture (`plantedMoaiManagedPaths`) verified at the source level, never from a count the test
derives by observing the function it is checking.

**Preserved through the re-baseline** (the auditor named these exemplary and asked that they
survive): AC-UDS-014's coverage construction with its 94.4% positive control, and the AC-UDS-007
(a) range note rejecting the backward-widened `sed` window. Both are intact; only their embedded
figures and coordinates were refreshed.

**Ceiling override.** This round consumed iteration 3 of 3. The user explicitly approved one
override of the plan-auditor ceiling for a bounded re-baseline round, which is the auditor's own
recommendation (its §10 option 3, reframed as "a bounded re-baseline round, not another
audit-remedy round"). No requirement's intent changed in this round.

### Epic run order (dependency sequencing)

`depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]` is a **real** dependency: E1's REQ-RIL2-015/016
(backup-before-delete for `defs.DeprecatedPaths`) is the inherited precondition behind this SPEC's
registry row 12 / §C.1 row 12, which this SPEC deliberately does not re-specify.

The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
`status: completed`. **E1 has reached that state, so the gate now passes on its own.** Measured:

```
$ git show origin/main:.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md | grep -E '^(id|status|version):'
id: SPEC-UPDATE-REINSTALL-LOOP-002
version: 0.4.0
status: completed
```

E1 landed as run PR #1261 (`beeb0ebc2`) and sync PR #1264 (`8cc108ddb`); both are ancestors of this
branch's HEAD. The dependency was satisfied by sequencing, exactly as planned — no bypass was used
and **none is needed**: `--ignore-deps` must NOT be passed on this SPEC's `/moai run`, because the
gate it would bypass is already open, and passing it would suppress a real check for no benefit.

| Order | SPEC | Status | Gate |
|---|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | **completed** | cleared — REQ-RIL2-015/016 landed |
| 2 | **`SPEC-UPDATE-DATA-SURVIVAL-001`** (this SPEC, E2) | draft | `depends_on` satisfied; run-ready after Implementation Kickoff Approval |
| 3+ | `SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-CONFIG-TIER-PERSIST-001`, `SPEC-CONFIG-KEY-HONESTY-001`, `SPEC-UPDATE-DOC-DRIFT-001` | draft | no `depends_on` edge to this SPEC |

The earlier contingency — running M1 and M3-M6 while leaving M2's registry row 12 open until E1
closed — is now moot. All six milestones may proceed in order; registry row 12 records E1's
REQ-RIL2-015 assignment against a requirement that has shipped.

## §E.2 Run-phase Evidence

### M1 — Failure contract: recovery manifest + restore entry point

Measured on branch `feat/SPEC-UPDATE-DATA-SURVIVAL-001`, code baseline `a2fc68e73`
(SPEC-artifact-only ahead of `8cc108ddb` = `origin/main`). Every row below is the observed output
of a command run in this milestone against this tree.

#### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UDS-001 (a) | PASS | `go test -run 'TestUpdateFailure_WritesRecoveryManifest' -count=1 -v ./internal/cli/` | `--- PASS: TestUpdateFailure_WritesRecoveryManifest (0.01s)` / `ok github.com/modu-ai/moai-adk/internal/cli 0.496s` |
| AC-UDS-001 (b) | PASS | `sed -n '/^var plantedMoaiManagedPaths = \[\]string{/,/^}/p' internal/cli/update_recovery_manifest_test.go \| grep -cE '^\s+"'` | `6` (AC requires ≥ 1) |
| AC-UDS-002 | PASS | `go test -run 'TestRestore_ProceedsWithoutProjectMarker' -count=1 -v ./internal/cli/` | `--- PASS: TestRestore_ProceedsWithoutProjectMarker (0.01s)` |
| AC-UDS-003 | PASS | `go test -run 'TestUpdate_RejectsTreeWithoutProjectMarker' -count=1 -v ./internal/cli/` | `--- PASS: TestUpdate_RejectsTreeWithoutProjectMarker (0.01s)` |
| AC-UDS-004 | PASS | `go test -run 'TestRestore_IdempotentAndRefusesForeignDir' -count=1 -v ./internal/cli/` | `--- PASS: TestRestore_IdempotentAndRefusesForeignDir (0.01s)` |
| AC-UDS-018 (cross-cutting) | PASS | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |

#### RED evidence (captured before GREEN)

Build-level RED (symbols absent):

```
internal/cli/update_recovery_manifest_test.go:81:11: undefined: newRecoveryGuard
internal/cli/update_restore_test.go:81:5: undefined: checkProjectMarker
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

Runtime RED against no-op stubs:

```
--- FAIL: TestUpdateFailure_WritesRecoveryManifest (0.01s)
    clause 1: recovery manifest not found at .../.moai-backups/20260801_151945/recovery-manifest.txt
--- FAIL: TestRestore_ProceedsWithoutProjectMarker (0.00s)
--- FAIL: TestUpdate_RejectsTreeWithoutProjectMarker (0.00s)
--- FAIL: TestRestore_IdempotentAndRefusesForeignDir (0.00s)
```

#### AC-UDS-001 non-vacuity — four observed mutations (`go test -overlay`, no tree mutation)

| Mutation | Clause that failed | Observed |
|---|---|---|
| `plantedMoaiManagedPaths` emptied | (b) + the in-test non-empty guard | command (b) → `0`; `plantedMoaiManagedPaths must be non-empty …` |
| paths declared but never created | clause 3 pre | `clause 3 pre: planted path .claude/settings.json absent before CleanMoaiManagedPaths` (×6) |
| automatic rollback injected into the failing stage | clause 4 | `clause 4: .claude/settings.json reappeared after the failing call returned (automatic rollback is prohibited by REQ-UDS-020)` (×6) |
| production guard stops writing the manifest file | clause 1 | `clause 1: recovery manifest not found at …/recovery-manifest.txt` |

#### Extend-vs-separate decision (plan.md §H two-owners hazard)

The restore entry point **EXTENDS** `backup.RestoreMoaiConfig`; it is not a second restore path.
`backup.RestoreFromBackupDir` adds only the preconditions the two mid-run callers
(`update_clean_install.go`, `update_template_sync.go`) do not need — a marker-file check
(REQ-UDS-024) and an idempotency pass (REQ-UDS-023) — then delegates the copy/merge semantics to
the single existing owner. The decision is recorded in a code comment at the entry point.

#### Idempotency finding (REQ-UDS-023)

`RestoreMoaiConfig` raw-copies a section whose target is ABSENT but re-serialises a section whose
target EXISTS through the YAML merge, which canonicalises key order and indentation. Measured on a
wiped config directory: pass 1 → `moai:\n  version: 3.0.0\n  template_version: 3.0.0\n`;
pass 2 → `moai:\n    template_version: 3.0.0\n    version: 3.0.0\n`; pass 3 → identical to pass 2.
A single pass therefore left the tree one step short of the merge fixed point, and AC-UDS-004
failed with two differing tree hashes. `RestoreFromBackupDir` now runs the merge to its fixed point
(two passes), scoped to the user-invoked restore; the mid-run callers keep single-pass behaviour
(NFR-UDS-005).

#### Coverage

| Package | Baseline (`a2fc68e73`, scratch worktree) | After M1 |
|---|---|---|
| `internal/cli` | `coverage: 75.8% of statements` | `coverage: 75.8% of statements` |
| `internal/cli/update/backup` | `coverage: 88.6% of statements` | `coverage: 88.9% of statements` |

#### Production wiring

- `update.go` — `--restore <dir>` branch placed BEFORE the project-marker gate (the scoped
  exemption, REQ-UDS-022/025); the gate itself extracted to `checkProjectMarker`.
- `update_template_sync.go` — every step runs under the recovery guard; `Clean Managed Paths` is
  the destructive stage.
- `update_clean_install.go` — the guard enters the destructive region at the Step 4 removal loop;
  the eight post-Step-4 error returns route through it.

#### Gaps (not verified in M1)

- `AC-UDS-003` drives the extracted `checkProjectMarker` predicate, not a full `runUpdate`
  invocation; that the gate is still *called* by `runUpdate` rests on reading the call site, not on
  an executed assertion.
- The `--restore` CLI flag is not exercised end-to-end by a test; the entry point beneath it is.
- The clean-reinstall path's manifest wiring is not covered by an executed failure test — only the
  template-sync stage sequencer is.

#### Lint

`golangci-lint run --timeout=5m ./internal/cli/...` → `0 issues.`; `gofmt -l` over the eight edited
files → no output; `go vet ./internal/cli/...` → exit 0.

### M2 — Destructive-target registry + `.moai/memory/` backup + comment reconciliation

Measured on branch `feat/SPEC-UPDATE-DATA-SURVIVAL-001`, pre-M2 tree `8556bc9e0`, M2 code commit
`a46652923`. Every row below is the observed output of a command run in this milestone against this
tree.

#### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UDS-005 | PASS | `go test -run 'TestDestructiveTargetRegistry_CoversAllSites' -count=1 -v ./internal/cli/` | `--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.01s)` / `ok github.com/modu-ai/moai-adk/internal/cli 0.744s` |
| AC-UDS-005 (§C.4 falsification — the gate) | PASS | scratch worktree at `a46652923`, unregistered `udsFalsifyProbe` injected into `update_cleanup.go`, then `go -C /tmp/uds-falsify-c4-wt test -run 'TestDestructiveTargetRegistry_CoversAllSites' -count=1 -v ./internal/cli/` | `--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.01s)` naming `unregistered destructive site: internal/cli/update_cleanup.go udsFalsifyProbe has 1 os.RemoveAll/os.Rename call site(s) but no registry row` |
| AC-UDS-006 | PASS | `go test -run 'TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval' -count=1 -v ./internal/cli/update/deploy/` | `--- PASS: TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval (0.01s)`; sentinel bytes recovered from `…/.moai-backups/20260801_165222/legacy-memory/notes/keep.md` |
| AC-UDS-007 (a) | PASS | `sed -n '/brand + db directories/,/Path: *"\.moai\/project\/brand"/p' internal/defs/dirs.go \| grep -c 'SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001'` | `1` (baseline `0`; AC requires ≥ 1) |
| AC-UDS-007 (b) | PASS | `grep -c 'moai/project/db' internal/defs/dirs.go` | `2` — unchanged |
| AC-UDS-007 (c) | PASS | `go test -run 'TestDeprecatedPaths' -count=1 -v ./internal/defs/` | 8 `--- PASS` lines incl. `TestDeprecatedPathsTotalCount`, `TestDeprecatedPathsCategorySplit` |
| AC-UDS-010 (cross-cutting) | PASS | `go test -count=1 ./internal/cli/... ./internal/cli/update/... ./internal/defs/` | exit 0, 18 `ok` lines, 0 `--- FAIL` |
| AC-UDS-018 (cross-cutting) | PASS | `go build ./...` ; `GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 |
| AC-UDS-019 (a) | PASS | `git diff --name-only "$(git merge-base origin/main HEAD)"..HEAD -- internal/template/templates/ \| wc -l` | `0` — the debt below was closed by `027e16eb4`, which re-anchored the AC on the merge-base; the retired pinned form returned `8`, attributable in full to foreign commit `9ced435e9` |
| AC-UDS-019 (b) | PASS | `git status --porcelain internal/template/templates/ \| wc -l` | `0` |

#### Step 0 re-scan (plan.md §F M2)

The §C.0 scan was re-run before anything was encoded, rather than trusting the table:

```
$ grep -rn 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | wc -l
      18
```

**18 sites across 11 (file, function) pairs** — identical to the §C.0 table, and identical again
after M2 landed (the new `backupLegacyMemoryDir` / `copyTree` helpers use only `MkdirAll` /
`ReadFile` / `WriteFile`, so they add no destructive site). No divergence to report.

#### RED evidence (captured before GREEN)

The registry file was created with an **empty** slice first, so the guard's RED is a real `--- FAIL`
enumerated from source rather than a compile error — which also demonstrates, before any row exists,
that the enumeration cannot be registry-derived:

```
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.01s)
    unregistered destructive site: internal/cli/update.go ensureGlobalSettingsEnv has 1 … but no registry row
    unregistered destructive site: internal/cli/update/backup/backup.go BackupMoaiConfig has 3 … but no registry row
    unregistered destructive site: internal/cli/update/backup/backup.go CleanupOldBackups has 1 … but no registry row
    unregistered destructive site: internal/cli/update/deploy/deploy.go CleanMoaiManagedPaths has 3 … but no registry row
    unregistered destructive site: internal/cli/update/deploy/deploy.go MigrateLegacyMemoryDir has 2 … but no registry row
    unregistered destructive site: internal/cli/update_archive.go archiveLegacySkills has 1 … but no registry row
    unregistered destructive site: internal/cli/update_archive.go archiveSkill has 1 … but no registry row
    unregistered destructive site: internal/cli/update_clean_install.go runCleanReinstall has 1 … but no registry row
    unregistered destructive site: internal/cli/update_cleanup.go removeDeprecatedFile has 1 … but no registry row
    unregistered destructive site: internal/cli/update_namespace_protect.go backupUserOwnedNamespace has 3 … but no registry row
    unregistered destructive site: internal/cli/update_residue_cleanup.go runV3ResidueCleanup has 1 … but no registry row
    scanned 11 (file, function) pair(s) …; registry has 0 row(s)
```

`.moai/memory/` backup, against the pre-M2 source (`8556bc9e0`, scratch worktree, `go -C`):

```
--- FAIL: TestMigrateLegacyMemoryDir_BacksUpBeforeRemoval (0.01s)
    sentinel bytes not found anywhere under …/.moai-backups — .moai/memory/ was destroyed without a backup (REQ-UDS-008)
--- FAIL: TestMigrateLegacyMemoryDir_AbortsRemovalOnBackupFailure (0.01s)
    expected an error when the backup cannot be written, got nil
```

`dirs.go` group comment — the grep baseline is the RED-equivalent (this deliverable is verified by
AC-UDS-007's grep, not by a new Go test): (a) `0` before, `1` after; (b) `2` before and after.

#### §C.4 falsification mechanism (and why `-overlay` was not used)

The guard scans the on-disk tree with `go/parser`, so `go test -overlay` is invisible to it and step
3 would PASS for the wrong reason. Per the §C.4 caveat the injection was therefore applied in a
**scratch `git worktree`** driven with `go -C` (`git stash` prohibited by plan.md §D). Sequence
observed: unmutated PASS → inject `udsFalsifyProbe` (scan goes 18 → 19) → `--- FAIL` naming the
unregistered site → `git checkout --` restore → PASS. The worktree was disposed with
`git worktree remove`.

#### Coverage

Baseline re-measured in a scratch worktree at `8556bc9e0` (not carried over from M1's record):

| Package | Baseline (`8556bc9e0`) | After M2 | Delta |
|---|---|---|---|
| `internal/cli` | `coverage: 75.8% of statements` | `coverage: 75.8% of statements` | 0.0 pp |
| `internal/cli/update/deploy` | `coverage: 97.5% of statements` | `coverage: 94.3% of statements` | **−3.2 pp** |

The deploy drop is new uncovered code, not lost coverage: `MigrateLegacyMemoryDir` and
`backupLegacyMemoryDir` are both 100%, while `copyTree` sits at 76.5%. Its four uncovered blocks are
fault-injection-only error returns (`deploy.go:230` walk error, `:235` `filepath.Rel` error, `:248`
`os.ReadFile` error, `:251` per-file `MkdirAll` error). The package remains above the 85% threshold.

#### Template neutrality finding (AC-UDS-019 (a))

Command (a) returns `8`, not `0`. **None of it is this SPEC's.** Every SPEC-owned commit
(`89b2e4772`, `4f3b96fd8`, `a2fc68e73`, `4ddd35120`, `279ae7b97`, `8556bc9e0`, `a46652923`) touches
`0` files under `internal/template/templates/`. All 8 files come from the foreign commit
`9ced435e9` (`fix(agents): resolve GEARS-vs-Given-When-Then layer confusion…`, PR #1266), which
landed on `origin/main` after `8cc108ddb` and entered this branch through merge commit `2255165f5`.
`origin/main` is now `9ced435e9`, and re-measured against it the count is `0`:

```
$ git diff --name-only 9ced435e9..HEAD -- internal/template/templates/ | wc -l
       0
$ git merge-base origin/main HEAD
9ced435e922934fe68835a6f85943d3f8e330e1d
```

This is the same class of TOCTOU drift `acceptance.md` §A.4 anticipates: the `8cc108ddb` baseline
pin is stale because upstream moved. Re-baselining that pin is an `acceptance.md` body edit, which
run-phase does not own — carried as a blocker below.

#### Gaps (not verified in M2)

- `copyTree`'s four error returns are unexercised (see Coverage). A failure to read one file
  mid-copy would abort the backup and therefore the removal — the safe direction — but that path is
  reasoned, not executed.
- The registry's protection assignments are **documentation of where protection lives**, not
  executable assertions. The guard proves every site is registered and assigned; it does not prove
  the named protection actually runs. M3/M4 supply the executed evidence for rows
  `CleanMoaiManagedPaths` and `ensureGlobalSettingsEnv`.
- The scan is keyed on the literal selector `os.RemoveAll` / `os.Rename`. A destructive call reached
  through an alias (`import osx "os"`) or a wrapper would not be seen. No such call exists in scope
  today; the file-total-vs-function-total cross-check in the scanner catches only the
  outside-a-function case, not aliasing.
- ~~`AC-UDS-019 (a)` is recorded PASS-WITH-DEBT, not PASS.~~ **CLOSED by `027e16eb4`** — the AC was
  re-anchored on `$(git merge-base origin/main HEAD)` instead of the literal `8cc108ddb` pin, which
  had gone stale when foreign commit `9ced435e9` (PR #1266) entered this branch through merge
  `2255165f5`. The AC now reads `0`/`0` on this tree. The merge-base form is invariant under future
  merges; the pinned form was not.

#### Lint

`golangci-lint run --timeout=2m` → `0 issues.` — identical to the pre-flight baseline measured on
this tree before any M2 edit (`0 issues.`), so M2 introduces no new finding. `gofmt -l` over the
five M2 files → no output. (`internal/cli/update/deploy/deploy_test.go` is reported by `gofmt -l`
but is pre-existing and untouched by M2.)

### M3 — On-disk backup for the three in-memory-only files

Landed on `feat/SPEC-UPDATE-DATA-SURVIVAL-001`. Files: `internal/cli/update_disk_backup.go` (new),
`internal/cli/update_disk_backup_test.go` (new), `internal/cli/update_template_sync.go` (edited),
`internal/cli/update_clean_install.go` (edited).

#### Claim

`.claude/settings.json`, `.moai/status_line.sh`, and `.gitignore` reach the run-scoped backup
directory on disk before the first destructive step of **both** execution paths, and a backup-write
failure aborts before that step rather than logging and continuing. The in-memory merge-back is
retained unchanged (REQ-UDS-004).

#### Evidence — AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UDS-008 | PASS | `go test -run 'TestBackup_OnDiskBeforeFirstDestructiveStep' -count=1 -v ./internal/cli/` | `--- PASS: TestBackup_OnDiskBeforeFirstDestructiveStep (0.00s)` / `ok github.com/modu-ai/moai-adk/internal/cli 0.606s` |
| AC-UDS-009 | PASS | `go test -run 'TestBackup_OnDiskCoverageParityAcrossPaths' -count=1 -v ./internal/cli/` | `--- PASS: TestBackup_OnDiskCoverageParityAcrossPaths (0.00s)` |
| AC-UDS-020 | PASS | `go test -run 'TestBackup_AbortsBeforeDestructionOnWriteFailure' -count=1 -v ./internal/cli/` | `--- PASS: TestBackup_AbortsBeforeDestructionOnWriteFailure (0.00s)` |
| AC-UDS-010 | PASS | `go test -count=1 ./internal/cli/... ./internal/cli/update/...` | exit 0, no `--- FAIL`; `ok .../internal/cli 154.923s` + 16 sibling packages `ok` |

M2 drift guard re-verified after the M3 edits: `go test -run
TestDestructiveTargetRegistry_CoversAllSites -count=1 -v ./internal/cli/` → `--- PASS:
TestDestructiveTargetRegistry_CoversAllSites (0.01s)`. The clean-reinstall removal loop moved inside
a closure, and the registry's (file, enclosing function, count) key still resolves it.

#### Evidence — RED before GREEN (TDD, template §E8)

Captured against a compiling stub, so each RED is a real `--- FAIL`, never a build error or the
`[no tests to run]` exit-0 vacuity `acceptance.md` §A.2 rejects:

```
--- FAIL: TestBackup_OnDiskBeforeFirstDestructiveStep (0.00s)
    update_disk_backup_test.go:95: on-disk backup did not survive the destructive step:
    stat in-memory-backups/.claude/settings.json: no such file or directory

--- FAIL: TestBackup_OnDiskCoverageParityAcrossPaths (0.00s)
    update_disk_backup_test.go:128: walk in-memory-backups: lstat in-memory-backups:
    no such file or directory

--- FAIL: TestBackup_AbortsBeforeDestructionOnWriteFailure (0.00s)
    update_disk_backup_test.go:199: expected an error when the on-disk backup write fails
```

#### Baseline-attribution

Measured on branch `feat/SPEC-UPDATE-DATA-SURVIVAL-001`, HEAD `8f799dcd5`, before any M3 edit:

```
$ go build ./...                                                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                                → exit 0
$ golangci-lint run --timeout=3m                                          → 0 issues.
$ grep -rn 'backup-write-failed' internal/cli/ --include='*.go' | wc -l   → 0
$ grep -rn 'settings.json' internal/cli/update/backup/ --include='*.go' \
    | grep -v _test | wc -l                                               → 0
```

Post-M3 on the same tree: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
`golangci-lint run --timeout=3m` → `0 issues.` (no NEW finding against the 0-issue baseline);
`grep -rn 'backup-write-failed' internal/cli/ --include='*.go' | grep -v _test | wc -l` → `1`.

Coverage: `go test -cover ./internal/cli/... ./internal/cli/update/...` → `internal/cli` 75.8%
(package-wide, pre-existing level), `internal/cli/update/backup` 88.9%. Per-function on the new
file: `inMemoryOnlyBackupTargets` 100.0%, `guardFirstDestructiveStep` 100.0%,
`backupInMemoryOnlyFiles` 83.3%, `ensureRunBackupDir` 75.0%.

The `settings.json`-in-backup-package baseline stays `0` by design: the mechanism lives in
`internal/cli/update_disk_backup.go`, not in `internal/cli/update/backup/`, so no second backup root
is introduced and `merge.go` is untouched.

#### Gaps (not verified in M3)

- **AC-UDS-009's filesystem half is per-fixture, not per-production-path.** Driving `runUpdate`'s
  template-sync flow and `runCleanReinstall` end-to-end for a set comparison was out of proportion to
  the milestone, so the test asserts (1) set equality across two independent fixtures through the
  shared mechanism and (2) a source scan proving BOTH `update_template_sync.go` and
  `update_clean_install.go` route their first destructive step through `guardFirstDestructiveStep`.
  Half (2) is what actually catches drift — deleting or bypassing either call site fails the test —
  but it is a static scan, not an executed path.
- `ensureRunBackupDir`'s `MkdirAll` failure branches (both arms) are unexercised; a failure there
  aborts before destruction, which is the safe direction, but the path is reasoned, not executed.
- The read-error branch of `backupInMemoryOnlyFiles` (a target that exists but cannot be read, e.g.
  a permission error) is unexercised. Only the write-failure branch is driven, via the
  `diskBackupWriteFile` seam.
- The on-disk copies are **written but never read back by production code**. M1's restore entry point
  applies the `.moai/config` backup; wiring it to also reapply `in-memory-backups/` is not in M3's
  scope and no AC requires it. A stranded operator currently recovers these three files by copying
  them out of the backup directory by hand.

#### Residual risk

- `diskBackupWriteFile` is a package-level seam, so `TestBackup_AbortsBeforeDestructionOnWriteFailure`
  must stay non-parallel. It restores the original through `t.Cleanup`; a future parallel test in
  `package cli` that also replaces it would race.
- The backup directory grows by three small files per run. `CleanupOldBackups` retention already
  applies to the same root, so this is bounded, but the retention count was not re-tuned for the
  added payload.
- `guardFirstDestructiveStep` is invoked inside the clean-reinstall path's recovery region: a
  backup-write failure there is wrapped by `recovery.fail("step 4: remove deprecated paths", …)`, so
  the recovery manifest names the removal step even though nothing was removed. The sentinel in the
  wrapped error still identifies the real cause.

### M4 — HOME deletion-radius pinning

Landed on `feat/SPEC-UPDATE-DATA-SURVIVAL-001`. Files: `internal/cli/update.go` (edited),
`internal/cli/update_home_radius_test.go` (new).

#### Claim

`ensureGlobalSettingsEnv` resolves HOME through the injectable `userHomeDirFn` seam, its removal
target is the single named symbol `globalMoaiHooksDir(homeDir)`, and a guard pins the deletion
radius to `<HOME>/.claude/hooks/moai` — a user-owned sibling at `<HOME>/.claude/hooks/user-hook.sh`
survives. The guard achieves HOME isolation through the seam, not through a process-wide env
mutation, and does not touch the operator's real home directory.

#### Evidence — AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UDS-011 (a) | PASS | `sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go \| grep -c 'userHomeDirFn()'` | `1` (baseline `0`) |
| AC-UDS-011 (b) | PASS | `go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/` | `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius (0.00s)` / `ok github.com/modu-ai/moai-adk/internal/cli 0.969s` |
| AC-UDS-012 | PASS | `go test -overlay=/tmp/uds-falsify/overlay.json -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/` | `--- FAIL: TestEnsureGlobalSettingsEnv_HooksRemovalRadius (0.00s)` / `FAIL github.com/modu-ai/moai-adk/internal/cli 0.980s` |
| AC-UDS-013 | PASS | before-snapshot → test → after-snapshot → `diff`; plus the two mechanism greps | `--- PASS:` line; `diff-exit=0` with no diff output; `t.Setenv("HOME"` grep → `0`; `userHomeDirFn()` grep → `1` |
| AC-UDS-018 | PASS | `go build ./...` ; `GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 |
| AC-UDS-019 | PASS | `BASE=$(git merge-base origin/main HEAD); git diff --name-only $BASE..HEAD -- internal/template/templates/ \| wc -l` | `0` (merge-base `9ced435e922934fe68835a6f85943d3f8e330e1d`); working tree also `0` |

#### Evidence — falsification (AC-UDS-012, `acceptance.md` §C.1)

The overlay copy drops the `"moai"` segment from `globalMoaiHooksDir`'s `filepath.Join`, widening the
radius to `<HOME>/.claude/hooks`. Mutated line, verbatim:

```
843:	return filepath.Join(homeDir, defs.ClaudeDir, "hooks")
```

Guard output under that overlay, verbatim:

```
=== RUN   TestEnsureGlobalSettingsEnv_HooksRemovalRadius
    update_home_radius_test.go:51: user-owned sibling …/001/.claude/hooks/user-hook.sh must survive,
    stat err = stat …/001/.claude/hooks/user-hook.sh: no such file or directory
--- FAIL: TestEnsureGlobalSettingsEnv_HooksRemovalRadius (0.00s)
```

The recorded baseline this refutes: the audit lens ran the same widening against the pre-M4 suite and
observed `ok github.com/modu-ai/moai-adk/internal/cli 18.305s` — every test passed with the radius
widened. The falsification was re-run after the only post-GREEN test edit (a comment rewording) and
still produced `--- FAIL`.

**Which assertion discriminates.** Because the fixture builds its removal target from the same named
symbol the production code uses (REQ-UDS-013), the widening mutation moves *both* the create-target
and the remove-target. The "moai subdirectory is gone" assertion therefore still passes under the
mutation; the **sibling-survival** assertion is the one that flips. That is the intended coupling —
asserting against the named symbol is what makes the guard track the production radius rather than a
re-derived constant — but it means the guard has one load-bearing assertion, not two.

#### Evidence — RED before GREEN (TDD, template §E8)

RED here is a **build failure**, not a `--- FAIL` assertion. Verbatim:

```
$ go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius' -count=1 -v ./internal/cli/
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/update_home_radius_test.go:27:18: undefined: globalMoaiHooksDir
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

This is weaker RED evidence than M3's compiling-stub `--- FAIL`, and the weakness was deliberate: the
alternative — adding `globalMoaiHooksDir` first, then running the guard before the seam substitution
— would have executed `ensureGlobalSettingsEnv` against `userHomeDir()`, i.e. the operator's **real**
`$HOME`, and issued `os.RemoveAll` on the real `~/.claude/hooks/moai`. A compiling-stub RED is not
reachable for this guard without that hazard. The assertion-level failure evidence the compiling-stub
RED would have supplied is instead supplied by AC-UDS-012's overlay `--- FAIL` above, which
demonstrates the same property (the guard fails when the radius is wrong) against real production
code.

#### Baseline-attribution

Measured on branch `feat/SPEC-UPDATE-DATA-SURVIVAL-001`, worktree
`.claude/worktrees/e2-data-survival`, HEAD `997679fc4`, before any M4 edit:

```
$ sed -n '/^func ensureGlobalSettingsEnv/,/^}/p' internal/cli/update.go | grep -c 'userHomeDirFn()'
0
$ grep -rn 'globalHooksDir' internal/cli/ --include='*_test.go' | wc -l
0
```

Post-M4 on the same tree: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
`go vet ./...` exit 0; `golangci-lint run --timeout=3m` → `0 issues.` (no NEW finding against the
0-issue baseline); `go test ./internal/cli/...` exit 0 (`ok .../internal/cli 245.946s` + 16 sibling
packages `ok`); `go test ./...` exit 0 with `0` lines matching `^FAIL`.

Targeted per-function coverage (`go test -run 'TestEnsureGlobalSettingsEnv_HooksRemovalRadius'
-covermode=set ./internal/cli/`): `globalMoaiHooksDir` 100.0%, `ensureGlobalSettingsEnv` 18.9% (the
guard exercises only the pre-settings-read prefix — the function returns early when no global
settings file exists).

#### Gaps (not verified in M4)

- **AC-UDS-013's filesystem diff is non-discriminating on this machine for the deletion direction.**
  `~/.claude/hooks` does not exist here (`test -d ~/.claude/hooks` → `DIR ABSENT`), so both snapshots
  are empty (0 lines) and `diff-exit=0` is reached trivially with respect to *deletion*. The check is
  not fully vacuous — had the test created anything under the real `~/.claude/hooks`, the
  after-snapshot would be non-empty and the diff would fail — so it still discriminates the
  *creation* direction. On an operator machine that does have `~/.claude/hooks` populated, the same
  command discriminates both directions. This is a property of the environment, not of the AC.
- **Three other `userHomeDir()` call sites remain unrouted through the seam**:
  `update_clean_install.go:410`, `update_template_sync.go:227`, `update_template_sync.go:272`.
  REQ-UDS-013 and AC-UDS-011 (a) name only `ensureGlobalSettingsEnv`, and the AC's `sed` window is
  scoped to that function body, so leaving them is in scope-discipline — but it means the seam
  substitution is **not uniform** across the update subsystem, and a future guard for any of those
  three paths will have to do its own seam work first.
- **The guard covers one radius, not the class.** It pins `ensureGlobalSettingsEnv`'s target only.
  Registry row 13 (`~/.claude/hooks/moai`) is the sole HOME-scoped destructive site M4 addresses; no
  guard was added for other outside-project deletions, and none is required by an M4 AC.
- **The `os.Stat` guard preceding the `os.RemoveAll` is not separately exercised.** The test always
  creates the directory, so the "target absent → skip" branch is covered only incidentally by
  pre-existing tests (`update_test.go:2009`), not by this guard.

#### Residual risk

- **`userHomeDirFn` is a package-level var, so this test must stay non-parallel.** It is declared at
  `glm_tools.go:123` and `glm_tools_test.go`'s `setupToolsTestHome` helper (47 call sites) reassigns
  the same variable. Confirmed by direct measurement, not assumed: `grep -c "t.Parallel()"
  internal/cli/glm_tools_test.go` → `0`, so no current caller runs parallel and no race exists today.
  A future `t.Parallel()` added to any `setupToolsTestHome` caller — or to this test — would race
  with it. Both sites restore the original via `t.Cleanup`, which orders the restore correctly but
  does not prevent concurrent reassignment.
- **The named-symbol coupling is a single point of failure in both directions.** It is what makes the
  guard track production (the point of REQ-UDS-013), but it also means a mutation that changes the
  symbol's return value moves the fixture with it — so the guard detects a *wrong radius* only via the
  sibling assertion. A mutation that instead bypassed `globalMoaiHooksDir` entirely and inlined a
  different path in `ensureGlobalSettingsEnv` would be caught (the fixture and the production target
  would diverge), but this direction was not separately falsified.
- **`userHomeDirFn`'s default still reads `$HOME` first** (`homedir.go:14`). Existing tests that
  inject HOME via the environment (`update_test.go:1498`/`:1528`,
  `coverage_improvement_test.go:4719`/`:4755`) therefore keep working unchanged — verified by the
  green `./internal/cli/...` run — but the seam does not *remove* the env-based path, so a future
  test could still reach production through `$HOME` and reintroduce the NFR-UDS-002 hazard.

### M5 — Non-vacuous user-area safety guard

Landed on `feat/SPEC-UPDATE-DATA-SURVIVAL-001`. Files: `internal/cli/update_safety_test.go` (rewritten).
No production code changed in this milestone.

#### Claim

`TestMoaiUpdate_PreservesUserArea` now drives the real production entry point
`deploy.CleanMoaiManagedPaths(projectRoot, io.Discard)` instead of a fake defined in the test file.
The `simulateMoaiUpdate` fake is deleted (definition, not renamed). The guard asserts in both
directions — user-owned areas byte-identical before and after, AND at least one managed target
actually removed — and it FAILS when the production entry point is mutated to touch
`.claude/agents/harness/`.

#### Evidence — AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UDS-014 (a) | PASS | `grep -cE '^func simulateMoaiUpdate\|^func simulate[A-Za-z]*Update' internal/cli/update_safety_test.go` | `0` (baseline `1`) |
| AC-UDS-014 (b) | PASS | `go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -covermode=set -coverpkg=./internal/cli/...,./internal/cli/update/... -coverprofile=/tmp/uds-m5.out ./internal/cli/` then `go tool cover -func` filtered to the 4 registry functions, `grep -vE '[[:space:]]0\.0%$'`, `test -s` | `test -s` passed; surviving lines: `.../update/deploy/deploy.go:29: CleanMoaiManagedPaths 66.7%` and `.../update/deploy/deploy.go:144: MigrateLegacyMemoryDir 26.9%` (baseline: all four `0.0%`, `test -s` printed `VACUOUS: guard executed no registry production function`) |
| AC-UDS-014 (c) | PASS | `go test -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/` | `--- PASS: TestMoaiUpdate_PreservesUserArea (0.01s)` / `ok github.com/modu-ai/moai-adk/internal/cli 0.838s` |
| AC-UDS-015 | PASS | `go test -overlay=/tmp/uds-falsify-m5/overlay.json -run 'TestMoaiUpdate_PreservesUserArea' -count=1 -v ./internal/cli/` | `--- FAIL: TestMoaiUpdate_PreservesUserArea (0.00s)` / `FAIL github.com/modu-ai/moai-adk/internal/cli 0.810s` |
| AC-UDS-018 | PASS | `go build ./...` ; `GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 |
| AC-UDS-019 | PASS | `BASE=$(git merge-base origin/main HEAD); git diff --name-only $BASE..HEAD -- internal/template/templates/ \| wc -l` | `0` (merge-base `835ea7b9125bc34770384f626f27026e9eca64f1`) |

#### Evidence — falsification (AC-UDS-015, `acceptance.md` §C.2)

The overlay copy of `internal/cli/update/deploy/deploy.go` adds one `cleanTarget` entry for the
user-owned path `.claude/agents/harness`, inserted after the `RulesMoaiSubdir` entry:

```go
// MUTATION (AC-UDS-015 falsification only — never committed).
{
    displayPath: filepath.Join(defs.ClaudeDir, "agents", "harness"),
    fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, "agents", "harness"),
},
```

Guard output under that overlay, verbatim:

```
=== RUN   TestMoaiUpdate_PreservesUserArea
    update_safety_test.go:137: user area changed: …/001/.claude/agents/harness
        pre:  map[ios-architect.md:a2cad9d5e0d950aedef6a81f7226d9c186293debdecafe2aff013e57da4c711d]
        post: map[]
--- FAIL: TestMoaiUpdate_PreservesUserArea (0.00s)
```

The overlay `Replace` key is `internal/cli/update/deploy/deploy.go` (module-root-relative) — a
different file from M4's overlay (`internal/cli/update.go`). The scratch directory
`/tmp/uds-falsify-m5/` was removed after the run.

The recorded baseline this refutes: `acceptance.md` AC-UDS-015 records that the same class of
mutation against the pre-M5 test **passed**, because `simulateMoaiUpdate` never called production
code at all. That baseline was NOT re-measured here — see Gaps.

#### Evidence — RED before GREEN (TDD, template §E8)

**This milestone has no RED step, and the omission is structural rather than skipped.** M5 deletes a
vacuous guard and replaces it with a non-vacuous one against **unchanged production code**; there is
no new behaviour to specify, so there is no state in which the replacement guard legitimately fails
against the real tree. Writing a deliberately-wrong assertion first would manufacture a RED with no
diagnostic value.

The assertion-level failure evidence RED would normally supply is supplied instead by AC-UDS-015's
overlay `--- FAIL` above: it demonstrates the same property (this guard fails when production code
violates the invariant) against real production code. That is the stronger form of the same
evidence, since it exercises the actual entry point rather than a stub.

#### Which user paths are genuinely at risk vs which survive trivially

Determined empirically by reading the `cleanTarget` list inside `CleanMoaiManagedPaths` (7 entries:
`.claude/settings.json`, `.claude/commands/moai`, `.claude/agents/moai`, `.claude/skills/moai*` (glob),
`.claude/rules/moai`, `.claude/output-styles/moai`, `.claude/hooks/moai`) plus the two unconditional
removals (`.moai/config/` entirely, then `MigrateLegacyMemoryDir` on `.moai/memory/`):

| User-owned fixture path | Risk from THIS entry point | Why |
|---|---|---|
| `.claude/skills/harness-ios-patterns/` | **Genuine** | It sits inside the directory the `.claude/skills/moai*` glob enumerates. A widened glob (`moai*` → `*`) or a sibling-prefix change sweeps it. Its survival depends on the glob pattern, not on the directory being out of reach. |
| `.claude/agents/harness/` | **Moderate** | Not enumerated today (the target is the sibling `.claude/agents/moai`), but it is one `cleanTarget` entry away — which is exactly the AC-UDS-015 mutation. Survival depends on the target list, not on structural distance. |
| `.moai/harness/run-extension.md`, `.moai/harness/main.md` | **Trivial** | `CleanMoaiManagedPaths` never walks `.moai/harness/` at all — it touches only `.moai/config/` and `.moai/memory/`. Asserting their survival against this entry point is weak evidence; they survive because nothing looks at them. |

The guard keeps all three assertions (preserving the retired test's intent), but only the first two
carry discriminating power against `CleanMoaiManagedPaths`. The `.moai/harness/` rows would only
become load-bearing against an entry point that walks `.moai/` more broadly — `runCleanReinstall`,
which this guard does not drive.

#### Baseline-attribution

Measured on branch `feat/SPEC-UPDATE-DATA-SURVIVAL-001`, worktree `.claude/worktrees/e2-data-survival`,
HEAD `e8eeab462`, before any M5 edit (the four registry functions all at `0.0%` while the test
reported PASS — the defect):

```
$ grep -cE '^func simulateMoaiUpdate|^func simulate[A-Za-z]*Update' internal/cli/update_safety_test.go
1
$ go tool cover -func=/tmp/uds-m5.out | grep -E 'CleanMoaiManagedPaths|MigrateLegacyMemoryDir|runCleanReinstall|BackupMoaiConfig'
.../update/backup/backup.go:27:        BackupMoaiConfig        0.0%
.../update/deploy/deploy.go:29:        CleanMoaiManagedPaths   0.0%
.../update/deploy/deploy.go:144:       MigrateLegacyMemoryDir  0.0%
.../update_clean_install.go:137:       runCleanReinstall       0.0%
```

Post-M5 on the same tree: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
`go vet ./...` exit 0; `golangci-lint run --timeout=3m` → `0 issues.` (no NEW finding against the
0-issue baseline); `go test ./internal/cli/...` exit 0 (`ok .../internal/cli` + 16 sibling packages
`ok`); `go test ./...` exit 0 with `0` lines matching `^FAIL|^--- FAIL`
(`ok .../internal/cli 262.854s`).

Evidence logs persisted under `.moai/state/verify/m5/` (`ac014b-covered.txt`,
`ac014c-guard-pass.log`, `ac015-falsification.log`, `g1-build.log`, `g2-cli-test.log`,
`g3-full-test.log`, `g4-lint.log`, `g5-vet.log`, `g6-winbuild.log`).

**Mid-run tree movement.** Between the gate runs above and the M5 commit, a merge of `origin/main`
(`68be0a27c`, bringing `origin/main` tip `835ea7b91` in) landed on this branch from another actor.
It touched **2 files, both under `.github/workflows/`; zero `.go` files**
(`git show --name-only --format='' 68be0a27c | grep -c '\.go$'` → `0`), so the Go-toolchain gate
results measured at `e8eeab462` still describe the committed tree. Re-verified after the merge:
`go build ./...` exit 0, `--- PASS: TestMoaiUpdate_PreservesUserArea (0.02s)`, and AC-UDS-019 still
`0` against the (unchanged) merge-base `835ea7b9125bc34770384f626f27026e9eca64f1`. The full suite,
lint, vet, and windows cross-build were NOT re-run post-merge — see Gaps.

#### Gaps (not verified in M5)

- **Only ONE of the four registry functions is driven by intent.** `CleanMoaiManagedPaths` (66.7%)
  is the entry point; `MigrateLegacyMemoryDir` (26.9%) is covered only incidentally as its tail call,
  and on the no-op branch (`.moai/memory/` absent in the fixture, so it returns early). The other
  two — `runCleanReinstall` and `BackupMoaiConfig` — remain at `0.0%` under this guard.
  AC-UDS-014 (b) requires at least one above `0.0%`, so this satisfies the AC, but the protection is
  **narrower than "the update subsystem preserves user areas"** would suggest: it covers one
  destructive step of one path, not the clean-reinstall orchestration or the backup step.
- **The AC-UDS-015 baseline claim was not re-measured.** `acceptance.md` records that the same
  mutation passed against the pre-M5 test. Since M5 deletes that test, the baseline could not be
  re-run on this tree; it is carried from the acceptance record, not observed here. The *forward*
  direction (the replacement guard FAILS under the mutation) WAS observed directly.
- **`MigrateLegacyMemoryDir`'s destructive branch is unexercised.** The fixture creates no
  `.moai/memory/`, so the rename branch and the both-exist backup-then-remove branch (the REQ-UDS-008
  path M2/M3 addressed) are not touched by this guard.
- **The glob-widening mutation was not falsified.** The AC-UDS-015 mutation adds a new `cleanTarget`;
  it does not widen the existing `.claude/skills/moai*` glob. The genuine-risk claim for
  `.claude/skills/harness-ios-patterns/` above rests on reading the glob's scope, not on an observed
  failing mutation of the pattern itself.
- **Four of the seven gates were not re-run after the mid-run `origin/main` merge.** `go test ./...`,
  `golangci-lint`, `go vet`, and the windows cross-build were measured at `e8eeab462`, before merge
  commit `68be0a27c`. The merge contains zero `.go` files (measured), so their verdicts are expected
  to hold — but "expected to hold" is an inference from the merge's file list, not a re-observation.
  `go build` + the M5 guard WERE re-run post-merge and passed.
- **`.moai/config/` removal is not asserted.** The fixture creates `.moai/config/config.yaml` and
  the managed-removal assertion covers it, but no assertion distinguishes "removed by the unconditional
  `os.RemoveAll(configDir)`" from "removed by a cleanTarget" — the guard asserts absence, not mechanism.

#### Residual risk

- **The managed-removal assertion is satisfied by ANY one target being gone.** It loops over six
  managed paths and requires `removed > 0` plus per-path absence; a mutation that removed only a
  subset would be caught by the per-path `t.Errorf`, but a mutation that *reordered* or *renamed*
  targets while still deleting all six would pass. The assertion pins outcome, not target identity.
- **`snapshotDir` now treats a missing root as an empty map** (changed from `t.Fatalf`). This makes
  a deleted user directory surface as a readable snapshot mismatch instead of an unrelated
  "no such file" fatal — but it also means a fixture bug that silently fails to create a user
  directory would produce `pre == post == empty` and pass. The fixture writes are `t.Fatal`-guarded,
  so a creation failure aborts before the comparison; the risk is bounded by that guard, not
  eliminated by the snapshot function.
- **The guard runs against `t.TempDir()`, so it never observes the real project tree.** It cannot
  detect a production path-resolution bug that only manifests against an absolute project root
  (the `filepath.Join(cwd, absPath)` hazard documented in CLAUDE.local.md §6). That class of defect
  is out of this guard's reach.
- **Package-level state: none.** Unlike M4's `userHomeDirFn`, this guard mutates no package-level
  variable and takes `projectRoot` as a parameter, so it carries no parallel-test hazard. Verified
  by inspection of the rewritten file: no `t.Setenv`, no package-var assignment.

### M6 — `mergeBackPreserveInventory` partial-restore reporting and coverage

Landed on `feat/SPEC-UPDATE-DATA-SURVIVAL-001`. Files: `internal/cli/update_preserve_inventory.go`
(three failure returns + a `restored` counter), `internal/cli/update_preserve_partial_test.go` (new).

#### Claim

`mergeBackPreserveInventory` now reports the boundary of a partial restore: each of its three
failure returns names the file it stopped at AND the count of files already restored
(`restored %d/%d before failure`). The three branches — stat, `MkdirAll`, `copyFile` — are each
reached deterministically by a subtest asserting a distinct error substring, and the guard fails
against the pre-M6 production code.

Authoring note: the M6 delegation was interrupted mid-run by a session limit after the code and
test were written but before commit and before this section was authored. The orchestrator
verified the surviving working-tree artifacts independently and completed the milestone; every
row below is an orchestrator-observed measurement, not a carried-over agent claim.

#### Evidence — AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UDS-016 (branches) | PASS | `go test -run 'TestMergeBackPreserveInventory_PartialRestore' -count=1 -v ./internal/cli/` | `--- PASS` for all three subtests: `stat_failure_unreadable_backup_dir`, `mkdirall_failure_parent_is_regular_file`, `reports_restored_count`; `ok github.com/modu-ai/moai-adk/internal/cli 0.805s` |
| AC-UDS-016 (coverage) | PASS | `go test -covermode=set -coverprofile=/tmp/uds-cov-m6.out ./internal/cli/` then `go tool cover -func` filtered to `mergeBackPreserveInventory` | `update_preserve_inventory.go:400: mergeBackPreserveInventory 94.1%` — strictly greater than the 64.3% baseline |
| AC-UDS-017 | PASS | `go test -run 'TestMergeBackPreserveInventory_PartialRestore/reports_restored_count' -count=1 -v ./internal/cli/` | `--- PASS: TestMergeBackPreserveInventory_PartialRestore/reports_restored_count (0.18s)` — the subtest-path `--- PASS` line is present, so the `-run` selector matched (a bare `ok` with no matching line would be a vacuous pass) |
| AC-UDS-018 | PASS | `go build ./...` / `GOOS=windows GOARCH=amd64 go build ./...` | `build-exit=0` / `win-build-exit=0` |
| AC-UDS-019 | PASS | `BASE=$(git merge-base origin/main HEAD); git diff --name-only $BASE..HEAD -- internal/template/templates/ \| wc -l` | `0` (merge-base `835ea7b9125bc34770384f626f27026e9eca64f1`); uncommitted template edits also `0` |

#### Evidence — falsification

No AC required an overlay falsification for M6, but the guard's non-vacuity was verified anyway,
because a test that asserts a substring the production code never emitted would be the exact
failure mode this SPEC exists to remove. The pre-M6 `update_preserve_inventory.go` (from `HEAD`)
was overlaid under the M6 test:

```
$ git show HEAD:internal/cli/update_preserve_inventory.go > /tmp/uds-verify-m6/reverted.go
$ go test -overlay=/tmp/uds-verify-m6/overlay.json \
    -run 'TestMergeBackPreserveInventory_PartialRestore' -count=1 -v ./internal/cli/
    update_preserve_partial_test.go:92:  error does not report the already-restored count:
    update_preserve_partial_test.go:109: error does not report the already-restored count:
    update_preserve_partial_test.go:129: error does not report the already-restored count:
--- FAIL: TestMergeBackPreserveInventory_PartialRestore (0.21s)
    --- FAIL: .../stat_failure_unreadable_backup_dir (0.15s)
    --- FAIL: .../mkdirall_failure_parent_is_regular_file (0.05s)
    --- FAIL: .../reports_restored_count (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.847s
```

All three subtests fail against the pre-M6 code, each for the counter-specific reason. The guard
depends on the M6 change; it cannot pass without it. `/tmp/uds-verify-m6` was removed after the run.

#### Baseline-attribution

Measured in worktree `.claude/worktrees/e2-data-survival`, branch
`feat/SPEC-UPDATE-DATA-SURVIVAL-001`, HEAD `6ad2f9fa3`, before any M6 edit:

```
$ go test -covermode=set -coverprofile=/tmp/uds-cov-base.out ./internal/cli/ && \
    go tool cover -func=/tmp/uds-cov-base.out | grep mergeBackPreserveInventory
.../internal/cli/update_preserve_inventory.go:400:	mergeBackPreserveInventory		64.3%
```

The uncovered blocks were the three failure returns at `:416` (stat), `:420` (`MkdirAll`), and
`:424` (`copyFile`). No pre-existing test reached any of them — `update_preserve_inventory_test.go`
covers only the success path (`:211`) and the empty-`backupDir` guard (`:238`).

Regression gates re-measured on the M6 tree: `golangci-lint run --timeout=3m` → `0 issues.`;
`go vet ./...` → exit 0; `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` →
exit 0; `go test ./...` → exit 0, `0` lines matching `^FAIL|^--- FAIL`.

#### Counter semantics

`restored` counts entries **actually written to the project**. A `continue`d entry — one missing
from the backup, which the function skips by design as a defensive race guard — does NOT increment
the counter, because it was not restored. The reported figure is therefore "files successfully
restored before the failure", not "inventory entries traversed". `total` is `len(inv.Files)`, the
full inventory size, so `restored 1/2` reads as "1 of 2 inventory entries restored before stopping".

#### Gaps (not verified in M6)

- **The stat subtest is platform-skipped.** `stat_failure_unreadable_backup_dir` skips on Windows
  (permission-bit denial is not modelled there) and when running as root (root bypasses permission
  bits, so the branch would not be reached and the subtest would pass for the wrong reason). On the
  Windows CI job that branch is therefore **unverified** — only the `MkdirAll` and `copyFile`
  subtests run on every platform. This is a deliberate trade: the alternative (an injectable
  `os.Stat` seam) would add a package-level test-only variable the SPEC does not require.
- **The `continue` branch (backup missing, `os.ErrNotExist`) is not covered by these subtests.**
  It is the one path through the loop that neither restores nor fails; no AC requires it.
- **Residual uncovered blocks at 94.1%.** The function is not at 100%; the remaining uncovered
  statements were not enumerated. The AC requires "strictly greater than 64.3%", which is met.
- **The error-message wording is not asserted verbatim.** The subtests assert substrings
  (`"stat backup blocked/second.md"`, `"restored 1/2"`), not the full formatted string, so a
  reformatting that preserved both substrings would pass.

#### Residual risk

- **`restored++` placement is load-bearing and untested in isolation.** The counter increments only
  after `copyFile` succeeds. If a future edit moved it above the copy, `restored` would over-count
  by one and the subtests would fail (they assert `1/2`, not merely a non-zero count) — so the
  guard does cover that specific regression, but only at the `1/2` fixture size. A larger
  off-by-one introduced elsewhere in the loop would not necessarily surface.
- **The stat subtest mutates filesystem permissions.** It `chmod 0o000`s a directory under
  `t.TempDir()` and restores the mode in `t.Cleanup`. If the test panics between the chmod and the
  cleanup registration, `t.TempDir()` cleanup would fail to descend into the unreadable directory.
  The registration is immediately after the chmod, so the window is one statement wide.
- **Shared-branch concurrency.** A merge from another actor landed on this branch during M5
  (`68be0a27c`, two `.github/workflows/` YAML files, zero `.go` files). The M6 delegation was
  itself interrupted mid-run. Both are evidence that this branch is not exclusively held; a merge
  landing between the final verification and a future push would re-open the same window.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-08-01

All six milestones are complete and committed on `feat/SPEC-UPDATE-DATA-SURVIVAL-001`:

| Milestone | Subject | Covers |
|---|---|---|
| M1 | Failure contract — recovery manifest + restore entry point | REQ-UDS-019 ~ 025 |
| M2 | Destructive-target registry + `.moai/memory/` backup + drift guard | REQ-UDS-006 ~ 010 |
| M3 | On-disk backup for the three in-memory-only files + failure abort | REQ-UDS-001 ~ 005 |
| M4 | HOME deletion-radius pinning via the `userHomeDirFn` seam | REQ-UDS-011 ~ 014 |
| M5 | Non-vacuous user-area safety guard (fake deleted, real entry point driven) | REQ-UDS-015 ~ 018 |
| M6 | Partial-restore reporting + failure-branch coverage | REQ-UDS-026 ~ 028 |

Three of the six milestones (M4, M5, M6) carry a demonstrated falsification — the guard was
observed FAILING against a mutation or against the pre-change production code — so their
non-vacuity is evidence, not assertion. M3's abort path is proven by a call-recording spy rather
than an inferred unchanged tree.

Cross-cutting gates on the final tree: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64
go build ./...` exit 0 (AC-UDS-018); `go test ./...` exit 0 with zero `FAIL` lines;
`golangci-lint run` → `0 issues.`; `go vet ./...` exit 0; template neutrality `0`
against merge-base `835ea7b91` (AC-UDS-019).

Known carried debt, stated for the sync-phase auditor rather than buried: the M5 guard drives one
of the four registry functions (the other three remain at `0.0%` under it); AC-UDS-013's
real-HOME diff is trivially satisfied on a machine without `~/.claude/hooks`; three
`userHomeDir()` call sites outside `ensureGlobalSettingsEnv` remain unrouted; and the M6 stat
branch is unverified on Windows. Each is recorded in its milestone's Gaps section above.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters (M3):

- tier: M (3 artifacts; plan.md §H)
- scope (file count): ~6-8 (`update_template_sync.go`, `update_clean_install.go`, one new
  `internal/cli/update/backup/*.go` on-disk writer, plus 2-3 new `_test.go` files)
- domain count: 1 (Go source in the `internal/cli` update subsystem)
- file language mix: 100% Go
- concurrency benefit: LOW — coding-heavy, sequential milestone with a shared call path across
  two files that both mutate the same backup mechanism

Mode evaluation:

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 `trivial` | no | New production code + failure-path contract; not a typo-class change |
| 2 `background` | no | Write-capable implementation work, not read-only analysis |
| 3 `agent-team` | no | RETIRED (tombstone) |
| 4 `parallel` | no | Single domain (<3) and coding-heavy — Anthropic coding-task parallelism caveat |
| 5 `sub-agent` | **yes** | Default fallback; coding-heavy single-domain milestone |
| 6 `workflow` | no | Far below the ~30-file mechanical-transform threshold; not a uniform transform |

Decision: sub-agent

Justification: M3 is coding-heavy work in one domain (`internal/cli` update subsystem) touching
both execution paths through a single shared on-disk backup mechanism. Anthropic's coding-task
parallelism caveat makes sequential sub-agent delegation the correct default; the milestone's
failure-path contract (AC-UDS-020 clause 2, the call-recording spy) requires one coherent author
rather than a fan-out. Mode 6 is excluded on scope (~6-8 files vs the ~30-file threshold) and on
transformation kind (new behaviour, not a uniform mechanical rewrite).
