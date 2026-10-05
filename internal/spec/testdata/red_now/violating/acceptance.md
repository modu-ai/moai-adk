# FIXTURE (violating) — a prose-only RED cell, and a laundered malformed command

This fixture is not a SPEC. It carries exactly two planted defects:

1. `AC-VIO-001` is release-blocking and its RED-now cell is prose only — no
   command, no stdout, no exit code, no pin. This is the element violation.
2. `E-02` is a piped command, so it is not a single invocation. It is cited by a
   **regression-guard** criterion, which is the class-laundering mutant (M-4): a
   check scoped to release-blocking criteria never reaches it.

Document-level tree pin: `deadbee1234`.

## D.0 Evidence ledger

```
E-01  ls internal/spec/testdata/red_now/violating/absent.txt
      stdout: (empty)
      exit:   1

E-02  grep -c "alpha" internal/spec/testdata/red_now/violating/target.txt | wc -l
      stdout: 1
      exit:   0
```

## D. AC Matrix

| AC | Class | Verifies | Scenario | RED-now proof | Green path |
|----|-------|----------|----------|---------------|------------|
| **AC-VIO-001** | release-blocking | REQ-VIO-001 | **Given** the tree **when** the check runs **then** it holds. | The check would obviously be red today, because the feature has not been built. | M1 lands it. |
| **AC-VIO-002** | release-blocking | REQ-VIO-002 | **Given** the tree **when** the artifact is looked up **then** it exists. | **E-01** — the artifact does not exist. | M1 creates it. |
| **AC-VIO-003** | **regression-guard** — *RED cell: none, deliberately* | REQ-VIO-003 | **Given** the tree **when** it is counted **then** the count still holds. | No RED. Measured green: **E-02**. | E-02 unchanged. |
