---
id: SPEC-TODO-CLASSIFY-DISPATCH-001
title: "LLM-classified card metadata at creation — priority·blocked·execution-mode on every card, a priority-sorted queue, mode-aware factory lane leases, and default-on auto-dispatch for -f lanes"
version: "0.5.0"
status: in-progress
created: 2026-09-29
updated: 2026-10-04
author: manager-spec (card t1332)
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/homestate, internal/hook, internal/config"
lifecycle: spec-anchored
tags: "todo, factory, classification, priority, execution-mode, serial, parallelizable, auto-dispatch, lease, queue-sort, card-t1332"
tier: M
card: t1332
amendment_of: SPEC-TODO-CLASSIFY-DISPATCH-001
depends_on: [SPEC-FACTORY-SELF-DISPATCH-001]
related_specs: [SPEC-MANAGER-TODO-001, SPEC-TODO-HOLD-STATE-001, SPEC-AUTONOMY-CONTRACT-001, SPEC-KANBAN-TODO-CLI-001, SPEC-TODO-SQLITE-001]
---

# SPEC: LLM-classified card metadata at creation

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-29 | Initial plan-phase authoring (card t1332, operator directive 2026-09-29), measured in worktree `.moai/worktrees/t1332`, branch `WT-card-autodispatch`, at HEAD `145c3d98c` (== local develop tip). The measured surface basis is exported at `.moai/reports/t1332/surface-notes.md`. Card premises verified against the tree: the t1240 self-dispatch surface (`factory next`/`stage`/`complete`) exists on branch `WT-factory-self-dispatch` (tip `d43e50bb3`, 28 commits ahead of develop at plan time — the card said 25; the difference is an absorb of develop, and the surface matches the card's description) and is NOT merged to develop; `internal/cli/contract_decide.go` and the `homestate`/`contract` decider vocabulary are the t1261 layer this SPEC reuses rather than re-plans; t1306's `/moai:todo --auto` serial cycle is merged on develop. No card premise was falsified. |
| 0.2.0 | 2026-09-29 | **Leader-ruling revision** (plan-audit PASS 0.875 iter 1; D1 + OD rulings folded, one repair pass). **D1 (provenance-only)**: the t1240-only `--clear-policy` citations in plan.md §B.5/§F M4 and surface-notes.md items 5/8 now carry the `WT-factory-self-dispatch:` prefix (the flag is absent from this tree's develop — measured 0 `ClearPolicy` hits); no REQ/AC content changed. **OD-1 REVISED (operator ruling 2026-09-29)**: pipeline exclusivity REJECTED; serial-card mutual exclusivity ADOPTED — a serial card blocks only OTHER SERIAL cards, served one at a time in priority order; parallelizable selection is unaffected while a serial card is in flight. Rationale recorded: a full `next` refusal costs half the throughput benefit that classification exists to deliver. REQ-TCD-008 rewritten, REQ-TCD-009 extended, AC-TCD-008 re-derived (three clauses). **OD-3 REVISED (same ruling)**: decider-failure default is now `serial` (fail-safe — a parallelizable failure default could run true-serial cards concurrently and violate ordering; the serial default costs throughput only); REQ-TCD-003 rewritten. The new serial failure-default vs REQ-TCD-014's parallelizable absent-field read default is a recorded TENSION, flagged for the lead in plan.md §C — not silently resolved. Provenance: leader Jev doctrine-fallback rulings, noul below threshold (OD-1 0.31, OD-3 0.36), dated 2026-09-29. REQ and AC counts unchanged (14 / 14). |
| 0.3.0 | 2026-09-29 | **Tension resolved (leader ruling 2026-09-29, OD-3 extension).** REQ-TCD-014's absent-field READ default for the MODE axis flips `parallelizable` → `serial`; the PRIORITY axis default (`normal`) and BLOCKED axis default (`false`) are untouched. Rationale recorded: absent = unclassified = conservative treatment; a parallelizable absence default would silently bypass OD-3's serial failure default (a fail-safe bypassed through silence). plan.md §C FLAGGED item becomes RESOLVED with this provenance. No new read-default AC added (auditor optional D4 stays run-phase discretion); no existing AC asserted the old parallelizable absence default (measured: zero default-value hits in acceptance.md). REQ and AC counts unchanged (14 / 14). |
| 0.3.1 | 2026-09-29 | **OD-2 resolved to judgment-file transport (leader ruling 2026-09-29, card t1332 remediation F2).** REQ-TCD-012 reworded to shipped reality: the product provides the decider seam (`--classification-file`, the sole injection path) with the default decider (`DefaultCardDecider`, identity `default`) as the shipped judgment backend; the `Decider(llm)` implementation is recorded out of scope for this SPEC (deferred to a follow-up card). plan.md §C OD-2 moves open → RESOLVED with the judgment-file option named as the adopted candidate (the plan's own OD-2 candidate list anticipated this). The Jev-isolation clause is untouched; no other REQ strengthened or weakened. Provenance: operator standing-delegation path, Jev ask noul 0.8 (gate 0.50), 2026-09-29. REQ and AC counts unchanged (14 / 14). |
| 0.4.0 | 2026-10-03 | **In-place amendment of the `completed` SPEC (card t1407; operator decision 2026-10-03, relayed through the factory leader; manager-spec re-delegation per D-NEW-1).** Three serial-slot behaviors of `moai factory next` now differ from the literal text of REQ-TCD-008 / AC-TCD-008(b): (1) a serial card whose lease has expired does not hold the serial slot; (2) a merely `assigned` serial card holds the slot against every arm that takes a NEW card, but not against a lane leasing the card assigned to itself; (3) `failed` is terminal and releases the slot. A known limitation (non-atomic slot read versus claim) is recorded and left open. The original REQ-TCD-008 and AC-TCD-008 text is byte-unchanged; `## Amendments` carries the structured record and supersedes that text where they conflict. No requirement deleted, no id renumbered, no AC mapping changed (14 / 14). progress.md §E.4 `sync_commit_sha` is left at the prior close. Status moves `completed → in-progress` per the SSOT amendment transition (`spec-frontmatter-schema.md` § Status Enum). |
| 0.5.0 | 2026-10-04 | **In-place amendment of the `completed` SPEC to record a supersession (card t1458; decision DL-4 of SPEC-FACTORY-ATOMIC-LEASE-001 §H, default "record it"; manager-spec re-delegation per D-NEW-1).** SPEC-FACTORY-ATOMIC-LEASE-001 (0.3.1, `completed`) lists in its §E which statements of the 0.4.0 "Known limitation" paragraph it supersedes; `## Amendments` carries one new entry that records them. No requirement text, no REQ-TCD-008 / AC-TCD-008 wording and no id is changed; counts stay 14 / 14. progress.md is not touched here. Status moves `completed → in-progress` per the SSOT amendment transition and returns to `completed` in the re-close commit. |

> **Provenance discipline.** Every `file:line` citation was measured at HEAD `145c3d98c` in this
> worktree. t1240-branch citations are prefixed `WT-factory-self-dispatch:` and were read via
> `git show` — that branch is NEVER merged or checked out by this SPEC's plan phase, and its
> absorption into develop is a run-phase entry precondition (REQ-TCD-013).

## Amendments

**2026-10-03 — v0.4.0 — in-place amendment of the prior `completed` SPEC (card t1407).**

- Transition: `completed → in-progress` per the SSOT amendment contract (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum, `completed → in-progress (amendment)`); `amendment_of: SPEC-TODO-CLASSIFY-DISPATCH-001` (self-referential). The schema forces this status transition for an in-place amendment of a `completed` SPEC; it was not a free choice.
- Prior completed version: **0.3.1** — closed 2026-09-29 (card t1332 sync lane).
- `prior_completed_sha: 34f09f34d` — the prior close's `sync_commit_sha` (this SPEC's progress.md §E.4); that field is left unmodified.
- Reason: operator decision 2026-10-03, relayed through the factory leader (card t1407). The serial exclusivity of REQ-TCD-008 / AC-TCD-008(b) is implemented in `internal/cli/factory_card.go` (`factoryNextSelectAndLease`, `factorySerialSlotFree`, `factorySerialSlotHeld`); the factory leader observed the literal text deadlocking the factory in run tm9i7y — eight leader-assigned serial cards, every `factory next` refused. That observation was not re-measured by the amending card, which reproduces the two-card shape in a fixture; the deadlock reproduction test is `TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned`.
- Scope — three behaviors, each superseding the original text where they conflict:

  1. **Expired lease.** A serial card in a lease-holding state whose lease has expired (`homestate.Card.LeaseExpired(now)`, `internal/homestate/card_record.go`; expired when `now` is at or after `lease_expires_at`, and an unparseable expiry reads as expired) does NOT hold the serial slot. An expired lease is collected lazily, so the row keeps its lease-holding state after its lane is gone; the original text counts any non-terminal serial card as holding, and this narrows REQ-TCD-008 and AC-TCD-008(b). The selection clock is read before the record snapshot, so a renewal landing between the two cannot be read as an expiry. A live lease holds the slot in every lease-holding state.
  2. **`assigned` serial cards (operator option B).** An `assigned` serial card (no lease) still holds the slot against every arm that takes a NEW card (selection arms b, b2, c). The one exception is arm (a): when a lane leases the card assigned TO ITSELF, sibling cards that are merely `assigned` are not counted; a sibling that is actually in flight (live lease) still blocks it, and `picked` rows still count. Effect: serial cards leader-assigned to several lanes no longer deadlock each other, yet remain served one at a time, because the first lease then holds the slot against the rest. Ordering consequence: arm (a) serves a lane's own assigned serial card in the order the lanes call `factory next`, not in priority order — it does not consult priority — so the priority-ordered serving of REQ-TCD-008 applies to the arms that take a NEW card (b, b2, c).
  3. **`failed` is terminal.** `homestate.IsTerminalCardState` lists `done`, `failed`, `abandoned`, and no transition leaves `failed`; the positive releasing-state enumeration of `factorySerialSlotFree` now includes `failed`. REQ-TCD-008 already says the terminal states re-admit serial selection, enumerated positively — this makes the enumeration complete, it does not add a new rule. Non-terminal parked states (`blocked`, `needs-decision`) and `picked` rows still hold the slot.

- AC-TCD-008(b), 한국어 판독: 원문은 high serial 이 terminal 에 도달하기 전까지 어떤 레인도 low serial 카드를 임대하지 못한다고 적었는데, 이를 아래처럼 고쳐 읽는다. 새 serial 카드를 가져가는 경로에서는 다른 serial 카드가 임대 중이거나 `assigned`·`picked` 상태이면 슬롯이 막힌다. 다만 임대가 만료된 카드와 `failed` 로 끝난 카드는 슬롯을 쥐지 않는다. 레인이 자신에게 배정된 카드를 임대하는 경로에서는 단지 `assigned` 인 형제 카드가 슬롯을 막지 않으며, 실제로 임대 중인 형제만 막는다.
- **Known limitation (not closed by this amendment; the operator will take it as a separate follow-up).** The serial slot is read from a record snapshot and the claim happens afterwards, so two lanes selecting concurrently can both lease a serial card. The factory record and the todo queue are two stores, so an atomic lease across both is future work. This amendment makes no atomicity claim. Under option B, two lanes each leasing their OWN assigned serial card at the same moment is now an ordinary path, so this non-atomic window is reached more often than before; it is still not closed here. The atomic lease is a separate follow-up card (the leader named it t1458; that id was given by the leader and is not verified in the repository).
- **Residual not repaired.** An ownerless `picked` serial row and an `assigned` serial row wait on each other: arm (a) does not lease the assigned card past a `picked` sibling (only merely-`assigned` siblings are ignored), and the picked row cannot be taken through arm (b) while the assigned card holds the slot. This follows from the scope of the operator ruling and is read from `factoryNextSelectAndLease`; the two halves are pinned by `TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling` (arm a) and the arm-b subtest of `TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms`.
- Original text: REQ-TCD-008 in §B and AC-TCD-008 in `acceptance.md` are unchanged; this entry supersedes them where they conflict. No requirement deleted, no id renumbered, no AC mapping changed. Requirement and AC counts unchanged: 14 / 14.
- Evidence: branch `WT-factory-serial-slot-stale-lease`, `git log --oneline 7109e0900..HEAD` — `250c03899` (RED, expired lease), `abae30f67` and `e4e967ef8` (behavior 1), `e2123a271` (RED, failed row), `c28eb4d70` (behavior 3), `fb5217710` (RED, own assigned card), `9499ec834` (fix, behavior 2), and `3cb71dee8` (test-only: picked-sibling and parallelizable controls). Tests, all in `internal/cli/factory_serial_slot_stale_test.go`: `TestFactoryNextExpiredLeaseReleasesSerialSlot`, `TestFactoryNextSerialSlotLeaseExpiryBoundary`, `TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState` (behavior 1); `TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned`, `TestFactoryNextOwnAssignedSerialCardBlockedByLiveSerialLease`, `TestFactoryNextAssignedSerialCardStillHoldsSlotAgainstNewTakes`, `TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms`, `TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling`, `TestFactoryNextParallelizableLeasesBesideLiveSerial` (behavior 2 and its controls); `TestFactoryNextFailedSerialRowReleasesSlot` (behavior 3).

**2026-10-04 — v0.5.0 — supersession record for SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458).**

- Transition: `completed → in-progress` per the SSOT amendment contract (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum, `completed → in-progress (amendment)`); `amendment_of: SPEC-TODO-CLASSIFY-DISPATCH-001` (self-referential, set by the 0.4.0 amendment and kept). The SPEC returns to `completed` in the re-close commit of the same delegation.
- Prior completed version: **0.4.0** — the amendment re-close of card t1407 (commit `f628fb2d8`, 2026-10-03).
- `prior_completed_sha: 34f09f34d` — the `sync_commit_sha` that this SPEC's progress.md §E.4 carries (the first close, v0.3.1). The amendment re-close `f628fb2d8` is recorded in the same file under `amendment_sync_commit_sha`, a key the era parser does not read. Both commits were confirmed present with `git cat-file -e`.
- Reason: SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458, 0.3.1) states in its §E which clauses of this SPEC it supersedes, and records each through the Amendments mechanism so that a reader of this SPEC does not read them as open (its §H DL-4: default "record it", the operator may veto). This entry records the supersession as that SPEC states it; it does not re-measure it.
- Scope — two statements of the 0.4.0 "Known limitation" paragraph above are superseded, and one is left alone:
  1. **Superseded for every lease arm** (REQ-FAL-002, REQ-FAL-005 of SPEC-FACTORY-ATOMIC-LEASE-001): "The serial slot is read from a record snapshot and the claim happens afterwards, so two lanes selecting concurrently can both lease a serial card. The factory record and the todo queue are two stores, so an atomic lease across both is future work." The same paragraph's parenthetical that the leader named the follow-up t1458 without it being verified in the repository is answered too: the card exists and is the SPEC named above.
  2. **Superseded** (REQ-FAL-005): "two lanes each leasing their OWN assigned serial card at the same moment is now an ordinary path … it is still not closed here."
  3. **Untouched.** The "Residual not repaired" paragraph (an ownerless `picked` serial row and an `assigned` serial row waiting on each other) is neither repaired nor altered by SPEC-FACTORY-ATOMIC-LEASE-001 and stays exactly as written above.
