---
id: SPEC-FACTORY-RECORD-REPAIR-001
title: "Factory record repair: lease renewal, one-shot expired-lease recovery, and the pushed-to-done edge (first batch, narrowed)"
version: "0.3.0"
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

## Scope

This SPEC is the first batch of the factory record repair for card t1621. Leader ruling `d-20261010T053435Z-006f` narrowed the card body to three items, and this SPEC carries them:

1. **Lease renewal.** A renew verb with its MCP tool, the lease-duration configuration key, and a watchdog renewal as its first step.
2. **One-shot recovery of an expired own lease.** Collect, re-lease, and apply the requested transition in one step, with a refusal that names the recovery command. The re-lease is a lease acquisition, so REQ-FRR-011 (worktree path at acquisition) applies to it.
3. **Pushed-to-done edge opening with out-of-band write detection.** Leader ruling `d-20261010T060606Z-c5c4` places this item in this card (option a). REQ-FRR-016 to REQ-FRR-019 carry it: the pushed-to-done edge, its refusals, and the `moai factory doctor` check for writes that bypass the transition API.

The SPEC changes `internal/homestate` and the factory verbs and MCP tools that call it. This SPEC (card t1621) owns the lease renewal verb and the TTL configuration key. Card t1618 references this SPEC for them and does not own them.

## Background

The factory record (`factory.db`) keeps one row per card and enforces a version-checked state machine. Four defects in that record bear on this batch:

- **Lease expiry (item 1).** A card lease lasts 15 minutes from acquisition and is extended only when a stage transition succeeds. No verb renews a lease on its own, and the lane watchdog never reads the lease (lens-4 defects 1 and 27).
- **Expired own lease (item 2).** An expired lease is reclaimed only when a later call reaches the card, and that call commits the reclaim and then returns an error. The next verb refuses a holder's own expired row as "in flight" (lens-4 defects 1 and 9; the card body cites `card_transition.go:311-316`).
- **Unstamped lease (decision `d-20261009T182157Z-8ccf`).** A lease does not record the card's worktree path. Because the re-lease of item 2 is also a lease acquisition, this SPEC records the path at acquisition.
- **Remote done edge (item 3).** When a remote exists, a card cannot reach `done` through the merged-local edge, which is refused while a remote is configured, and the reserved ci-green edge is the only other route (`card_transition.go:156-158` and `:203-208`; the card body cites `:207-209` and `:668-670`, which match this tree). The card body reports that 75 of the 76 done rows were written outside the transition API. That figure is the card body's claim; this SPEC did not measure it.

The sync-audit re-entry, the holder evidence re-stamp, and the empty-stage guard that the first revisions carried are retired by `d-20261010T053435Z-006f` (HISTORY 0.2.0). The verdict file format and the ceiling engine belong to t1618 and are not changed by this SPEC.

## Requirements

The requirements use GEARS notation. Each active requirement maps to at least one criterion in `acceptance.md`, and the mapping is in `plan.md` §E. The active requirements are REQ-FRR-001 to REQ-FRR-007, REQ-FRR-011, and REQ-FRR-014 to REQ-FRR-019; REQ-FRR-016 to REQ-FRR-019 are item (3) (HISTORY 0.3.0). The identifiers REQ-FRR-008 to -010, -012, and -013 are retired and are not reused (HISTORY 0.2.0).

### Item (1): lease renewal, TTL key, and watchdog first step (lens-4 defects 1 and 27)

- **REQ-FRR-001** When the lane session that holds a live card lease runs the factory renew verb for that card, the factory record shall extend the lease to the configured lease duration counted from the renewal, and the verb shall report the new expiry time.
- **REQ-FRR-002** The factory record shall expose the renewal as an MCP tool whose success and refusal results match the factory renew verb for the same card and caller.
- **REQ-FRR-003** The factory record shall take the card lease duration from a configuration key whose default value is 15 minutes.
- **REQ-FRR-004** When a renewal is requested for a lease that has already expired, the factory record shall refuse the renewal and leave the card's state, version, and event history unchanged.
- **REQ-FRR-005** When the lane stall watchdog starts a pass, the watchdog shall renew its own card lease before it measures lane progress. The renewal is the one write the watchdog adds. Its guarantee is limited: a renewal at the start of each watchdog pass keeps the lease alive only if consecutive passes start within the lease duration. The standing recheck cadence is 20 minutes and the lease is 15 minutes, so the gap between passes is not covered by this requirement. This SPEC does not change the cadence or the default lease duration; the gap is recorded as evidence for the TTL design, which this SPEC owns (card t1621).

### Item (2): one-shot recovery of an expired own lease (lens-4 defect 9)

- **REQ-FRR-006** When the holder of an expired card lease requests a transition through the factory stage verb, or through the factory next verb where that verb is reachable (run from the parent checkout), the factory record shall collect the expired lease, re-lease the card to the same holder, and apply the requested transition in one step, so that a refused transition leaves the expired row unchanged.
- **REQ-FRR-007** When a card lease is refused as expired, whether the refused request is a transition or a renewal, the refusal shall name the recovery command that the holder runs to re-lease the card: the factory stage verb (`moai factory stage`), which a holder inside its card worktree can run. REQ-FRR-006 applies that verb's recovery in one step. The factory next verb also recovers from the parent checkout but is not the named command, because it refuses to run from a card worktree.
- **REQ-FRR-011** When a lane acquires the lease on a card, including the re-lease of REQ-FRR-006, the factory record shall first materialize the card worktree that the factory provides for that card, and then record that worktree path in the same transaction that writes the holder, heartbeat, and expiry. If materialization fails, the factory record shall refuse the acquisition and leave the card unchanged. The factory record shall refuse a later acquisition that would record a different worktree path for the card, and shall not hold a lease on a card that has no recorded worktree path. The acquisition is the T3 transition.

