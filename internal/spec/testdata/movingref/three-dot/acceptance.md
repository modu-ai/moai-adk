# SPEC-MRGA-001 — Acceptance Criteria (three-dot form)

## PRESERVE criteria

AC-MRG-004. The three-dot form is NOT a safe exemption: `spec.md` §B.2 measured
the identical wrong answer from both the two-dot and three-dot forms in this
tree, because merge-base is not stable under upstream advance.

| AC | Command | Expected |
|---|---|---|
| AC-X | `git diff --stat origin/main...HEAD -- internal/` | unchanged |