- Original text: REQ-TCD-008 in §B and AC-TCD-008 in `acceptance.md` are unchanged, as are the three behaviors of the 0.4.0 entry. The two superseded statements stay in place above as the record of what was known on 2026-10-03; this entry supersedes them where they conflict. No requirement deleted, no id renumbered, no AC mapping changed. Requirement and AC counts unchanged: 14 / 14.
- Evidence: SPEC-FACTORY-ATOMIC-LEASE-001 `spec.md` §E (the supersession table) and REQ-FAL-002 / REQ-FAL-005; its AC-FAL-014 is the criterion this entry satisfies for this SPEC.

---

## §A Context

### A.1 The gap: the queue is an unordered pile to every machine reader

Today a card carries no judgment about how urgent it is or how it may be executed. The queue item
is five frozen fields plus additive evidence stamps (`backlog_store.go:83-110`); order is insertion
order, and every machine consumer reads that order as priority. `moai factory next`'s
auto-promotion arm takes "the first `queued` item in slice order"
(`WT-factory-self-dispatch:internal/cli/factory_card.go`, arm (c)) — the oldest card, whatever its
urgency, and regardless of whether the lane running it can proceed concurrently with others. The
operator's judgment lives — at best — in card prose, which machines do not read (the same
text-invisibility defect SPEC-TODO-HOLD-STATE-001 §A.1 recorded for holds).

