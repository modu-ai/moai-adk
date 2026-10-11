---
id: SPEC-FACTORY-RECORD-REPAIR-001
title: "Factory record repair: lease renewal, one-shot expired-lease recovery, and watchdog renewal (narrowed to the remaining items)"
version: "0.4.0"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec (card t1621)
priority: P0
phase: "v3.2.0 target"
module: "internal/homestate"
lifecycle: spec-anchored
tags: "factory-record, lease, renew, recovery, card-t1621"
tier: M
depends_on: [SPEC-FACTORY-RECORD-001]
related_specs: [SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-FACTORY-SELF-DISPATCH-001]
---

# SPEC: Factory record repair (first batch, narrowed)

## HISTORY

- 0.1.0 — 2026-10-10 — Plan-phase artifacts authored for card t1621 (Class C, Tier M), on branch `WT-factory-record-fix` from HEAD `2aab5f797`. Scope is the first batch only (card brief R1 to R6). Evidence: lens-4 (factory evidence and leases) defects 1, 7, 9, and 27 and the interim part of defect 8; lens-2 (audit verdict engine) D1, marked P0; leader decision records `d-20261009T235640Z-3653`, `d-20261009T180009Z-ae0f`, and `d-20261009T182157Z-8ccf`. Priority P0 because the batch carries lens-4 defect 1 and lens-2 D1, both marked P0. The brief's defect numbers were checked against the lens-4 table, and the corrections are recorded in `plan.md` §B. Requirement count 15, criterion count 16 (Tier M ceilings 16 / 16).
- 0.1.1 — 2026-10-10 — repair round 1 after plan audit iteration 1 (FAIL, 0.56, 17 blocking defects); operator decisions are listed in decision-index.md.
- 0.2.0 — 2026-10-10 — Narrowed to the three surviving items of card t1621's re-scope, recorded as leader ruling `d-20261010T053435Z-006f`, which reconciles operator decision `d-20261010T053034Z-1a19` with the operator-approved queue re-scope. The card body now reads (Korean, verbatim): "범위를 팩토리 v2(3.2-2-2) 선행 수리로 한정: (1) 임대 renew 동사 + TTL 설정 키 + 워치독 첫 단계 갱신 (2) 만료 수거 원자화(수거+재임대+전이 한 번에, ErrLeaseExpired에 복구 명령) (3) pushed→done 엣지 개방 + 대역 외 쓰기 탐지(factory doctor). (4)~(12)는 3.2-2-2(8상태·하트비트)가 상태 집합 교체로 흡수." Retired, because each serves only old items (4) or (7) of the body: REQ-FRR-008, -009, and -010 and AC-FRR-008, -009, and -010 serve item (7), which names "plan-audit/sync-audit 에서 판정이 파싱 불가면 앞뒤 출구 없음 — 홀더용 감사 재진입 엣지" (no exit forward or back; holder audit re-entry edge); REQ-FRR-012 and -013 and AC-FRR-012 and -013 serve item (4), which names the interim guard "그 전까지 SPEC 카드의 빈 stage→run(T4b …) 우회 가드"; AC-FRR-016 counts only the three rows that item (7)'s requirements add. Kept with their item: REQ-FRR-001 to -007 (items 1 and 2), REQ-FRR-011 (the re-lease of item 2 is a lease acquisition), REQ-FRR-014, and REQ-FRR-015. Item (3) has no requirement in this revision and is open (decision-index Q21). Repairs: B1 fixed (surviving criteria reclassified as regression guards, the release-gate wording removed; verdict 1a19, Q1). B5 fixed (option i: materialize, then record and lease in one transaction). B2, B3, and B4 retired with their subjects. Verdicts from 1a19 applied to Q1, Q8, Q9, Q10 (ii, at T3), and Q18. Q2, Q4, Q5, Q6, Q7, and Q13 are marked not applicable (retired subjects). Identifiers REQ-FRR-008 to -010, -012, -013 and AC-FRR-008 to -010, -012, -013, -016 are retired and not reused.
- 0.3.0 — 2026-10-10 — Item (3) written, per leader ruling `d-20261010T060606Z-c5c4` (option a: this card carries the pushed-to-done opening and the out-of-band write detection; t1616 does not own the done transition). Added REQ-FRR-016 to REQ-FRR-019 and AC-FRR-017 to AC-FRR-021, each classified as a regression guard. REQ-FRR-016 opens `pushed` → `done` under the receipt and containment conditions. REQ-FRR-017 refuses that edge when a condition fails and leaves the other edges as they are. REQ-FRR-018 and REQ-FRR-019 give the `moai factory doctor` out-of-band write check and its read-only, empty-sweep reporting. The transition-count pin `TestFR_AC005_TransitionTableEdgeCount` is re-authored in place from 71 to 72 accepted pairs. The narrowing of `d-20261010T053435Z-006f` is unchanged: items (4) to (12) stay retired, and REQ-FRR-008, -009, -010, -012, -013 and AC-FRR-008, -009, -010, -012, -013, -016 are not reused. Q2, Q4, Q5, Q6, Q7, and Q13 stay not applicable, as confirmed by `d-20261010T060606Z-c5c4`. The 1a19 verdicts (`d-20261010T053034Z-1a19`) are unchanged. The 75-of-76 figure is the card body's claim and was not re-measured. Requirement count 14, criterion count 15 (Tier M ceilings 16 / 16).
- 0.4.0 — 2026-10-10 — Repair after plan audit iteration 3 (FAIL 0.75, six blocking findings B1 to B6; card t1621), per operator-delegated decision `d-20261010T064450Z-68b2` (영실이 판단, option (b): repair B1, B3, and B6; transfer B2, B4, and B5 to 3.2-2-2 (t1639); one final audit, and a further FAIL is admitted as debt). It resolves hold `d-20261010T064355Z-4fb1`. (B1) The active requirements are renumbered to a contiguous REQ-FRR-001 to REQ-FRR-010. REQ-FRR-011 becomes REQ-FRR-008, REQ-FRR-014 becomes REQ-FRR-009, and REQ-FRR-015 becomes REQ-FRR-010; REQ-FRR-001 to -007 keep their numbers. The identifier map in Requirements keeps every old identifier with its retirement record. Amendment of the no-reuse rule (this row supersedes the 0.2.0 wording and decision-index routing item 1): the retired REQ-FRR-008, -009, and -010 are reused for the renumbered requirements with the meanings the map gives; REQ-FRR-012, -013, and the transferred REQ-FRR-016 to -019 are not reused in this SPEC; criterion identifiers are not renumbered. (B2, B4, B5) Transferred out of the active body, with reasons in `progress.md` § G: REQ-FRR-016 to -019 and AC-FRR-017 to -021 (item 3); the record-call transaction boundary (the placement clause of REQ-FRR-008, plan M1 item 5, plan M3 item 1, and plan § C item 5); the non-holder expiry ordering. REQ-FRR-014 keeps only the live-lease refusal (now REQ-FRR-009), REQ-FRR-004 is narrowed to the holder's own expired lease, and the pin of accepted transition pairs (72 and 369) is removed from the body. (B3) AC-FRR-015's Given is restated through the configured key `workflow.factory_lease.duration`, read by the factory handle when OpenFactory runs; the queue claim stays at 15 minutes, and the test observes the expiry the claim writes. (B6) REQ-FRR-005 renews the watchdog's card lease at intervals no longer than one half of the lease duration key; a refused renewal is recorded and the measurement still runs (named in AC-FRR-005); the 20-minute cadence is removed from this SPEC. Requirements 10, criteria 10 (Tier M ceilings 16 / 16).

