---
id: SPEC-FACTORY-LEASE-RENEW-001
title: "Factory lease renewal: renew verb, lease configuration keys, and watchdog renewal (card t1663)"
version: "0.1.1"
status: in-progress
created: 2026-10-10
updated: 2026-10-10
author: manager-spec (card t1663)
priority: P0
phase: "v3.2.0 target"
module: "internal/homestate"
lifecycle: spec-anchored
tags: "factory-lease, lease-renew, watchdog, lease-config, card-t1663"
tier: M
depends_on: [SPEC-FACTORY-RECORD-001]
related_specs: [SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-FACTORY-RECORD-REPAIR-001]
---

# SPEC: Factory lease renewal (renew verb, lease configuration keys, watchdog renewal)

## HISTORY

- 0.1.0 — 2026-10-10 — Plan-phase artifacts authored for card t1663 (Class C, Tier M), on branch `WT-factory-lease-renew` from HEAD `13cc4a2f6`. The card body sets the scope: a factory lease renew verb, a configuration key for the card lease duration, and watchdog renewal with a renewal period of at most one half of the duration, the value of which comes from a configuration key. The revision builds in the three findings of the draft's plan audit iteration 4 (`.moai/reports/t1621/plan-audit-iter4.md`, items D1, D3, and D4) from its first revision. D1: the renewal-period bound is stated in REQ-FLR-007 and enforced by a refusal at factory open, with behavioural checks in AC-FLR-007 and AC-FLR-012; the residual-risk paragraph, decision-index Q1, and progress.md §E.1 state the same period. D3: one sentence for the `Operator verdict:` line, used identically in decision-index.md and acceptance.md. D4: acceptance.md and its Definition of Done use one vocabulary, in which a predicted RED is not an observed RED and a criterion is recorded as PASS only from its GREEN output. Draft requirements REQ-FRR-006, REQ-FRR-007, and REQ-FRR-008 are transferred to card t1639 (3.2-2-2) and are not authored here. Priority P0 is inherited from the draft batch, which carries lens-4 defects 1 and 27 (draft HISTORY 0.1.0). Requirement count 12, criterion count 13 (Tier M ceilings 16 / 16). This is the first plan audit of this SPEC.
- 0.1.1 — 2026-10-10 — Repair of plan audit iteration 1 (`.moai/reports/t1663/plan-audit.md`, verdict FAIL, eight blocking findings). FLR-B1: the stall window is counted from the last observed channel change (REQ-FLR-013), because a comparison with the previous snapshot cannot give a stall verdict at a wake period below N; the snapshot gains one additive field, `last_change_at`. FLR-B2: the lane startup instruction in four locales, its two pinning tests, and the cadence sentence of `factory-dispatch.md` move to the renewal period (M3). FLR-B3: the armed wake starts at minute 0 (`*/N`), and AC-FLR-013 computes the largest gap for every legal interval. FLR-B4: AC-FLR-005 checks the acquisition, transition-extension, and renewal clauses. FLR-B5: the M4 RED is the output of three named mutants, and the predicted "verb not registered" RED is withdrawn. FLR-B6: the snapshot field `renew_result` is named, and the renew verb reports its refusal class in JSON. FLR-B7: the reaped-state assertion of `TestRenewLeaseRefusalsAndExpiry` is rewritten under REQ-FLR-004 in M2. FLR-B8: AC-FLR-009 reads the configured values. Point (d): AC-FLR-011 checks the refusal class and the same-pass measurement, and AC-FLR-013 checks the computed cadence. FLR-O1 (a negative-duration case in AC-FLR-006), FLR-O2 (Q5 names the cadence contradiction), and FLR-O3 (the third M4 duration pair) are answered. Added: REQ-FLR-013, AC-FLR-014, AC-FLR-015. Requirement count 13, criterion count 15 (Tier M ceilings 16 / 16). Decision-index Q6 (the stall window reference) is opened, and Q1 gains the armed form. The draft SPEC-FACTORY-RECORD-REPAIR-001 is still not modified.

## Scope

This SPEC carries the three items of card t1663. A lane session that holds a card lease renews it through a new factory verb. The card lease duration and the watchdog renewal period come from configuration keys. The lane stall watchdog renews its own card lease at the start of each pass, and it counts its stall window from the last observed channel change.