### A.2 What this SPEC adds — three pieces, exactly

Per the operator directive (card t1332):

1. **Creation-time classification.** `moai todo add` classifies every admitted card through a
   decider seam and records the judgment as card metadata: a priority ordinal, a blocked flag
   (urgency and blocked-ness), and an execution mode (`serial | parallelizable`). The queue is
   kept sorted by priority.
2. **Mode-aware factory dispatch.** `factory next` picks up cards in priority order;
   parallelizable cards allow distinct cards to be leased by distinct lanes concurrently; a serial
   card excludes other SERIAL cards — while one is in flight, no other serial card is leased
   (served one at a time in priority order) and parallelizable selection is unaffected; the
   factory record/status shows which lane holds which card.
3. **Auto-dispatch default-on for `-f` lane sessions.** A `-f` lane self-dispatches by default;
   manual specification remains a launcher flag option.

### A.3 The decider seam, in the existing vocabulary

The product already carries a decider-identity vocabulary this SPEC reuses, not duplicates: the
kickoff contract's closed set `human | llm | llm+jev` (`internal/contract/seal.go:55-56`), the
"jev is never a sole decider" refusal (`internal/contract/kickoff/decide.go:111-112`), and the
`homestate.Decider` identity recorded on card decisions (`internal/homestate/card_record.go:45-46`).
This SPEC defines the classification seam (a `Decider` abstraction for card classification); its
product implementation is the `--classification-file` judgment-file injection with
`DefaultCardDecider` (identity `default`) as the shipped judgment backend and a deterministic
absence-of-judgment fallback, while the `Decider(llm)` implementation is deferred out of scope for
this SPEC (REQ-TCD-012, leader ruling 2026-09-29 OD-2).
`scripts/jev/` is LOCAL-ONLY tooling (AGENTS.local.md §29) and is never wired into any product
path; the local dogfood seam is stated in plan.md §D.3.

