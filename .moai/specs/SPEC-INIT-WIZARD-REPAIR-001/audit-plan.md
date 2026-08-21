# Plan-Audit — SPEC-INIT-WIZARD-REPAIR-001 (card t174)

Canonical full report: `.moai/reports/plan-audit/SPEC-INIT-WIZARD-REPAIR-001-review-1.md` (iteration 1/2, Tier M). This file carries the verdict surface for the SPEC directory.

**Verdict: FAIL — Overall score 1.0 (gate-driven FAIL: two must-pass failures override the aggregate).**

All ground-truth disposition claims were re-verified against the tree (WT-init-wizard-repair @ bb4dff662) and every one held — including wizard reachability (init.go:555/565), the 3 dead-wiring functions + the 2 additional dead links (init.go:322-327 validated-then-discarded; no opts→yaml consumer), the 3 flag-key readers + auto_merge's absent reader, init.go:210 unconditional overwrite, and both stale comments. RESTORE dispositions are justified on all items; "restored-but-unread, documented" is the right call for auto_merge.

## Blocking (must fix before run-phase)

1. **MP-7 clarification gate** — 2 unresolved `[NEEDS CLARIFICATION]` markers at plan.md:11-12 (USER-scope settings write; update-wizard prompt step). Lead must resolve both via AskUserQuestion and replace the markers with recorded decisions. Both recommendations ("wire it") are supported by the verified facts.
2. **MP-5 / D7-4** — SPEC-WT-DOC-001 has `status: archived` and is cited as design authority (spec.md:14/72/118, plan.md:12) without an explicit reconciliation statement. Capability verified live in-tree (this is administrative archive, not retirement) — add the one-sentence reconciliation to §6.
3. **progress.md §E.1** — add the literal `plan_status: audit-ready` + `plan_complete_at` fields (plan→run gate precondition).

## Optional (non-blocking)

- acceptance.md:3 "byte-identical" preamble comparator vs AC-006/AC-009's template-baseline — align wording.
- REQ-005 `Where` → `When` (semantic nicety; structural GERS match holds).
- frontmatter `related_specs` is a non-schema extra field (harmless).

Tier M ceilings respected (9 REQ / 10 AC ≤ 16/16). Re-audit (iteration 2, final for Tier M) is scoped to the delta above.
