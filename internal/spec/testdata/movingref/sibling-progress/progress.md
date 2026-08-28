# SPEC-MRGA-001 — Progress

## §E.2 Run-phase Evidence

AC-MRG-009. This fixture's `spec.md` is clean; the flagged shape lives only here.
`SPECDoc.Body` carries spec.md alone, so a body-only rule would report nothing
against this SPEC while appearing to work.

- PRESERVE check: `git diff --name-only origin/develop -- internal/hook` | empty (unchanged)