## §B Requirements (GEARS)

> Value sets used throughout: `priority ∈ {high, normal, low}`; `blocked ∈ {true, false}`;
> `mode ∈ {serial, parallelizable}`; decider identity reuses the contract vocabulary
> (`llm` for the product implementation; the deterministic fallback records `default`).

- **REQ-TCD-001 (Ubiquitous)** — The `moai todo add` path shall classify every admitted card
  through the card-classification decider seam at creation time, inside the same locked write that
  appends the card, so no card becomes visible to a machine selector without a recorded
  classification.

- **REQ-TCD-002 (Ubiquitous)** — The queue item shall carry the classification as ONE additive
  nullable field on `BacklogItem` (priority, blocked, mode, decider identity, classified-at stamp,
  one-line reason) following the `Landing`/`PickedAt` additive discipline: a pointer, `omitempty`,
  absent-field marshal byte-identical to the pre-SPEC record, and the SQLite mirror added by a
  guarded `ALTER TABLE ... ADD COLUMN` (`backlog_sqlite.go:480-550` precedent).

- **REQ-TCD-003 (Event-driven)** — When the classification decider is unavailable or fails, the
  add path shall admit the card with the fail-safe default (priority `normal`, blocked `false`,
  mode `serial`, decider identity `default`), emit one stderr notice naming the fallback, and
  record the decider identity — the serial default is the fail-safe direction, because a
  wrongly-parallel default could run true-serial cards concurrently and violate the ordering the
  mode exists to protect, while a wrongly-serial default costs throughput only; admission never
  blocks on classification.

