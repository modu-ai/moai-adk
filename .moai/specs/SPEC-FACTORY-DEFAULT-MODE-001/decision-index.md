# Decision Index: SPEC-FACTORY-DEFAULT-MODE-001

Decision gate: ON (`.moai/config/sections/interview.yaml` `interview.decision_gate: on`).
Board records are cited from `/Users/goos/.moai/db/moai-adk-go-1bd3d038/decisions/board.jsonl`;
each citation is pinned by the verbatim record line in a fenced block, with the sha256 digest
prefix of that line.

---

### Q1: Which mechanism carries the absent-classification default — read-time derivation or an add-time stamp of existing records?

- Label: DECIDED
- Authority anchor: `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/spec.md` § HISTORY, v0.3.0 row
  (2026-09-29) — the absent-field READ default is carried by the read derivation (the
  REQ-TCD-014 seam); the row re-valued that derivation, establishing the seam as the mechanism.
  This SPEC re-values it under the 2026-10-19 operator ruling; the mechanism question is
  unchanged.
- Why unresolved: the default VALUE is new here (it reverses v0.3.0), so the mechanism it rides
  had to be confirmed before authoring; the anchor's row decides the mechanism, not the value.
- Operator verdict:

### Q2: Does the decider-failure fail-safe stay serial when the absent-read default flips?

- Label: DECIDED
- Authority anchor: `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/spec.md` § HISTORY, v0.2.0 row
  (OD-3 REVISED, 2026-09-29) — the failure default is `serial` because "a wrongly-parallel
  default could run true-serial cards concurrently". This SPEC preserves that ruling
  (REQ-FDM-003) and only splits the absent-read value away from it.
