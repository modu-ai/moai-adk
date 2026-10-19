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
| AC-FDM-001 | REQ-FDM-001 | P1 | release-blocking | `go test ./internal/factory -run TestEffectiveClassificationAbsentDefaultsParallelizable` |
| AC-FDM-002 | REQ-FDM-001 | P2 | release-blocking | `go test ./internal/cli -run TestQueueAbsentCardClassificationDefault` |
| AC-FDM-003 | REQ-FDM-002 | P1 | release-blocking | `go test ./internal/cli -run TestTodoAddDefaultClassificationParallelizable` |
| AC-FDM-004 | REQ-FDM-003 | P1 | release-blocking | `go test ./internal/cli -run TestTodoAddDeciderFailureFailsafeSerial` |
| AC-FDM-005 | REQ-FDM-004 | P1 | release-blocking | `go test ./internal/cli -run TestTodoClassifyQueuedCard` |
| AC-FDM-006 | REQ-FDM-004 | P1 | release-blocking | `go test ./internal/cli -run TestTodoClassifyPickedCard` |
| AC-FDM-007 | REQ-FDM-006 | P1 | release-blocking | `go test ./internal/cli -run TestTodoClassifyRefusals` |
| AC-FDM-008 | REQ-FDM-005 | P2 | release-blocking | `go test ./internal/cli -run TestFactorySelectionReadsReclassifiedMode` |
| AC-FDM-009 | REQ-FDM-007 | P1 | regression-guard | `go test ./internal/cli -run TestSerialExclusivityAfterDefaultFlip` |
| AC-FDM-010 | REQ-FDM-001 (C3) | P2 | regression-guard | `go test ./internal/factory -run TestRecordedClassificationNotRewritten` |

## Two-cell adoption state (plan-close tree `74ad197380`)

- AC-FDM-001…-008 are **RED-now** on the plan-close tree for the stated reason each: the
  absent-read derivation returns `serial` today (AC-001/-002 fail on the mode assertion), the
  default decider answers `serial` (AC-003), the verb does not exist (AC-005/-006/-007: command
  not found is the red), and no test observes post-reclassify selection reads (AC-008: test
  absent). The green path cell for each is its milestone in plan.md §F (M1: -001…-004; M2:
  -005…-007; M3: -008).
- AC-FDM-009 and AC-FDM-010 are **regression-guards**: green at plan close (serial exclusivity
  and recorded-judgment stability already hold — `TestFactoryNextParallelizableLeasesBesideLiveSerial`
  and the recorded-judgment read path), kept green through the flip. They are not release-blocking
  because they assert today's invariant survives, not that new behavior lands.

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

## §E Quality gates

- `go test ./internal/cli ./internal/factory` green; coverage on both packages ≥ 85%
  (`quality.yaml` `test_coverage_target`).
- `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` passes.
- LSP zero-error gate on touched packages per `quality.yaml` `lsp_quality_gates.run`.

## §F Definition of Done

1. All ten ACs pass with their verification commands observed green in this run (E1 evidence
   quoted verbatim in progress.md §E.2; AC-FDM-009/-010 observed green as regression guards).
2. The fallback notice count regression (AC-FDM-004) and refusal byte-identity (AC-FDM-007) are
   test-observed, not inspection-asserted.
3. plan.md §E self-verification items E1–E7 each carry observed output.
4. progress.md §E.3 audit-ready signal populated; §E.4 left for the sync phase.