## Scope

This SPEC is the first batch of the factory record repair for card t1621. Leader ruling `d-20261010T053435Z-006f` narrowed the card body to three items. Operator-delegated decision `d-20261010T064450Z-68b2` then transferred the third item and two sub-parts of the first batch to 3.2-2-2 (t1639). This SPEC carries the four remaining items:

1. **Lease renew verb.** The renew verb and its MCP tool, with renewal authority held by the holder of a live lease (REQ-FRR-001, REQ-FRR-002, REQ-FRR-009).
2. **Lease-duration key.** The configuration key `workflow.factory_lease.duration`, and the queue claim's independence from it (REQ-FRR-003, REQ-FRR-010).
3. **Watchdog renewal.** The lane stall watchdog's renewal of its own card lease, as its first step (REQ-FRR-005).
4. **Atomic collection of the holder's own expired row.** The holder's refused renewal of an expired lease, and the one-step recovery through the stage verb or the parent-checkout next verb, with a refusal that names the recovery command. The re-lease is a lease acquisition under REQ-FRR-008 (REQ-FRR-004, REQ-FRR-006, REQ-FRR-007, REQ-FRR-008).

Transferred out of this SPEC by 68b2 (HISTORY 0.4.0): the pushed-to-done edge and out-of-band write detection (old item (3)); the record-call transaction boundary; and the non-holder expiry ordering. The moved identifiers and the reasons are in `progress.md` § G.