- Why unresolved: the parent explicitly unified the two defaults ("the same defaults by
  ruling"); flipping one half required confirming the other half's ruling stands.
- Operator verdict:

### Q3: How does the reclassify verb keep the factory record's mode in step — a second store write inside the same lock, or a single-store property?

- Label: FOUNDER
- Class: implementation-level
- Why unresolved: the card text ("factory 쪽 mode도 같은 잠금 쓰기로 맞춤") presumes a
  factory-side mode field to sync. Measured at HEAD `74ad197380`: `homestate.Card`
  (`internal/homestate/card_record.go:100`) carries no mode field, and every factory-side reader
  derives the mode from the queue at read time (`internal/cli/factory_card.go:839-842`,
  `:1397`, `:3155-3160`). The cited parent-SPEC body (REQ-TCD-010's "read from the queue"
  comment) is not a HISTORY/Amendments row, so the register offers no citable anchor; the
  resolution below is a measured-tree judgment, recorded here for the audit trail.
- Default: single locked queue write; factory-side readers observe by derivation (rule: the
  option that preserves current behavior — no second store exists today)
- Alternate: add a mode field to the factory record and dual-write it inside the verb's lock
- Operator verdict: DEFAULT-APPLIED 2026-10-19T03:20:00Z manager-spec (card t1737, worktree
  .moai/worktrees/t1737)

### Q4: Are the ~29 recorded serial-default cards batch-migrated, or reclassified per card through the new verb?

- Label: DECIDED
- Authority anchor: board `d-20261019T015400Z-02f4#6c1e0d60572f` and board
  `d-20261019T020637Z-0c51#f54e2862f2b8` (operator rulings, 2026-10-19), pinned verbatim:
- Why unresolved: the default flip alone leaves the recorded `serial` rows in place; whether a
  bulk rewrite accompanies the flip was open until the operator's same-day records were checked.
- Operator verdict:

```
{"id":"d-20261019T015400Z-02f4","scope":"card:t1737","kind":"ruling","decided_by":"operator (AskUserQuestion 10-19: t1737 직렬 예외); leader (session d1aa9eb6) for shape","evidence_refs":"todo classification counts 10-19 (29 serial/default); board a672; internal/cli/factory_card.go:335,943 serial gating on main","ladder_path":"operator serial-slot exception (like ad53 for t1885)","body":"Operator granted t1737 (SPEC-FACTORY-DEFAULT-MODE-001, scope extended with an official reclassify verb) a one-time serial-slot exception: it takes the serial slot ahead of the waiting serial cards (t1888, t1889, t1877, t1880, t1881, t1883, t1673). Mechanics: the slot is held by any in-flight serial lease (factory_card.go:335/943), so t1882 (lane-44, plan-audit) pauses — renewals stop, its lease lapses, tree preserved — and t1737 is leased by lane-18. t1882 resumes when t1737 reaches merging. Once t1737 lands, the leader reclassifies parallel-safe serial-default cards (first t1888, t1889) with the new verb.","created":"2026-10-19T01:54:00Z"}

{"id":"d-20261019T020637Z-0c51","scope":"standing","kind":"standing-rule","decided_by":"operator (chat 10-19: '카드 순차/병렬 잘 판단해서 설계'); leader (session d1aa9eb6) for procedure","evidence_refs":"board a672, 02f4; todo classification counts 10-19 (29 serial/default)","ladder_path":"operator standing instruction","body":"Card issuance must carry an explicit serial/parallel judgment. The leader adds every card with --classification-file (decider=human) whose mode is decided from the expected file set: parallelizable unless it edits the same files/shared state as an in-flight card or a slot-gated surface (factory/todo core, release pipeline, shared migrations), in which case serial with the conflicting card named in reason. The reason field states the file set and the conflict check. No card is issued on the fail-safe serial default; a card that slipped through is reclassified (t1737 verb once landed; until then reissue via classification-file, as t1888→t1892, t1889→t1893). Class C cards get --spec at pick time (cannot be added after assign).","predicate":"every card the leader issues or picks","created":"2026-10-19T02:06:37Z"}
```

### Q5: Which card states does the classify verb admit — queued and picked only, or a wider set?

- Label: FOUNDER
- Class: product-level
- Why unresolved: the operator's card text scopes the verb to "queued·picked 카드 대상", but the
  card body lives in the queue store, not in a committed register artifact, so the settlement
  cannot be pinned per the authority register. Fail-closed: the row goes to the operator at the
  kickoff gate rather than being authored on an unverifiable citation. The spec'd admission set
  (REQ-FDM-004) follows the card text; this row is the operator's chance to widen or confirm it.
- Default: queued and picked only (rule: the option with the smaller user-visible surface)
- Alternate: all non-terminal queue states (hold, dropped, …)
- Operator verdict:

### Q6: When an in-flight (picked) card is reclassified to `serial` while another serial card is in flight, does the verb refuse the write or record an operator-acknowledged overlap? (round-2 D2)

- Label: FOUNDER
- Class: implementation-level
- Why unresolved: the round-1 audit measured that the verb as specified admits the unserialized
  overlap — the codex required backend reproduced two factory-recorded serial cards in
  run/leased coexisting after a conforming locked write (receipt
  `rcpt-c9f6d2f8c1f45eb7e134711c`, tree `f6742080a6`). No committed register artifact settles
  the verb's conflict behavior: the operator rulings cited in Q4 scope the verb's existence and
  the issuance discipline, not this edge, so the row fails closed to the register; the default
  below is applied per the published Default rule so the plan can close.
- Default: refuse with a recovery-naming message and write nothing (rule: the option that
  preserves current behavior — the serial-slot exclusivity already refuses a second concurrent
  serial lease at the selection gate (`factory_card.go:943`) and the nominated lease
  (`:1406`), so refusal extends the existing invariant instead of adding a bypass flag; it is
  also the option with the smaller user-visible surface — no new flag)
- Alternate: record the reclassification and let both serial cards run, gated by an explicit
  operator-acknowledged flag (`--allow-overlap`)
- Operator verdict: DEFAULT-APPLIED 2026-10-19T03:28:48Z manager-spec (card t1737, worktree
  .moai/worktrees/t1737)
