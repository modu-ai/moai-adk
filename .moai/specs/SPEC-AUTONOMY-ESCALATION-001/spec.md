---
id: SPEC-AUTONOMY-ESCALATION-001
title: "Contract-mode escalation detector: mechanical detection of the six escalate_on classes plus operational trips, reported as an escalation record without blocking or mutating the queue"
version: "0.2.0"
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
  copied). `escalate_on` must hold exactly the six tokens, so per-class disabling was removed;
  class 1 consumes A1's own verify reasons; class 6 reads push authority from `push-develop`.
  The dependency tag on contract-reading requirements became 「A1 plan-audit 통과본으로 재확인」.
- **2026-09-26** — v0.2.0 after plan-audit iteration 1 (FAIL 0.79) and the lead rulings of the
  same day (§H). Lane-owned defects D4-D16 repaired: card base and per-language sub-kind coverage
  (REQ-AE-009), operation/turn definitions and a readable-signed-contract budget source
  (REQ-AE-012), detector state file as an effect (REQ-AE-003), REQ-AE-018 restated as a
  detector-side prohibition, not-observed set widened (REQ-AE-019), explicit hook budget (C4),
  O10. Lead rulings folded in: contract resolution from the worktree root (REQ-AE-002, D1),
  post-signing immutability and a `contract-void` operational escalation instead of silent
  disarming (REQ-AE-021, REQ-AE-023, D2/Q6), fixed exemptions plus `ownership.scratch`
  (REQ-AE-022, D3), record location and schema (REQ-AE-015, Q1), audit-retry cap from
  `budget.audit_retries` for both audits (REQ-AE-014, Q4), and the `frozen-files` union
  (REQ-AE-007, Q7). Requirements 20 → 23, criteria 22 → 25 (renumbered, mapping in
  acceptance.md §A).

## §B — Problem

### B.1 What the card asks for

A contract-mode run (a SPEC carrying a `contract.yaml`, schema owned by card t1234 / A1)
lets an agent proceed without per-step human questions **as long as it stays inside the
contract**. Staying inside is only a guarantee if leaving is detected mechanically. The
contract names six `escalate_on` classes; the card adds three operational trips, and the lead
ruling of 09-26 adds a fourth operational trip for a contract that goes void mid-run:

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
| 10 | contract-void | operational (lead ruling 09-26) |

When a class trips, the card becomes *needs-decision* and an escalation record captures the
observation, the options, and the line that tripped. Escalation is a **report, not a
question**: the agent continues other work.

### B.2 Premises that did not hold as stated (measured, `research.md` §B)

1. **No needs-decision card state exists, and the store forbids adding one.** The queue
   states are exactly `queued` / `picked` / `dropped` (`internal/kanban/backlog_store.go:61-65`);
   `internal/kanban/backlog_schema_freeze_test.go:3-5,70` pins the three-state CHECK and
   states the landing must "admit no fourth state". Queue doctrine additionally forbids a
   machine acting on a card (`kanban-dispatch.md` § Entry into the board). The
   needs-decision marking is therefore a **record beside the queue** (REQ-AE-015).
2. **The AC-snapshot guard is not reusable as a detector.** `scripts/ac-baseline/check-staged.sh:1-4`
   declares itself a local-only dev tool with no template mirror; it compares an AC count
   against a corpus snapshot (`.moai/reports/t338/ac-count-baseline.txt`, line 25), not a
   per-SPEC contract hash. Class 1 consumes A1's own verify of the contract's recorded hash and
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

Under `autonomy.mode: contract`, every class in §B.1 is detected by a mechanical predicate,
each trip yields exactly one attributable escalation record naming the line that tripped, and
no loss of detection happens silently. Under the distributed default `autonomy.mode: guided`,
nothing observable changes.

## §D — Requirements (GEARS)

