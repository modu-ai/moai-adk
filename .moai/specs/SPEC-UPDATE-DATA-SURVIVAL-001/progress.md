# SPEC-UPDATE-DATA-SURVIVAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]
code_baseline: d5336214e
worktree_head_at_revision: a8b42e112
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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