The SPEC changes `internal/homestate` and the factory verbs and MCP tools that call it. This SPEC (card t1621) owns the lease renewal verb and the TTL configuration key. Card t1618 references this SPEC for them and does not own them.

## Background

- **Lease expiry (item 1).** A card lease lasts 15 minutes from acquisition and is extended only when a stage transition succeeds. No verb renews a lease on its own, and the lane watchdog never reads the lease (lens-4 defects 1 and 27).
- **Expired own lease (item 2).** An expired lease is reclaimed only when a later call reaches the card, and that call commits the reclaim and then returns an error. The next verb refuses a holder's own expired row as "in flight" (lens-4 defects 1 and 9; the card body cites `card_transition.go:311-316`).
- **Unstamped lease (decision `d-20261009T182157Z-8ccf`).** A lease does not record the card's worktree path. Because the re-lease of item 2 is also a lease acquisition, this SPEC records the path at acquisition.
- **Remote done edge (item 3, transferred).** When a remote exists, a card cannot reach `done` through the merged-local edge, which is refused while a remote is configured, and the reserved ci-green edge is the only other route (`card_transition.go:156-158` and `:203-208`; the card body cites `:207-209` and `:668-670`, which match this tree). The card body reports that 75 of the 76 done rows were written outside the transition API. That figure is the card body's claim; this SPEC did not measure it. These facts are kept as evidence for 3.2-2-2; this SPEC does not carry the item.

The sync-audit re-entry, the holder evidence re-stamp, and the empty-stage guard that the first revisions carried are retired by `d-20261010T053435Z-006f` (HISTORY 0.2.0). The verdict file format and the ceiling engine belong to t1618 and are not changed by this SPEC.

## Requirements

The requirements use GEARS notation. Each active requirement maps to at least one criterion in `acceptance.md`, and the mapping is in `plan.md` §E. The active requirements are REQ-FRR-001 to REQ-FRR-010, contiguous after the renumbering in HISTORY 0.4.0. The identifier map below gives the current number of each requirement and the retirement record of each retired number.

### Lease renewal, lease-duration key, and watchdog renewal (lens-4 defects 1 and 27)

- **REQ-FRR-001** When the lane session that holds a live card lease runs the factory renew verb for that card, the factory record shall extend the lease to the configured lease duration counted from the renewal, and the verb shall report the new expiry time.
- **REQ-FRR-002** The factory record shall expose the renewal as an MCP tool whose success and refusal results match the factory renew verb for the same card and caller.
- **REQ-FRR-003** The factory record shall take the card lease duration from the configuration key `workflow.factory_lease.duration`, whose default value is 15 minutes.
- **REQ-FRR-004** When the holder of a card lease requests a renewal of that lease after it has expired, the factory record shall refuse the renewal and leave the card's state, version, and event history unchanged. (Narrowed to the holder in 0.4.0, B5.)
- **REQ-FRR-005** When the lane stall watchdog starts a pass, the watchdog shall renew its own card lease before it measures lane progress, and it shall renew that lease at intervals no longer than one half of the card lease duration read from `workflow.factory_lease.duration`. When the factory record refuses a renewal, the watchdog shall record the refusal and shall still measure lane progress in that pass. The renewal is the one write the watchdog adds. (Rewritten in 0.4.0, B6.)

### One-shot recovery and acquisition of the holder's own expired lease (lens-4 defect 9)

- **REQ-FRR-006** When the holder of an expired card lease requests a transition through the factory stage verb, or through the factory next verb where that verb is reachable (run from the parent checkout), the factory record shall collect the expired lease, re-lease the card to the same holder, and apply the requested transition in one step, so that a refused transition leaves the expired row unchanged.
- **REQ-FRR-007** When a card lease is refused as expired, whether the refused request is a transition or a renewal, the refusal shall name the recovery command that the holder runs to re-lease the card: the factory stage verb (`moai factory stage`), which a holder inside its card worktree can run. REQ-FRR-006 applies that verb's recovery in one step. The factory next verb also recovers from the parent checkout but is not the named command, because it refuses to run from a card worktree.
- **REQ-FRR-008** When a lane acquires the lease on a card, including the re-lease of REQ-FRR-006, the factory record shall materialize the card worktree that the factory provides for that card and record that worktree path as part of the acquisition. If materialization fails, the factory record shall refuse the acquisition and leave the card unchanged. The factory record shall refuse a later acquisition that would record a different worktree path for the card, and shall not hold a lease on a card that has no recorded worktree path. The acquisition is the T3 transition. (Renumbered from REQ-FRR-011 in 0.4.0. The same-transaction placement is removed; its transaction boundary is transferred with B2, `progress.md` § G.)

