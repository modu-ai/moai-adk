# Decision index: SPEC-CARD-AGE-ALARM-001

Revision 0.1.0, authored at plan phase for card t1838 (Tier M). The decision gate is on (`.moai/config/sections/interview.yaml:6`, `interview.decision_gate: on`), so this index is owed alongside the Tier artifacts. It carries no status field: the lifecycle lives in `spec.md` alone.

The card is an operator delegation ("운영자 위임 10-17") whose design input is the leader-authored redesign doc at `.moai/reports/card-policy-redesign-20261017/design.md`. That doc is **not committed in this branch's tree** (measured: absent from the worktree at base `7faf75fdf`; it exists in the primary checkout only). It is therefore cited below as the card's normative input, never as a committed authority anchor — every DECIDED row below is anchored instead on a measurement of committed code at `7faf75fdf`, which the row names.

## How to read a row

- **Label** is one of `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, or `FOUNDER`.
- **Class** is `product-level` or `implementation-level`. A row with no Class line is product-level.
- **Authority anchor** names a committed file and section (measured at the tree this SPEC branches from), or states that none was found.
- **Why unresolved** states what the operator must settle.
- **Default** and **Alternate** appear only where a ranking exists.

An open FOUNDER row blocks the autonomous Kickoff until its `Operator verdict:` line is written. No row below is an open FOUNDER: Q1-Q4 are settled by committed evidence or policy, and Q5 records the values question with its confirmation surface named (the OFF-by-default enable flag) rather than leaving it open — see the row for the exact residual that remains the operator's to override.

## Summary

| Row | Subject | Label | Class | Operator verdict: line |
|---|---|---|---|---|
| Q1 | Gate-wait age anchor: events table, `updated_at`, or a new column | DECIDED | implementation-level | not required (DECIDED by committed schema measurement) |
| Q2 | Last-commit signal source: render-time git read vs a recorded store timestamp | DECIDED | implementation-level | not required (DECIDED by committed schema measurement + G1 boundary) |
| Q3 | Invalid alarm key: degrade the section, or refuse the status command | POLICY-COVERED | implementation-level | not required (POLICY-COVERED) |
| Q4 | Alarm surfaces: which outputs render the alarm | DECIDED | implementation-level | not required (DECIDED by committed surface measurement) |
| Q5 | Default band values and the OFF flag | DECIDED | product-level | not required at plan close — see row (values become binding only at enablement; a config edit overrides them with no code change) |

## Routing

1. **FOUNDER rows with recorded operator verdicts:** none.
2. **DEFAULT-APPLIED at plan close:** none.
3. **POLICY-COVERED, no verdict needed:** Q3.
4. **DECIDED:** Q1, Q2, Q4 (committed-evidence anchors) and Q5 (see the row's own residual statement).

## Rows

### Q1: Where does the gate-wait age anchor come from — the events table, `cards.updated_at`, or a new `gate_entered_at` column?

Label: DECIDED
Class: implementation-level
Authority anchor: `internal/homestate/factory.go:79-85` (the events DDL — append-only, one row per record act) and `internal/homestate/card_transition.go:901-908` (`appendEvent`). The disqualifying counter-anchor for `updated_at` alone: `internal/homestate/card_picked.go:197` — field-only writes append a `card.fields` event and stamp `updated_at` with no transition, so `updated_at` is a generic write timestamp, not a state-entry timestamp. A new column is disqualified by the design's row-preserving stage discipline (design doc §6: "row-preserving migration" is for stages that must migrate; G1 adds no state).
Why resolved: the events table already records the exact entry moment of every state — every transition appends `card.transition` with payload `to` and `created_at`, creation included (`card_picked.go:173`), and lazy expiry into `blocked` appends `lease.expired` whose payload also carries `to` (`card_transition.go:434-438`). Deriving the anchor from events is exact and survives future non-transition writers; `updated_at` remains the fallback for legacy rows with no events (REQ-CAA-003).
Default: none needed (one option survives measurement).
Operator verdict: not required (DECIDED by committed schema measurement)

### Q2: Where does the last-commit signal come from — a render-time git read, or a timestamp recorded in the store?

Label: DECIDED
Class: implementation-level
Authority anchor: `internal/homestate/card_record.go:100-138` — the `Card` struct's only commit-shaped columns are `EvidenceSHA` (the single run→sync commit, T10) and `MergeSHA` (T16); no per-commit or last-commit timestamp exists. `internal/homestate/card_record.go:121` carries `WorktreePath`, the recorded pointer the git read resolves against.
Why resolved: exactly one recorded source of commit timestamps exists — git itself, against the card's recorded worktree. Recording a new store timestamp was rejected on two grounds: REQ-CAA-007 keeps G1 display-only, and the only write path that could carry it today is the lease-renewal path, which stage A (process-identity ownership) is about to redesign — growing it now would create migration surface for G1's successor stages. The render-time read satisfies the card's boundary as recorded-timestamps-only: it reads commit timestamps git already recorded; it probes no pid, session, or process identity (spec.md §Boundaries).
Default: none needed (measurement leaves one source).
Operator verdict: not required (DECIDED by committed schema measurement + G1 boundary)

### Q3: When an alarm key fails its bound, does the status command degrade the alarm section, or refuse the whole command?

Label: POLICY-COVERED
Class: implementation-level
Authority anchor: `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline ("Never let an advisory failure block the path") and the measured in-command precedent `internal/cli/factory_card.go:3131-3136` — the quota-classification read already degrades its cells rather than failing the status surface, explicitly so "the lease rows survive a queue read fault".
Why resolved: the alarm is an advisory block on the leader's situational read; a misconfigured display threshold must not blind the leader to the card rows (contrast the lease keys, `internal/homestate/factory_lease.go:14-18`, which refuse the factory OPEN because they govern writes — a different surface with a different failure cost). REQ-CAA-002 encodes degrade-with-one-warning-line.
Default: none needed (policy decides).
Operator verdict: not required (POLICY-COVERED)

