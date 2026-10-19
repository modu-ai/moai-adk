---
id: SPEC-FACTORY-DEFAULT-MODE-001
title: "Implementation plan — default-mode split and the todo classify verb"
version: "0.1.0"
created: 2026-10-19
updated: 2026-10-19
author: manager-spec (card t1737)
---

# Plan: SPEC-FACTORY-DEFAULT-MODE-001

## §A Context

### A.1 Implementation surface

TDD mode (quality.yaml `constitution.development_mode: tdd`); RED-GREEN-REFACTOR per milestone.
Estimated 5–7 files including tests (the t1714 split proposal's C3 estimate, still accurate):

| File | Change |
|------|--------|
| `internal/factory/classification.go` | Split the default: the absent-read derivation becomes `parallelizable`; the fail-safe keeps `serial`. Two named constructors (or one constructor + one absent-read constant) replace the single shared `DefaultCardClassification` call sites. |
| `internal/cli/todo_classify.go` | Fail-safe branch promotes the serial fail-safe (text of `:41` unchanged); the absent/default decider path feeds from the new absent-read default. |
| `internal/cli/todo_classify_cmd.go` (new) | `newTodoClassifyCmd` — the verb body: one `store.Mutate`, state gate, human judgment record. |
| `internal/cli/todo.go` | Register `newTodoClassifyCmd()` in the `:342` AddCommand list. |
| `internal/cli/factory_card.go` | No behavior change expected; regression guards only. |
| `internal/cli/*_test.go` (new/extended) | AC tests per §F. |

### A.2 C1-safety rationale (predecessor t1735 skipped by operator exception)

The t1714 split proposal made C3 depend on C1 with two reasons, both re-examined at HEAD
`74ad197380`:

1. *"Before C1, the global check counts only serial holders (`factory_card.go:335`), so
   unlabelled cards would stop holding the serial slot and nothing else would hold them."*
   — True, and it is the intended behavior change, adopted by the operator with eyes open: the
   measured cost of the conservative default (board a672: 29 serial-default cards, 6 waiting on
   one serial slot, lane renewal loop) is itself a live defect. The exposure it opens — a
   true-serial card slipping through unlabelled and running beside others — is bounded three
   ways **without C1**: (a) the fail-safe stays `serial` (REQ-FDM-003), so a decider failure
   still lands conservative; (b) issuance is now explicit by standing rule (board
   `d-20261019T020637Z-0c51`: every card issued with `--classification-file`, decider=human,
   serial whenever it edits the same files/shared state as an in-flight card or a slot-gated
   surface), so the unlabelled population is shrinking by policy; (c) the verb (REQ-FDM-004)
   gives the leader a same-day correction path for anything that slipped through, replacing
   direct database edits. The parallelizable dispatch path itself is not new machinery —
   explicitly classified parallelizable cards already lease beside serial cards today (pinned by
   `TestFactoryNextParallelizableLeasesBesideLiveSerial`); the flip widens the population using
   an existing, tested path.
2. *"The fail-safe's unknown-domain effect (REQ-SLT-003) also needs C1's slot decision."* —
   REQ-SLT-003 belongs to C1's domain-slot design; this SPEC touches no slot or domain
   machinery (C4: no schema change), so the dependency does not bind.

Residual risk accepted: two concurrently-running unlabelled cards may overlap on shared files
and meet at the integration merge. That collision is detectable (the integration-window overlap
refusal exists today) and recoverable (reclassify + re-dispatch), versus the standing wedge the
flip removes. C1 may still land later; nothing here precludes it — C1's domain slots would give
parallelizable cards file-domain separation and further shrink this residual.

### A.3 Key design decisions (what and why, no function signatures)

- **D1 — The default applies at the read-derivation seam, plus the default decider's answer.**
  Absence of the field means `parallelizable` wherever absence is derived today (the
  `EffectiveCardClassification` absent branch and the queue-absent path of
  `factoryQueueClassification`), and a fresh admission without a supplied judgment records
  `parallelizable` through `DefaultCardDecider`. No add-time rewrite of existing records: the
  store holds legacy absent rows and recorded `serial` rows that a read-time rule covers
  uniformly, and recorded judgments stay verbatim (C3). The seam is already the sole derivation
  point (its MX:ANCHOR), so the flip is one-seam, not per-consumer.
- **D2 — The two defaults become two values.** The parent ruling unified them ("the two
  defaults are the same defaults by ruling, not by coincidence", `classification.go:94-97`);
  the operator's 2026-10-19 ruling separates them: absent-read → `parallelizable`, fail-safe →
  `serial`. Implementation keeps one named source per value so the pairing cannot silently
  re-merge; the fallback notice keeps naming `mode=serial` (C7).
- **D3 — "Factory-side sync in the same locked write" is a single-store property.** The card
  text assumed a factory-side mode field to keep in step; measured: `homestate.Card` carries no
  mode field and every factory-side reader derives from the queue at read time
  (`factory_card.go:839-842`, `:1397`, `:3155-3160`). The requirement is therefore stated as an
  observable (REQ-FDM-005: readers observe the new mode from their next pass, no second write)
  and is satisfied by the verb performing its update inside ONE `Mutate` acquisition. No second
  store exists to reconcile; introducing one would be new machinery this SPEC does not need.
- **D4 — The verb follows the hold-verb contract.** `store.Mutate` callback, refusal by early
  return (nothing written, file byte-identical), `human` decider fixed (C6), reason required,
  closed-set mode, target states queued|picked with recovery advice in the refusal message. An
  in-flight (picked) target requested `serial` while another distinct in-flight card holds the
  serial slot is refused the same way — the slot-conflict refusal of REQ-FDM-004 / AC-FDM-011
  (decision-index Q6, DEFAULT-APPLIED), reusing the selection gate's own predicate
  (`factorySerialInFlightExcluding`, `factory_card.go:324-340`) so the verb adds no second
  serial-overlap policy.

## §B Known Issues

- The ~29 recorded serial-default cards stay `serial` after this SPEC until individually
  reclassified (C3; board 02f4 names the leader's per-card plan, first t1888/t1889). This is
  decided, not an open issue; it is listed so no reviewer reads the residual queue waiting as a
  defect of this SPEC.
- The serial-slot read/claim non-atomicity noted in the parent SPEC's 0.4.0 amendment is out of
  scope here; the default flip does not touch the slot read (REQ-FDM-007 guards the semantics).

## §C Pre-flight

- P1 — Confirm at run entry that no `classify` subcommand appeared on main since HEAD
  `74ad197380` (`grep -rn '"classify"' internal/cli/*.go` returns zero hits — no classify
  subcommand is wired; measured at `f6742080a6`: no output, exit 1. Corroborating:
  `newTodoClassifyCmd` is absent from `internal/cli/todo.go` (exit 1), and the installed
  binary answers `Unknown command "classify" for "moai todo"` with exit 1. The run-entry check
  re-measures in case main moved).
- P2 — Confirm the parent SPEC's completed tests still pass before the flip
  (`go test ./internal/factory -run Classification` and `go test ./internal/cli -run
  TodoClassify`) — these name the baseline the split intentionally changes; the run updates the
  affected assertions rather than deleting them.
- P3 — Fallback-path closure check (card scope item 3): after M2, the leader's correction path
  for a fail-safe admission is the verb itself — exercised by AC-FDM-006's picked-card scenario
  (a fail-safe-admitted card that was leased is exactly a picked card).
- P4 — No `[NEEDS CLARIFICATION]` markers remain in this plan (decision-index rows Q1–Q5 carry
  the only open judgments; Q5 rides the kickoff gate per the decision gate being on).

## §D Constraints

Inherited from spec.md §C (C1–C7). Additional implementation constraints:

- IC1 — Tests run with the repo's standard isolation (t.TempDir; no OTEL env in parallel tests;
  no TMPDIR under $HOME).
- IC2 — The verb's output is one line on success (`classified <id> mode=<m>` shape), mirroring
  the hold verb's one-line confirmation, so supervising readers can count mutations.
- IC3 — No exported symbol beyond what the split requires; the two default constructors are the
  only new factory-package API surface.

## §E Self-Verification

- E1 — AC PASS/FAIL matrix from `go test ./internal/cli ./internal/factory -run
  'Classification|TodoClassify|FactorySelection|SerialExclusivity'` output, verbatim.
- E2 — Full package suites for touched packages: `go test ./internal/cli ./internal/factory`.
- E3 — Coverage on touched packages meets the 85% gate (`go test -cover ./internal/factory
  ./internal/cli`).
- E4 — Subagent-boundary: no grep hit for AskUserQuestion in any new file.
- E5 — `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` passes.
- E6 — Byte-identity of refusals observed by test (AC-FDM-007), not by inspection.
- E7 — Fallback notice count regression (AC-FDM-004) observed as exactly one stderr line.

## §F Milestones

Priority-ordered, decision-reversibility-ordered: the semantic default split (highest
change-likelihood, widest reader surface) first; mechanical registration and docs last.

- **M1 (RED→GREEN) — the default split (REQ-FDM-001/002/003; AC-FDM-001…004).**
  RED: extend/adjust the classification tests to assert the absent-read `parallelizable`
  derivation, the default decider's `parallelizable` answer, and the fail-safe `serial`
  promotion with exactly one notice — all red at HEAD `74ad197380` for the stated reason (the
  derivation currently returns `serial`). GREEN: implement D1/D2 in `classification.go` +
  `todo_classify.go`; update the affected parent-SPEC assertions to the new rule (recorded
  judgments unchanged assertions stay).
- **M2 (RED→GREEN) — the classify verb (REQ-FDM-004/006; AC-FDM-005…007, -011).**
  RED: `TestTodoClassifyQueuedCard`, `TestTodoClassifyPickedCard`, `TestTodoClassifyRefusals`,
  `TestTodoClassifySerialSlotConflictRefusal` (verb absent at HEAD — red because the command
  does not exist). GREEN: `todo_classify_cmd.go` + registry line at `todo.go:342`; hold-verb
  refusal contract; human decider fixed; reason required; closed-set mode; the slot-conflict
  refusal (REQ-FDM-004's in-flight serial clause, decision-index Q6).
- **M3 (guard-addition — green-now cells kept green; no RED phase) — factory-side observation
  and the exclusivity guard (REQ-FDM-005/007; AC-FDM-008…010).** Adds the tests without a RED
  stage: the selection-honors-reclassified-mode test (AC-FDM-008, release-blocking; its
  subject event cannot occur until M2's verb exists, and once it can the behavior is expected
  green-by-construction since readers already derive through the seam), plus the serial
  exclusivity (AC-FDM-009) and recorded-judgment-stability (AC-FDM-010) guards that are
  green-now and stay green through the flip. Any fix needed here is a bug fix, not new
  machinery.
- **M4 — docs and close-out.** CHANGELOG entry; MX tag review on the touched seam (the
  `EffectiveCardClassification` ANCHOR's reason text mentions the serial default — refresh the
  comment text, not the tag); `moai spec lint --strict` clean; progress.md §E.2/§E.3 evidence.

## §G Anti-Patterns

- Do NOT rewrite recorded classifications in a migration (C3) — the verb is the only writer of
  human judgments.
- Do NOT derive the default inline at a consumer (`mode == factory.ClassModeSerial` string tests
  against the pointer) — every reader goes through the seam (its MX:ANCHOR exists for exactly
  this).
- Do NOT let the fail-safe branch read the absent-read default — the two values must not
  re-merge (D2).
- Do NOT widen the verb's admission set beyond queued|picked without an operator ruling
  (decision-index Q5).

## §H Cross-References

- Parent lineage: SPEC-TODO-CLASSIFY-DISPATCH-001 § HISTORY v0.2.0 (OD-3 REVISED — fail-safe
  serial) and v0.3.0 (absent-read default flipped to serial — reversed here by operator ruling).
- Operator records: board `d-20261019T015400Z-02f4` (t1737 ruling), board
  `d-20261019T014804Z-a672` (measured wedge), board `d-20261019T020637Z-0c51` (issuance rule).
- Split proposal: `.moai/worktrees/t1714/.moai/reports/t1714/split-proposal.md` §1 (read-only).
- Sibling: SPEC-FACTORY-CONFLICT-SLOTS-001 (parent index; proposed split not executed on this
  lineage — disposition-only reference, see spec.md §E), SPEC-FACTORY-CONFLICT-DOMAIN-001
  (C1, skipped by exception), SPEC-TCD-LLM-DECIDER-001 (LLM decider, untouched).
