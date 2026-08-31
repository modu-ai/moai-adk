# FIXTURE (ledger) — every command lives in the ledger, none in a table cell

This fixture is not a SPEC. It is the carrier-relocation mutant (M-5): the
malformed command `E-02` appears in **no** table cell at all, so a form check
scoped to table cells finds nothing to check and reports zero findings. A
carrier-independent check reports it.

Document-level tree pin: `deadbee1234`.

## D.0 Evidence ledger

```
E-01  ls internal/spec/testdata/red_now/ledger/absent.txt
      stdout: (empty)
      exit:   1

E-02  grep -c "alpha" internal/spec/testdata/red_now/ledger/target.txt | head -1
      stdout: 0
      exit:   0
```

## D. AC Matrix

| AC | Class | Verifies | Scenario | RED-now proof | Green path |
|----|-------|----------|----------|---------------|------------|
| **AC-LDG-001** | release-blocking | REQ-LDG-001 | **Given** the tree **when** the artifact is looked up **then** it exists. | **E-01** — the artifact does not exist. | M1 creates it. |
| **AC-LDG-002** | release-blocking | REQ-LDG-002 | **Given** the target **when** it is counted **then** the count is 1. | **E-02** — the count is 0 today. | M1 lands the token. |
