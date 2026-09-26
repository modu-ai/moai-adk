---
id: SPEC-AUTONOMY-ESCALATION-001
title: "Contract-mode escalation detector: mechanical detection of the six escalate_on classes plus three operational trips, reported as an escalation record without blocking or mutating the queue"
version: "0.1.1"
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
- **2026-09-26** — v0.1.1 contract-field alignment (card t1235). Every contract field this SPEC
  references is aligned to the A1 schema draft, `design.md` § Contract Schema of
  SPEC-AUTONOMY-CONTRACT-001 at commit `8f77d9a33` (not yet plan-audited; referenced, not
  copied). Three reshapes followed from the draft: `escalate_on` must hold exactly the six tokens,
  so per-class disabling is gone and REQ-AE-020 now governs a contract that fails A1 verify;
  class 1 consumes A1's own verify reasons instead of re-measuring; `actions` is a closed token
  vocabulary, so class 6 reads push authority from `push-develop` alone. The dependency tag on
  contract-reading requirements became 「A1 plan-audit 통과본으로 재확인」; the full delta is §F.

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
   is not. Detection class 1 consumes A1's own verify of the contract's recorded hash and
   count (§F O7).
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

20 requirements. Requirements whose predicate reads a `contract.yaml` field or an A1 verify
result carry the tag 「A1 plan-audit 통과본으로 재확인」; every such field is listed once, in §F,
against the A1 draft at `8f77d9a33`.

### D.1 — Activation and inertness

- **REQ-AE-001** (Capability gate) — Where `workflow.autonomy.mode` is absent, empty,
  unrecognized, or `guided`, the escalation detector shall perform no detection, write no
  escalation record, and add no subprocess to any hook invocation, so that hook output is
  byte-identical to the pre-change behavior.