### Q4: Which surfaces render the alarm?

Label: DECIDED
Class: implementation-level
Authority anchor: `internal/cli/factory_handoff_recover.go:56-61` — the factory subcommand registry, where `status` is the only card-level situational read; `internal/cli/mcp_server.go:619-628` — the MCP factory tools are `factory_msg_*` only, with no factory-status tool; a sweep for a separate leader-panel/board renderer (`상황판`, `situational`, `leader board`) over `internal/` found no code surface.
Why resolved: the leader's situational panel IS the `moai factory status` read (text for the glance, `--json` for programmatic consumption) — both surfaces named in REQ-CAA-006. Adding an MCP tool, a new command, or skill-body edits (the foreman and watchdog skills are mirrored templates) would grow G1 beyond display-only for zero added visibility; the alarm section is self-describing where the leader already looks.
Default: none needed (measurement leaves one surface pair).
Operator verdict: not required (DECIDED by committed surface measurement)

### Q5: What default band values ship, and does the enable flag default OFF?

Label: DECIDED
Class: product-level (the template `workflow.yaml` ships the block, so the values are template defaults)
Authority anchor for the OFF flag: the card body ("자동 조치 없음") and the design doc §6 stage discipline ("config flag default off"), which the delegation carries; for the value status, no committed authority fixes numbers — this row records that honestly.
Why resolved (shape): the flag ships OFF, so the values are inert until the operator enables the alarm; enablement is the operator's confirmation point. The four values — `gate_warn_after: 6h`, `gate_alert_after: 24h`, `stall_warn_after: 2h`, `stall_alert_after: 8h` — are UNMEASURED chosen starting points in the quota-gate tradition (`internal/config/defaults.go:89-98`): gate waits are human-scale (the incident inventory's leader-repair stalls ran ~5.5h; multi-day strandings exist), working-card stalls are machine-scale (a 15m lease with 5m renewal makes 2h without a stage move or commit notable, 8h serious). G1's stated purpose is to produce the calibration data that retunes them (design §6: G2 is "with G1 data"), so the defaults are the instrument's initial scale, not a commitment.
Residual left to the operator: if the operator prefers different bands before first enablement, editing `workflow.card_age_alarm.*` in the project config changes them with no code change; a preference recorded any time up to enablement overrides this row at zero cost. Surfaced to the leader in the plan-phase report.
Default: none recorded (no ranking rule applies to unmeasured calibration values).
Operator verdict: not required at plan close — see row (values become binding only at enablement; a config edit overrides them with no code change)