### Cross-cutting

- **REQ-FRR-014** When a caller that does not hold the lease requests a renewal of a live or expired lease, or a recovery or any transition on a card whose lease has expired, the factory record shall refuse the request and leave the card's state, version, and event history unchanged. (Serves items 1 and 2.)
- **REQ-FRR-015** Where the card lease duration is configured, the factory record shall keep the queue claim lease duration unchanged. (Serves item 1's TTL key.)

### Item (3): pushed-to-done edge and out-of-band write detection (leader ruling `d-20261010T060606Z-c5c4`, option a)

- **REQ-FRR-016** When the human decider requests the `done` transition for a card in state `pushed`, and the request carries the card's backlog identity, and the leader approval receipt of the card verifies against the card's current version, evidence SHA, and owner label, and the card's merge commit is contained in the remote-tracking ref of the integration branch as that ref stands at the request, the factory record shall move the card to `done` in one transaction and record the move in the card's event history. (Serves item 3, the edge opening.)
- **REQ-FRR-017** When a request for the `done` transition of a card in state `pushed` fails any condition of REQ-FRR-016, the factory record shall refuse the request, name the failing condition in the refusal, and leave the card's state, version, and event history unchanged. The merged-local to `done` edge shall remain refused while a remote is configured, and the reserved `pushed` to `ci-green` edge shall remain refused. (Serves item 3, the refusals and the unchanged edges.)
- **REQ-FRR-018** When the operator runs `moai factory doctor`, the factory doctor shall compare, for each card of the factory record it reads (every run, or the single run the caller names), the card's recorded state and version with the card's event history, and shall report each card where they disagree, naming the card, the run, and both values. The history is the events that name the card and carry a version (`card.transition`, `card.fields`, `card.dispatch-owner`, and `lease.expired`). The recorded version agrees when it is the highest version in that history. The recorded state agrees when it is the state set by the card's latest event that sets a state (`card.transition` or `lease.expired`). A card with no version-bearing event is one such disagreement. A card recorded as `done` with no event that moves it into `done` is reported as written outside the transition API. The factory doctor shall exit with a non-zero status when it reports a disagreement. (Serves item 3, the out-of-band write detection.)
- **REQ-FRR-019** The factory doctor shall not write to the factory record. It shall report the number of cards it compared. A run in which it compares no cards shall be reported as an empty sweep and shall exit with a non-zero status. The factory doctor shall report a clean result only when it compared at least one card and reported no disagreement. (Serves item 3, the visibility of the check.)

## Exclusions (What NOT to Build)

The following items are out of scope for this SPEC.

### Out of Scope — Retired items of the card body (absorbed by 3.2-2-2)

- Old item (4), lens-4 defect 8: run migration (a new run row reading stage, evidence, SPEC id, and kickoff decisions from the prior run), and the interim guard on the empty-stage to run edge that the card body names as its stopgap (REQ-FRR-012 and REQ-FRR-013 are retired).
- Old item (7), lens-4 defect 7: holder audit re-entry at plan-audit and sync-audit, including the sync-audit to run edge and the holder evidence re-stamp (REQ-FRR-008, REQ-FRR-009, and REQ-FRR-010 are retired), and the non-holder re-audit part of defect 7.
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

- t1616 (lane-13): the local main landing. It does not own the done transition; leader ruling `d-20261010T060606Z-c5c4` places the pushed-to-done opening in this card (item 3).
- t1618 (picked; this lane has not verified its progress): owns the ceiling engine and the verdict file format. It references this SPEC for the lease renewal verb and the TTL configuration key, which this SPEC owns. This SPEC does not change the verdict parser.
- Remote push and pull requests: this repository lands locally, with no push and no pull request from the lane. The leader's batch push is outside this SPEC.

### Out of Scope — Item (3) boundaries

- The pushed → ci-green edge and the CI verdict reader that would admit it: that edge stays reserved, and this SPEC does not open it.
- Repair of an out-of-band write: `moai factory doctor` reports disagreements and never writes or repairs the record.
- Comparison of lease, heartbeat, expiry, and evidence fields by the doctor (see Residual risk).
- An MCP tool for the doctor, and any schedule that runs it.

## Residual risk

- **Lens-2 D1 (P0) stays open.** The empty-stage edge (T4b) admits a card with a SPEC id without a plan verdict. Its interim guard was part of old item (4), which `d-20261010T053435Z-006f` absorbs into 3.2-2-2. Until that lands, this edge is unguarded in this revision.
- **Item (3) detection limits.** The doctor compares state and version only. The lease renewal (`card_transition.go:465`) writes heartbeat and expiry with no version bump and no event, so an out-of-band change to those fields is not detected. The doctor is an on-demand verb, and nothing in this SPEC makes it run, so a write made while it does not run is invisible to it until it runs (verification-completeness §1.3). How often it runs is an open choice in decision-index Q21.
- The standing awaken cadence of 20 minutes is longer than the 15-minute lease, so a renewal at each watchdog pass alone does not keep a lease alive between passes (plan.md § Verified premises, item (h)). The gap is recorded as evidence for the TTL design, which this SPEC owns (card t1621).
- AC-FRR-005 checks the watchdog's procedure text, not its runtime behaviour, so a text-only change passes it.
