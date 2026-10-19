# SPEC-V3R6-DOCS-PREREWRITE-SWEEPC-001 — Progress

## §A Premise record (plan phase, 2026-10-19, tree 74ad197380a0711cdb9816d4e37c94d29af730ea)

Measured premises of card t1723 (lane-31). Each row is a command run in this tree; the full premise check is `.moai/reports/t1723/premise-check.md` (local evidence, not committed).

| # | Premise | Observed | Reading |
|---|---------|----------|---------|
| P1 | The twelve target files exist | `git ls-files` expands the three pathspecs to 12 paths (acceptance.md E-SWC-1) | holds |
| P2 | Sources `child-cards.md` and `split-proposal.md` of the t1655 tree | neither exists; the t1655 tree is gone | does NOT hold (acceptance.md G1); nothing taken from them |
| P3 | Umbrella SPEC is the authority | `SPEC-V3R6-DOCS-PREREWRITE-CLEANUP-001` is in this tree (commit bdb7b764) | holds, under that id |
| P4 | Q13 and Q17 | open rows in the umbrella; operator verdict recorded on the board while this plan was written: `d-20261019T015908Z-220a`, 2026-10-19T01:59:08Z | both met for this child (decision-index.md, Gate) |
| P5 | "Different lines" from the README child | the three Korean E-01 lines are the Korean CG sentence README REQ-PRC-004 targets (acceptance.md E-SWC-8) | does NOT hold for those three lines |
| P6 | The matches are stale citations | 12 matches, all Class N (a retirement notice), 0 Class C | the row Q-SWC-1 is open |

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-19
tier: M
requirements: 5
acceptance_criteria: 6
open_rows: [Q-SWC-1]
open_items: [OI-1]
note: "ready for the plan audit; no audit has run on these artifacts, and Kickoff needs an operator verdict on Q-SWC-1"
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
