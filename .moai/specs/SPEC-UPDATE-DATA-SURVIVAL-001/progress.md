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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
