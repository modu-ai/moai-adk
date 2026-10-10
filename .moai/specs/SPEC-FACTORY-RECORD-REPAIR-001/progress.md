## §E.1 Plan-phase Audit-Ready Signal

_<pending plan re-audit>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Transfer record (3.2-2-2, t1639)

Record: operator-delegated decision `d-20261010T064450Z-68b2` (영실이 판단, delegated by the operator; option (b)). It resolves hold `d-20261010T064355Z-4fb1`. It moves the item-(3) subject of `d-20261010T060606Z-c5c4` out of this SPEC, and it leaves the retirements of `d-20261010T053435Z-006f` in place. The 1a19 verdicts (`d-20261010T053034Z-1a19`) stay on their surviving rows. The Operator verdict line of Q21 stays empty.

The moved identifiers and sub-parts are carried by 3.2-2-2 (t1639). None of them is active in this SPEC.

### Moved requirements and criteria (item 3)

| Moved identifier | Was | Reason for the move |
|---|---|---|
| REQ-FRR-016 | pushed → done edge under the receipt and containment conditions (B4) | Its trigger, "the human decider", has no surface that can make the request. `decideOne` (`internal/cli/factory_card.go:2803`, request at `:2893`) takes only merged-local and merged-PR cards, and abandon, block, resume, and unblock do not reach done. Plan M8 adds no decider surface. The delegated decision replaces the edge with 3.2-2-2's landing-to-done machine transition. It is outside the four remaining items. |
| REQ-FRR-017 | refusals of the pushed → done request; the other edges keep their refusals | It refuses the edge that REQ-FRR-016 names, so it moves with that edge. |
| REQ-FRR-018 | `moai factory doctor` out-of-band write detection | The doctor is not one of the four remaining items (68b2). |
| REQ-FRR-019 | doctor read-only and empty-sweep reporting | The same reason as REQ-FRR-018. |
| AC-FRR-017 | verifies REQ-FRR-016 | Verifies a moved requirement. |
| AC-FRR-018 | verifies REQ-FRR-017; carries the regression guards for merged-local → done and ci-green → done | Verifies a moved requirement. The guards sit in the same criterion. |
| AC-FRR-019 | count pin of 72 accepted and 369 refused | The pin counts the pushed → done edge, and no remaining item needs it (the verdict's remove-unless-it-serves rule). The code test `TestFR_AC005_TransitionTableEdgeCount` stays unchanged at 71 accepted and 370 refused; re-authoring it moves with the edge. |
| AC-FRR-020 | verifies REQ-FRR-018 | Verifies a moved requirement. |
| AC-FRR-021 | verifies REQ-FRR-019 | Verifies a moved requirement. |

Milestone M8 and plan pre-flight item 6 move with item (3). Plan premises (l) stay in `plan.md` as evidence for 3.2-2-2.

### Moved sub-parts of the first batch (B2, B4, B5)

| Sub-part | Where it was | Reason for the move |
|---|---|---|
| B2: record-call transaction boundary and the order of the expiry check against it | REQ-FRR-011 placement clause (now REQ-FRR-008); plan M1 item 5 mechanism; plan M3 item 1 materialization call; plan § C item 5 retry check; plan § F "Closed before M1" mechanism bullet | `RecordCardWorktree` opens its own write transaction (`internal/homestate/card_transition.go:353-357`, `card_worktree.go:33`), so a record call inside the lease transaction is a nested write (codex P1, audit iteration 3, plan.md:182). `factoryEnsureCardWorktree` lives in `internal/cli`, which `transitionTx` cannot call. The transaction-scoped recorder is 3.2-2-2's design. REQ-FRR-008 keeps its outcome clauses, because the re-lease in REQ-FRR-006 is an acquisition and must not leave an unstamped lease (decision `d-20261009T182157Z-8ccf`). |
| B4: decider path to done | REQ-FRR-016; plan M8; decision-index Q21, item 2 | As in the first table. |
| B5: non-holder expiry ordering | REQ-FRR-014's expired-lease and non-holder transition clauses (now REQ-FRR-009, narrowed to a live lease); AC-FRR-014's expired-lease case and its test `TestNonHolderCannotReapOrRenew`; plan M2 item 1 and its RED bullet | `transitionTx` and `RenewLease` collect an expired lease before any holder check (`card_transition.go:309-316`, `:452-460`). A refusal with no write for a non-holder would remove that collection, and a card whose holder does not return would then be stranded (audit iteration 3, B5). The delegated decision keeps only the holder's own expired row in this card. |

### Narrowed in this SPEC, not moved

- REQ-FRR-004: the refusal of an expired renewal now applies to the holder only. B5 cites it as the other half of the non-holder collection, and the delegated decision keeps only the holder's own expired row. The decision names REQ-FRR-014 explicitly; the narrowing of REQ-FRR-004 is the author's reading, recorded for the operator to confirm.
- REQ-FRR-009 (old REQ-FRR-014): narrowed to a non-holder's renewal of a live lease.
- REQ-FRR-008 (old REQ-FRR-011): the placement clause is removed; the outcome clauses are kept.

### Changed in this SPEC, not moved

- REQ-FRR-005 (B6): the renewal interval is bounded at no more than one half of `workflow.factory_lease.duration`; a refused renewal is recorded and the measurement still runs (AC-FRR-005); the 20-minute cadence is removed from spec.md (audit lines 59 and 117). The wake schedule in `auto-semantics.md:108` is not changed by this SPEC.
- AC-FRR-015 (REQ-FRR-010, B3): the Given is restated through `workflow.factory_lease.duration` and the factory handle; the queue claim stays at 15 minutes.