### Cross-cutting

- **REQ-FRR-009** When a caller that does not hold a card lease requests a renewal of that lease while it is live, the factory record shall refuse the request and leave the card's state, version, and event history unchanged. (Renumbered from REQ-FRR-014 in 0.4.0 and narrowed to a live lease; the expired-lease clauses are transferred with B5.)
- **REQ-FRR-010** Where the card lease duration is configured, the factory record shall keep the queue claim lease duration unchanged. (Renumbered from REQ-FRR-015 in 0.4.0; serves the lease-duration key.)

### Identifier map (0.4.0 renumbering, retired and moved identifiers)

Current requirements, with the number each carried in 0.3.0:

| Current | 0.3.0 number | Change in 0.4.0 |
|---|---|---|
| REQ-FRR-001 | REQ-FRR-001 | none |
| REQ-FRR-002 | REQ-FRR-002 | none |
| REQ-FRR-003 | REQ-FRR-003 | key named |
| REQ-FRR-004 | REQ-FRR-004 | narrowed to the holder (B5) |
| REQ-FRR-005 | REQ-FRR-005 | renewal interval bound and refusal outcome (B6) |
| REQ-FRR-006 | REQ-FRR-006 | none |
| REQ-FRR-007 | REQ-FRR-007 | none |
| REQ-FRR-008 | REQ-FRR-011 | renumbered; same-transaction placement removed (B2) |
| REQ-FRR-009 | REQ-FRR-014 | renumbered; narrowed to a live lease (B5) |
| REQ-FRR-010 | REQ-FRR-015 | renumbered |

Retired and moved identifiers. None of them is an active requirement in this SPEC. Each meaning is the meaning the identifier had when it was active.

| Identifier | Record | Status in this SPEC | Meaning when active |
|---|---|---|---|
| REQ-FRR-008 (0.2.0 and 0.3.0 numbering) | `d-20261010T053435Z-006f` (0.2.0) | reused: current REQ-FRR-008 is the old REQ-FRR-011 | sync-audit to run through a card-scoped ruling (old item 7; decision-index Q4) |
| REQ-FRR-009 (0.2.0 and 0.3.0 numbering) | `d-20261010T053435Z-006f` (0.2.0) | reused: current REQ-FRR-009 is the old REQ-FRR-014 | three reason tokens for the sync-audit ruling (old item 7; Q4) |
| REQ-FRR-010 (0.2.0 and 0.3.0 numbering) | `d-20261010T053435Z-006f` (0.2.0) | reused: current REQ-FRR-010 is the old REQ-FRR-015 | holder re-stamp refused on an audited-SHA mismatch (old item 7; Q5) |
| REQ-FRR-012 | `d-20261010T053435Z-006f` (0.2.0) | not reused | interim guard on the empty-stage to run edge (old item 4; Q6, Q7) |
| REQ-FRR-013 | `d-20261010T053435Z-006f` (0.2.0) | not reused | residual route for an empty SPEC id (old item 4; Q13) |
| REQ-FRR-016 | `d-20261010T060606Z-c5c4` (0.3.0), moved by `d-20261010T064450Z-68b2` | not active here; moved to 3.2-2-2 | pushed → done edge under the receipt and containment conditions (item 3; B4) |
| REQ-FRR-017 | `d-20261010T060606Z-c5c4`, moved by `d-20261010T064450Z-68b2` | not active here | refusals of the pushed → done request, other edges unchanged (item 3) |
| REQ-FRR-018 | `d-20261010T060606Z-c5c4`, moved by `d-20261010T064450Z-68b2` | not active here | `moai factory doctor` out-of-band write detection (item 3) |
| REQ-FRR-019 | `d-20261010T060606Z-c5c4`, moved by `d-20261010T064450Z-68b2` | not active here | `moai factory doctor` read-only and empty-sweep reporting (item 3) |

No-reuse rule, amended by HISTORY 0.4.0: an identifier retired in 0.2.0 or moved in 0.4.0 is not reused in this SPEC, with one exception. REQ-FRR-008, REQ-FRR-009, and REQ-FRR-010 are reused by the renumbering, each with the current meaning given in the first table. REQ-FRR-012, REQ-FRR-013, and REQ-FRR-016 to REQ-FRR-019 remain unused.

## Exclusions (What NOT to Build)

The following items are out of scope for this SPEC.

### Out of Scope — Retired items of the card body (absorbed by 3.2-2-2)

