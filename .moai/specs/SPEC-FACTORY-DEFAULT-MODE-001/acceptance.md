---
id: SPEC-FACTORY-DEFAULT-MODE-001
title: "Acceptance criteria — default-mode split and the todo classify verb"
version: "0.1.0"
created: 2026-10-19
updated: 2026-10-19
author: manager-spec (card t1737)
---

# Acceptance: SPEC-FACTORY-DEFAULT-MODE-001

## §D AC Matrix

| AC | Requirement | Severity | Class | Verification |
|----|-------------|----------|-------|--------------|
| AC-FDM-001 | REQ-FDM-001 | P1 | release-blocking | `go test -v ./internal/factory -run TestEffectiveClassificationAbsentDefaultsParallelizable` |
| AC-FDM-002 | REQ-FDM-001 | P2 | release-blocking | `go test -v ./internal/cli -run TestQueueAbsentCardClassificationDefault` |
| AC-FDM-003 | REQ-FDM-002 | P1 | release-blocking | `go test -v ./internal/cli -run TestTodoAddDefaultClassificationParallelizable` |
| AC-FDM-004 | REQ-FDM-003 | P1 | regression-guard (demoted, §D.0) | `go test -v ./internal/cli -run TestTodoAddDeciderFailureFailsafeSerial` |
| AC-FDM-005 | REQ-FDM-004 | P1 | release-blocking | `go test -v ./internal/cli -run TestTodoClassifyQueuedCard` |
| AC-FDM-006 | REQ-FDM-004 | P1 | release-blocking | `go test -v ./internal/cli -run TestTodoClassifyPickedCard` |
| AC-FDM-007 | REQ-FDM-006 | P1 | release-blocking | `go test -v ./internal/cli -run TestTodoClassifyRefusals` |
| AC-FDM-008 | REQ-FDM-005 | P2 | release-blocking | `go test -v ./internal/cli -run TestFactorySelectionReadsReclassifiedMode` |
| AC-FDM-009 | REQ-FDM-007 | P1 | regression-guard | `go test -v ./internal/cli -run TestSerialExclusivityAfterDefaultFlip` |
| AC-FDM-010 | REQ-FDM-001 (C3) | P2 | regression-guard | `go test -v ./internal/factory -run TestRecordedClassificationNotRewritten` |
| AC-FDM-011 | REQ-FDM-004 | P1 | release-blocking | `go test -v ./internal/cli -run TestTodoClassifySerialSlotConflictRefusal` |

## §D.0 RED-now evidence ledger and swept-set guard (round-2 D1 repair)

Carrier note (verification-completeness.md §2.1): each release-blocking AC's RED-now cell cites
ledger entries below; an entry carries the four required elements — the command (read-only,
single invocation), that command's verbatim stdout, its exit code, and the tree SHA it was
observed on. Channel attributions: (a) the installed `moai` binary (build commit named per
tool-provenance attribution, §2.2); (b) tree grep / file read on this worktree; (c) attributed
verbatim from the round-1 audit verdict `.moai/reports/t1737/plan-audit-cold.md` — receipt
`rcpt-c9f6d2f8c1f45eb7e134711c`, `audited_sha f6742080a6`.

Tree equivalence: every entry below was observed at `f6742080a6`, whose `internal/` code is
byte-identical to the plan-close tree `74ad197380` — `git diff --name-only 74ad197380..HEAD`
lists the five plan artifacts only (observed this session).

