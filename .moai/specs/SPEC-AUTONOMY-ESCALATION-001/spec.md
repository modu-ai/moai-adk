---
id: SPEC-AUTONOMY-ESCALATION-001
title: "Contract-mode escalation detector: mechanical detection of the six escalate_on classes plus operational trips, reported as an escalation record without blocking or mutating the queue"
version: "0.3.0"
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
  the measured table (asset → exists? → file:line) lives in `research.md` §B.
- **2026-09-26** — v0.1.1 contract-field alignment to the A1 schema draft, SPEC-AUTONOMY-CONTRACT-001
  `design.md` § Contract Schema at commit `8f77d9a33` (referenced, not copied). Dependency tag
  「A1 plan-audit 통과본으로 재확인」 introduced.
- **2026-09-26** — v0.2.0 after plan-audit iteration 1 (FAIL 0.79): lane-owned defects D4-D16
  repaired; lead rulings 09-26 #1-#6 folded in (§H).
- **2026-09-26** — v0.2.1 added two A3 preconditions (push serializer, contract-sign guard).
  **Moved out at v0.3.0** (§K).
- **2026-09-26** — v0.3.0 after plan-audit iteration 2 (FAIL 0.82), the last-iteration revision.
  Lead rulings 09-26 (2) folded in (§H): the contract resolver narrows by card id and then by SPEC
  status (N1, N3); `contract-void` is detected before the resolver can disarm (N2); the record
  format becomes one Markdown file per class and fingerprint with a YAML frontmatter, and gains
  revoke kinds for A3 (§I); the two A3 preconditions and the mission-validator projection move to
  card t1245 (§K). Requirements renumbered contiguously in document order, 25 → 23; criteria
  25 → 23. Old → new mapping: acceptance.md §A.

## §B — Problem

### B.1 What the card asks for

A contract-mode run (a SPEC carrying a `contract.yaml`, schema owned by card t1234 / A1)
lets an agent proceed without per-step human questions **as long as it stays inside the
contract**. Staying inside is only a guarantee if leaving is detected mechanically. The
contract names six `escalate_on` classes; the card adds three operational trips, and lead ruling
09-26 #1b adds a fourth operational trip for a contract that goes void mid-run:

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
| 10 | contract-void | operational |

When a class trips, the card becomes *needs-decision* and an escalation record captures the
observation, the options, and the line that tripped. Escalation is a **report, not a
question**: the agent continues other work.

### B.2 Premises that did not hold as stated (measured, `research.md` §B)

1. **No needs-decision card state exists, and the store forbids adding one.** The queue
   states are exactly `queued` / `picked` / `dropped` (`internal/kanban/backlog_store.go:61-65`);
   `internal/kanban/backlog_schema_freeze_test.go:3-5,70` pins the three-state CHECK and
   states the landing must "admit no fourth state". Queue doctrine additionally forbids a
   machine acting on a card (`kanban-dispatch.md` § Entry into the board). The
   needs-decision marking is therefore a **record beside the queue** (REQ-AE-018).
2. **The AC-snapshot guard is not reusable as a detector.** `scripts/ac-baseline/check-staged.sh:1-4`
   declares itself a local-only dev tool; it compares an AC count against a corpus snapshot
   (line 25), not a per-SPEC contract hash. Class 1 consumes A1's own verify (§F O7).
3. **`graph_file_api` reads the working tree only.** `internal/graph/codequery.go:65-81`
   opens the file at `projectRoot/relPath`; it has no commit parameter. The before-side of a
   comparison is extracted from the base commit's blob by another route (design.md §C.4).

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

23 requirements, numbered in document order. Requirements whose predicate reads a
`contract.yaml` field, an A1 verify result, or an A1-owned configuration key carry the tag
「A1 plan-audit 통과본으로 재확인」; every such field is listed once, in §F.

