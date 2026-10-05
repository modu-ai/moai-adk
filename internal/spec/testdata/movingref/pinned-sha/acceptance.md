# SPEC-MRGA-001 — Acceptance Criteria (R1: pinned)

## PRESERVE criteria

The AC-MRG-001 row with the anchor's resolved value recorded on the same line —
the R1 remedy, and exactly the shape filter 4 of `spec.md` §B.3 removes (a line in
a git-command context that carries a 7-40 hex SHA).

The moving-ref token is deliberately RETAINED. A fixture that deleted it would
fail conjuncts 1+2 instead of being exempted by REQ-MRG-008, and the criterion's
stated mutation (delete the hex-SHA exclusion branch) could then not turn it red.

| AC | Command | Expected |
|---|---|---|
| AC-X | `git diff --name-only origin/main -- internal/` at d566ecc7511e1954e3aeb1dff3a60afa5be1089b | empty (unchanged) |