- Old item (4), lens-4 defect 8: run migration (a new run row reading stage, evidence, SPEC id, and kickoff decisions from the prior run), and the interim guard on the empty-stage to run edge that the card body names as its stopgap (old REQ-FRR-012 and REQ-FRR-013 are retired).
- Old item (7), lens-4 defect 7: holder audit re-entry at plan-audit and sync-audit, including the sync-audit to run edge and the holder evidence re-stamp (old REQ-FRR-008, REQ-FRR-009, and REQ-FRR-010 are retired; the current REQ-FRR-008 to -010 are different requirements), and the non-holder re-audit part of defect 7.
- Old item (5), lens-4 defect 5: changing the SPEC id while a lease is held (the SPEC id is immutable once picked).
- Old item (6), lens-4 defect 6: accepting the documented kickoff audit-ready line form.
- Old item (8), lens-4 defects 4 and 12: the stage-to-kickoff path that commits the transition and then returns an error, and the stage verb that opens the assignment-to-lease edge outside the hold, serial, and quota gates.
- Old item (9), lens-4 defects 16, 17, and 18: needs-decision and failed producers, leader reassignment and collection of expired cards, and lease columns left on abandoned rows.
- Old item (10), lens-4 defect 15: a reclassification verb for queued cards.
- Old item (11), lens-4 §7 case (d): the stage verb's help text and artifact-absent message.
- Old item (12), lens-4 defect 23: help and description drift across documents and help text.

### Out of Scope — Lens-4 defects not in the card body

- Broker defects (lens-4 defects 3, 11, and 21): message handover on a generation change, token rotation with null lists, and the byte-based truncation rule.
- Lens-4 defects 10, 13, 14, 19, 20, 22, 24, 25, and 26 are not addressed by this SPEC and have no follow-up assigned yet.

### Out of Scope — Adjacent cards and remote landing

- t1616 (lane-13): the local main landing. It does not own the done transition, and this SPEC does not place the pushed-to-done opening in it; that opening is transferred to 3.2-2-2 by `d-20261010T064450Z-68b2`.
- t1618 (picked; this lane has not verified its progress): owns the ceiling engine and the verdict file format. It references this SPEC for the lease renewal verb and the TTL configuration key, which this SPEC owns. This SPEC does not change the verdict parser.
- Remote push and pull requests: this repository lands locally, with no push and no pull request from the lane. The leader's batch push is outside this SPEC.

### Out of Scope — Item (3), transferred to 3.2-2-2 (t1639)

- The pushed → done edge and its decider path, with the refusals of that request (old REQ-FRR-016 and REQ-FRR-017; B4). This SPEC adds no decider surface.
- The `moai factory doctor` out-of-band write detection, its empty-sweep reporting, and its schedule (old REQ-FRR-018 and REQ-FRR-019).
- The record-call transaction boundary of the lease acquisition and the order of the expiry check against it (B2). REQ-FRR-008 states the outcome; the placement is 3.2-2-2's.
- The non-holder expiry ordering (B5): what a caller other than the holder receives on an expired lease, and whether that caller's request collects the lease. Only the holder's own expired row remains in this SPEC.
- The pushed → ci-green edge and the CI verdict reader that would admit it. That edge stays reserved, and this SPEC does not open it.
- The re-authored count pin of accepted transition pairs (72 accepted and 369 refused). The count test in the code is unchanged at 71 accepted and 370 refused.

## Residual risk

- **Lens-2 D1 (P0) stays open.** The empty-stage edge (T4b) admits a card with a SPEC id without a plan verdict. Its interim guard was part of old item (4), which `d-20261010T053435Z-006f` absorbs into 3.2-2-2. Until that lands, this edge is unguarded in this revision.
- **Non-holder expiry ordering (B5, transferred).** This SPEC does not change the order in which the transition path and the renewal path judge an expired lease: a caller other than the holder still has the lease collected before the holder check (`card_transition.go:309-316` and `:452-460`). A card whose holder does not return is therefore not restored by this SPEC; its restoration is the transferred item (old item 9, 3.2-2-2).
- **Watchdog wake schedule (REQ-FRR-005).** The renewal interval is bounded at one half of the lease duration. The watchdog runs only when a lane wakes it, and the standing wake schedule (`.claude/rules/moai/workflow/auto-semantics.md:108`) is not changed by this SPEC. A wake schedule slower than the bound lets the lease lapse between passes; whether the schedule changes is an open question (`progress.md` § G).
- AC-FRR-005 checks procedure text, not runtime behaviour, so a text-only change passes it.