Definitions used below. The **worktree root** is the top-level directory of the git worktree
containing the tool call's working directory. The **card id** is the base name of that worktree
directory (`kanban-dispatch.md`: the worktree directory keeps the card id). The **card's SPEC**
is the `spec_id` the queue store records for that card id (`internal/kanban/backlog_store.go:81`).
The **resolved contract** is the one `contract.yaml` REQ-AE-002 selects. An **operation** is one
write-capable or Bash tool call observed at PostToolUse — the unit A1's mission projection maps
to `MaxOperations`; a **turn** is one Stop hook event; an **audit retry** is one audit verdict
file for the card beyond the first. The **card state file** is
`<worktree root>/.moai/state/escalation/<card-id>.json` (design.md §C.6).

### D.1 — Activation, contract resolution, and effects

- **REQ-AE-001** (Capability gate) — Where `workflow.autonomy.mode` is absent, empty,
  unrecognized, or `guided`, the escalation detector shall perform no detection, write no
  escalation record, and add no subprocess to any hook invocation, so that hook output is
  byte-identical to the pre-change behavior. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-002** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract`,
  when a hook or checkpoint fires, the escalation detector shall first run the contract-void
  check of REQ-AE-017 and then resolve the contract in two layers inside one resolver: (a) take
  the card id from the worktree directory name and the card's SPEC from the queue store; (b)
  within that SPEC, count its `contract.yaml` only if it carries a signature block and the SPEC's
  `status` is neither `completed` nor `archived`. Exactly one surviving contract arms detection
  against it; every other outcome — no card id, no queue record, no `spec_id`, a terminal SPEC
  status, an unsigned or absent contract — appends one `not-armed` line naming the layer and
  the reason to the detector audit log. The branch name shall never be an input.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-003** (Ubiquitous) — The escalation detector shall never deny, ask about, or
  alter a tool call; its only effects shall be the escalation records (which carry the
  needs-decision marking), the detector audit log, and the card state file (budget counters,
  failure-fingerprint history, and the last observed contract state).
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
  `frozenInstructionFiles` (matched by base name), and the contract's `ownership.never` — the
  escalation detector shall trip class `invariant-violation` (sub-kind `frozen-file`)
  regardless of the calling agent's identity. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-008** (Event-driven) — When a write-capable tool call targets a path inside the
  worktree root that matches no `ownership.write` pattern, or matches any `ownership.never`
  pattern, and is not exempt under REQ-AE-013, the escalation detector shall trip class
  `ownership-move`, with a `never` match taking precedence in the record.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-009** (Capability gate + event) — Where `workflow.autonomy.escalation.new_api_detector`
  is `graph`, when the on-demand checkpoint compares the card base to HEAD and observes a new
  package directory, a new exported declaration, a new CLI verb, a new MCP tool name, or a new
  configuration key, the escalation detector shall trip class `new-architecture-or-api` and
  list every observed addition by kind and path. The **card base** is the merge-base of HEAD
  with the integration branch the project's git strategy configuration names (the development
  branch where one is configured, otherwise the default branch); where no such branch resolves
  or the merge-base is empty, class 4 is not-observed (REQ-AE-022). Sub-kind coverage is per
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
- **REQ-AE-012** (State-driven) — While the resolved contract carries a signature block, when a
  write-capable tool call targets that SPEC's `contract.yaml` or `acceptance.md`, the escalation
  detector shall trip class `ownership-move` even where `ownership.write` covers the path (lead
  ruling 09-26 #1a: both files are implicitly `never` after signing). 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-013** (Ubiquitous) — The escalation detector shall exempt from class 3 every write
  under `.moai/reports/<card-id>/`, under `.moai/state/`, under the operating system's temporary
  directory, under the session scratchpad, under the auto-memory store, and under any glob in the
  contract's optional `ownership.scratch`. A write outside the worktree root that none of these
  covers shall trip `ownership-move` when every exemption root was determined, and shall be listed
  as not-observed when one or more roots could not be determined (design.md §C.9).
  「A1 plan-audit 통과본으로 재확인」

### D.3 — Operational trips

- **REQ-AE-014** (Event-driven) — When the card's observed turn count, operation count, or
  audit-retry count exceeds the contract's `budget.turns` / `budget.operations` /
  `budget.audit_retries` (or, where no resolved contract with a readable `budget` block exists,
  the `workflow.autonomy.escalation.budget_default` value), the escalation detector shall trip
  class `budget-exceeded` naming the exceeded dimension. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-015** (Event-driven) — When the same failure fingerprint (normalized failing
  command plus diagnostic key) is observed on three consecutive failing attempts with no
  intervening success of that command, the escalation detector shall trip class
  `same-diagnostic-repeat`.
- **REQ-AE-016** (Event-driven) — When a plan-audit or a sync-audit verdict file for the card
  records FAIL at iteration `budget.audit_retries + 1` of that audit, the escalation detector
  shall trip class `audit-fail-at-retry-cap`; no separate configuration key supplies this
  ceiling (lead ruling 09-26 #5). 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-017** (Event-detected) — When the card state file records a contract as observed
  signed-valid earlier in this card, and the contract is now absent, without its signature block,
  or reported by A1 verify as `signed-invalid` for a reason other than the four acceptance
  reasons of REQ-AE-005, the escalation detector shall — before REQ-AE-002 resolution can report
  not-armed — write exactly one escalation record of class `contract-void` naming the observed
  condition, record in the same record and in the audit log that contract classes 2-6 are
  disarmed for that card, and keep classes 7-10 active; no `escalate_on` token is added. On the
  PreToolUse path only absence and a missing signature block are checked (file stat and block
  presence); the digest check runs at the commit checkpoint (C4). 「A1 plan-audit 통과본으로 재확인」

### D.4 — Reporting and marking

- **REQ-AE-018** (Ubiquitous) — The escalation detector shall write each escalation record as
  one Markdown file at `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` whose YAML
  frontmatter conforms to the schema in §I, shall treat a card as needs-decision exactly when at
  least one of its records of kind `contract` or `operational` has `status: open`, and shall not
  add a state to, rewrite, reorder, or drop any queue item.
- **REQ-AE-019** (Event-driven) — When any class trips, the escalation record shall carry the
  class, the observation (event or command plus verbatim evidence), at least two options for the
  decider, the tripped line as `contract.yaml:<line>` with the matching `escalate_on` token (or
  the configuration key or verify reason, for the operational classes), the SPEC ID, the card
  id, and the HEAD commit at detection. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-020** (State-driven) — While an escalation record for the same class and fingerprint
  exists with `status: open`, the escalation detector shall not write a second record for that
  trip and shall increment the existing record's `occurrences`; when the existing record has
  `status: resolved`, a new trip shall be written as `<class>-<fingerprint>-<n>.md` with the next
  free ordinal, leaving the resolved record and its `decider` unchanged.
- **REQ-AE-021** (Ubiquitous) — The escalation detector shall present an escalation as a
  record, shall not route it through the user question channel, and shall not block, pause,
  or require acknowledgement before the triggering agent's next tool call.
- **REQ-AE-022** (Ubiquitous) — The escalation detector shall label every detection it could not
  complete (an unavailable CI verdict, an unsupported language for declaration extraction or
  registration recognition, an unresolvable card base, an unreadable contract field, an
  undetermined exemption root, a `constitution:` invariant, which has no mechanical violation
  signal, a command-kind invariant that no tool call executed since the previous checkpoint) as
  not-observed, and shall never render not-observed as agreement or as absence of a trip.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AE-023** (Event-detected) — When the contract resolved at a first observation (no
  signed-valid state yet recorded in the card state file) is `signed-invalid` for any reason
  other than the four acceptance reasons of REQ-AE-005, the escalation detector shall not arm
  contract classes 2-6 for that card, shall append a `not-armed` line and a warning line
  carrying the verify reason codes, and shall keep operational classes 7-9 active against
  `workflow.autonomy.escalation.budget_default`. 「A1 plan-audit 통과본으로 재확인」

## §E — Constraints

- **C1 — Default preserves today.** `workflow.autonomy.mode` and
  `workflow.autonomy.escalation.budget_default` belong to A1's configuration (A1 draft
  § Configuration: default `guided`, invalid value → `guided` plus a warning); this SPEC reads
  them and adds only `workflow.autonomy.escalation.new_api_detector`. No behavior, file, or hook
  output changes under `guided` (REQ-AE-001). `MOAI_AUTONOMY_TIER`
  (`internal/config/envkeys.go:148`) keeps its meaning and is neither read nor renamed.
- **C2 — Name distinct from the harness escalation block.** `harness.escalation`
  (`internal/config/types.go:1199-1243`) already names harness-level escalation
  (minimal→thorough). The keys read here live under `workflow.autonomy.escalation` and the two
  are never merged.
- **C3 — Detection only.** This SPEC adds no denial of any kind. No existing gate is rewired,
  weakened, or strengthened; the destructive denylist, deny rules, and branch guard keep
  deciding exactly as they do now.
- **C4 — Hook budget.** Under `contract` mode, detector work on the PreToolUse path is
  in-memory path and string matching plus file stats, bounded by the existing hook bind budget;
  the HEAD commit REQ-AE-019 records on that path is read from the repository's HEAD and ref
  files, and the queue lookup of REQ-AE-002 reads the queue store once per hook, never through a
  subprocess. The commit checkpoint inside a PostToolUse hook is limited to in-process work (A1
  verify, verdict-file reads) within the hook's configured timeout; a timeout is a REQ-AE-004
  fault. Class 4 runs only at the on-demand checkpoint (open question Q2), never inside a hook.
- **C5 — Template neutrality.** Any template content added carries no card id, SPEC id,
  internal date, or commit SHA and favors no programming language.
- **C6 — Heuristic honesty.** Class 4 is a heuristic; its misses are expected and are
  covered by sync-audit and second review, not by this SPEC.

## §F — Dependency on A1

This SPEC does not design the contract schema and does not copy it. **Source of every field
below:** the A1 draft, SPEC-AUTONOMY-CONTRACT-001 `design.md` § Contract Schema at commit
`8f77d9a33` (branch `WT-contract-schema`; read-only copy for this card at
`.moai/reports/t1235/a1-design-8f77d9a33.md`). The plan-audit of iteration 2 compared a later A1
commit, `4208a3a3b`, which this lane did not read. That draft has **not** been plan-audited;
every requirement tagged 「A1 plan-audit 통과본으로 재확인」 and every row of §F.1 is re-checked
against the plan-audit-passed version before run-phase M2.

### F.1 Fields consumed

| Draft field / surface | Consumed by |
|---|---|
| `.moai/specs/<SPEC-ID>/contract.yaml` (file location) and presence of the `signature` block | REQ-AE-002, REQ-AE-012, REQ-AE-017 |
| verify `state` (`unsigned` / `signed-valid` / `signed-invalid`) and `reasons[]` | REQ-AE-005, REQ-AE-017, REQ-AE-023 |
| `acceptance.sha256`, `acceptance.ac_count` (with verify's measured counterparts) | REQ-AE-005 |
| `invariants[]` — kinds `constitution:<glob>`, `frozen-files`, command (any other string) | REQ-AE-006, REQ-AE-007, REQ-AE-022 |
| `ownership.write[]`, `ownership.never[]` | REQ-AE-007, REQ-AE-008, REQ-AE-012 |
| `ownership.scratch[]` (R2) | REQ-AE-013 |
| `review.second_model` | REQ-AE-010 |
| `actions[]` closed vocabulary; `push-develop` token | REQ-AE-011 |
| `budget.turns`, `budget.operations`, `budget.audit_retries` | REQ-AE-014, REQ-AE-016 |
| `escalate_on[]` six tokens (the record's tripped-line token) | REQ-AE-019 |
| contract lifecycle relative to SPEC `status` (R7) | REQ-AE-002 |
| `workflow.autonomy.mode`, `workflow.autonomy.escalation.budget_default` (A1 § Configuration) | REQ-AE-001, REQ-AE-014, REQ-AE-023 |

### F.2 Open items and requests to A1

- **O1 — `escalate_on` is fixed, not selective.** The draft requires exactly the six tokens;
  per-class disabling is gone.
- **O2 — `invariants` has three kinds.** `constitution:<glob>` has no mechanical violation
  signal here (not-observed, REQ-AE-022). `frozen-files` is defined by lead ruling 09-26 #6 (R4).
- **O3 — `actions` names capabilities, not push targets.** Push authority is read from
  `push-develop` only; tags and releases have no token and always trip class 6.
- **O4 — `budget` is filled at sign time.** The `budget_default` fallback applies only where no
  resolved contract with a readable budget exists.
- **O5 — Configuration ownership.** `mode` and `escalation.budget_default` are A1's keys; this
  SPEC adds only `escalation.new_api_detector`.
- **O6 — Line addressability is not provided.** Verify output carries no source line numbers;
  the record's contract line reference is mapped by this detector from the file text.
- **O7 — Acceptance measurement already exists in A1.** Class 1 consumes A1's verify result.
- **O8 — Fields present in the draft and not assumed by the design-source:** `schema_version`,
  `spec_id`, `plan_audit`, `signature`. A signed contract that turns invalid mid-run escalates as
  `contract-void` (REQ-AE-017).
- **O10 — Who runs command invariants.** The draft calls command-kind invariants "opaque to A1,
  run by A2/run-phase". This SPEC never executes them (C4); it observes executions and reports an
  unexecuted invariant as not-observed. If the plan-audit-passed A1 makes execution an A2
  obligation, REQ-AE-006 and C4 need amendment.

(O9, the mission-validator projection, moved to card t1245 — §K.)

- **R1 — Immutable after signing.** Once `signature` is present, the SPEC's `contract.yaml` and
  `acceptance.md` are implicitly `never`, and the signature carries a hash for tamper detection.
  Consumed by REQ-AE-012, REQ-AE-017.
- **R2 — `ownership.scratch[]`.** Optional glob set of additional write exemptions. REQ-AE-013.
- **R3 — `budget.audit_retries` scope.** Bounds plan-audit and sync-audit alike. REQ-AE-014,
  REQ-AE-016.
- **R4 — `frozen-files` definition.** Union of the zone registry's Frozen-zone target files, the
  hook's `frozenInstructionFiles` (base-name match), and `ownership.never`. REQ-AE-007.
- **R7 — A contract becomes terminal when its SPEC is completed.** A1 states that a signed
  contract whose SPEC `status` is `completed` (or `archived`) no longer binds any run, so a
  resolver may ignore it. Consumed by REQ-AE-002 (lead ruling 09-26 (2) #1).

(R5 and R6 concerned the moved A3 preconditions and went with them to card t1245 — §K.)

Also depends on A1 landing on `develop` before run-phase begins (card text: "plan 은 병행 가능,
run 은 A1 develop 병합 뒤").

## §G — Out of Scope

### Out of Scope — rewiring or adding gates

- Turning any escalation trip into a deny, an ask, or a block (that is track A3).
- Adding a plain-push-to-`main` or `git tag` deny rule; class 6 reports the attempt only.
- Push serialization and the contract-sign deny (moved to card t1245, §K).

### Out of Scope — the closure report

- Aggregating escalation records into a closure or completion report (track A4).

### Out of Scope — the contract schema and revocation

- Defining, validating, signing, or versioning `contract.yaml` (track A1, card t1234); §F.2
  lists requests only.
- Detecting violation of a `constitution:<glob>` invariant; it is reported as not-observed.
- Implementing `moai contract revoke` or writing revoke records; §I only reserves their kind
  for A3.

### Out of Scope — queue state changes and the F1 record layer

- Adding a `needs-decision` value to the queue state enum, or any queue column; the resolver only
  reads the queue.
- Automatically resolving, dismissing, or acting on an escalation.
- Migrating escalation records into the factory worktree↔card record (F1 ingests the files later).

### Out of Scope — mission-validator projection

- Projecting the contract onto `mission.MissionContract` and reusing `ValidateMissionDecision`
  (moved to card t1245, §K).

### Out of Scope — reviving the ac-baseline guard as a product feature

- Mirroring `scripts/ac-baseline/` into the template or wiring it into this detector.

### Out of Scope — a CI poller

- Fetching CI results from a remote. Class 5 consumes a recorded CI verdict when one exists and
  reports not-observed otherwise.

## §H — Lead rulings

Each ruling is recorded with the reason it was needed; the reasons are the plan-audit findings it
closes (`.moai/reports/t1235/plan-audit-iter1.md`, `plan-audit-iter2.md`).

| Ruling | Decision | Reason | Carried by |
|---|---|---|---|
| lead ruling 09-26 #1a | After signing, `contract.yaml` and `acceptance.md` are implicitly `never`. | D2: editing the contract, which `ownership.write` must cover, could turn detection off with no record. | REQ-AE-012, R1 |
| lead ruling 09-26 #1b | Contract loss is never silently disarmed: one escalation through the operational path, no seventh `escalate_on` token, disarming recorded. | D2 / Q6. | REQ-AE-017 |
| lead ruling 09-26 #2 | Resolve from the worktree root, never the branch name; not-armed outcomes are logged; resolver is one function. | D1. | REQ-AE-002 (narrowed by (2) #1) |
| lead ruling 09-26 #3 | Fixed default exemptions plus optional `ownership.scratch`. | D3. | REQ-AE-013, R2 |
| lead ruling 09-26 #4 | Needs-decision lives in an escalation record beside the queue; F1 ingests later; M1 does not wait. | Q1. Its path and format are superseded by (2) #6. | REQ-AE-018 |
| lead ruling 09-26 #5 | No new audit-ceiling key; `budget.audit_retries` for both audits. | Q4. | REQ-AE-016, R3 |
| lead ruling 09-26 #6 | `frozen-files` = registry Frozen targets ∪ `frozenInstructionFiles` ∪ `ownership.never`. | Q7. | REQ-AE-007, R4 |
| lead ruling 09-26 (2) #1 | Resolver narrows in two layers: the worktree directory name is the card id and selects that card's SPEC; within it, count only signed contracts of SPECs not `completed`/`archived`; arm on exactly one, log not-armed otherwise; one function; ask A1 for R7. | N1: signed contracts stay in the tree after their SPEC closes, so counting every signed contract converged to permanent not-armed from the second contract card on. N3/Q9: no card id source existed, so card-evidence writes were not exempt. | REQ-AE-002, REQ-AE-013, R7 |
| lead ruling 09-26 (2) #2 | Push serialization uses the `moai slot` `push-develop` lease; nothing changes here because of (2) #3. | N6. | §K |
| lead ruling 09-26 (2) #3 | Split: REQ-AE-024, REQ-AE-025, AC-AE-022..025, and the mission-validator projection move to card t1245. | N5: the receipt-path criterion could only turn green after A3, which starts after this SPEC — this SPEC could never complete. N7: the projection had no owner. Tier L ceilings left no room for the N1/N2 criteria. | §K |
| lead ruling 09-26 (2) #4 | `contract-void` is detected before the resolver decides not-armed, from the card state file. | N2: with the resolver first, a removed signature or deleted file ended as not-armed and class 10 never ran — the silent disarm D2 was meant to close. | REQ-AE-017, design.md §C.6 |
| lead ruling 09-26 (2) #6 | Record path `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md` with YAML frontmatter (card, class, fingerprint, contract line ref, status open\|resolved, decider); revoke kinds for A3. Supersedes the `escalations/<ts>.json` form of #4. | A per-class, per-fingerprint file makes dedup a file lookup and gives A3 and F1 one stable, parseable surface. | REQ-AE-018, REQ-AE-020, §I |

(Ruling (2) #5 was a stale reference repair in design.md; it changes no requirement.)

## §I — Escalation record format

One file per class and fingerprint:
`.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`, with `-<n>` appended for a re-trip
after resolution (REQ-AE-020). `<fingerprint>` is the first 16 lowercase hex characters of the
SHA-256 of the class plus the normalized observation.

### I.1 Frontmatter (machine-readable, YAML)

| Field | Type | Rule |
|---|---|---|
| `schema_version` | int | `1` |
| `card` | string | card id (worktree directory name) |
| `spec` | string | the resolved contract's SPEC ID, or `""` when no contract was resolved |
| `kind` | string | `contract`, `operational`, or `revoke` |
| `class` | string | a §B.1 class name for `contract` / `operational`; a revoke class (I.2) for `revoke` |
| `fingerprint` | string | 16 lowercase hex characters, equal to the file name's fingerprint |
| `contract_ref` | string | `contract.yaml:<line>` for contract classes; `config:<key>` or `verify:<reason>` for operational classes; `""` for revoke records without a line |
| `escalate_on` | string | the matching token for contract classes, otherwise `""` |
| `status` | string | `open` or `resolved` |
| `decider` | string | `""` while open; the person or role that resolved it once resolved |
| `occurrences` | int | ≥ 1 |
| `head_sha` | string | HEAD at first detection |
| `detected_at` | string | UTC RFC 3339, first detection |
| `updated_at` | string | UTC RFC 3339, last change |
| `not_observed` | string list | every detection this record could not complete (REQ-AE-022) |

The body carries three Markdown sections in order: `## Observation` (event or command plus
verbatim evidence), `## Options` (at least two numbered options), `## Not observed`.

