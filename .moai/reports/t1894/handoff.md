# t1894 — handoff (lane-31, 2026-10-19)

Written under the operator's standing rules f28a and a919 (finish the executing card, stop only new-card selection, hand off with a named follow-up owner) and the leader's safe-handoff criteria. This is a HOLD handoff at the plan stage, not a completion: nothing is merged, no PR exists, the plan has not passed audit.

## Card and tree

| Field | Value |
|---|---|
| Card / SPEC | t1894 / SPEC-CARD-MODE-JUDGMENT-001 (Tier L, Class C, Template-First, docs only; code is t1737's) |
| Worktree | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894` |
| Branch | `WT-fail-safe-serial` (base `74ad197380`) |
| Plan HEAD at the last audit | `cb71f2b475263dc8b0921e37836dcc8208a62f71` (SPEC files only) |
| HEAD after this handoff | the commit that adds `verdict iter2 + plan-hold.md + handoff.md`: read it with `git rev-parse HEAD` in the worktree; the lane reports the SHA to the leader in the same turn it is created |
| Commit chain | `55d4f1486a` plan artifacts, `cc109a75b6` Q3 ruling d314, `b7b44433c3` iteration 0 verdict + premise record, `4834700468` correction round 1, `d31fb64239` iteration 1 verdict, `cb71f2b475` correction round 2, then the handoff commit |
| Remote preservation | pushed to `origin/WT-fail-safe-serial` under board d-20261019T002431Z-18ba after the handoff commit (no PR; `git ls-remote --heads origin WT-fail-safe-serial` gives the SHA) |
| Uncommitted | none at the handoff commit; the tree is clean |

## Audit result

Plan audit rounds (Tier L ceiling 3, all used): 0.65 (9 blocking), 0.79 (4 blocking), 0.88 (1 blocking, D23), all FAIL because the required codex gate failed. `moai spec ceiling` prints: ceiling HIT at 3 rounds, disposition hold, split required. Reports: `plan-audit.md`, `plan-audit-iter1.md`, `plan-audit-iter2.md` (receipts `rcpt-39ba8947c14db9823fd47f7e`, `rcpt-2adbd9564960a6f7da294042`, `rcpt-55ffb196ab252308a7c4a3be`). Hold record and options: `plan-hold.md`.

## Unresolved

- **D23 BLOCKING**: the path half of AC-MJ-020 exempts `.moai/`, so a widening in a tracked `.moai/policies/**` or `.moai/workflows/gtd.md` copy passes every release-blocking criterion while four SPEC sentences claim otherwise. Fix and probes: `plan-audit-iter2.md` Recommendation and Operational Notes.
- D25 (optional, order-blind comparison; moved code fence), D24 (control cell E-013c depends on the grep engine).
- **Q7** (FOUNDER, empty): what establishes "the operator has delegated dispatch". Blocks only the autonomous form of the plan to run Kickoff.
- Run phase not started: the Template-First edits, the `make build` sync and the guard tests are Go work and wait for the resource slot order (leader order: t1885, t1899, t1876, the pre-PR scope tests, then the run cards), load at or below 32.

## Who decides what

The ceiling decision (exception, split, or accept) is the operator's, relayed by the leader. Options A/B/C and the lane's recommendation (A) are in `plan-hold.md`. Follow-up owner: the leader (decision), then a lane that leases t1894 again (lane-31 or a GLM lane).

## Lease, slot, renewal

- Lease: `moai factory renew t1894` last printed `renewed until 2026-10-19T06:30:16Z`; the lane stops renewing after this handoff and lets it lapse (board a672: a held card is not renewed, expiry is harmless). Re-lease with `moai factory next --card t1894 --run tn1p4e` from the parent checkout; if the lapsed record refuses both `renew` and `next`, the leader's reassign round-trip (done once at 05:1xZ) clears it.
- Slot: none held. No Go command, build or test was run for this card.
- Renewal cron `e87cd435` (session-only, t1894 only): deleted by the lane at this handoff. No other cron or scheduled job exists in this session.
- Running work: none (all spawned agents have returned; no writer is active).

## How to resume

1. Leader records the ceiling disposition on the board (exception / split / accept).
2. If exception A: lease t1894 again (above), `EnterWorktree` the path above, spawn `manager-spec` for the scoped correction (D23 plus D25/D24), commit with explicit pathspecs, recompute the hash with the MCP `audit_cache compute_hash`, spawn a new `plan-auditor` (file name `plan-audit-iter3.md`, one `audit_multi`, project_root = the worktree), then `moai spec ceiling` read-only. PASS: the Kickoff still needs the operator's Q7 answer; then run phase in slot order.
3. If split B: new cards per `plan-hold.md`; this card's SPEC stays as the record.

## Other cards on this lane (unchanged)

- t1865: final sync hold (`sync-hold-final.md`), origin `WT-update-ownership-repair` at `563745ea66`, lease lapsed, operator decision pending (F1/F2).
- t1723: closed under board 9558 (`closure.md`), origin `WT-cli-windows-guide` at `752948a5f`.
