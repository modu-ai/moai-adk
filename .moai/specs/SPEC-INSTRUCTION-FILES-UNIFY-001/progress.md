# SPEC-INSTRUCTION-FILES-UNIFY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: L (5 artifacts + progress.md). Artifacts written: spec.md, plan.md, acceptance.md,
  design.md, research.md.
- SPEC ID regex check executed as Bash, output `PASS`.
- Requirements: 25 (Tier L ceiling 25). Acceptance criteria: 25 (ceiling 25). Both at
  ceiling — any further requirement or criterion is a signal to split the SPEC, not to
  relax the budget.
- Status: `draft`. Plan phase only — no implementation, no commits by this agent.
- Run phase is blocked on card t1175 landing on develop (plan.md §B).
- Lead directives 1 and 2 (2026-09-26) applied at spec.md v0.2.0: the M0 P7 line-3
  ancestor-discovery observation is labelled synthetic-fixture / UNCONFIRMED throughout and
  serves as no premise; the `.tmpl` invariant (card t925) and the nested-sum 32,768-byte
  budget are REQ-IFU-024 / REQ-IFU-025 with their own criteria; the worktree duplicate-load
  is scoped out to card t1219.
- Two run-phase measurements are mandatory and carry criteria whichever way they come out:
  AC-IFU-021 (real linked worktree, ancestor discovery) and AC-IFU-022 (Codex discovery of
  `AGENTS.local.md`, with tail sentinels so silent truncation is observable).
- AC-IFU-022's command verb was verified in the plan phase, not inherited: `codex debug --help`
  against **codex-cli 0.157.0** lists `prompt-input` ("Render the model-visible prompt input
  list as JSON"). A run under a different codex-cli version re-confirms before relying on it.
- AC-IFU-004 asserts the Codex-discovered **filename set** recursively over the template tree,
  not a path the criterion already expects — a `-maxdepth 1` or path-presence form would keep
  passing after a rename back to `AGENTS.md`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