1. **Lease renew verb.** The factory renew verb and its MCP tool. Renewal authority belongs to the holder of a live lease. An expired renewal is refused with no write (REQ-FLR-001 to REQ-FLR-004).
2. **Lease configuration keys.** `workflow.factory_lease.duration` (the card lease duration, default 15 minutes) and `workflow.factory_lease.renew_interval` (the watchdog renewal period, default 5 minutes). Each is refused at factory open when its value falls outside its bound. The queue claim lease stays at 15 minutes (REQ-FLR-005 to REQ-FLR-009).
3. **Watchdog renewal, standing wake, and stall window.** The lane stall watchdog renews its own card lease before it measures lane progress, and it records a refused renewal and still measures. Its standing wake is armed at the renewal period as a minute step that starts at minute 0. Its stall verdict counts the N-minute window from the last observed channel change (REQ-FLR-010 to REQ-FLR-013).

The SPEC changes `internal/homestate`, the factory verbs and MCP tools that call it, the `workflow` configuration section and its template, the lane stall watchdog skill, the standing-wake and stall paragraphs of `auto-semantics.md`, the cadence sentence of `factory-dispatch.md`, and the lane startup instruction of `internal/hook/session_start_factory_i18n.go` with its two pinning tests. The draft SPEC-FACTORY-RECORD-REPAIR-001 is not modified by this SPEC. Its requirements REQ-FRR-001 to REQ-FRR-005, REQ-FRR-009, and REQ-FRR-010 cover the same items, so a ruling decides how the two drafts relate (decision-index Q5).

## Background

- **The lease duration is compiled.** A card lease lasts for the compiled constant `DefaultFactoryLeaseDuration`, 15 minutes (`internal/config/defaults.go`), read through the package variable `FactoryLeaseDuration` (`internal/homestate/card_record.go:93`). No configuration key reads it. A lease is extended only when a stage transition succeeds (`internal/cli/factory_card.go:1737` is the only non-test caller of the renewal, and it renews right after the transition has extended the lease), and no verb renews a lease on its own.
- **The standing watchdog wake cannot keep a lease alive.** The standing wake is `CronCreate` with `7,27,47 * * * *`, every 20 minutes (`.claude/rules/moai/workflow/auto-semantics.md:108`). A 20-minute gap is longer than the 15-minute lease, so a stalled lane's lease lapses between passes.
- **An expired renewal writes today.** `RenewLease` reaps an expired lease before it checks the holder (`internal/homestate/card_transition.go:441-456`). A refused renewal therefore bumps the card version and appends a lease event.
- **Why the renewal period is at most half the duration.** A renewal at time r sets the expiry to r plus the duration D. A pass that runs late by d after its scheduled time still renews only while d is less than D minus the renewal period s. At s equal to one half of D, that margin equals one renewal period. The expiry instant itself counts as expired (`internal/homestate/card_record.go:143-146`), so the margin is open at its upper end.
- **The queue claim is separate.** The queue claim lease uses the same 15-minute constant directly (`internal/factory/backlog_store.go:1244`) and does not read the card lease duration.
- **The stall window and the wake period.** The stall verdict compares two observations: the previous snapshot and the current one (`.claude/rules/moai/workflow/auto-semantics.md` §4). The previous snapshot is replaced on every pass. At a wake period shorter than the N-minute window, the previous observation is always younger than N, so that comparison never gives a stall verdict. The doctrine states the rule as "no change across the N-minute window" (auto-semantics §3). The comparison in §4 matches that statement only when the period is N or more.

## Requirements

The requirements use GEARS notation. Each active requirement maps to at least one criterion in `acceptance.md`, and the mapping is in `plan.md` §E. The active requirements are REQ-FLR-001 to REQ-FLR-013.

### Lease renew verb

- **REQ-FLR-001** When the lane session that holds a live card lease runs the factory renew verb for that card, the factory record shall extend the lease to the configured card lease duration counted from the renewal, shall report the new expiry time, and shall leave the card's state and version unchanged.
- **REQ-FLR-002** When a lane calls the MCP renewal tool for a card, the factory record shall return the same success fields and the same refusal class as the factory renew verb for the same card, caller, and lease state.
- **REQ-FLR-003** When a caller other than the holder requests renewal of a live card lease, the factory record shall refuse the request and shall leave the card's state, version, and event history unchanged.
- **REQ-FLR-004** When the holder requests renewal of its card lease after the lease has expired, the factory record shall refuse the renewal and shall leave the card's state, version, and event history unchanged.

### Lease configuration keys