- **REQ-TCD-004 (Event-driven)** — When a classification is supplied explicitly to `todo add`
  (judgement file or standard input, mirroring `contract decide --judgement`), the add path shall
  validate it against the closed value sets before the locked write, refuse an out-of-set value as
  a usage error with nothing written, and record the supplying decider's identity; the decider
  identity `jev` shall be refused on every product classification path.

- **REQ-TCD-005 (Ubiquitous)** — The queue shall be kept sorted by the classification key —
  non-blocked before blocked, then priority `high > normal > low`, then insertion order stable
  within a rank — and the sort shall be re-established inside the same locked write as the add
  that may change it; no other verb reorders ranked cards.

- **REQ-TCD-006 (Ubiquitous)** — The queue position `add` prints, the `todo list` rendering, and
  the web queue read shall reflect the priority-sorted order, so every consumer that reads
  position reads priority.

- **REQ-TCD-007 (Event-driven)** — When a lane runs `factory next` and selection reaches the
  auto-promotion arm, the factory shall promote and lease the highest-ranked eligible queued card
  in sorted order, and shall never auto-select a `blocked` card (an operator pick or unblock is
  the only path that dispatches one).

- **REQ-TCD-008 (Ubiquitous)** — A `serial` card shall hold mutual exclusivity against other
  serial cards: while a serial card is recorded in the factory run in a non-terminal state,
  `factory next` shall lease no serial card for any lane — serial cards are served one at a time
  in priority order, and a lane finding only serial candidates blocks on the existing no-card
  exit — while parallelizable candidates remain leasable by other lanes throughout; the terminal
  states that re-admit serial selection shall be enumerated positively, never as a negation.