23 requirements. Requirements whose predicate reads a `contract.yaml` field, an A1 verify
result, or an A1-owned configuration key carry the tag 「A1 plan-audit 통과본으로 재확인」;
every such field is listed once, in §F, against the A1 draft at `8f77d9a33`.

Definitions used below. The **worktree root** is the top-level directory of the git worktree
containing the tool call's working directory. The **resolved contract** is the one
`contract.yaml` REQ-AE-002 selects. An **operation** is one write-capable or Bash tool call
observed at PostToolUse — the unit A1's mission projection maps to `MaxOperations`; a **turn**
is one Stop hook event; an **audit retry** is one audit verdict file for the card beyond the
first.

### D.1 — Activation and contract resolution

- **REQ-AE-001** (Capability gate) — Where `workflow.autonomy.mode` is absent, empty,
  unrecognized, or `guided`, the escalation detector shall perform no detection, write no
  escalation record, and add no subprocess to any hook invocation, so that hook output is
  byte-identical to the pre-change behavior. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-002** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`,
  when a hook or checkpoint fires, the escalation detector shall resolve the contract by
  locating the worktree root from the tool call's working directory and counting the
  `.moai/specs/*/contract.yaml` files under it that carry a signature block: exactly one arms
  detection against it; zero appends one `not-armed` line to the detector audit log; two or
  more append one `not-armed` line plus one warning line naming every candidate. The branch
  name shall never be an input to resolution. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-003** (Ubiquitous) — The escalation detector shall never deny, ask about, or
  alter a tool call; its only effects shall be the escalation records (which carry the
  needs-decision marking), the detector audit log, and the detector state file (the budget
  counters, the failure-fingerprint history, and the last observed contract state).
- **REQ-AE-004** (Event-detected) — When the escalation detector faults internally (a
  panic, an unreadable input, a timeout), the detector shall let the tool call proceed,
  shall append a `not-checked` line naming the class and the fault, and shall not write an
  escalation record for that fault.

### D.2 — The six contract classes

- **REQ-AE-005** (Event-driven) — When a checkpoint's A1 verify of the resolved contract reports
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
  `frozen-files`, when a write-capable tool call targets a path in the frozen set — the union
  of the constitution zone registry's Frozen-zone target files, the hook's
  `frozenInstructionFiles`, and the contract's `ownership.never` — the escalation detector
  shall trip class `invariant-violation` (sub-kind `frozen-file`) regardless of the calling
  agent's identity. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-008** (Event-driven) — When a write-capable tool call targets a path inside the
  worktree root that matches no `ownership.write` pattern, or matches any `ownership.never`
  pattern, and is not exempt under REQ-AE-022, the escalation detector shall trip class
  `ownership-move`, with a `never` match taking precedence in the record.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-009** (Capability gate + event) — Where `workflow.autonomy.escalation.new_api_detector`
  is `graph`, when the on-demand checkpoint compares the card base to HEAD and observes a new
  package directory, a new exported declaration, a new CLI verb, a new MCP tool name, or a new
  configuration key, the escalation detector shall trip class `new-architecture-or-api` and
  list every observed addition by kind and path. The **card base** is the merge-base of HEAD
  with the integration branch the project's git strategy configuration names (the development
  branch where one is configured, otherwise the default branch); where no such branch resolves
  or the merge-base is empty, class 4 is not-observed (REQ-AE-019). Sub-kind coverage is per
  language: new package and new exported declaration apply to every language the declaration
  extractor supports; new CLI verb, new MCP tool name, and new configuration key apply only to
  languages for which the detector carries a registration recognizer, and every other language
  reports those three sub-kinds as not-observed.
- **REQ-AE-010** (Event-driven) — When a recorded verdict pair for the card disagrees —
  an `audit_multi` result whose `disagreement_flag` is present and true, a verdict from the
  contract's `review.second_model` opposite to the first verdict, or a recorded CI failure for
  a head whose local verification was recorded as passing — the escalation detector shall trip
  class `contradictory-evidence`. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-011** (Event-driven) — When a Bash tool call is a push, a tag creation or tag push,
  a release creation, a force push, or a match for the existing destructive-command denylist,
  and it is not a push of `develop` authorized by the `push-develop` token in the contract's
  `actions`, the escalation detector shall trip class `irreversible-action` before the command
  executes, whether or not the existing denylist then denies it. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-021** (State-driven) — While the resolved contract carries a signature block, when a
  write-capable tool call targets that SPEC's `contract.yaml` or `acceptance.md`, the escalation
  detector shall trip class `ownership-move` even where `ownership.write` covers the path (lead
  ruling 09-26: both files are implicitly `never` after signing). 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-022** (Ubiquitous) — The escalation detector shall exempt from class 3 every write
  under the resolved card's evidence directory `.moai/reports/<card-id>/` (or
  `.moai/reports/<SPEC-ID>/` when the card id is unresolved), under `.moai/state/`, under the
  operating system's temporary directory, under the session scratchpad, under the auto-memory
  store, and under any glob in the contract's optional `ownership.scratch`; a write outside the
  worktree root that none of these exemptions covers shall trip `ownership-move`.
  「A1 plan-audit 통과본으로 재확인」

### D.3 — Operational trips

- **REQ-AE-012** (Event-driven) — When the card's observed turn count, operation count, or
  audit-retry count exceeds the contract's `budget.turns` / `budget.operations` /
  `budget.audit_retries` (or, where no resolved contract with a readable `budget` block exists,
  the `workflow.autonomy.escalation.budget_default` value), the escalation detector shall trip
  class `budget-exceeded` naming the exceeded dimension. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-013** (Event-driven) — When the same failure fingerprint (normalized failing
  command plus diagnostic key) is observed on three consecutive failing attempts with no
  intervening success of that command, the escalation detector shall trip class
  `same-diagnostic-repeat`.
- **REQ-AE-014** (Event-driven) — When a plan-audit or a sync-audit verdict file for the card
  records FAIL at iteration `budget.audit_retries + 1` of that audit, the escalation detector
  shall trip class `audit-fail-at-retry-cap`; no separate configuration key supplies this
  ceiling (lead ruling 09-26). 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-023** (Event-detected) — When a contract recorded in the detector state file as
  signed-valid is later observed absent, without its signature block, or `signed-invalid` for
  any reason other than the four acceptance reasons of REQ-AE-005, the escalation detector shall
  write exactly one escalation record of operational class `contract-void` naming the observed
  condition and stating that contract classes 2-6 are disarmed for that SPEC, shall append the
  disarming to the audit log, and shall keep classes 7-10 active; no `escalate_on` token is
  added. 「A1 plan-audit 통과본으로 재확인」

### D.4 — Reporting and marking

- **REQ-AE-015** (Ubiquitous) — The escalation detector shall write each escalation record as
  one JSON file at `.moai/reports/<card-id>/escalations/<timestamp>.json` (or under
  `.moai/reports/<SPEC-ID>/escalations/` with `card_id: "unresolved"` when the card id is
  unresolved) conforming to the schema in §I, shall treat a card as needs-decision exactly when
  at least one of its records has `status: open`, and shall not add a state to, rewrite,
  reorder, or drop any queue item.
- **REQ-AE-016** (Event-driven) — When any class trips, the escalation record shall carry the
  class, the observation (event or command plus verbatim evidence), at least two options for the
  decider, the tripped line as `contract.yaml:<line>` with the matching `escalate_on` token (or
  the configuration key or verify reason, for the operational classes), the SPEC ID, the card id
  or `unresolved`, and the HEAD commit at detection. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-017** (State-driven) — While an open escalation record with the same class and the
  same observation fingerprint exists for the card, the escalation detector shall not write a
  second record for that trip and shall increment the existing record's occurrence count.
- **REQ-AE-018** (Ubiquitous) — The escalation detector shall present an escalation as a
  record, shall not route it through the user question channel, and shall not block, pause,
  or require acknowledgement before the triggering agent's next tool call.
- **REQ-AE-019** (Ubiquitous) — The escalation detector shall label every detection it could not
  complete (an unavailable CI verdict, an unsupported language for declaration extraction or
  registration recognition, an unresolvable card base, an unreadable contract field, a
  `constitution:` invariant, which has no mechanical violation signal, a command-kind invariant
  that no tool call executed since the previous checkpoint) as not-observed, and shall never
  render not-observed as agreement or as absence of a trip. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-020** (Event-detected) — When the contract resolved at a first observation (no
  signed-valid state yet recorded for it) is `signed-invalid` for any reason other than the four
  acceptance reasons of REQ-AE-005, the escalation detector shall not arm contract classes 2-6
  for that SPEC, shall append a `not-armed` line and a warning line carrying the verify reason
  codes, and shall keep operational classes 7-9 active against
  `workflow.autonomy.escalation.budget_default`. 「A1 plan-audit 통과본으로 재확인」

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
  in-memory path and string matching bounded by the existing hook bind budget; the HEAD commit
  REQ-AE-016 records on that path is read from the repository's HEAD and ref files, never
  through a subprocess. The commit checkpoint that runs inside a PostToolUse hook is limited to
  in-process work (A1 verify, verdict-file reads) within the hook's configured timeout; a
  timeout is a REQ-AE-004 fault. Class 4, which needs git blob reads and declaration
  extraction, runs only at the on-demand checkpoint (surface per open question Q2), never
  inside a hook.
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
version before run-phase M2, and so is every row of §F.1 whether or not a tag cites it.

### F.1 Fields consumed (names as in the `8f77d9a33` draft)

| Draft field / surface | Consumed by |
|---|---|
| `.moai/specs/<SPEC-ID>/contract.yaml` (file location) and presence of the `signature` block | REQ-AE-002, REQ-AE-021, REQ-AE-023 |
| verify `state` (`unsigned` / `signed-valid` / `signed-invalid`) and `reasons[]` | REQ-AE-005, REQ-AE-020, REQ-AE-023 |
| `acceptance.sha256`, `acceptance.ac_count` (with verify's measured counterparts) | REQ-AE-005 |
| `invariants[]` — kinds `constitution:<glob>`, `frozen-files`, command (any other string) | REQ-AE-006, REQ-AE-007, REQ-AE-019 |
| `ownership.write[]`, `ownership.never[]` | REQ-AE-007, REQ-AE-008, REQ-AE-021 |
| `ownership.scratch[]` (requested, R2 — not in the draft) | REQ-AE-022 |
| `review.second_model` (`codex` / `glm` / `none`) | REQ-AE-010 |
| `actions[]` closed vocabulary; `push-develop` token | REQ-AE-011 |
| `budget.turns`, `budget.operations`, `budget.audit_retries` | REQ-AE-012, REQ-AE-014 |
| `escalate_on[]` six tokens (used as the record's tripped-line token) | REQ-AE-016 |
| `workflow.autonomy.mode`, `workflow.autonomy.escalation.budget_default` (A1 § Configuration) | REQ-AE-001, REQ-AE-012, REQ-AE-020 |

### F.2 Open items — where the draft differs from the assumptions, and requests to A1

- **O1 — `escalate_on` is fixed, not selective.** The draft requires exactly the six tokens
  (`escalate_on_incomplete` otherwise); per-class disabling is gone.
- **O2 — `invariants` has three kinds.** `constitution:<glob>` has no mechanical violation
  signal here (not-observed, REQ-AE-019). `frozen-files` is now defined by lead ruling (R4).
- **O3 — `actions` names capabilities, not push targets.** Push authority is read from
  `push-develop` only (which A1 additionally gates on `workflow.autonomy.contract.push_develop`);
  tags and releases have no token and always trip.
- **O4 — `budget` is filled at sign time.** A signed contract always carries it; the
  `budget_default` fallback applies only where no resolved contract with a readable budget exists.
- **O5 — Configuration ownership.** `mode` and `escalation.budget_default` are A1's keys;
  this SPEC adds only `escalation.new_api_detector`.
- **O6 — Line addressability is not provided.** The draft's `show --json` / `verify --json`
  object carries no source line numbers; REQ-AE-016's `contract.yaml:<line>` is mapped by this
  detector from the file text.
- **O7 — Acceptance measurement already exists in A1.** A1 specifies an acceptance hash rule
  and a Go port of the AC counter with a parity test; class 1 consumes A1's verify result.
- **O8 — Fields present in the draft and not assumed by the design-source:** `schema_version`,
  `spec_id`, `plan_audit`, `signature` (with `contract_digest_mismatch`). A signed contract that
  turns invalid mid-run now escalates as `contract-void` (REQ-AE-023, lead ruling 09-26).
- **O9 — Mission projection.** The draft projects the contract onto `mission.MissionContract`
  so that A2 "can reuse `ValidateMissionDecision`" (present at `internal/mission/policy.go:200`).
  This SPEC does not commit to that reuse; it is a design option (design.md §C.6).
- **O10 — Who runs command invariants.** The draft describes command-kind invariants as
  "opaque to A1, run by A2/run-phase". This SPEC never executes them (C4, design.md §C.2); it
  observes executions the agent performs and reports an unexecuted invariant as not-observed
  (REQ-AE-019). If the plan-audit-passed A1 schema makes execution an A2 obligation, REQ-AE-006
  and C4 need amendment.

Requests to A1 that follow from the lead rulings of 09-26 (§H):

- **R1 — Immutable after signing.** A1 states that once `signature` is present, the SPEC's
  `contract.yaml` and `acceptance.md` are implicitly `never` for the run, and that the
  signature carries a hash that lets a consumer detect tampering (the draft's
  `signature.contract_sha256` / `contract_digest_mismatch` may already satisfy the hash half).
  Consumed by REQ-AE-021 and REQ-AE-023.
- **R2 — `ownership.scratch[]`.** An optional glob set of additional write exemptions.
  Consumed by REQ-AE-022.
- **R3 — `budget.audit_retries` scope.** A1 states the field bounds plan-audit and sync-audit
  retries alike. Consumed by REQ-AE-012 and REQ-AE-014.
- **R4 — `frozen-files` definition.** A1 defines the token as the union of the constitution zone
  registry's Frozen-zone target files, the hook's `frozenInstructionFiles`, and the contract's
  `ownership.never`. Consumed by REQ-AE-007.

Also depends on A1 landing on `develop` before run-phase begins (card text: "plan 은 병행 가능,
run 은 A1 develop 병합 뒤").

## §G — Out of Scope

### Out of Scope — rewiring or adding gates

- Turning any trip into a deny, an ask, or a block (that is track A3).
- Adding a plain-push-to-`main` or `git tag` deny rule; class 6 reports the attempt only.

### Out of Scope — the closure report

- Aggregating escalation records into a closure or completion report (track A4).

### Out of Scope — the contract schema

- Defining, validating, signing, or versioning `contract.yaml` (track A1, card t1234); §F.2
  lists requests only.
- Detecting violation of a `constitution:<glob>` invariant; it is reported as not-observed.

### Out of Scope — queue state changes and the F1 record layer

- Adding a `needs-decision` value to the queue state enum, or any queue column.
- Automatically resolving, dismissing, or acting on an escalation.
- Migrating escalation records into the factory worktree↔card record (F1 ingests the files later).

### Out of Scope — reviving the ac-baseline guard as a product feature

- Mirroring `scripts/ac-baseline/` into the template or wiring it into this detector.

### Out of Scope — a CI poller

- Fetching CI results from a remote. Class 5 consumes a recorded CI verdict when one
  exists and reports not-observed otherwise.

## §H — Lead rulings 09-26

Each ruling is recorded with the reason it was needed; the reasons are the plan-audit
iteration-1 findings (`.moai/reports/t1235/plan-audit-iter1.md`) it closes.

| # | Ruling | Reason | Carried by |
|---|---|---|---|
| 1a | After signing, `contract.yaml` and `acceptance.md` are implicitly `never`; a write to either trips `ownership-move`. | D2: an agent could edit the contract, which `ownership.write` must cover, and turn detection off without any record. | REQ-AE-021, R1 |
| 1b | A contract that disappears, loses its signature, or fails its hash is never silently disarmed: exactly one escalation through the operational path, no seventh `escalate_on` token, and the disarming itself recorded. | D2 / Q6: the former REQ-AE-020 disarmed with only a log line, contradicting REQ-AE-019. | REQ-AE-023 |
| 2 | Resolve the contract from the tool call's worktree root: count signed `.moai/specs/*/contract.yaml`; one arms, zero is not-armed, two or more is not-armed plus a warning; every not-armed outcome is an observable line; never the branch name; resolver isolated for a later move to the F1 record. | D1: no rule said which SPEC's contract a tool call is judged against. | REQ-AE-002, design.md §C.8 |
| 3 | Fixed default exemptions (card evidence dir, `.moai/state/`, OS temp, session scratchpad, auto-memory store) plus an optional contract `ownership.scratch`. | D3: ordinary writes outside `ownership.write` would keep every card permanently needs-decision. | REQ-AE-022, R2 |
| 4 | The needs-decision record lives at `.moai/reports/<card>/escalations/<ts>.json` with a schema in this SPEC; F1 ingests these files later; M1 does not wait for F1. | Q1 blocked M1. | REQ-AE-015, §I |
| 5 | No new configuration key for the audit ceiling; `budget.audit_retries` applies to plan-audit and sync-audit. | Q4: no sync-audit ceiling source existed. | REQ-AE-014, R3 |
| 6 | `frozen-files` = registry Frozen-zone target files ∪ `frozenInstructionFiles` ∪ `ownership.never`. | Q7: the draft left the set undefined. | REQ-AE-007, R4 |

## §I — Escalation record schema

One file per record, `.moai/reports/<card-id>/escalations/<timestamp>.json`, where
`<timestamp>` is the detection time in UTC as `YYYYMMDDTHHMMSSZ` followed by a short
fingerprint suffix so two trips in one second do not collide.

| Field | Type | Rule |
|---|---|---|
| `schema_version` | int | `1` |
| `card_id` | string | resolved card id, or `"unresolved"` |
| `spec_id` | string | the resolved contract's SPEC ID, or `""` for an operational trip with no resolved contract |
| `class` | string | one of the ten class names in §B.1 |
| `kind` | string | `contract` or `operational` |
| `tripped` | object | `{ "file": "contract.yaml", "line": <int>, "token": "<escalate_on token>" }` for contract classes; `{ "config_key": "<key>" }` or `{ "verify_reason": "<code>" }` for operational classes |
| `observation` | object | `{ "event": "<hook or checkpoint>", "command": "<string or empty>", "evidence": "<verbatim excerpt>" }` |
| `options` | string[] | at least two entries |
| `not_observed` | string[] | every detection this record could not complete (REQ-AE-019) |
| `head_sha` | string | HEAD at detection |
| `detected_at` | string | UTC RFC 3339 |
| `fingerprint` | string | stable hash of `class` + normalized observation (REQ-AE-017) |
| `occurrences` | int | ≥ 1 |
| `status` | string | `open` when written; changing it is a human act outside this SPEC |

A card is needs-decision exactly when one of its records has `status: open` (REQ-AE-015).