- **REQ-FLR-005** The factory record shall take the card lease duration for every lease acquisition, transition extension, and renewal from the configuration key `workflow.factory_lease.duration`, whose default value is 15 minutes.
- **REQ-FLR-006** When the configured value of `workflow.factory_lease.duration` is not a positive duration, the factory record shall refuse to open and shall name `workflow.factory_lease.duration` in the refusal.
- **REQ-FLR-007** When the configured value of `workflow.factory_lease.renew_interval` is not a whole number of minutes from 1 to 59, or is longer than one half of the configured card lease duration, the factory record shall refuse to open and shall name `workflow.factory_lease.renew_interval` and the bound that it violates in the refusal.
- **REQ-FLR-008** Where `workflow.factory_lease.duration` is configured, the queue claim lease duration shall remain 15 minutes.
- **REQ-FLR-009** The factory record shall take the watchdog renewal period from the configuration key `workflow.factory_lease.renew_interval`, whose default value is 5 minutes, and the factory status surface shall report the effective value of both keys.

### Watchdog renewal, standing wake, and stall window

- **REQ-FLR-010** When the lane stall watchdog starts a pass, the watchdog shall renew the lane's own card lease through the factory renew verb before it measures lane progress.
- **REQ-FLR-011** When a watchdog renewal does not extend the lease, the watchdog shall record the class that the renew verb reports in its JSON output in the snapshot field `renew_result`, one of `refused-expired`, `refused-not-holder`, or `error`, and shall still measure lane progress in the same pass. A renewal that extends the lease records `renewed` in the same field.
- **REQ-FLR-012** While a lane is armed with the standing watchdog wake, the watchdog wake schedule shall be a minute step of `workflow.factory_lease.renew_interval` that starts at minute 0, the expression `*/N * * * *` for an interval of N minutes, so that no two consecutive scheduled wakes are further apart than one renewal period, including across the hour boundary.
- **REQ-FLR-013** When the lane stall watchdog judges a scheduled pass, the watchdog shall report a stall only when all three channels are unchanged since the previous snapshot and at least N minutes (15, as auto-semantics §3 states) have passed since the last observed channel change, so that a lane with no channel change is reported as stalled no later than N minutes plus one renewal period after that change.

The snapshot record is `.moai/state/watchdog/<card-id>.json` (auto-semantics §4). Its format changes by two additive fields and by no other change: `renew_result` (REQ-FLR-011) and `last_change_at`. `last_change_at` is the RFC3339 observation time of the last pass at which a channel changed. The first snapshot sets it to its own observation time, and it is carried forward unchanged while no channel changes. No existing field changes its name or meaning.

### Lineage to the draft

The draft is SPEC-FACTORY-RECORD-REPAIR-001 at HISTORY 0.4.0. The table gives each active requirement's draft number and the change.

- Lineage, REQ-FLR-001: draft REQ-FRR-001, no change.
- Lineage, REQ-FLR-002: draft REQ-FRR-002, no change.
- Lineage, REQ-FLR-003: draft REQ-FRR-009, renumbered; live lease only, as in the draft.
- Lineage, REQ-FLR-004: draft REQ-FRR-004, no change.
- Lineage, REQ-FLR-005: draft REQ-FRR-003; default and reading tied to the factory open path.
- Lineage, REQ-FLR-006: new; refusal for a non-positive duration.
- Lineage, REQ-FLR-007: draft REQ-FRR-005 (bound); bound split out and enforced at open (D1).
- Lineage, REQ-FLR-008: draft REQ-FRR-010, renumbered with the same meaning.
- Lineage, REQ-FLR-009: new; renewal-period key and status report.
- Lineage, REQ-FLR-010: draft REQ-FRR-005 (renew first), no change.
- Lineage, REQ-FLR-011: draft REQ-FRR-005 (refusal recorded), with the snapshot field named here (FLR-B6).
- Lineage, REQ-FLR-012: draft REQ-FRR-005 (standing wake, left open in the draft); the wake is a minute step from minute 0 (FLR-B3).
- Lineage, REQ-FLR-013: new; the draft states no reference for the stall window (FLR-B1).

## Exclusions (What NOT to Build)

The following items are out of scope for this SPEC.

### Out of Scope — Transferred to card t1639 (3.2-2-2)

- Atomic collection of the holder's own expired lease through the factory stage verb or the parent-checkout next verb, in one step with the requested transition (draft REQ-FRR-006).
- The refusal text that names the recovery command for an expired lease, and the receipt that a caller other than the holder gets on an expired lease (draft REQ-FRR-007).
- The worktree path recorded as part of lease acquisition, with its refusal when materialization fails or when a different path is recorded (draft REQ-FRR-008).
- The non-holder expiry ordering: the order in which the transition path and the renewal path judge an expired lease for a caller other than the holder (the expired clauses of draft REQ-FRR-014; card t1621 item B5).