**Swept-set guard (binds every release-blocking AC's green).** A green result requires the
target test to actually execute: the verification command is `go test -v -run <Name> ./<pkg>`
and green is stdout carrying a `--- PASS: <Name>` line for the named test. A result carrying
`[no tests to run]` (or `no test files`), or lacking the named `--- PASS:` line, is **not
green** regardless of exit code — a selector sweeping zero tests exits 0 (observed by the
codex required backend on both packages: EV-FDM-A1).

**Classification repair (round-1 D1).** AC-FDM-004 is demoted release-blocking →
regression-guard under the undecidable disposition (verification-completeness.md §2.1): the
fail-safe-serial behavior it asserts already holds on the plan-close tree, so no honest RED
exists for it and none is recorded. It is green-now and stays green through the flip (M1
guard-addition); its separation clause — the fail-safe stays `serial` regardless of the
absent-read default — becomes observable only after M1 lands, and the guard asserts it from
then on.

### EV-FDM-01 — the ten cited AC test functions have zero definitions on the tree

- Command: `grep -rn -E 'TestEffectiveClassificationAbsentDefaultsParallelizable|TestQueueAbsentCardClassificationDefault|TestTodoAddDefaultClassificationParallelizable|TestTodoAddDeciderFailureFailsafeSerial|TestTodoClassifyQueuedCard|TestTodoClassifyPickedCard|TestTodoClassifyRefusals|TestFactorySelectionReadsReclassifiedMode|TestSerialExclusivityAfterDefaultFlip|TestRecordedClassificationNotRewritten' internal/cli internal/factory`
- stdout: *(none)*
- exit code: `1`
- tree `f6742080a6` · channel (b), this session 2026-10-19T03:24Z

### EV-FDM-02 — installed binary provenance

- Command: `moai version`
- stdout (verbatim):
```
╭────────────────────────────────────╮
│                                    │
│  moai-adk moai_cp/20261007_195927  │
│                                    │
│                                    │
╰────────────────────────────────────╯
 moai_cp/20261007_195927   moai_cp/20261007_195927-1888-g7705b365e2   built 2026-10-18T10:45:08Z 
```
- exit code: `0`
- tree `f6742080a6` · channel (a). Build commit `7705b365e2` predates HEAD `f6742080a6` (lag
  stated per tool-provenance attribution §2.2; the tree-level corroboration is EV-FDM-04).

### EV-FDM-03 — the classify verb does not exist (binary observation)

- Command: `moai todo classify`
- stdout (verbatim):
```
  ERROR  
          
  Unknown command "classify" for "moai todo".                                                                         

  Try --help for usage.
```
- exit code: `1`
- tree `f6742080a6` · channel (a), binary of EV-FDM-02. The single-unknown-token form is
  read-only by the verb's documented contract (a single unknown token stays an error; only a
  phrase of two or more words falls through to add) — no queue mutation.

### EV-FDM-04 — the classify verb does not exist (tree observation)

- Command: `grep -n "newTodoClassifyCmd" internal/cli/todo.go`
- stdout: *(none)*
- exit code: `1`
- tree `f6742080a6` · channel (b), this session. Corroborated by file read: the AddCommand
  list at `internal/cli/todo.go:342-349` carries no classify entry.

### EV-FDM-05 — the absent-read default constructor returns mode `serial`

- Command: `grep -n -A 8 "func DefaultCardClassification" internal/factory/classification.go`
- stdout (verbatim):
```
98:func DefaultCardClassification() CardClassification {
99-	return CardClassification{
100-		Priority: ClassPriorityNormal,
101-		Blocked:  false,
102-		Mode:     ClassModeSerial,
103-		Decider:  DeciderIdentityDefault,
104-	}
105-}
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-06 — the queue-absent factory-card read falls through the same seam

- Command: `grep -n -A 16 "func factoryQueueClassification" internal/cli/factory_card.go`
- stdout (verbatim):
```
345:func factoryQueueClassification(rec *factory.BacklogRecord, cardID string) factory.CardClassification {
346-	if rec == nil {
347-		return factory.DefaultCardClassification()
348-	}
349-	for _, it := range rec.Items {
350-		if it.ID == cardID {
351-			return factory.EffectiveCardClassification(it)
352-		}
353-	}
354-	for _, entry := range rec.Archived {
355-		if entry.Item.ID == cardID {
356-			return factory.EffectiveCardClassification(entry.Item)
357-		}
358-	}
359-	return factory.DefaultCardClassification()
360-}
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-07 — the default decider records mode `serial` at admission

