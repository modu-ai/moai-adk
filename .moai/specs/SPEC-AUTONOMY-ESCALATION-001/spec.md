---
id: SPEC-AUTONOMY-ESCALATION-001
title: "Contract-mode escalation detector: mechanical detection of the six escalate_on classes plus three operational trips, reported as an escalation record without blocking or mutating the queue"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "autonomy,contract,escalation,detector,hook,needs-decision"
tier: L
related_specs: [SPEC-AUTONOMY-TIERS-001, SPEC-ACSNAPSHOT-COMMIT-GUARD-001]
---

# SPEC-AUTONOMY-ESCALATION-001 — Contract-mode escalation detector

## §A — History

- **2026-09-26** — v0.1.0 plan-phase draft (card t1235, contract-based autonomous harness
  track A2). Base tree `develop` at `ca1d5dc43`. Every "existing asset" premise in the
  card's design table was measured against this tree before any requirement was written;
  the measured table (asset → exists? → file:line) lives in `research.md` §B. Three
  premises did not hold as stated and reshaped scope — see §B.2.

## §B — Problem

### B.1 What the card asks for

A contract-mode run (a SPEC carrying a `contract.yaml`, schema owned by card t1234 / A1)
lets an agent proceed without per-step human questions **as long as it stays inside the
contract**. Staying inside is only a guarantee if leaving is detected mechanically. The
contract names six `escalate_on` classes; the card adds three operational trips:

| # | Class | Kind |
|---|---|---|
| 1 | acceptance-change | contract |
| 2 | invariant-violation | contract |
| 3 | ownership-move | contract |
| 4 | new-architecture-or-api | contract |
| 5 | contradictory-evidence | contract |
| 6 | irreversible-action | contract |
| 7 | budget-exceeded | operational |
| 8 | same-diagnostic-repeat (3 consecutive) | operational |
| 9 | audit-fail-at-retry-cap | operational |

When a class trips, the card becomes *needs-decision* and an escalation report records the
observation, the options, and the contract line that tripped. Escalation is a **report, not
a question**: the agent continues other work.

### B.2 Premises that did not hold as stated (measured, `research.md` §B)

1. **No needs-decision card state exists, and the store forbids adding one.** The queue
   states are exactly `queued` / `picked` / `dropped` (`internal/kanban/backlog_store.go:61-65`);
   `internal/kanban/backlog_schema_freeze_test.go:3-5,70` pins the three-state CHECK and
   states the landing must "admit no fourth state". Queue doctrine additionally forbids a
   machine acting on a card (`kanban-dispatch.md` § Entry into the board). The
   needs-decision marking is therefore specified as a **record beside the queue**, not a
   queue state (REQ-AE-015).
2. **The AC-snapshot guard is not reusable as a detector.** `scripts/ac-baseline/check-staged.sh:1-4`
   declares itself a local-only dev tool with no template mirror; it compares an AC count
   against a corpus snapshot (`.moai/reports/t338/ac-count-baseline.txt`, line 25), not a
   per-SPEC contract hash. Its *counting rule* is reusable in principle; its *mechanism*
   is not. Detection class 1 is scoped against the contract's own recorded hash and count.
3. **`graph_file_api` reads the working tree only.** `internal/graph/codequery.go:65-81`
   opens the file at `projectRoot/relPath` via `os.Lstat`; it has no commit parameter. A
   before/after comparison therefore needs the before-side extracted from the base commit's
   blob by some other route (design.md §C.4).

Four further assets exist only as doctrine or partial coverage (frozen-file guard scoped to
the harness-learner identity only; `audit_multi` disagreement is a tri-state pointer; the
deny list blocks force-push but not a plain push to `main` or `git tag`; super-advisor E1 is
a doctrine row with no counter). Each is recorded with file:line in `research.md` §B.

## §C — Goal

Under `autonomy.mode: contract`, every one of the nine classes is detected by a mechanical
predicate, and each trip yields exactly one attributable escalation record naming the tripped
contract line. Under the distributed default `autonomy.mode: guided`, nothing observable
changes.

## §D — Requirements (GEARS)

20 requirements. Requirements whose predicate reads a `contract.yaml` field carry the tag
「A1 스키마 확정 후 재조정」; the assumed fields are listed once, in §F.

### D.1 — Activation and inertness

- **REQ-AE-001** (Capability gate) — Where `workflow.autonomy.mode` is absent, empty,
  unrecognized, or `guided`, the escalation detector shall perform no detection, write no
  escalation record, and add no subprocess to any hook invocation, so that hook output is
  byte-identical to the pre-change behavior.