- **REQ-AE-002** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`,
  when a hook or checkpoint fires for a SPEC whose `contract.yaml` is absent or whose A1 verify
  state is `unsigned`, the escalation detector shall perform no contract-class detection and
  shall append one `not-armed` line naming the SPEC and the reason to the detector audit log.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-003** (Ubiquitous) — The escalation detector shall never deny, ask about, or
  alter a tool call; its only effects shall be the escalation record, the needs-decision
  record, and the detector audit log.
- **REQ-AE-004** (Event-detected) — When the escalation detector faults internally (a
  panic, an unreadable input, a timeout), the detector shall let the tool call proceed,
  shall append a `not-checked` line naming the class and the fault, and shall not write a
  needs-decision record for that fault.

### D.2 — The six contract classes

- **REQ-AE-005** (Event-driven) — When a checkpoint's A1 verify of a signed contract reports
  any of `acceptance_hash_mismatch`, `ac_count_mismatch`, `ac_count_ambiguous`, or
  `acceptance_missing`, the escalation detector shall trip class `acceptance-change`, citing the
  recorded `acceptance.sha256` / `acceptance.ac_count` beside the measured values.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-006** (Event-driven) — When a Bash tool call whose command string equals a
  command-kind entry of the contract's `invariants` (an entry that is neither a
  `constitution:` glob nor `frozen-files`) completes with a non-zero exit status, the
  escalation detector shall trip class `invariant-violation` (sub-kind `command`).
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-007** (Capability gate + event) — Where the contract's `invariants` contains
  `frozen-files`, when a write-capable tool call targets a path in the frozen-zone file set
  that token resolves to, the escalation detector shall trip class `invariant-violation`
  (sub-kind `frozen-file`) regardless of the calling agent's identity.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-008** (Event-driven) — When a write-capable tool call targets a path that
  matches no `ownership.write` pattern, or matches any `ownership.never` pattern, the
  escalation detector shall trip class `ownership-move`, with a `never` match taking
  precedence in the report. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-009** (Capability gate + event) — Where `workflow.autonomy.escalation.new_api_detector`
  is `graph`, when a checkpoint compares the card base to HEAD and observes a new package
  directory, a new exported declaration, a new CLI verb, a new MCP tool name, or a new
  configuration key, the escalation detector shall trip class `new-architecture-or-api`
  and list every observed addition by kind and path.
- **REQ-AE-010** (Event-driven) — When a recorded verdict pair for the card disagrees —
  an `audit_multi` result whose `disagreement_flag` is present and true, a verdict from the
  contract's `review.second_model` opposite to the first verdict, or a recorded CI failure for
  a head whose local verification was recorded as passing — the escalation detector shall trip
  class `contradictory-evidence`. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-011** (Event-driven) — When a Bash tool call is a push, a tag creation or tag push,
  a release creation, a force push, or a match for the existing destructive-command denylist,
  and it is not a push of `develop` authorized by the `push-develop` token in the contract's
  `actions`, the escalation detector shall trip class `irreversible-action` before the command
  executes. 「A1 plan-audit 통과본으로 재확인」

### D.3 — The three operational trips

- **REQ-AE-012** (Event-driven) — When the card's observed turn count, operation count, or
  audit-retry count exceeds the contract's `budget.turns` / `budget.operations` /
  `budget.audit_retries` (or, where no signed-valid contract supplies them, the
  `workflow.autonomy.escalation.budget_default` value), the escalation detector shall trip
  class `budget-exceeded` naming the exceeded dimension. 「A1 plan-audit 통과본으로 재확인」
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
  detection. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-017** (State-driven) — While an escalation record with the same class and the
  same observation fingerprint is already open for the card, the escalation detector shall
  not write a second report for that trip and shall increment an occurrence count on the
  existing record.
- **REQ-AE-018** (Ubiquitous) — The escalation detector shall present an escalation as a
  report and shall not route it through the user question channel; the agent that
  triggered it may continue work not blocked by the tripped class.
- **REQ-AE-019** (Ubiquitous) — The escalation report shall label every detection that the
  detector could not complete (an unavailable CI verdict, an unsupported language for
  declaration extraction, an unreadable contract field, a `constitution:` invariant, which has
  no mechanical violation signal) as not-observed, and shall never render not-observed as
  agreement or as absence of a trip. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-020** (Event-detected) — When A1 verify reports a SPEC's contract as
  `signed-invalid` for any reason other than the four acceptance reasons of REQ-AE-005, the
  escalation detector shall not arm contract classes 2-6 for that SPEC, shall append a
  `not-armed` line carrying the verify reason codes, and shall keep operational classes 7-9
  active against `workflow.autonomy.escalation.budget_default`. 「A1 plan-audit 통과본으로 재확인」

## §E — Constraints

- **C1 — Default preserves today.** `workflow.autonomy.mode` and
  `workflow.autonomy.escalation.budget_default` belong to A1's configuration (A1 draft
  § Configuration at `8f77d9a33`: default `guided`, invalid value → `guided` plus a warning);
  this SPEC reads them and adds only `workflow.autonomy.escalation.new_api_detector`. No
  behavior, file, or hook output changes under `guided` (REQ-AE-001). `MOAI_AUTONOMY_TIER`
  (`internal/config/envkeys.go:148`) keeps its meaning (hook block strength) and is neither read
  nor renamed by this SPEC.
- **C2 — Name distinct from the harness escalation block.** `harness.escalation`
  (`internal/config/types.go:1199-1243`) already names harness-level escalation
  (minimal→thorough). The keys read here live under `workflow.autonomy.escalation` and the two
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

This SPEC does not design the contract schema and does not copy it. **Source of every field
below:** the A1 draft, SPEC-AUTONOMY-CONTRACT-001 `design.md` § Contract Schema at commit
`8f77d9a33` (branch `WT-contract-schema`; read-only copy for this card at
`.moai/reports/t1235/a1-design-8f77d9a33.md`). That draft has **not** been plan-audited; every
requirement tagged 「A1 plan-audit 통과본으로 재확인」 is re-checked against the plan-audit-passed
version before run-phase M2.

### F.1 Fields consumed (names as in the `8f77d9a33` draft)

| Draft field / surface | Consumed by |
|---|---|
| `.moai/specs/<SPEC-ID>/contract.yaml` (file location) | REQ-AE-002, REQ-AE-020 |
| verify `state` (`unsigned` / `signed-valid` / `signed-invalid`) and `reasons[]` | REQ-AE-002, REQ-AE-005, REQ-AE-020 |
| `acceptance.sha256`, `acceptance.ac_count` (with verify's measured counterparts) | REQ-AE-005 |
| `invariants[]` — kinds `constitution:<glob>`, `frozen-files`, command (any other string) | REQ-AE-006, REQ-AE-007, REQ-AE-019 |
| `ownership.write[]`, `ownership.never[]` | REQ-AE-008 |
| `review.second_model` (`codex` / `glm` / `none`) | REQ-AE-010 |
| `actions[]` closed vocabulary; `push-develop` token | REQ-AE-011 |
| `budget.turns`, `budget.operations`, `budget.audit_retries` | REQ-AE-012 |
| `escalate_on[]` six tokens (used as the report's tripped-line token) | REQ-AE-016 |
| `workflow.autonomy.mode`, `workflow.autonomy.escalation.budget_default` (A1 § Configuration) | REQ-AE-001, REQ-AE-012, REQ-AE-020 |

### F.2 Open items — where the draft differs from the design-source assumptions

- **O1 — `escalate_on` is fixed, not selective.** The draft requires exactly the six tokens
  (`escalate_on_incomplete` otherwise). The design-source implied per-class selection; that is
  gone, and REQ-AE-020 now governs an invalid contract instead of a disabled class.
- **O2 — `invariants` has three kinds.** `constitution:<glob>` has no mechanical violation
  signal in this SPEC (not-observed, REQ-AE-019). `frozen-files` resolves to "the constitution
  frozen-zone file set", which the draft does not enumerate; whether it equals the hook's
  `frozenInstructionFiles` + prefix zones (research.md P2) is unresolved.
- **O3 — `actions` names capabilities, not push targets.** Push authority is read from
  `push-develop` only (which A1 additionally gates on `workflow.autonomy.contract.push_develop`);
  tags and releases have no token and always trip.
- **O4 — `budget` is filled at sign time.** A signed contract always carries it; the
  `budget_default` fallback applies only where no signed-valid contract exists.
- **O5 — Configuration ownership.** `mode` and `escalation.budget_default` are A1's keys;
  this SPEC adds only `escalation.new_api_detector`.
- **O6 — Line addressability is not provided.** The draft's `show --json` / `verify --json`
  object carries no source line numbers; REQ-AE-016's `contract.yaml:<line>` must be mapped by
  this detector from the file text.
- **O7 — Acceptance measurement already exists in A1.** A1 specifies an acceptance hash rule
  and a Go port of the AC counter with a parity test; class 1 consumes A1's verify result rather
  than re-measuring (resolves former open question Q3).
- **O8 — Fields present in the draft and not assumed by the design-source:** `schema_version`,
  `spec_id`, `plan_audit`, `signature` (with `contract_digest_mismatch`). A signed contract that
  turns invalid mid-run for a non-acceptance reason is handled by REQ-AE-020 as not-armed;
  whether it should instead escalate is open question Q6.
- **O9 — Mission projection.** The draft projects the contract onto `mission.MissionContract`
  so that A2 "can reuse `ValidateMissionDecision`" (present at `internal/mission/policy.go:200`).
  This SPEC does not commit to that reuse; it is a design option (design.md §C.6).

Also depends on A1 landing on `develop` before run-phase begins (card text: "plan 은 병행 가능,
run 은 A1 develop 병합 뒤").

## §G — Out of Scope

### Out of Scope — rewiring or adding gates

- Turning any trip into a deny, an ask, or a block (that is track A3).
- Adding a plain-push-to-`main` or `git tag` deny rule; class 6 reports the attempt only.

### Out of Scope — the closure report

- Aggregating escalation records into a closure or completion report (track A4).

### Out of Scope — the contract schema

- Defining, validating, signing, or versioning `contract.yaml` (track A1, card t1234).
- Detecting violation of a `constitution:<glob>` invariant; it is reported as not-observed.

### Out of Scope — queue state changes

- Adding a `needs-decision` value to the queue state enum, or any queue column.
- Automatically resolving, dismissing, or acting on an escalation.

### Out of Scope — reviving the ac-baseline guard as a product feature

- Mirroring `scripts/ac-baseline/` into the template or wiring it into this detector.

### Out of Scope — a CI poller

- Fetching CI results from a remote. Class 5 consumes a recorded CI verdict when one
  exists and reports not-observed otherwise.
