# SPEC-CAPACITY-RECLAIM-001 — Progress

status: draft
plan_status: audit-ready
plan_complete_at: 2026-10-14T02:03:04Z

## §E.1 Plan-phase Audit-Ready Signal

Plan artifacts (spec.md, plan.md, acceptance.md, decision-index.md) are revised to SPEC v0.2.0 on card worktree HEAD `e2658f30127375e75e0c606d09ca97be2c551cf2` (short `e2658f3012`, branch `WT-capacity-reclaim`). This is the FINAL authorized repair round: leader decision d-20261010T080627Z-987a allows exactly one repair inside fix_scope and one final delta re-audit after it; a FAIL on that audit returns the card to the operator.

Basis. The revision answers the iter3 verdict (`.moai/reports/t1631/plan-audit-iter3.md`: FAIL 0.56, 11 blocking, MP-8) restricted to card A's subset of the operator's narrowed scope (decision recorded on the card queue record, 2026-10-10): D1 (AC-CAP-008's approved-branch four-element cell EV-007, command unescaped, verbatim output, exit per the §0 rule, tree pin), D2 (dissolved — its criterion moved to card B with the sweep scope), D3 (AC-CAP-004 restored to RB with a four-element cell; the G-EXIT premise withdrawn), D4 (the §0 exit-capture rule restated from this session's runner observations; the `bash -c` wrapper withdrawn; no "runner-captured" label without a runner-printed status), D7 (plan §D.1 rotation mechanics deleted; retention ownership and the ordering guarantee stated instead), D8 (new AC-CAP-009: the archive-before-rewrite characterization test on the default branch), D9 (new AC-CAP-012: the SessionStart split behavior), D11 (§F absent-key residual rewritten: no notice exists in that case). D5, D6, and D10 left this SPEC by excision, not by rewrite: the worktree-sweep scope (disposal narrowing, last-activity field, report fields, `moai worktree list`, bulk-run evidence, 10 GiB target) moved to card B; the doctor cap read-path (`internal/cli/doctor_disk.go:155`, `:261`) and the compiled-constant removal moved to card D; the inbox rotation moved to card t1653.

Absorb and pins. The card worktree merged local `main` (merge commit `e2658f3012`, 689 commits, zero conflicts, tree clean) and every pin this SPEC carries was re-measured on that tree in this session. Of the measured paths, only `internal/config/defaults.go` (constants moved 451→456, 455→460) and the template `workflow.yaml` (hygiene mode line 318→331) differ from the iter3-audited tree `ab9f5a6d3`; all other cited lines reproduce unchanged. The template `.gitignore` still holds no rule for the state files or archives; the five state files are still tracked, now totalling 132,585,731 B across 191 files.

Scope addition flagged for operator ratification. The merged tree carries a new tracked monthly archive (`learning-history/archive/2026-09.jsonl.gz`, 2,142,019 B; six `.gz` archives now tracked, 4,929,798 B total), which pushed the post-untracking residual to 5,230,544 B — above the operator's 5,000,000-byte completion criterion 1. REQ-CAP-010 (new) adds the archive untracking (local copies kept; template rule `.moai/harness/learning-history/archive/*.jsonl.gz`), restoring the criterion (residual 300,746 B). This widens the M2 untracking set from five to eleven files and was not in the operator's enumerated list; it is forced by the operator's own PASS definition on the merged tree and is recorded in spec §A v0.2.0, §E A14, and §F.

Q3. The clone-measurement gate is satisfied: the per-clone recipe was executed on 2026-10-10 and recorded at `.moai/reports/t1631/q3-clone-measurement.md`; the untracking is authorized by the narrowed-scope decision. B1 lands through the card's normal landing path.

Counts. 10 requirements (REQ-CAP-001–010) and 13 acceptance criteria (AC-CAP-001–013), against the Tier M ceilings of 16 and 16. AC renumbering and the v0.1.2→v0.2.0 mapping are recorded in acceptance.md §0.

v0.2.1 — the authorized one-clause D12 repair (operator decision d-20261014T025128Z-0aa1, kind ceiling-exception: repair only D12, then exactly one re-audit; another FAIL returns the card to hold): acceptance.md AC-CAP-012 clause (1) now seeds one `verify` residue-class target in the fixture project and requires its `"unit":"gc"` / `"mode":"apply"` audit row, making the cache/GC apply half observable (D12, `.moai/reports/t1631/plan-audit-final.md`); premise cells EV-015 (no hook fixture seeds the class) and EV-016 (the GC's decision-gated apply-row write, `gc.go:395–404`) were observed on this tree `111493b71`, EV-008 re-observed there unchanged, and the evidence carriers they cite were updated accordingly. Nothing outside the clause changed: REQ-CAP-010 and every other REQ, AC, and milestone are untouched, no `internal/` code changed, and the REQ-CAP-010 ratification remains with the operator. The re-audit is the next act.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