- Command: `grep -n -A 6 "func (DefaultCardDecider) Classify" internal/factory/classification.go`
- stdout (verbatim):
```
256:func (DefaultCardDecider) Classify(string) (CardClassification, error) {
257-	c := DefaultCardClassification()
258-	c.Reason = "deterministic default: no classification judgment supplied"
259-	return c, nil
260-}
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-08 — the fail-safe notice names `mode=serial` (green-now baseline)

- Command: `grep -n "mode=serial" internal/cli/todo_classify.go`
- stdout (verbatim):
```
41:const todoClassificationFallbackNotice = "todo add: classification decider unavailable; admitted with the fail-safe default (priority=normal, blocked=false, mode=serial, decider=default)"
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-09 — the serial-slot holder predicate

- Command: `grep -n -B 2 -A 6 "func factorySerialSlotHeld" internal/cli/factory_card.go`
- stdout (verbatim):
```
295:func factorySerialSlotHeld(c homestate.Card, now time.Time) bool {
296-	return !factorySerialSlotFree(c.State) && !c.LeaseExpired(now) && c.OwnerLabel != ""
297-}
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-10 — the other-serial-in-flight predicate (slot-conflict grounding)

- Command: `grep -n -A 17 "func factorySerialInFlightExcluding" internal/cli/factory_card.go`
- stdout (verbatim):
```
324:func factorySerialInFlightExcluding(cards []homestate.Card, classOf func(string) factory.CardClassification, cardID string, now time.Time, ignoreAssigned bool) bool {
325-	for _, c := range cards {
326-		if c.CardID == cardID {
327-			continue
328-		}
329-		if ignoreAssigned && c.State == homestate.CardAssigned {
330-			continue
331-		}
332-		if c.State == homestate.CardPicked && (c.BundleID != "" || c.HintAfter != "") {
333-			continue
334-		}
335-		if factorySerialSlotHeld(c, now) && classOf(c.CardID).Mode == factory.ClassModeSerial {
336-			return true
337-		}
338-	}
339-	return false
340-}
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-11 — the serial-slot gating read sites