- **REQ-AE-002** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`,
  when a hook or checkpoint fires for a SPEC that has no readable `contract.yaml`, the
  escalation detector shall perform no class detection and shall append one `not-armed`
  line naming the SPEC and the reason to the detector audit log. 「A1 스키마 확정 후 재조정」
- **REQ-AE-003** (Ubiquitous) — The escalation detector shall never deny, ask about, or
  alter a tool call; its only effects shall be the escalation record, the needs-decision
  record, and the detector audit log.
- **REQ-AE-004** (Event-detected) — When the escalation detector faults internally (a
  panic, an unreadable input, a timeout), the detector shall let the tool call proceed,
  shall append a `not-checked` line naming the class and the fault, and shall not write a
  needs-decision record for that fault.

### D.2 — The six contract classes

- **REQ-AE-005** (Event-driven) — When a checkpoint observes that the SPEC's
  `acceptance.md` sha256 or its AC count differs from the value recorded in the contract,
  the escalation detector shall trip class `acceptance-change`. 「A1 스키마 확정 후 재조정」
- **REQ-AE-006** (Event-driven) — When a Bash tool call whose command string equals a
  contract invariant command completes with a non-zero exit status, the escalation
  detector shall trip class `invariant-violation` (sub-kind `command`). 「A1 스키마 확정 후 재조정」
- **REQ-AE-007** (Event-driven) — When a write-capable tool call targets a path in the
  frozen instruction set or under a frozen-zone prefix, the escalation detector shall trip
  class `invariant-violation` (sub-kind `frozen-file`) regardless of the calling agent's
  identity.
- **REQ-AE-008** (Event-driven) — When a write-capable tool call targets a path that
  matches no `ownership.write` pattern, or matches any `ownership.never` pattern, the
  escalation detector shall trip class `ownership-move`, with a `never` match taking
  precedence in the report. 「A1 스키마 확정 후 재조정」
- **REQ-AE-009** (Capability gate + event) — Where `workflow.autonomy.escalation.new_api_detector`
  is `graph`, when a checkpoint compares the card base to HEAD and observes a new package
  directory, a new exported declaration, a new CLI verb, a new MCP tool name, or a new
  configuration key, the escalation detector shall trip class `new-architecture-or-api`
  and list every observed addition by kind and path.
- **REQ-AE-010** (Event-driven) — When a recorded verdict pair for the card disagrees —
  an `audit_multi` result whose `disagreement_flag` is present and true, a second review
  verdict opposite to the first, or a recorded CI failure for a head whose local
  verification was recorded as passing — the escalation detector shall trip class
  `contradictory-evidence`.
- **REQ-AE-011** (Event-driven) — When a Bash tool call's command matches the
  irreversible-action pattern set (a push whose target ref is not listed in the contract's
  `actions`, a tag creation or tag push, a release creation, a force push, or a pattern
  from the existing destructive-command denylist), the escalation detector shall trip
  class `irreversible-action` before the command executes. 「A1 스키마 확정 후 재조정」

### D.3 — The three operational trips

- **REQ-AE-012** (Event-driven) — When the card's observed turn count, operation count, or
  audit-retry count exceeds the contract's `budget` value (or, absent one, the
  `workflow.autonomy.escalation.budget_default` value), the escalation detector shall trip
  class `budget-exceeded` naming the exceeded dimension. 「A1 스키마 확정 후 재조정」
- **REQ-AE-013** (Event-driven) — When the same failure fingerprint (normalized failing
  command plus diagnostic key) is observed on three consecutive failing attempts with no
  intervening success of that command, the escalation detector shall trip class
  `same-diagnostic-repeat`.
- **REQ-AE-014** (Event-driven) — When a plan-audit or sync-audit verdict file for the card
  records FAIL at an iteration equal to the tier-resolved retry ceiling, the escalation
  detector shall trip class `audit-fail-at-retry-cap`.

### D.4 — Reporting and marking

- **REQ-AE-015** (Ubiquitous) — The escalation detector shall mark a tripped card as
  needs-decision through a record stored outside the queue store, shall not add a state
  to, rewrite, reorder, or drop any queue item, and shall leave the queue store's schema
  identical.
- **REQ-AE-016** (Event-driven) — When any class trips, the escalation detector shall write
  one escalation report carrying: the class, the observation (command or event plus
  verbatim evidence), at least two options for the decider, the tripped contract line as
  `contract.yaml:<line>` with the matching `escalate_on` token (or the config key, for the
  operational classes), the SPEC ID, the card id when resolvable, and the HEAD commit at
  detection. 「A1 스키마 확정 후 재조정」
- **REQ-AE-017** (State-driven) — While an escalation record with the same class and the
  same observation fingerprint is already open for the card, the escalation detector shall
  not write a second report for that trip and shall increment an occurrence count on the
  existing record.
- **REQ-AE-018** (Ubiquitous) — The escalation detector shall present an escalation as a
  report and shall not route it through the user question channel; the agent that
  triggered it may continue work not blocked by the tripped class.
- **REQ-AE-019** (Ubiquitous) — The escalation report shall label every detection that the
  detector could not complete (an unavailable CI verdict, an unsupported language for
  declaration extraction, an unreadable contract field) as not-observed, and shall never
  render not-observed as agreement or as absence of a trip.
- **REQ-AE-020** (Event-detected) — When a class listed in the detector's nine is absent
  from the contract's `escalate_on`, the escalation detector shall not trip that contract
  class and shall record in the audit log that the class was disabled by the contract;
  operational classes 7-9 shall remain active regardless. 「A1 스키마 확정 후 재조정」

## §E — Constraints

- **C1 — Default preserves today.** The distributed template ships
  `workflow.autonomy.mode: guided`. No behavior, file, or hook output changes under it
  (REQ-AE-001). `MOAI_AUTONOMY_TIER` (`internal/config/envkeys.go:148`) keeps its meaning
  (hook block strength) and is neither read nor renamed by this SPEC.
- **C2 — Name distinct from the harness escalation block.** `harness.escalation`
  (`internal/config/types.go:1199-1243`) already names harness-level escalation
  (minimal→thorough). The new keys live under `workflow.autonomy.escalation` and the two
  are never merged.
- **C3 — Detection only.** No existing gate is rewired, weakened, or strengthened; the
  destructive denylist, deny rules, and branch guard keep deciding exactly as they do now.
- **C4 — Hook budget.** Under `contract` mode, detector work on the PreToolUse path is
  bounded by the existing hook bind budget; work that needs a subprocess or a git history
  read runs only at checkpoints, never per tool call.
- **C5 — Template neutrality.** Any template content added carries no card id, SPEC id,
  internal date, or commit SHA and favors no programming language.
- **C6 — Heuristic honesty.** Class 4 is a heuristic; its misses are expected and are
  covered by sync-audit and second review, not by this SPEC.

## §F — Dependency on A1

This SPEC does not design the contract schema. The fields below are **assumptions** taken
from the draft in `.moai/reports/t1235/design-source.md`; every requirement tagged
「A1 스키마 확정 후 재조정」 is re-read once card t1234's schema lands on `develop`.

| Assumed field | Assumed shape | Consumed by |
|---|---|---|
| file location | `.moai/specs/<SPEC-ID>/contract.yaml` | REQ-AE-002 |
| `acceptance.file` | path string | REQ-AE-005 |
| `acceptance.sha256` | hex digest of the acceptance file | REQ-AE-005 |
| `acceptance.ac_count` | integer | REQ-AE-005 |
| `invariants[]` | list of strings; command entries are literal command strings | REQ-AE-006 |
| `ownership.write[]` | glob list | REQ-AE-008 |
| `ownership.never[]` | glob list | REQ-AE-008 |
| `actions[]` | token list naming permitted external actions / push targets | REQ-AE-011 |
| `budget.{turns,operations,audit_retries}` | integers | REQ-AE-012 |
| `escalate_on[]` | token list from the six contract class names | REQ-AE-016, REQ-AE-020 |
| line addressability | each field sits on a distinct source line | REQ-AE-016 |

Also depends on A1 landing before run-phase begins (card text: "plan 은 병행 가능, run 은 A1
develop 병합 뒤").

## §G — Out of Scope

### Out of Scope — rewiring or adding gates

- Turning any trip into a deny, an ask, or a block (that is track A3).
- Adding a plain-push-to-`main` or `git tag` deny rule; class 6 reports the attempt only.

### Out of Scope — the closure report

- Aggregating escalation records into a closure or completion report (track A4).

### Out of Scope — the contract schema

- Defining, validating, or versioning `contract.yaml` (track A1, card t1234).

### Out of Scope — queue state changes

- Adding a `needs-decision` value to the queue state enum, or any queue column.
- Automatically resolving, dismissing, or acting on an escalation.

### Out of Scope — reviving the ac-baseline guard as a product feature

- Mirroring `scripts/ac-baseline/` into the template or wiring it into this detector.

### Out of Scope — a CI poller

- Fetching CI results from a remote. Class 5 consumes a recorded CI verdict when one
  exists and reports not-observed otherwise.