### Out of Scope — Items retired or left to other cards

- The pushed-to-done edge and the `moai factory doctor` out-of-band write detection (draft REQ-FRR-016 to REQ-FRR-019, moved by decision d-20261010T064450Z-68b2).
- Any change to the queue claim beyond keeping its lease at 15 minutes (REQ-FLR-008).
- Any change to the audit-verdict parser (`internal/auditverdict/`) or to the ceiling engine, which card t1618 owns.
- Any change to the stall verdict's N-minute length, to its three measurement channels, or to its first condition (no channel change since the previous snapshot). The second condition's reference point changes under REQ-FLR-013 (decision-index Q6); no other change to the stall verdict is in scope.

### Out of Scope — Delivery

- Remote push, pull requests, and any remote operation from the lane. This repository lands locally, and the leader's batch push is outside this SPEC.

## Residual risk

- **The standing wake is armed outside the factory record (REQ-FLR-012).** The lane arms its wake with `CronCreate`, and the factory record does not observe the armed expression. A wake armed with a stale or mistyped expression lets the lease lapse between passes. The lapse shows at the next pass as a refused renewal, which REQ-FLR-011 records. The holder then recovers through the stage verb, which this SPEC does not change. AC-FLR-013 checks the form and its gap, not the armed entry.
- **The renewal period in this revision.** The working value is 5 minutes (REQ-FLR-009; decision-index Q1). The bound is one half of the duration key: 7 minutes 30 seconds at the 15-minute default, and 7 minutes is the whole-minute maximum there (REQ-FLR-007). The late-pass margin is the duration minus the period: 10 minutes at 5 minutes, and 8 minutes at 7 minutes. Decision-index Q1 is ruled at this value (d-20261010T095818Z-c501), so these figures stand. progress.md §E.1 states the same period.
- **The bound is enforced at open and checked behaviourally.** A value outside the bound stops the factory from opening (REQ-FLR-007). AC-FLR-007 checks the refusal and the boundary, and AC-FLR-012 checks the margin. A text-only change to the bound cannot pass AC-FLR-007.
- **A holder whose lease has lapsed cannot renew it.** REQ-FLR-004 refuses the renewal with no write, and this SPEC does not restore the lease, because that restoration belongs to the transferred items. decision-index Q3 is ruled as refusal with no write (d-20261010T095731Z-5ddc), so the current reap-then-refuse behaviour does not stay.
- **The cost of the cadence.** Each standing wake is a watchdog turn, and a wake after the prompt-cache window re-writes the session prefix (`.claude/rules/moai/workflow/cache-aware-execution.md` directive 9). decision-index Q1 records this trade-off. A start at minute 0 fires together with other top-of-hour schedules, which the draft's off-minute start avoided. REQ-FLR-012 takes that collision for its form-level gap guarantee; an off-minute start needs a gap check for each value in force (decision-index Q1, sub-question d). It is a cost, not a correctness risk.
- **The stall window reference changes a verdict that the leader reads.** Under REQ-FLR-013 a stall is reported at the first scheduled pass at or after N minutes without a channel change, which the previous-snapshot comparison never reports at a period below N. The reference is ruled at decision-index Q6 (d-20261010T095731Z-5ddc), and REQ-FLR-013 is the rule. At a period of N or more the two rules give the same verdict on every pass. That is a derivation from the rule text; AC-FLR-015 checks it at one period, 20 minutes.
- **Procedure and rule checks are text or model checks.** AC-FLR-010, the procedure half of AC-FLR-011, and AC-FLR-013 check the watchdog procedure and the standing-wake text. AC-FLR-014 checks the rendered lane rule. AC-FLR-015 checks a model of the stall rule and a text pin on auto-semantics §4. A text-only change passes the text checks, and no criterion observes a live lane's watchdog turn.
- **The refusal record is written by the model, not by code.** No code writes the watchdog snapshot; the lane writes it by following the procedure. AC-FLR-011 checks the class the renew verb reports and the procedure's copy step. It does not observe a written snapshot file.
- **Liveness is simulated.** AC-FLR-012 uses an injected clock, and AC-FLR-015 simulates the stall rule over a deterministic sequence. Neither measures real wake timing.
- **Overlap with the draft.** SPEC-FACTORY-RECORD-REPAIR-001 describes the same renew, duration key, and watchdog behaviour. Until a ruling retires or narrows its overlapping requirements, two drafts describe one behaviour (decision-index Q5).
