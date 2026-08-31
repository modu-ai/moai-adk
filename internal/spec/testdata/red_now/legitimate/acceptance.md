# FIXTURE (legitimate) — every release-blocking cell carries the four elements

This fixture is not a SPEC. It exists so the pass direction of the form check is
observed rather than assumed.

Document-level tree pin: `deadbee1234`. It binds every cell below that carries no
pin of its own.

## D.0 Evidence ledger

```
E-01  grep -c "alpha" internal/spec/testdata/red_now/legitimate/target.txt
      stdout: 0
      exit:   1

E-02  ls internal/spec/testdata/red_now/legitimate/absent.txt
      stdout: (empty)
      exit:   1
```

## D. AC Matrix

| AC | Class | Verifies | Scenario | RED-now proof | Green path |
|----|-------|----------|----------|---------------|------------|
| **AC-LEG-001** | release-blocking | REQ-LEG-001 | **Given** the target **when** it is scanned **then** the token is present. | **E-01** — the token is absent, so the count is 0 at exit 1. | M1 adds the token; E-01 returns 1 at exit 0. |
| **AC-LEG-002** | release-blocking | REQ-LEG-002 | **Given** the tree **when** the artifact is looked up **then** it exists. | **E-02** — the artifact does not exist. | M1 creates it; E-02 exits 0. |
| **AC-LEG-003** | **regression-guard** — *RED cell: none, deliberately* | REQ-LEG-003 | **Given** the tree **when** it is scanned **then** the property still holds. | No RED, and no invented one. Measured green: **E-01**. | E-01 unchanged. |
