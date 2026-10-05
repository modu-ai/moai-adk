# SPEC-MRGA-001 — Acceptance Criteria (bare marker)

## PRESERVE criteria

Fixture (c) of AC-MRG-003. The marker is present but its reason is empty, so it
does NOT suppress: the rule reports the marker as incomplete instead. A bare
marker would make silencing cheaper than fixing, which inverts the incentive the
exemption exists to set.

| AC | Command | Expected |
|---|---|---|
<!-- moving-ref-ok: -->
| AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) |
