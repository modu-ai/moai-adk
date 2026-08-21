# SPEC-AGENTS-MD-CANON-001 — Research

## Sources

- `.moai/reports/t82/measurement.md` — the measured baseline (surface bytes, token estimate, codex
  0.147.0 binary-symbol observations, t91 status). Primary source for `spec.md` §A.1.
- `internal/config/token_budget_guard.go` — the surface definition (`alwaysLoadedSurface()`), the
  budget constant, and the comment recording the 75,000 → 76,000 temporary raise.
- `CLAUDE.md` §9 — live evidence that `@`-imports resolve in this repo.
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/` — closed; source of the inherited stub + lazy-companion
  pattern and the budget guard.

## What was measured for this SPEC

The contract-layer sizing in `spec.md` §A.2 and `design.md` §1 is original to this SPEC. Commands
and outputs are recorded in `progress.md` §E.1.

Headline: the verbatim `[HARD]` contract across the always-loaded rules and `CLAUDE.md` is
**32,543 B** — 4.0× the card's ~8 KiB root target, and 99.3 % of a presumed 32,768 B shared cap.

## What was NOT measured (carried as preconditions)

Three premises belong to card t91 (M0), whose report does not exist:

1. the real default of `project_doc_max_bytes`;
2. the merge scope of nested `AGENTS.md`;
3. whether truncation is visible or silent.

A fourth is added here: whether `project_doc_max_bytes` is settable from project scope rather than
only per-user. All four are `spec.md` §D.1 entry gates.

No codex model invocation was performed. Every codex claim in these artifacts rests on binary
symbol observation, per the measurement report's own stated limit.

## Prior-art note

`SPEC-ALWAYS-LOADED-DIET-001` demonstrated the pattern this SPEC scales up: `goal-directive.md`
went to a 6,531 B stub with a 17,334 B lazy companion — a 72 % always-loaded reduction with no
obligation moved off the always-loaded surface. That precedent is why REQ-AMC-002 is stated as a
constraint rather than an aspiration: the pattern already exists and already respects it.
