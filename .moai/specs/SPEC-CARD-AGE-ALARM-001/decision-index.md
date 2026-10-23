# Decision index: SPEC-CARD-AGE-ALARM-001

Revision 0.1.2 — rows Q1-Q5 authored at plan phase (0.1.0); rows Q6-Q8 record the plan-audit iteration 1 repairs (verdict `.moai/reports/t1838/plan-audit-iter1.md`); rows Q9-Q10 record the plan-audit iteration 2 rule-C correction (verdict `.moai/reports/t1838/plan-audit-iter2.md`), for card t1838 (Tier M). The decision gate is on (`.moai/config/sections/interview.yaml:6`, `interview.decision_gate: on`), so this index is owed alongside the Tier artifacts. It carries no status field: the lifecycle lives in `spec.md` alone.

The card is an operator delegation ("운영자 위임 10-17") whose design input is the leader-authored redesign doc at `.moai/reports/card-policy-redesign-20261017/design.md`. That doc is **not committed in this branch's tree** (measured: absent from the worktree at base `7faf75fdf`; it exists in the primary checkout only). It is therefore cited below as the card's normative input, never as a committed authority anchor — every DECIDED row below is anchored instead on a measurement of committed code at `7faf75fdf`, which the row names.

## How to read a row

- **Label** is one of `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, or `FOUNDER`.
- **Class** is `product-level` or `implementation-level`. A row with no Class line is product-level.
- **Authority anchor** names a committed file and section (measured at the tree this SPEC branches from), or states that none was found.
- **Why unresolved** states what the operator must settle.
- **Default** and **Alternate** appear only where a ranking exists.

An open FOUNDER row blocks the autonomous Kickoff until its `Operator verdict:` line is written. No row below is an open FOUNDER: Q1-Q4 are settled by committed evidence or policy, Q5 records the values question with its confirmation surface named (the OFF-by-default enable flag) rather than leaving it open — see the row for the exact residual that remains the operator's to override — and Q6-Q8 (added at the iter1 repair) are settled by measured substrate or committed precedent, and Q9-Q10 (added at the iter2 rule-C correction) are settled by audit findings reproduced through the real card API plus committed substrate and precedent.

## Summary

| Row | Subject | Label | Class | Operator verdict: line |
|---|---|---|---|---|
| Q1 | Gate-wait age anchor: events table, `updated_at`, or a new column | DECIDED | implementation-level | not required (DECIDED by committed schema measurement) |
| Q2 | Last-commit signal source: render-time git read vs a recorded store timestamp | DECIDED | implementation-level | not required (DECIDED by committed schema measurement + G1 boundary) |
| Q3 | Invalid alarm key: degrade the section, or refuse the status command | POLICY-COVERED | implementation-level | not required (POLICY-COVERED) |
| Q4 | Alarm surfaces: which outputs render the alarm | DECIDED | implementation-level | not required (DECIDED by committed surface measurement) |
| Q5 | Default band values and the OFF flag | DECIDED | product-level | not required at plan close — see row (values become binding only at enablement; a config edit overrides them with no code change) |
| Q6 | Does a same-stage re-entry (expiry → re-acquire → T4) refresh the stage-checkpoint anchor | DECIDED | implementation-level | not required (DECIDED by audit finding + committed transition-table measurement) |
| Q7 | How a future-dated signal timestamp is handled | DECIDED | implementation-level | not required (DECIDED by committed precedent `QuotaClockSkewTolerance`) |
| Q8 | Where AC-CAA-006's RED observation points | DECIDED | implementation-level | not required (repair of an acceptance-wording defect; DECIDED by audit finding + plan pin) |
| Q9 | Does the no-progress-age basis admit the card's creation event (bounding the age against inherited base commits and pool-state cycling) | DECIDED | implementation-level | not required (DECIDED by audit finding + committed event-table measurement) |
| Q10 | Does the timestamp-validity rule bind every derived age or only the three signal timestamps | DECIDED | implementation-level | not required (DECIDED by audit finding + committed precedent `QuotaClockSkewTolerance`) |

## Routing

1. **FOUNDER rows with recorded operator verdicts:** none.
2. **DEFAULT-APPLIED at plan close:** none.
3. **POLICY-COVERED, no verdict needed:** Q3.
4. **DECIDED:** Q1, Q2, Q4 (committed-evidence anchors), Q5 (see the row's own residual statement), Q6-Q8 (plan-audit iteration 1 repairs, each anchored on a measured substrate or a committed precedent), and Q9-Q10 (plan-audit iteration 2 rule-C correction, anchored on audit findings reproduced through the real card API plus committed substrate and precedent).

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

### Q6: Does a same-stage re-entry — the expiry → re-acquire → T4 cycle — refresh the stage-checkpoint anchor (and with it the no-progress age)?

Label: DECIDED
Class: implementation-level
Authority anchor: plan-audit iteration 1 finding D1 (`.moai/reports/t1838/plan-audit-iter1.md`; the codex required-backend leg reproduced expiry → re-acquire → `plan` re-entry through the real transition API on `1cac5de4f`, observing the anchor age drop 9h1m → 1m) and the committed transition table: `internal/homestate/card_transition.go:418-421` (`applyLeaseExpiry` returns an expired card to `assigned` with its stage kept), `:137-142` (the T4 edges), `:811-818` (`stageAdmits` admits only a resume into the stage the card last reached).
Why resolved: under the 0.1.0 predicate ("most recent `card.transition` whose `to` is a pipeline state") every cycle of the crash-loop pattern reset the no-progress age, masking precisely the stuck work the SPEC exists to surface — activity was counting as progress. REQ-CAA-004 (0.1.1) counts stage changes only: an event qualifies when its `to` is a pipeline stage and differs from the `to` of the card's most recent prior pipeline-stage event, so the T4 re-entry (which by construction re-enters the kept stage) never qualifies, while stage changes in both directions — including the backward T6/T12 audit-FAIL returns — still do. The prior-stage comparison is deliberately against the prior stage event, not the prior row state: the intermediate T3 (`to: leased`) would otherwise defeat the exclusion. Regression-pinned by plan.md M2 (f), `TestStageCheckpointSameStageReEntry`, covering both directions.
Default: none needed (the transition table admits exactly one sound predicate).
Operator verdict: not required (DECIDED by audit finding + committed transition-table measurement)

### Q7: How is a future-dated signal timestamp handled?

Label: DECIDED
Class: implementation-level
Authority anchor: `internal/config/defaults.go:82-87` — `QuotaClockSkewTolerance` (REQ-QAS-005): a record whose capture time lies in the future beyond the tolerance is treated as unknown, because "a record from another machine or a skewed clock must not win by being 'newest'". The house already guards this exact hazard class on the quota surface; the omission at 0.1.0 was an inconsistency with precedent (plan-audit iteration 1 finding D2).
Why resolved: REQ-CAA-004 (0.1.1) adds the validity rule — a signal timestamp in the future beyond a compiled clock-skew tolerance renders that signal `unknown`, never an age of zero or negative — and REQ-CAA-005 excludes unknown signals from the newest-of, so a skewed commit can neither yield a negative age nor suppress a band. The tolerance ships as the compiled constant `CardAgeAlarmClockSkewTolerance` in `internal/config/defaults.go`, mirroring the precedent's shape. It is deliberately NOT a `workflow.card_age_alarm.*` key: the quota precedent is a compiled non-key value, the tolerance is not an operator-relevant band, and REQ-CAA-002's no-literal rule is satisfied because the constant is the one sanctioned home of the number. Named test: plan.md M3 (f), `TestFactoryStatusAgeAlarmFutureCommit` (the commit read is the realistic future-dated source, being the one signal that crosses machine boundaries; store-written timestamps are stamped by the reading machine's own clock, and the M2 seeds go through the real transition API, which stamps now).
Default: none needed (precedent decides shape and placement).
Operator verdict: not required (DECIDED by committed precedent `QuotaClockSkewTolerance`)

### Q8: Where does AC-CAA-006's RED observation point?

Label: DECIDED
Class: implementation-level
Authority anchor: plan-audit iteration 1 finding D3 — the 0.1.0 E-006 cell grepped `internal/cli/factory_card.go` for the hyphenated token `age-alarm`, but the plan's three additive touches to that file (the `AgeAlarms` field with its underscore JSON tag, one compute call, one render call) do not entail that token ever appearing there, so a compliant M3 could leave the grep at zero and falsify the document-level flip-exactness claim.
Why resolved: E-006 is retargeted to `grep -rln "age-alarm" internal/cli/` — empty, exit 1, re-measured on tree `1cac5de4fa4c0f45942c8e2e037de1bd7f2fac2d` — and the flip is made plan-entailed in both directions: plan.md M3 (2) pins the text alarm section's tag-prefixed lines to the literal token `age-alarm`, and plan §C places the renderer in the new file `internal/cli/factory_age_alarm.go`. Nothing else under `internal/cli/` carries the token today (measured), so only M3's own renderer flips the cell.
Default: none needed (one observation is plan-entailed).
Operator verdict: not required (repair of an acceptance-wording defect; DECIDED by audit finding + plan pin)

### Q9: Does the no-progress-age basis admit the card's creation event — bounding the age against inherited base commits and pool-state cycling?

Label: DECIDED
Class: implementation-level
Authority anchor: plan-audit iteration 2 findings B1/B2 (`.moai/reports/t1838/plan-audit-iter2.md`; the codex required-backend leg reproduced both through the real card transition API on `c14bb152a` — a fresh picked card reading `SPEC_no_progress_age=48h` against an `actual_card_age=0s` via its 48h-old inherited base commit, and a 10-cycle pre-stage expiry/re-acquire loop whose stuck time of 2h41m read as `1m` through the fallback) and the committed creation-event substrate `internal/homestate/card_picked.go:173` (creation is a `card.transition` with payload `from: ""`, `to: picked`, appended only in the card-not-found branch — exactly one per card with events, necessarily that card's first event).
Why resolved: without the creation member, (B2) a fresh card's no-progress age could predate its existence through the inherited base HEAD's committer timestamp — systematic false-alert noise on every newly-leased card whenever local main's HEAD is older than the band, precisely the quiet periods the panel exists for — and (B1) a pre-pipeline crash loop reset the age on every expiry → re-acquire cycle, because the 0.1.1 both-unknown fallback read the most recent `card.transition` and every cycle appends a fresh T3 (`assigned` → `leased`, `internal/homestate/card_transition.go:136`). REQ-CAA-005 (0.1.2) measures from the newest of {stage checkpoint, commit, creation}: creation is by construction the card's earliest event, so the iter1 D1 post-entry predicate is unchanged, while the both-unknown path now anchors at creation — pool-state cycling is activity, not progress (design principle 2), and the most-recent-transition fallback (the reset vector) is withdrawn, leaving `updated_at` as the legacy-row fallback only. Regression-pinned by plan.md M2 (g), `TestNoProgressAgeCreationAnchor`, covering the crash-loop mutant, the inherited-base mutant, and the stuck-still-alerts true positive.
Default: none needed (audit finding + substrate measurement leave one sound predicate).
Operator verdict: not required (DECIDED by audit finding + committed event-table measurement)

### Q10: Does the timestamp-validity rule bind every derived age, or only the three signal timestamps?

Label: DECIDED
Class: implementation-level
Authority anchor: plan-audit iteration 2 finding B3 (`.moai/reports/t1838/plan-audit-iter2.md`; reproduced through the real `RecordPicked` API on `c14bb152a` with a future stamp — a fallback transition dated tomorrow yielded a `−24h` no-progress age) and the committed precedent `internal/config/defaults.go:82-87` (`QuotaClockSkewTolerance`, REQ-QAS-005: "a record from another machine or a skewed clock must not win by being 'newest'").
Why resolved: the 0.1.1 rule bound only the three signal timestamps; the both-unknown fallback, the `updated_at` fallback, and the REQ-CAA-003 gate-wait anchor subtracted store-written timestamps unchecked, so a rolled-back clock or a DB read from another machine could produce negative ages — the exact hazard class iter1 D2 established as blocking, incompletely repaired. REQ-CAA-003/004/005 (0.1.2) extend the rule to every timestamp the displayed ages derive from: the REQ-CAA-005 basis members (stage checkpoint, commit, and the creation event), the legacy `updated_at` fallback, and the gate-wait anchor (transition/lease-expired event `created_at`, `updated_at` fallback). A future-dated source beyond `CardAgeAlarmClockSkewTolerance` renders that source `unknown`; an age with no valid basis member renders `unknown`; no derived age is ever negative. Named test: plan.md M2 (h), `TestAgeAnchorsFutureTimestamps`.
Default: none needed (precedent + audit finding decide shape and reach).
Operator verdict: not required (DECIDED by audit finding + committed precedent `QuotaClockSkewTolerance`)