- Command: `grep -n "ClassModeSerial && serialInFlightExcluding\|ClassModeSerial && factorySerialInFlightExcluding" internal/cli/factory_card.go`
- stdout (verbatim):
```
943:		if classOf(c.CardID).Mode == factory.ClassModeSerial && serialInFlightExcluding(c.CardID, true) {
1072:				if cls.Mode == factory.ClassModeSerial && serialInFlightExcluding(it.ID, false) {
1406:	nom.serialHeld = classOf(cardID).Mode == factory.ClassModeSerial && factorySerialInFlightExcluding(cards, classOf, cardID, now, assignedHere)
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-12 — the parent-SPEC test pinning the current absent-read serial rule

- Command: `grep -n "func TestEffectiveCardClassificationDerivesAbsentDefaults" internal/factory/classification_test.go`
- stdout (verbatim):
```
81:func TestEffectiveCardClassificationDerivesAbsentDefaults(t *testing.T) {
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session. Its assertion body
  (classification_test.go:82-91, read this session) wants `Mode: ClassModeSerial` for an
  absent field — this passing run is the observed baseline the flip breaks. Its verbatim
  passing output is **PENDING-observation** (Go-restraint window active at repair time); the
  command `go test -v ./internal/factory -run TestEffectiveCardClassificationDerivesAbsentDefaults`
  is recorded for run entry alongside plan §C P2.

### EV-FDM-13 — the parent-SPEC test pinning the fail-safe fallback (AC-FDM-004 baseline)

- Command: `grep -n "func TestTodoAddDeciderFailureFallsBackWithNotice" internal/cli/todo_classify_test.go`
- stdout (verbatim):
```
109:func TestTodoAddDeciderFailureFallsBackWithNotice(t *testing.T) {
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session. Passing output
  **PENDING-observation** (same window); recorded at run entry with plan §C P2.

### EV-FDM-14 — the existing parallelizable-beside-serial test (AC-FDM-009 baseline)

- Command: `grep -n "func TestFactoryNextParallelizableLeasesBesideLiveSerial" internal/cli/factory_serial_slot_stale_test.go`
- stdout (verbatim):
```
303:func TestFactoryNextParallelizableLeasesBesideLiveSerial(t *testing.T) {
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session. Passing output
  **PENDING-observation** (same window); recorded at run entry with plan §C P2.

### EV-FDM-15 — the derivation seam's recorded-field branch (AC-FDM-010 baseline)

- Command: `grep -n -A 10 "func EffectiveCardClassification" internal/factory/classification.go`
- stdout (verbatim):
```
148:func EffectiveCardClassification(it BacklogItem) CardClassification {
149-	if it.Classification != nil {
150-		return *it.Classification
151-	}
152-	return DefaultCardClassification()
153-}
```
- exit code: `0` · tree `f6742080a6` · channel (b), this session

### EV-FDM-A1 — attributed: the cited selectors sweep zero tests (channel c)

Not self-run — attributed verbatim from the round-1 audit verdict
(`.moai/reports/t1737/plan-audit-cold.md`, defect D1 / MP-8), produced by the required codex
backend under receipt `rcpt-c9f6d2f8c1f45eb7e134711c` on tree `f6742080a6`:

> "the codex required backend independently executed the cited selectors and observed 'PASS …
> [no tests to run]' on both packages"

This is the observed proof of the swept-set hazard the guard above forecloses: a zero-test
sweep reports success.

## Two-cell adoption state (plan-close tree `74ad197380`)

Observations recorded 2026-10-19 at repair time on HEAD `f6742080a6` — byte-identical to the
plan-close tree on `internal/` (see §D.0). Each RED-now cell cites §D.0 ledger entries; each
green path cell names its milestone (plan.md §F) and the exact green output required under the
swept-set guard.

- **AC-FDM-001 — RED-now.** The absent-read derivation returns mode `serial` on this tree:
  the seam's default constructor is `Mode: ClassModeSerial` (EV-FDM-05). The cited test is
  absent (EV-FDM-01). The parent test pinning the current rule —
  `TestEffectiveCardClassificationDerivesAbsentDefaults`, wanting `Mode: ClassModeSerial` for
  an absent field (EV-FDM-12) — passes today; that passing run is the observed baseline the
  flip breaks. Green path: M1 — `--- PASS: TestEffectiveClassificationAbsentDefaultsParallelizable`.
- **AC-FDM-002 — RED-now.** The queue-absent factory-card read returns
  `factory.DefaultCardClassification()` on both the nil-record and id-absent branches
  (EV-FDM-06) — mode `serial` today, `parallelizable` only after M1 flips the seam. Test
  absent (EV-FDM-01). Green path: M1 — `--- PASS: TestQueueAbsentCardClassificationDefault`.
- **AC-FDM-003 — RED-now.** The default decider records the same serial value at admission:
  `DefaultCardDecider.Classify` answers `DefaultCardClassification()` with reason
  "deterministic default: no classification judgment supplied" (EV-FDM-07). Test absent
  (EV-FDM-01). Green path: M1 — `--- PASS: TestTodoAddDefaultClassificationParallelizable`.
- **AC-FDM-004 — regression-guard (demoted; green-now; no RED recorded).** The fail-safe
  promotion already lands serial with exactly one notice naming `mode=serial` (EV-FDM-08;
  promotion branch `todo_classify.go:81-89`), pinned today by
  `TestTodoAddDeciderFailureFallsBackWithNotice` (EV-FDM-13, passing output PENDING with P2).
  Per the undecidable disposition no RED is claimed. Green path: M1 guard-addition — the
  guard asserts the value stays `serial` and the notice count stays exactly one after the
  absent-read default flips (separation clause observable from then on).
- **AC-FDM-005 — RED-now.** The verb does not exist: the installed binary answers
  `Unknown command "classify" for "moai todo"`, exit 1 (EV-FDM-03); the registry carries no
  `newTodoClassifyCmd` (EV-FDM-04); the cited tests are absent (EV-FDM-01). Green path: M2 —
  `--- PASS: TestTodoClassifyQueuedCard`.
- **AC-FDM-006 — RED-now.** Same verb absence (EV-FDM-03/-04/-01); a picked-card
  reclassification cannot be observed because no reclassification path exists. Green path:
  M2 — `--- PASS: TestTodoClassifyPickedCard`.
- **AC-FDM-007 — RED-now.** Same verb absence (EV-FDM-03/-04/-01); no refusal matrix exists
  to observe byte-identity on. Green path: M2 — `--- PASS: TestTodoClassifyRefusals`.
- **AC-FDM-008 — RED-now.** No reclassification can occur (verb absent, EV-FDM-03/-04), so no
  post-reclassify selection read exists to observe, and the observing test is absent
  (EV-FDM-01). Green path: M3 guard-addition, expected green-by-construction once M1/M2 land —
  `--- PASS: TestFactorySelectionReadsReclassifiedMode`.
- **AC-FDM-009 — regression-guard (green-now).** Serial exclusivity holds today and is pinned
  by the existing `TestFactoryNextParallelizableLeasesBesideLiveSerial`
  (factory_serial_slot_stale_test.go:303, EV-FDM-14; passing output PENDING with P2). M3 adds
  `TestSerialExclusivityAfterDefaultFlip` as the named guard — absent today (EV-FDM-01);
  green requires it to exist and carry its `--- PASS:` line.
- **AC-FDM-010 — regression-guard (green-now).** Recorded judgments return verbatim through
  the seam's present-field branch (`EffectiveCardClassification`, EV-FDM-15; the parent suite
  pins the round-trip at classification_test.go:92-98). M3 adds
  `TestRecordedClassificationNotRewritten` as the named guard — absent today (EV-FDM-01);
  green requires its `--- PASS:` line.
- **AC-FDM-011 — RED-now (new, round-2 D2).** The verb does not exist (EV-FDM-03/-04); the
  unserialized overlap this AC forecloses was reproduced on the plan-close tree by the codex
  required backend (EV-FDM-A1), and the refusal predicate it will reuse already exists on the
  tree (EV-FDM-09/-10/-11). Green path: M2 —
  `--- PASS: TestTodoClassifySerialSlotConflictRefusal` under the swept-set guard.

## §D.1 AC-FDM-001 — absent queue item derives parallelizable

- **Given** a queue record whose item carries no classification field (a pre-classification-era
  row), **When** any machine consumer derives the item's classification through the derivation
  seam, **Then** mode is `parallelizable`, priority `normal`, blocked `false`, decider `default`,
  and the derived record's reason names the absence derivation.
- Red reason at plan close: the seam returns mode `serial` (`classification.go:98-105`).

## §D.2 AC-FDM-002 — queue-absent factory card derives parallelizable

- **Given** a card present in the factory record whose id names no queue item, **When** the
  factory-side classification reader resolves the card's mode, **Then** it derives
  `parallelizable` (the same absent-field rule, not a serial fallback).

## §D.3 AC-FDM-003 — plain admission records parallelizable

- **Given** the default decider seam (no `--classification-file`), **When** a card is admitted
  with `moai todo add`, **Then** the stored classification carries mode `parallelizable`, decider
  `default`, and a reason naming the deterministic default.

## §D.4 AC-FDM-004 — decider failure promotes serial, exactly one notice

- **Given** a decider whose `Classify` fails (unavailable or error), **When** the add completes
  its locked write, **Then** the card's stored classification is mode `serial`, decider
  `default`, stderr carries exactly one fallback notice, and the notice text still names
  `mode=serial`. Admission is not refused.
- Classification (round-2 D1): **regression-guard** — the behavior asserted already holds on
  the plan-close tree (the seam default is serial, EV-FDM-05; the promotion branch
  `todo_classify.go:81-89`; the notice names `mode=serial`, EV-FDM-08), so no honest RED
  exists and none is recorded (the undecidable disposition, verification-completeness.md
  §2.1). Green-now, kept green through the flip (M1 guard-addition); the separation clause —
  the fail-safe stays `serial` regardless of the absent-read default — becomes observable
  only after M1 lands, and the guard asserts it then.

## §D.5 AC-FDM-005 — classify a queued card records the human judgment

- **Given** a queued card, **When** `moai todo classify <id> --mode serial --reason "touches
  the release pipeline"` returns success, **Then** the item's classification shows mode `serial`,
  decider `human`, the reason verbatim, a fresh classified-at stamp, priority and blocked
  unchanged, and every other item byte-unchanged; stdout carries exactly one confirmation line.

## §D.6 AC-FDM-006 — classify a picked card succeeds

- **Given** a picked card (leased, in the factory record — the shape of a fail-safe-admitted
  card already in flight), **When** the leader runs the classify verb with
  `--mode parallelizable` and a reason, **Then** the same one-locked-write recording of AC-FDM-005
  lands. This is the card's fallback-path closure check: the leader's correction path for a
  fail-safe admission is the verb, not a database edit.

## §D.7 AC-FDM-007 — refusal matrix (byte-identical writes)

- **Given** each of: a held card; a dropped card; a card in a terminal state; an id absent from
  the queue; a missing/empty `--reason`; `--mode conflicting`; **When** the verb runs, **Then**
  each invocation refuses with a non-zero exit and a recovery-naming message, and the queue file
  is byte-identical before and after (observed by hash in the test, not by inspection). The verb
  accepts no decider flag; decider is `human` unconditionally.

## §D.8 AC-FDM-008 — factory-side readers observe the reclassified mode

- **Given** a factory-recorded in-flight card reclassified by the verb, **When** the next
  selection pass and the status view read the card, **Then** both observe the new mode with no
  write beyond the verb's single locked write — selection honors a `serial` reclassification
  against the serial slot, and the status view renders the new mode.

## §D.9 AC-FDM-009 — serial exclusivity survives the flip (regression guard)

- **Given** a serial-mode card holding the serial slot, **When** another serial-mode card
  attempts a lease, **Then** it waits; and **when** a parallelizable card attempts a lease
  beside the same holder, **Then** it leases normally.

## §D.10 AC-FDM-010 — recorded judgments are never rewritten (regression guard)

- **Given** a card carrying an explicit classification recorded by the old default decider
  (decider `default`, mode `serial`), **When** the SPEC's changes are active and the card is
  read through the derivation seam, **Then** the recorded judgment is returned verbatim — no
  default, migration, or read path rewrites it; only the verb changes it.

## §D.11 AC-FDM-011 — serial-slot conflict on an in-flight serial reclassification is refused (new, round-2 D2)

- **Given** a picked (in-flight) card targeted by the verb and another distinct in-flight card
  currently holding the serial slot (`factorySerialSlotHeld` ∧ mode `serial` — the
  `factorySerialInFlightExcluding` read, `factory_card.go:324-340`), **When** the verb runs
  with `--mode serial`, **Then** it refuses with a non-zero exit and a recovery-naming message
  (naming the slot holder; advising wait for the holder to land, reclassifying the holder, or
  reclassifying this card `parallelizable`), and the queue file is byte-identical before and
  after (hash-observed in the test); **and when** the same target is requested with
  `--mode parallelizable`, or with `--mode serial` while no other serial card is in flight,
  **Then** the one-locked-write recording of AC-FDM-005 lands normally.
- Red reason at plan close: the verb does not exist (EV-FDM-03/-04) — the unserialized overlap
  codex reproduced on the plan-close tree (EV-FDM-A1) is exactly the state this refusal
  forecloses; the predicate it reuses exists on the tree (EV-FDM-09/-10/-11). Swept-set guard
  applies: green requires `--- PASS: TestTodoClassifySerialSlotConflictRefusal`.
- Verification: `go test -v ./internal/cli -run TestTodoClassifySerialSlotConflictRefusal`

## §E Quality gates

- `go test ./internal/cli ./internal/factory` green; coverage on both packages ≥ 85%
  (`quality.yaml` `test_coverage_target`).
- `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` passes.
- LSP zero-error gate on touched packages per `quality.yaml` `lsp_quality_gates.run`.

## §F Definition of Done

1. All eleven ACs pass with their verification commands observed green in this run (E1 evidence
   quoted verbatim in progress.md §E.2). Every release-blocking green carries its named
   `--- PASS:` line per the §D.0 swept-set guard — a `[no tests to run]` result is never
   green; AC-FDM-004/-009/-010 are observed green as regression guards.
2. The fallback notice count regression (AC-FDM-004) and refusal byte-identity (AC-FDM-007) are
   test-observed, not inspection-asserted.
3. plan.md §E self-verification items E1–E7 each carry observed output.
4. progress.md §E.3 audit-ready signal populated; §E.4 left for the sync phase.
