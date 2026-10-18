# Progress: SPEC-CARD-DURABLE-IDENTITY-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-CARD-DURABLE-IDENTITY-001
card: t1842
tier: L
base_sha: f4bcd4f2cc65342559b505f1e5a29df80f992894
spec_version: 0.2.0
req_count: 23
ac_count: 25
decision_rows: 7
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md
plan_status: audit-ready
plan_complete_at: 2026-10-18
release_target: v3.2.0
audit_ready: true
```

What the token means: the revised plan-phase artifact set is complete and ready for the next audit iteration. It is not an audit result. Plan audit iteration 1 returned FAIL 0.74 (`.moai/reports/t1842/plan-audit.md`, audited commit `1c0b594a0d7cca23a19e0256e38702ceb07f57dc`, hash `f48a8af397477554d847506010280f9db6a45ba9a5b084ec5922eec91af203a4`); this revision changed every hashed file, so that hash no longer binds the artifacts. **The new plan-artifact hash was not computed by the author**; the auditor computes it and the next verdict file carries it.

Plan-phase facts (manager-spec, card t1842, revision 0.2.0): 23 requirements and 25 acceptance criteria against the Tier L ceilings of 25 and 25 (at the ceiling for acceptance criteria; a further addition needs a tier-up or a split). Seven decision rows: Q1, Q3, Q4, Q5, Q6 carry `DEFAULT-APPLIED 2026-10-18T01:21:43Z`; Q2 and Q7 are `LANE-RULED` by lane-41 (decided_by=lane-41 claude; ladder_path=disk evidence: design document section 2 principle 6, plan-audit.md D3 and D10; operator away). The stamps were written after the last edit of every other hashed file.

### Revision record (after plan audit iteration 1)

- Per-defect dispositions D1 to D17: plan.md § J.
- Engine claims were re-probed under the project driver (`modernc.org/sqlite` v1.60.1, SQLite 3.53.4) in a scratch module, not on the command line: rollback with trigger names dropped and recreated (six distinct names, both tables protected), the `UPDATE OR REPLACE` journal guard, a column-adding migration beside the `cards` view, and the read-only open's sidecars (research.md § 6.1).
- Fresh live-database copy 2026-10-18T01:16:00Z (300 rows, 228 ids): strict conflict rule blocks 25 ids by owner labels and at most 27 with the lease rule; 0 ids have non-terminal rows in two active runs (research.md § 4.1).
- RED cells touched by the revision were run on the audited tree (`1c0b594a0`): E20a (empty, exit 1), E20b (empty, exit 1: the sweep is empty at RED, as intended), E25 (empty, exit 1). Earlier cells are unchanged.
- `moai spec lint --strict SPEC-CARD-DURABLE-IDENTITY-001`: `0 error(s), 0 warning(s)` plus one INFO `OwnershipTransitionUnmeasured` (the plan-phase commit has no agent trailer). A first run after the edits showed one `VacuousTestAssertion` warning at acceptance.md line 162 (a phrase in the E00 note); it was reworded and the rerun is clean.
- Traceability by a script of my own (`trace.py` in the scratchpad): 23 requirements, 0 uncovered, 0 orphan, 25 acceptance headings. Ordering verb (the plan-auditor definition's awk, copied to the scratchpad): `COLLECTED: 8 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7), 0 exit bindings, 35 ordering candidates`, no `CONFLICT:` line.
- Tool provenance: the lint ran on the installed `moai` v3.2.0-rc.36-t1822c (commit `ee981a932`), a strict ancestor of the tree HEAD; its results are advisory.

### Gaps

- No `go build` and no `go test` was run; no behavior of the planned code was exercised, and the four existing retire tests of AC-CDI-002 were not observed to pass.
- The driver probes used reduced schemas, not the 31-column `cards` DDL or the real triggers, and no cutover code exists to run; contention under `_txlock=immediate`, a live MCP server handle and the real write funnels are untested.
- Feasibility of REQ-CDI-023 (sealing inside the existing funnels) was established by reading `applyLeaseExpiry`, `commitTransition`, `ReassignCard` and `recordPickedTx` and by the single-transaction argument, not by running it; the operator `abandon` edge keeps the holder columns (`keepColumns`), so it ends no grant and seals nothing.
- The rule for the live-lease check of REQ-CDI-005 reads a clock at the cutover; whether a registered worker bound to a retired run exists today is a one-instant measurement (20 rows on the copy, all in the active run).
- Stage-A and card-t1852 facts are from unmerged worktrees and memory records; the stage-A iteration 1 to 3 defect bodies were read as headlines and opening sentences; cards t1835 and t1843 have no SPEC that was read (unchanged from revision 0.1.0).
- The `events` log stays mutable (spec R-10); displaced grants are preserved by the attempt archive only.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