- **REQ-TCD-009 (Ubiquitous)** — `parallelizable` cards shall allow concurrent multi-lane
  dispatch: distinct parallelizable cards may be held by distinct lanes at the same time —
  including while a serial card is in flight — and the version-checked lease edges
  (`ErrStaleVersion`/`ErrLeaseHolder` race retry) remain the only duplicate-prevention mechanism
  — a card is never leased twice.

- **REQ-TCD-010 (Ubiquitous)** — The factory record and the `factory status` surface shall show
  which lane holds which card together with the card's execution mode and priority (the
  `factoryCardView` extension), so a reader of status can see the mode-aware dispatch state
  without opening the queue.

- **REQ-TCD-011 (Where)** — Where a `-f` lane session is launched, auto-dispatch — the lane
  entering its `factory next` self-dispatch loop without a per-card lead routing step — shall be
  the default behavior, the launcher shall stamp the selection into the lane bootstrap, and a
  launcher opt-out flag shall select the manual mode; the default is recorded in code, and no new
  config key is introduced.

- **REQ-TCD-012 (Ubiquitous)** — The product shall provide the card-classification decider seam —
  the `--classification-file` judgment-file injection as the sole product injection path
  (REQ-TCD-004) — with the default decider (`DefaultCardDecider`, identity `default`) as the
  shipped judgment backend; the `Decider(llm)` implementation is out of scope for this SPEC
  (deferred to a follow-up card, leader ruling 2026-09-29 OD-2); and no product path
  (`internal/`, `pkg/`, `cmd/`, `internal/template/templates/`) shall reference `scripts/jev/`
  or any local-only tooling path.

