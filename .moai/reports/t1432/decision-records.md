# t1432 — decision records (lane-12)

Append-only, one grep-able line per decision, in the form of `.claude/rules/moai/workflow/auto-semantics.md` §10. Self-attested by the writing lane; the sync audit re-reads these lines.

## D-1 plan→run Kickoff (autonomous, auto-semantics §9.1)

decision record: decided_by=lane-12 orchestrator (session 783a9ff3, claude-sonnet-5-5) evidence_refs=.moai/reports/t1432/plan-audit-iter3.md#AUDIT-VERDICT:PASS(0.92,tier-M-threshold-0.80,audited_sha=c2cc32080) + .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001/progress.md#plan_status:audit-ready ladder_path=gate-row:plan-run-kickoff;§9.1-autonomous;counter_refs=none searched=keepset

### The four §9.1 conditions, each measured in this run on this tree

| Condition | Observation (this run) |
|---|---|
| Independent plan-audit verdict is PASS | `plan-audit-iter3.md` verdict PASS, overall 0.92, no blocking finding, iteration 3 of 3, cold auditor `claude-sonnet-5-5`; iterations 1 (FAIL 0.75) and 2 (FAIL 0.87) are in `plan-audit.md` and `plan-audit-iter2.md` |
| Plan phase records audit-ready status | `progress.md` lines 7-8: `plan_status: audit-ready`, `plan_complete_at: 2026-10-03` |
| Plan-artifact hashes unchanged since the verdict | `git log --format=%h c2cc32080..HEAD -- .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001 \| wc -l` printed `0` at tree `25e65626f`; the audited commit is `c2cc32080` |
| No blocker open | iteration-3 report lists no blocking finding; the leader's four process points were accepted in a cross-session message on 2026-10-03 |

### Keep-set check (the operator form stays for these; none applies)

- Environment-impossible work: no. Everything the run needs is available (go toolchain, darwin). Windows runtime is unavailable and the SPEC claims no Windows runtime observation, only `GOOS=windows` build and vet.
- Operator-held work: no. The lockfile-contract options (D1-C, D2-B, D2-C) and appender locking (D3-A) are non-default and unselected; the leader relayed on 2026-10-03 that the defaults stand.
- Irreversible operation on an external shared system: no. The run edits files on the card branch only; nothing is pushed, merged or published by this lane before the pre-merge report.

### Debt carried from the audit (accept-as-debt, the sync audit re-reads these)

N1 DoD-1 wording, N2 stale "untracked" note in G-7 and plan B5, N3 a false statement that CI never compiles Windows tests (`ci.yml` runs `GOOS=windows go vet ./...`), N4 AC-HRH-005 (e) skip granularity, N5 no end-to-end case for a user-owned link to a root-owned target, N6 mutant overlays must be regenerated at M4 and M6, O4, O8, O9. N7 (overlay JSON pointing at session scratch) is not debt: it is regenerated before M0. The SPEC body is not edited for N1-N3, because an edit would change the plan-artifact hash and invalidate this decision.

### Limits of this record

- It is written on the card evidence path, not on the home-surface decision board named in auto-semantics §11. No CLI verb that writes an autonomous decision record to that store was found (`moai factory decide` records only an operator decision and accepts `--decider human` only, so it was not used). That is a Gap for the sync audit.
- `mcp__moai__factory_next` returned "no card available" for this lane and the factory record shows t1432 as `assigned`, owner `lane-12`, lease none. The leader confirmed on 2026-10-03 that a leader-direct dispatch carries no lease; stage and complete transitions are reported with their output if refused, never worked around.
