# SPEC-AUDIT-EXPORT-CLAUSE-001 — progress

Card: t1059 · Branch: `WT-audit-export-clause` · Evidence: `.moai/reports/t1059/`

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-22
tier: M
artifacts: spec.md, plan.md, acceptance.md
spec_version: 0.3.1
```

v0.3.1 lands two bounded lead-review additions and reopens nothing else: the
`check-ignore` discriminant is promoted from a plan instrument note to SPEC §A.2a
with REQ-AEC-014 + AC-AEC-015, and §E records the basis for the 13-file count at
the Tier M band edge. Requirements 13 → 14, criteria 14 → 15, both absorbed out
of existing headroom. AC-AEC-015 is adopted as a **regression-guard**, not a
release-blocking criterion: it is vacuous against the current tree (0 matches
pre-implementation) and its observed red was taken against the v0.2.0 draft.

Plan-phase artifacts authored by manager-spec from card t1059. v0.3.0 rewrites
them for **direction A** — the operator ruled the 2026-09-14 directive canonical,
so lead-verdict evidence stays on disk and the `git add -f` permission clause is
withdrawn rather than repaired. The card now also enacts the ruling by withdrawing
the `.gitignore` negation, and absorbs the paraphrase surfaces v0.2.0 excluded.

All SPEC §A measurements were re-taken in this worktree at HEAD `64c7edbf3`
(local `develop` absorbed); none were carried over from the dispatch or from the
v0.2.0 draft. Two did not survive re-measurement and are recorded as such: the
paraphrase phrase `tracked** path` no longer exists in this tree (SPEC §A.6), and
`git check-ignore -v` was found to exit 0 on a negation match (plan §E).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
