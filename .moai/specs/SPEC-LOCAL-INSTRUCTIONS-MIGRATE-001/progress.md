# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- **NOT audit-ready.** This SPEC was created by carve, not by a completed plan phase. Card
  **t1259** owns completing the plan phase and taking it to plan-audit.
- Created 2026-09-26 by carve from `SPEC-INSTRUCTION-FILES-UNIFY-001` at commit `1140bcd1d`, by
  operator decision relayed through the lead. Tier: M (provisional — 3 artifacts + progress.md:
  spec.md, plan.md, acceptance.md).
- SPEC ID regex check executed as Bash against the Go `specIDPattern`
  (`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`), output `PASS`. Directory-collision check against
  `.moai/specs/` returned `0`.
- Requirements: **9** (`REQ-IFU-007~012`, `020~022`; Tier M ceiling 16). Acceptance criteria:
  **7** (`AC-IFU-007`, `-011`, `-013`, `-014`, `-015`, `-023`, `-024`; ceiling 16). All
  transferred **verbatim** from the parent SPEC; nothing dropped, nothing renumbered. The `IFU`
  infix is retained so existing cross-references and audit citations still resolve; the id gaps
  are the carve's footprint.
- Status: `draft`. No implementation, no commits by the carving agent.
- Repairs applied during the carve, both from the plan-audit of `1140bcd1d`:
  - **D3** — `REQ-IFU-021` now names the migrating copy (the `develop`-committed
    `CLAUDE.local.md`, per that file's own §0.1 discriminant) and states one unit throughout
    (characters, `wc -m`). `AC-IFU-007` now requires the before-and-after pair so it cannot be
    discharged by measuring an already-compliant copy. Measured 2026-09-26:
    `git show origin/develop:CLAUDE.local.md | wc -m` → `44,381`, so a reduction of at least
    4,381 characters is required and the canonical copy does **not** already pass.
  - **Frontmatter** — canonical 12-field schema applied at birth (`tags` as a quoted
    comma-separated string, `module`, `lifecycle` present), so this SPEC parses from its first
    commit rather than repeating the parent's `ParseFailure`.
- **Open, and owned by t1259** (recorded so none of it is absorbed silently):
  - The Tier judgment is provisional.
  - No plan-audit has run against this SPEC.
  - `plan.md`'s milestones are transferred, not re-sequenced for this SPEC's own dependency
    order.
  - The two **recorded debts** (`AC-IFU-011`, `AC-IFU-015`) are preserved as authored. The Tier L
    ceiling that forced each fold no longer binds here, so the headroom to unfold exists — the
    decision is t1259's and is unspent.
  - Whether `AC-IFU-015` should be promoted to blocking: it now carries a data-integrity outcome
    (`moai update` mutating a user's own file) that no blocking criterion in this SPEC covers.
  - This SPEC has no whole-change CI criterion of its own; the parent's `AC-IFU-025` stayed with
    the parent.
  - The parent SPEC's `research.md` Q4 (whether the launcher's fallback branch has a diagnostic
    surface for the advisory) is unanswered and is M1's first task.
- **Sequencing constraint carried from the carve:** `SPEC-INSTRUCTION-FILES-UNIFY-001` M2 must
  land before this SPEC's M1, because both edit the local-instruction loop in
  `internal/cli/codex_launcher.go` — that SPEC changes its iteration order, this one adds the
  advisory on its fallback branch (plan.md §B).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
