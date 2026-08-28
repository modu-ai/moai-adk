# SPEC-MRGA-001 — Acceptance Criteria

## PRESERVE criteria

The row below is the AC-MRG-001 true-positive shape: a moving ref inside a
git-command context, deciding an invariant claim, with no SHA pin and no
frozen-baseline variable.

| AC | Command | Expected |
|---|---|---|
| AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) |
