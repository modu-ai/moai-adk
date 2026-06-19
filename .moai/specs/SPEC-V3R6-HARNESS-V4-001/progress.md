# progress.md — SPEC-V3R6-HARNESS-V4-001

> Plan-phase progress stub. The §E skeleton below carries the canonical 4 placeholder headings (§E.1 through §E.4) per the manager-spec plan-phase artifact protocol. Only §E.1 is populated at plan-phase; §E.2-§E.4 are empty placeholder headings awaiting run-phase / sync-phase population by their owning agents (manager-develop owns §E.2/§E.3; manager-docs owns §E.4).

## §A. Plan-Phase Status

- **Artifact set**: 5 artifacts authored (spec.md / plan.md / acceptance.md / design.md / research.md) + this progress.md stub.
- **Tier**: L (5-artifact set).
- **Era**: V3R6 (3-phase plan→run→sync; MX Tag cross-cutting per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`).
- **Frontmatter**: 12 canonical fields + `era: V3R6` + `tier: L` + `depends_on` + `related_specs`.
- **SPEC ID pre-write self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | HARNESS ✓ | V4 ✓ | 001 ✓ → PASS`.

## §B. Plan-Phase Audit-Ready Signal

- **Pending**: plan-auditor verdict (to be populated by plan-auditor sub-agent at plan-phase gate).
- **Target**: ≥ 0.80 (Tier L threshold).
- **GEARS compliance**: all 13 REQs use GEARS notation (Where/While/When compound clauses + generalized `<subject>`).
- **OutOfScopeRule**: §B.2 contains 5 `### Out of Scope — <topic>` H3 sub-headings each with `-` bullets.
- **Frontmatter schema**: 12 canonical fields present; `created`/`updated` (NOT `created_at`/`updated_at`); `tags` comma-separated string (NOT `labels` array); `version` quoted string.

## §C. Run-Phase Entry Preconditions (for manager-develop)

- [ ] plan-auditor verdict ≥ 0.80 at plan-phase gate.
- [ ] User Implementation Kickoff Approval obtained (CLAUDE.local.md §19.1 / REQ-ATR-015).
- [ ] BI-001 pre-flight: Claude Code subdirectory-command resolution verified.
- [ ] BI-002 pre-flight: `moai update` preserve-logic surface enumerated.
- [ ] BI-003 pre-flight: dynamic-workflow runtime version verified (v2.1.154+).
- [ ] Pre-spawn sync check: `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` clean.

## §D. Milestone Progress (run-phase)

| Milestone | Status | Commit SHA | Notes |
|-----------|--------|------------|-------|
| M1 — `/moai:harness` NL entry + §24 namespace extension | pending | — | entry point + namespace protection |
| M2 — Builder Workflow `harness-build.js` (4 phases) | pending | — | core dynamic-workflow |
| M3 — manifest.json schema + Runner primitive-mapping | pending | — | manifest SSOT |
| M4 — `/harness:<name>` + lifecycle + orphan prevention | pending | — | execution + lifecycle |
| M5 — Conditional worktree-isolation | pending | — | sub-agent-granular |
| M6 — Migrate 4 specialists + legacy redirect + dogfooding | pending | — | migration + validation |

---

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-auditor verdict — to be populated at plan-phase gate>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