### I.2 Revoke kinds (reserved for A3)

`kind: revoke` records document a `moai contract revoke` (track A3). This SPEC reserves the kind
and two class names and never writes them:

| Class | Meaning |
|---|---|
| `revoke-operator` | a person revoked the contract directly |
| `revoke-on-decision` | a decider revoked the contract while resolving an open escalation record |

A revoke record carries `status: resolved` and a non-empty `decider`; it never makes a card
needs-decision (REQ-AE-018). A3 may add revoke classes; this table is the reserved minimum.

A card is needs-decision exactly when one of its `contract` or `operational` records has
`status: open`.

## §K — Moved to A2b (card t1245)

Lead ruling 09-26 (2) #3 moved the following out of this SPEC to card **t1245** (A2b). They are
not requirements of this SPEC any more; numbers below are the v0.2.1 identifiers.

| What it was | Former id (v0.2.1) |
|---|---|
| Push serializer on the `moai slot` `push-develop` lease (A3 precondition) | REQ-AE-024 |
| PreToolUse deny on agent-invoked `moai contract sign`, with the receipt-path allowance (A3 precondition) | REQ-AE-025 |
| Push serializer criteria (second push denied / first unaffected; release and stale reclaim) | AC-AE-022, AC-AE-023 |
| Contract-sign guard criteria (bypass shapes and controls; receipt path allowed) | AC-AE-024, AC-AE-025 |
| Their mechanics, milestone, and bypass-shape table | design.md §G, plan.md M6, spec.md §J |
| Receipt path; which window `push_requires_window` means | A1 requests R5, R6 |
| Projecting the contract onto `mission.MissionContract` and reusing `ValidateMissionDecision` (`internal/mission/policy.go:200`) | O9 / N7 — mission-validator projection |

Because of the split, plan-audit iteration-2 findings **N4** (sign-guard bypass through `script`,
`timeout`, `sudo` and similar wrappers) and **N8** (the sign deny contradicting the "nothing
changes under `guided`" promise) no longer apply to this SPEC; they travel with REQ-AE-025 to
card t1245. N5, N6, and N7 likewise moved with their items.
