# SPEC-MRGA-001 — Acceptance Criteria (R3: declared exemption)

## PRESERVE criteria

Fixture (b) of AC-MRG-003. The marker sits on the line immediately above the
flagged row and carries a non-empty reason, so the finding is suppressed.

The reason is modelled on the real subject-class corpus line `AC-COORD-016`
(`spec.md` §B.3): the ref token there is quoted subject matter, not a
measurement anchor.

| AC | Command | Expected |
|---|---|---|
<!-- moving-ref-ok: the command string is the subject, not an anchor -->
| AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) |