- **REQ-TCD-013 (Event-driven)** — When run-phase entry is attempted before local develop carries
  the t1240 self-dispatch surface (SPEC-FACTORY-SELF-DISPATCH-001 merged; `factory next` and
  `complete` present in develop's `internal/cli`), the run shall be blocked — the absorption order
  is a run-phase entry precondition, and the absorption source is branch
  `WT-factory-self-dispatch`.

- **REQ-TCD-014 (Event-driven)** — When a queue recorded before this SPEC is read, every
  classification consumer shall derive the no-judgment default for a card whose classification
  field is absent by positive default derivation at read — priority `normal`, blocked `false`,
  mode `serial` (leader ruling 2026-09-29, OD-3 extension: absent = unclassified = conservative
  treatment; a parallelizable absence default would silently bypass the serial failure default
  through silence) — never by treating absence as a fourth priority or a third mode.

## §C Constraints

- C1. The five-field per-item contract of REQ-TODO-013 keeps its names, types, and JSON tags; the
  classification field is additive after the `Landing` precedent (`backlog_store.go:92-100`).
- C2. The lead-side serial cycle `/moai:todo --auto` (SPEC-MANAGER-TODO-001, merged) is untouched;
  it consumes the queue in queue order and thereby inherits the sorted order without change.
- C3. Hold semantics (SPEC-TODO-HOLD-STATE-001) are untouched: a `hold` card stays invisible to
  every machine selector; sorting positions it, selection filters it.
- C4. Jev boundary: `scripts/jev/` is never referenced by product code (REQ-TCD-012); the existing
  `internal/jev` package stays behind its own `workflow.jev.enabled` default-off gate and gains no
  new consumer from this SPEC.
- C5. No new config key (REQ-TCD-011); the auto-dispatch default is a launcher code default.

## §D Dependencies

- **depends_on: SPEC-FACTORY-SELF-DISPATCH-001** (t1240) — the `factory next`/`stage`/`complete`
  lease surface, the version-checked claim edges, and the no-card exit this SPEC extends. Not yet
  merged to develop; absorption order is REQ-TCD-013.
- **related: SPEC-MANAGER-TODO-001** (t1306, merged) — the lead-side serial cycle that inherits
  the sorted queue. **SPEC-TODO-HOLD-STATE-001** (t1308, merged) — the fourth state and the
  positive-enumeration discipline REQ-TCD-008/014 follow. **SPEC-AUTONOMY-CONTRACT-001** — the
  decider vocabulary REQ-TCD-004/012 reuse. **SPEC-KANBAN-TODO-CLI-001 / SPEC-TODO-SQLITE-001** —
  the queue store and its additive-schema discipline.

## §E Out of Scope

### Out of Scope — lane self-lease verbs (t1240 / SPEC-FACTORY-SELF-DISPATCH-001)

- The `factory next`/`stage`/`complete` verb surface, the duplicate-prevention lease mechanics,
  and the admission/refusal asymmetry are t1240's deliverables; this SPEC only reorders the
  selection input and adds the mode-aware eligibility gate on top. Merging, checking out, or
  re-planning that branch is forbidden.

### Out of Scope — lead-side serial auto-pickup (t1306 / SPEC-MANAGER-TODO-001)

- The `/moai:todo --auto` foreman cycle, its evidence-judged completion, and its serial
  card-processing contract are merged and owned; no change here.

### Out of Scope — the kickoff contract decider layer (t1261)

- `contract decide`, the kickoff receipt, the R5 jev-sole refusal, and the
  `workflow.autonomy.kickoff.decider` config are owned there; this SPEC defines a separate
  classification seam and reuses only the identity vocabulary.

### Out of Scope — Jev product wiring

- `scripts/jev/` stays a local-only dev tool (AGENTS.local.md §29). No template mirror, no
  product-path reference, no new `internal/jev` consumer. The local dogfood seam is a plan.md
  concern (§D.3), not a deliverable.

### Out of Scope — scheduling intelligence

- The blocks/depends relation graph (t1309) stays record-only; no scheduler reads it. Priority is
  the classification's three-level ordinal — not a computed score, not a re-classification on
  edit, not a deadline system.

## §F Cross-references

- Measured surface basis: `.moai/reports/t1332/surface-notes.md`.
- Acceptance criteria: `acceptance.md` (AC-TCD-001..014, Given-When-Then, binary-testable).
- Implementation plan and open decisions: `plan.md`.
